package commands

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type randomHandler struct {
	cmd
	store    *store.Store
	songdata *songdata.Service
}

func NewRandomHandler(store *store.Store, songdata *songdata.Service) *randomHandler {
	return &randomHandler{
		cmd: cmd{
			cmds: []string{"random"},
			params: [][]param{
				{
					{
						name: "level",
					},
					{
						name:     "diff",
						optional: true,
					},
				},
			},
		},
		store:    store,
		songdata: songdata,
	}
}

func (h *randomHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, levelStr, ok := extractParamForward(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	_, diffStr, hasDiff := extractParamForward(params, 1)

	level, ok := parseLevel(levelStr)
	if !ok {
		sendReply(st, embedbuilder.UserError(fmt.Sprintf("Invalid level `%s`!", levelStr)), e)
		return true
	}

	diff := ""
	if hasDiff {
		diff, ok = parseDiffKey(diffStr)
		if !ok {
			sendInvalidDiffError(st, diffStr, e)
			return true
		}
	}

	chartList := make([]struct {
		*songdata.Song
		*songdata.Chart
	}, 0)

	if !hasDiff {
		for _, song := range h.songdata.GetData() {
			for _, chart := range song.Charts {
				if chart.Level == level {
					chartList = append(chartList, struct {
						*songdata.Song
						*songdata.Chart
					}{&song, &chart})
				}
			}
		}
	} else {
		for _, song := range h.songdata.GetData() {
			chart, ok := song.GetChart(diff)
			if ok && chart.Level == level {
				chartList = append(chartList, struct {
					*songdata.Song
					*songdata.Chart
				}{&song, &chart})
			}
		}
	}

	if len(chartList) == 0 {
		sendReply(st, embedbuilder.UserError(fmt.Sprintf("There are no Lv%s charts with the difficulty %s!", level, getFullDiffName(diff))), e)
		return true
	}

	randIdx := rand.IntN(len(chartList))
	selChart := chartList[randIdx]

	embedFields := []discord.EmbedField{
		{
			Name:  "Title",
			Value: selChart.Title,
		},
		{
			Name:  "Artist",
			Value: selChart.Artist,
		},
		{
			Name:  "Difficulty",
			Value: fmt.Sprintf("%s - Lv%s (%.1f) (v%s)", selChart.GetDiffDisplayName(), selChart.Level, selChart.CC, selChart.Ver),
		},
	}

	songEmbed := discord.Embed{
		Title:  "Randomly Selected Chart",
		Fields: embedFields,
	}

	sendReply(st, embedbuilder.Info(songEmbed), e)

	return true
}
