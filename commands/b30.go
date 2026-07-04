package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type b30Handler struct {
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewB30Handler(store *store.Store, db *database.Service, songdata *songdata.Service) *b30Handler {
	return &b30Handler{
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *b30Handler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "b30" {
		return false
	}

	st := h.store.Bot.State()

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

	scoresRepo := sess.GetScoresRepo()

	count, err := scoresRepo.GetUserPlayedChartCount(ctx, int64(e.Sender().ID))
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	if count == 0 {
		sendCommandErrorReply(st, "You don't have any scores saved!", e)
		return true
	}

	avgRt, avgScore, err := scoresRepo.GetBestScoreRatingsAverage(ctx, int64(e.Sender().ID), 30)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	entries, err := scoresRepo.GetBestScoresByUserWithOffset(ctx, int64(e.Sender().ID), 0, 5)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	embed := createB30Embed(h, avgRt, avgScore, entries, 0)
	components := createB30PageButtons(int64(e.Sender().ID), count, 0)

	sendInteractionResponse(st, embedbuilder.Info(embed), components, e)

	return true
}

func (h *b30Handler) HandleB30PageSelect(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	st := h.store.Bot.State()

	val := e.Data.(*discord.ButtonInteraction).CustomID

	params := strings.Split(string(val), ",")
	receiver := params[1]
	if receiver != "b30" {
		return false
	}

	userId, _ := strconv.ParseInt(params[0], 10, 64)
	offset, _ := strconv.Atoi(params[2])

	pageIdx := offset / 5

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	defer func() {
		err := sess.Conn.Close()
		if err != nil {
			logAndSendInteractionError(ctx, st, err, e)
		}
	}()

	scoresRepo := sess.GetScoresRepo()

	count, err := scoresRepo.GetUserPlayedChartCount(ctx, userId)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	avgRt, avgScore, err := scoresRepo.GetBestScoreRatingsAverage(ctx, userId, 30)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	entries, err := scoresRepo.GetBestScoresByUserWithOffset(ctx, userId, offset, 5)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	embed := createB30Embed(h, avgRt, avgScore, entries, offset)
	components := createB30PageButtons(userId, count, pageIdx)

	resp := api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds:     &[]discord.Embed{embedbuilder.Info(embed)},
			Components: (*discord.TopLevelComponents)(&components),
		},
	}

	st.RespondInteraction(e.ID, e.Token, resp)

	return true
}

func createB30Embed(h *b30Handler, avgRt float64, avgScore float64, entries []database.ScoreRecordRating, idx int) discord.Embed {
	entriesBuilder := strings.Builder{}

	for i, s := range entries {
		chart, song, _ := h.songdata.GetChartById(s.ChartId)

		fmt.Fprintf(&entriesBuilder, "%v. %v ▸ %v Lv%v (%.1f)\n  %v - **%.4f** (<t:%v:R>)\n  -# Score ID: %v\n",
			idx+i+1,
			song.AltTitle,
			chart.GetDiffDisplayName(),
			chart.Level,
			chart.CC,
			s.Score,
			s.Rating,
			s.Timestamp/1000,
			s.Id)
	}

	embed := discord.Embed{
		Title: "Highest Play Ratings from Saved Scores",
		Fields: []discord.EmbedField{
			{
				Name:  "Best-30 Stats",
				Value: fmt.Sprintf("**Average rating: %.4f**\nAverage score: %.2f", avgRt, avgScore),
			},
			{
				Name:  "Top Play Ratings",
				Value: entriesBuilder.String(),
			},
		},
	}

	return embed
}

func createB30PageButtons(userId int64, count int, pageIdx int) []discord.TopLevelComponent {
	prevOffset := (pageIdx - 1) * 5
	nextOffset := (pageIdx + 1) * 5

	return []discord.TopLevelComponent{
		&discord.ActionRowComponent{
			&discord.ButtonComponent{
				CustomID: discord.ComponentID(fmt.Sprintf("%v,b30,%v", userId, prevOffset)),
				Label:    "<",
				Disabled: prevOffset < 0,
			},
			&discord.ButtonComponent{
				CustomID: discord.ComponentID(fmt.Sprintf("%v,b30,%v", userId, nextOffset)),
				Label:    ">",
				Disabled: nextOffset >= count,
			},
		},
	}
}
