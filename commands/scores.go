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
	cmd
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewScoresHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *scoresHandler {
	return &scoresHandler{
		cmd: cmd{
			cmds: []string{"scores"},
			params: [][]param{
				{
					{
						name: "song",
					},
					{
						name: "diff",
					},
				},
			},
		},
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *scoresHandler) HandleTextCommand(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

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

	scoresRepo := sess.GetScoresRepo()

	count, err := scoresRepo.GetScoreCountByUserAndChart(ctx, int64(e.Author.ID), chart.Id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	if count == 0 {
		sendReply(st, embedbuilder.UserError("You don't have any scores saved for this chart!"), e)
		return true
	}

	bestScore, err := scoresRepo.GetBestScoreByUserAndChart(ctx, int64(e.Author.ID), chart.Id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	recentScores, err := scoresRepo.GetByUserAndChartWithOffset(ctx, int64(e.Author.ID), chart.Id, 0, 5)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	embed := createScoresEmbed(song, chart, bestScore, recentScores, 1)

	pageIdx := 1
	pageStart := 1
	pageOpts := []discord.SelectOption{}

	for pageStart <= count {
		pageOpts = append(pageOpts, discord.SelectOption{
			Label: fmt.Sprintf("Page %d (%d-%d)", pageIdx, pageStart, pageStart+4),
			Value: fmt.Sprintf("%v,%v,%v", e.Author.ID, chart.Id, (pageIdx-1)*5),
		})

		pageIdx += 1
		pageStart += 5
	}

	components := []discord.ContainerComponent{
		&discord.ActionRowComponent{
			&discord.StringSelectComponent{
				CustomID:    "ScorePageSelect",
				Options:     pageOpts,
				Placeholder: "Select page...",
			},
		},
	}

	sendReplyWithComponents(st, embedbuilder.Info(embed), components, e.ChannelID, e.ID)

	return true
}

func (h *scoresHandler) HandleScorePageSelect(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	st := h.store.Bot.State()

	val := e.Data.(*discord.StringSelectInteraction).Values[0]

	params := strings.Split(val, ",")
	userId, _ := strconv.ParseInt(params[0], 10, 64)
	chartId, _ := strconv.Atoi(params[1])
	offset, _ := strconv.Atoi(params[2])

	chart, song, _ := h.songdata.GetChartById(chartId)

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendInteractionError(ctx, st, err, e)
		return true
	}

	scoresRepo := sess.GetScoresRepo()
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

	resp := api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{embedbuilder.Info(embed)},
		},
	}

	st.RespondInteraction(e.ID, e.Token, resp)

	return true
}

func createScoresEmbed(song songdata.Song, chart songdata.Chart, best database.Score, recents []database.Score, idx int) discord.Embed {
	recentsBuilder := strings.Builder{}

	for i, s := range recents {
		recentsBuilder.WriteString(fmt.Sprintf("%v. %v (<t:%v:R>)\n  -# Score ID: %v\n", idx+i, s.Score, s.Timestamp/1000, s.Id))
	}

	embed := discord.Embed{
		Title: fmt.Sprintf("Saved Scores for %s - %s Lv%s", song.EscapedAltTitle(), chart.GetDiffDisplayName(), chart.Level),
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
