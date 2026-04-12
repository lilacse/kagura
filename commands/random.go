package commands

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type randomHandler struct {
	store    *store.Store
	songdata *songdata.Service
}

func NewRandomHandler(store *store.Store, songdata *songdata.Service) *randomHandler {
	return &randomHandler{
		store:    store,
		songdata: songdata,
	}
}

func (h *randomHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "random" {
		return false
	}

	st := h.store.Bot.State()

	level := data.Options.Find("level").String()
	diff := data.Options.Find("diff").String()

	hasLevel := level != ""
	hasDiff := diff != ""

	chartList := make([]struct {
		*songdata.Song
		*songdata.Chart
	}, 0)

	for _, song := range h.songdata.GetData() {
		for _, chart := range song.Charts {
			if hasLevel && chart.Level != level {
				continue
			}
			if hasDiff && chart.Diff != diff {
				continue
			}
			chartList = append(chartList, struct {
				*songdata.Song
				*songdata.Chart
			}{&song, &chart})
		}
	}

	if len(chartList) == 0 {
		errStr := strings.Builder{}
		errStr.WriteString("There are no charts matching the query:")
		if hasLevel {
			fmt.Fprintf(&errStr, " Lv%s", level)
		}
		if hasDiff {
			fmt.Fprintf(&errStr, " %s", getFullDiffName(diff))
		}
		errStr.WriteString("!")

		sendCommandErrorReply(st, errStr.String(), e)
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
	}

	if hasLevel || hasDiff {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Difficulty",
			Value: fmt.Sprintf("%s - Lv%s (%.1f) (v%s)", selChart.GetDiffDisplayName(), selChart.Level, selChart.CC, selChart.Ver),
		})
	}

	songEmbed := discord.Embed{
		Title:  "Randomly Selected Chart",
		Fields: embedFields,
	}

	res := embedbuilder.Info(songEmbed)
	sendCommandReply(st, res, e)

	return true
}
