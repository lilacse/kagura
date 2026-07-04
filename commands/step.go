package commands

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type stepHandler struct {
	store    *store.Store
	songdata *songdata.Service
}

func NewStepHandler(store *store.Store, songdata *songdata.Service) *stepHandler {
	return &stepHandler{
		store:    store,
		songdata: songdata,
	}
}

func (h *stepHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "step" {
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

	step, err := data.Options.Find("stat").FloatValue()
	if err != nil {
		sendCommandErrorReply(st, fmt.Sprintf("Invalid step `%s`!", data.Options.Find("stat").String()), e)
		return true
	}

	score, errStr, ok := parseShortScore(data.Options.Find("score").String())
	if !ok {
		sendCommandErrorReply(st, errStr, e)
		return true
	}

	ptt := chart.GetActualScoreRating(score)

	progress := (2.45*math.Sqrt(ptt) + 2.5) * (float64(step) / 50)
	floored := math.Floor(progress*10) / 10
	formula := fmt.Sprintf("(2.45 * sqrt(%.4f) + 2.5) * (%v / 50) = **%.4f** (shown as **%.1f**)", ptt, step, progress, floored)

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
				Name:   "Step stat",
				Value:  strconv.FormatFloat(step, 'f', -1, 64),
				Inline: true,
			},
			{
				Name: "Progress gained",
				Value: fmt.Sprintf(`%s

-# - There might be a ±0.1 difference in actual progress gained due to differences in calculation performed by the game.
-# - For partner progression bonuses, __add__ them to the value above before calculating Play+ and fragment boosts.
-# - For Play+ boost, __multiply__ the value by stamina used. For fragment boost, further __multiply__ the value by boost multiplier.`, formula),
			},
		},
	}

	res := embedbuilder.Info(embed)
	sendInteractionResponse(st, res, []discord.TopLevelComponent{}, e)

	return true
}
