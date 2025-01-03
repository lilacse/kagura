package unsave

import (
	"context"
	"fmt"
	"strconv"

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
	params, ok := commands.ExtractParamsString("unsave", e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()
	prefix := h.store.Bot.Prefix()

	_, idStr, ok := commands.ExtractParamReverse(params, -1)
	if !ok {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%sunsave [score id]`!", prefix)))
		return true
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid score ID `%s`!", idStr)))
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

	currRecs, err := scoresRepo.GetById(ctx, id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	if len(currRecs) == 0 || currRecs[0].UserId != int64(e.Author.ID) {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("You don't have a score with ID `%s`!", idStr)))
		return true
	}

	currRec := currRecs[0]
	chart, song, ok := songdata.GetChartById(currRec.ChartId)
	if !ok {
		logAndSendError(ctx, st, fmt.Errorf("chart id %v is not found in songdata", currRec.ChartId), e)
		return true
	}

	_, err = scoresRepo.Delete(ctx, id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	err = tx.Commit()
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	isCommit = true

	embed := discord.Embed{
		Title: "Score deleted",
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
				Value:  strconv.Itoa(currRec.Score),
				Inline: true,
			},
			{
				Name:   "Timestamp",
				Value:  fmt.Sprintf("<t:%v:R>", currRec.Timestamp/1000),
				Inline: true,
			},
		},
	}

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}

func logAndSendError(ctx context.Context, st *state.State, err error, e *gateway.MessageCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling save command: %s", err.Error()))
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Error(ctx, err.Error()))
}
