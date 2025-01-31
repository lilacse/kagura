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
	cmd
	store    *store.Store
	songdata *songdata.Service
}

func NewStepHandler(store *store.Store, songdata *songdata.Service) *stepHandler {
	return &stepHandler{
		cmd: cmd{
			cmds: []string{"step"},
			params: []param{
				{
					name: "stat",
				},
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
		},
		store:    store,
		songdata: songdata,
	}
}

func (h *stepHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, stepStr, ok := extractParamForward(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	params, scoreStr, ok := extractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	params, diffStr, ok := extractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	_, songStr, ok := extractParamReverse(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid step `%s`!", stepStr)))
		return true
	}

	score, errMsg, ok := parseScore(scoreStr)
	if !ok {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(errMsg))
		return true
	}

	matchSong := h.songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		sendSongQueryError(st, songStr, e)
		return true
	}

	song := matchSong[0]

	diffKey, ok := getDiffKey(diffStr)
	if !ok {
		sendInvalidDiffError(st, diffStr, e)
		return true
	}

	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistError(st, diffKey, song.AltTitle, e)
		return true
	}

	if chart.CC == 0.0 {
		sendCcUnknownError(st, diffKey, song.AltTitle, e)
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
				Value: fmt.Sprintf("%s - %s", song.Title, song.Artist),
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
				Value:  strconv.Itoa(step),
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

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}
