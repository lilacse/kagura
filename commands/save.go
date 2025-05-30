package commands

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type saveHandler struct {
	cmd
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewSaveHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *saveHandler {
	return &saveHandler{
		cmd: cmd{
			cmds: []string{"save"},
			params: [][]param{
				{
					{
						name: "song",
					},
					{
						name: "diff",
					},
					{
						name: "score",
					},
				},
			},
		},
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *saveHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, scoreStr, ok := extractParamBackwards(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	params, diffStr, ok := extractParamBackwards(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	_, songStr, ok := extractParamBackwards(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	score, errMsg, ok := parseFullScore(scoreStr)
	if !ok {
		sendReply(st, embedbuilder.UserError(errMsg), e)
		return true
	}

	matchSong := h.songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		sendSongQueryError(st, songStr, e)
		return true
	}

	song := matchSong[0]

	diffKey, ok := parseDiffKey(diffStr)
	if !ok {
		sendInvalidDiffError(st, diffStr, e)
		return true
	}

	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistError(st, diffKey, song.EscapedAltTitle(), e)
		return true
	}

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	defer func() {
		err := sess.Conn.Close()
		if err != nil {
			logAndSendError(ctx, st, err, e)
		}
	}()

	tx, err := sess.Conn.BeginTx(ctx, nil)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	isCommit := false

	defer func() {
		if !isCommit {
			err := tx.Rollback()
			if err != nil {
				logAndSendError(ctx, st, err, e)
			}
		}
	}()

	scoresRepo := sess.GetScoresRepo()

	userScores, err := scoresRepo.GetByUserAndChart(ctx, int64(e.Author.ID), chart.Id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	// we store up to the 30th recently saved score and the user's best score.

	scoreCount := len(userScores)
	if scoreCount >= 30 {
		best := userScores[0]
		for _, s := range userScores {
			if s.Score > best.Score {
				best = s
			}
		}

		slices.SortFunc(userScores, func(a, b database.Score) int {
			return int(a.Timestamp - b.Timestamp)
		})

		if best.Timestamp != userScores[0].Timestamp {
			scoresRepo.Delete(ctx, userScores[0].Id)
		} else if scoreCount > 30 {
			scoresRepo.Delete(ctx, userScores[1].Id)
		}
	}

	ts := time.Now()

	res, err := scoresRepo.Insert(ctx, int64(e.Author.ID), chart.Id, score, ts.UnixMilli())
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	newId, _ := res.LastInsertId()

	err = tx.Commit()
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	isCommit = true

	embed := discord.Embed{
		Title: "Score saved",
		Fields: []discord.EmbedField{
			{
				Name:  "Song",
				Value: fmt.Sprintf("%s - %s", song.EscapedTitle(), song.EscapedArtist()),
			},
			{
				Name:  "Chart",
				Value: fmt.Sprintf("%s - Lv%s (%s) (v%s)", chart.GetDiffDisplayName(), chart.Level, chart.GetCCString(), chart.Ver),
			},
			{
				Name:   "Score",
				Value:  strconv.Itoa(score),
				Inline: true,
			},
			{
				Name:   "Play Rating",
				Value:  chart.GetScoreRatingString(score),
				Inline: true,
			},
			{
				Name:   "Timestamp",
				Value:  fmt.Sprintf("<t:%v:R>", ts.Unix()),
				Inline: true,
			},
		},
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("Send `%sunsave %v` to delete this score.", h.store.Bot.Prefix(), newId),
		},
	}

	sendReply(st, embedbuilder.Info(embed), e)

	return true
}
