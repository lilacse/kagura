package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type unsaveHandler struct {
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewUnsaveHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *unsaveHandler {
	return &unsaveHandler{
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *unsaveHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "unsave" {
		return false
	}

	st := h.store.Bot.State()

	idStr := data.Options.Find("score_id").String()

	id, ok := parseScoreId(idStr)
	if !ok {
		sendCommandErrorReply(st, fmt.Sprintf("Invalid score ID `%s`!", idStr), e)
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

	currRecs, err := scoresRepo.GetById(ctx, id)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	if len(currRecs) == 0 || currRecs[0].UserId != int64(e.Sender().ID) {
		sendCommandErrorReply(st, fmt.Sprintf("You don't have a score with ID `%s`!", idStr), e)
		return true
	}

	currRec := currRecs[0]
	chart, song, ok := h.songdata.GetChartById(currRec.ChartId)
	if !ok {
		logAndSendCommandError(ctx, st, fmt.Errorf("chart id %v is not found in songdata", currRec.ChartId), e)
		return true
	}

	_, err = scoresRepo.Delete(ctx, id)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	err = tx.Commit()
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	isCommit = true

	embed := discord.Embed{
		Title: "Score deleted",
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

	res := embedbuilder.Info(embed)
	sendCommandReply(st, res, e)

	return true
}
