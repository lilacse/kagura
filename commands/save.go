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
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewSaveHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *saveHandler {
	return &saveHandler{
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *saveHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "save" {
		return false
	}

	st := h.store.Bot.State()

	query := data.Options.Find("song").String()
	matched := h.songdata.Search(query, 1)
	if len(matched) == 0 {
		sendSongQueryCommandError(st, query, e)
		return true
	}

	song := matched[0]

	diffKey := data.Options.Find("diff").String()
	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistCommandError(st, diffKey, song.EscapedAltTitle(), e)
		return true
	}

	score, errStr, ok := parseFullScore(data.Options.Find("score").String())
	if !ok {
		sendCommandErrorReply(st, errStr, e)
		return true
	}

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	defer func() {
		err := sess.Conn.Close()
		if err != nil {
			logAndSendCommandError(ctx, st, err, e)
		}
	}()

	tx, err := sess.Conn.BeginTx(ctx, nil)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	isCommit := false

	defer func() {
		if !isCommit {
			err := tx.Rollback()
			if err != nil {
				logAndSendCommandError(ctx, st, err, e)
			}
		}
	}()

	scoresRepo := sess.GetScoresRepo()

	userScores, err := scoresRepo.GetByUserAndChart(ctx, int64(e.Sender().ID), chart.Id)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
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

		slices.SortFunc(userScores, func(a, b database.ScoreRecord) int {
			return int(a.Timestamp - b.Timestamp)
		})

		if best.Timestamp != userScores[0].Timestamp {
			scoresRepo.Delete(ctx, userScores[0].Id)
		} else if scoreCount > 30 {
			scoresRepo.Delete(ctx, userScores[1].Id)
		}
	}

	ts := time.Now()

	insertRes, err := scoresRepo.Insert(ctx, int64(e.Sender().ID), chart.Id, score, ts.UnixMilli())
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	newId, _ := insertRes.LastInsertId()

	err = tx.Commit()
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
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
			Text: fmt.Sprintf("Send `/unsave %v` to delete this score.", newId),
		},
	}

	res := embedbuilder.Info(embed)
	sendCommandReply(st, res, e)

	return true
}
