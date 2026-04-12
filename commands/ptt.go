package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type pttHandler struct {
	cmd
	store    *store.Store
	songdata *songdata.Service
}

func NewPttHandler(store *store.Store, songdata *songdata.Service) *pttHandler {
	return &pttHandler{
		cmd: cmd{
			cmds: []string{"ptt", "rt"},
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
				{
					{
						name: "cc",
					},
					{
						name: "score",
					},
				},
			},
		},
		store:    store,
		songdata: songdata,
	}
}

func (h *pttHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "ptt" {
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

	if chart.CC == 0.0 {
		sendCcUnknownCommandError(st, diffKey, song.EscapedAltTitle(), e)
		return true
	}

	score, errStr, ok := parseShortScore(data.Options.Find("score").String())
	if !ok {
		sendCommandErrorReply(st, errStr, e)
		return true
	}

	ptt := chart.GetScoreRating(score)
	formula := getPttFormula(score, chart.CC, ptt)

	embed := discord.Embed{
		Fields: []discord.EmbedField{
			{
				Name:  "Song",
				Value: fmt.Sprintf("%s - %s", song.EscapedTitle(), song.EscapedArtist()),
			},
			{
				Name:  "Chart",
				Value: fmt.Sprintf("%s - Lv%s (%.1f) (v%s)", chart.GetDiffDisplayName(), chart.Level, chart.CC, chart.Ver),
			},
			{
				Name:  "Score",
				Value: strconv.Itoa(score),
			},
			{
				Name:  "Play Rating",
				Value: formula,
			},
		},
	}

	res := embedbuilder.Info(embed)
	sendCommandReply(st, res, e)

	return true
}

func getPttFormula(score int, cc float64, ptt float64) string {
	if score >= 10000000 {
		return fmt.Sprintf("%.1f + 2.0 = **%.4f**", cc, ptt)
	} else if score >= 9800000 && score < 10000000 {
		return fmt.Sprintf("%.1f + 1.0 + ((%v - 9800000) / 200000) = **%.4f**", cc, score, ptt)
	} else {
		if ptt >= 0.0 {
			return fmt.Sprintf("%.1f + (%v - 9500000) / 300000 = **%.4f**", cc, score, ptt)
		} else {
			return fmt.Sprintf("%.1f + (%v - 9500000) / 300000 = %.4f (considered as **0.0**)", cc, score, ptt)
		}
	}
}
