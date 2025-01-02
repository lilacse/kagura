package save

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/commands"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type handler struct {
	store *store.Store
	db    *database.DbService
}

func NewHandler(store *store.Store, db *database.DbService) *handler {
	return &handler{
		store: store,
		db:    db,
	}
}

func (h *handler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := commands.ExtractParamsString("save", e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, scoreStr, ok := commands.ExtractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), e)
		return true
	}

	params, diffStr, ok := commands.ExtractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), e)
		return true
	}

	_, songStr, ok := commands.ExtractParamReverse(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), e)
		return true
	}

	score, errMsg, ok := commands.ParseScore(scoreStr)
	if !ok {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(errMsg))
	}

	matchSong := songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("No matching song found for query `%s`!", songStr)))
		return true
	}

	song := matchSong[0]

	diffKey := ""
	switch strings.ToLower(diffStr) {
	case "pst", "past":
		diffKey = "pst"
	case "prs", "present":
		diffKey = "prs"
	case "ftr", "future":
		diffKey = "ftr"
	case "etr", "eternal":
		diffKey = "etr"
	case "byd", "beyond":
		diffKey = "byd"
	}

	if diffKey == "" {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid difficulty `%s`!", diffStr)))
		return true
	}

	chart := songdata.Chart{}
	for _, c := range song.Charts {
		if c.Diff == diffKey {
			chart = c
			break
		}
	}

	if chart.Id == 0 {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Difficulty %s does not exist for the song %s!", strings.ToUpper(diffKey), song.AltTitle)))
		return true
	}

	ptt := chart.ActualScoreRating(score)

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
				Name: "",
			},
			{
				Name:  "Song",
				Value: fmt.Sprintf("%s - %s", song.Title, song.Artist),
			},
			{
				Name:  "Chart",
				Value: fmt.Sprintf("%s - Lv%s (%.1f) (v%s)", chart.DiffDisplayName(), chart.Level, chart.CC, chart.Ver),
			},
			{
				Name:   "Score",
				Value:  strconv.Itoa(score),
				Inline: true,
			},
			{
				Name:   "Play Rating",
				Value:  fmt.Sprintf("%.4f", ptt),
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

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}

func sendFormatError(st *state.State, prefix string, e *gateway.MessageCreateEvent) {
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%ssave [song] [diff] [score]`!", prefix)))
}

func logAndSendError(ctx context.Context, st *state.State, err error, e *gateway.MessageCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling save command: %s", err.Error()))
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Error(ctx, err.Error()))
}
