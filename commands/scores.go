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

type scoresHandler struct {
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewScoresHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *scoresHandler {
	return &scoresHandler{
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *scoresHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "scores" {
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

	count, err := scoresRepo.GetScoreCountByUserAndChart(ctx, int64(e.Sender().ID), chart.Id)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	if count == 0 {
		sendCommandErrorReply(st, "You don't have any scores saved for this chart!", e)
		return true
	}

	bestScore, err := scoresRepo.GetBestScoreByUserAndChart(ctx, int64(e.Sender().ID), chart.Id)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	recentScores, err := scoresRepo.GetByUserAndChartWithOffset(ctx, int64(e.Sender().ID), chart.Id, 0, 5)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return true
	}

	embed := createScoresEmbed(song, chart, bestScore, recentScores, 0)
	components := createScoresPageButtons(int64(e.Sender().ID), chart.Id, count, 0)

	sendInteractionResponse(st, embedbuilder.Info(embed), components, e)

	return true
}

func (h *scoresHandler) HandleScorePageSelect(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	st := h.store.Bot.State()

	val := e.Data.(*discord.ButtonInteraction).CustomID

	params := strings.Split(string(val), ",")
	receiver := params[1]
	if receiver != "scores" {
		return false
	}

	userId, _ := strconv.ParseInt(params[0], 10, 64)
	chartId, _ := strconv.Atoi(params[2])
	offset, _ := strconv.Atoi(params[3])

	pageIdx := offset / 5

	chart, song, _ := h.songdata.GetChartById(chartId)

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

	count, err := scoresRepo.GetScoreCountByUserAndChart(ctx, userId, chart.Id)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	bestScore, err := scoresRepo.GetBestScoreByUserAndChart(ctx, userId, chartId)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	recentScores, err := scoresRepo.GetByUserAndChartWithOffset(ctx, userId, chartId, offset, 5)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	embed := createScoresEmbed(song, chart, bestScore, recentScores, offset)
	components := createScoresPageButtons(userId, chart.Id, count, pageIdx)

	resp := api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds:     &[]discord.Embed{embedbuilder.Info(embed)},
			Components: (*discord.ContainerComponents)(&components),
		},
	}

	st.RespondInteraction(e.ID, e.Token, resp)

	return true
}

func createScoresEmbed(song songdata.Song, chart songdata.Chart, best database.ScoreRecord, recents []database.ScoreRecord, idx int) discord.Embed {
	recentsBuilder := strings.Builder{}

	for i, s := range recents {
		fmt.Fprintf(&recentsBuilder, "%v. %v (<t:%v:R>)\n  -# Score ID: %v\n", idx+i+1, s.Score, s.Timestamp/1000, s.Id)
	}

	embed := discord.Embed{
		Title: fmt.Sprintf("Saved Scores for %s ▸ %s Lv%s", song.EscapedAltTitle(), chart.GetDiffDisplayName(), chart.Level),
		Fields: []discord.EmbedField{
			{
				Name:  "Best score",
				Value: fmt.Sprintf("%v (Play Rating %s)\n<t:%v:R>\n-# Score ID: %v", best.Score, chart.GetScoreRatingString(best.Score), best.Timestamp/1000, best.Id),
			},
			{
				Name:  "Recent scores",
				Value: recentsBuilder.String(),
			},
		},
	}

	return embed
}

func createScoresPageButtons(userId int64, chartId int, count int, pageIdx int) []discord.ContainerComponent {
	prevOffset := (pageIdx - 1) * 5
	nextOffset := (pageIdx + 1) * 5

	return []discord.ContainerComponent{
		&discord.ActionRowComponent{
			&discord.ButtonComponent{
				CustomID: discord.ComponentID(fmt.Sprintf("%v,scores,%v,%v", userId, chartId, prevOffset)),
				Label:    "<",
				Disabled: prevOffset < 0,
			},
			&discord.ButtonComponent{
				CustomID: discord.ComponentID(fmt.Sprintf("%v,scores,%v,%v", userId, chartId, nextOffset)),
				Label:    ">",
				Disabled: nextOffset >= count,
			},
		},
	}
}
