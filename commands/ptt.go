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

func (h *pttHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	ok := false
	params := ""
	for _, n := range h.cmds {
		params, ok = extractParamsString(n, e.Message.Content, h.store.Bot.Prefix())
		if ok {
			break
		}
	}

	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, scoreStr, ok := extractParamBackwards(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	score, errMsg, ok := parseShortScore(scoreStr)
	if !ok {
		sendReply(st, embedbuilder.UserError(errMsg), e)
		return true
	}

	_, ccStr, ok := extractParamBackwards(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	cc, isCc := parseCc(ccStr)
	if isCc {
		handleCcQuery(h, e, score, cc)
		return true
	}

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

	handleChartQuery(h, e, score, songStr, diffStr)
	return true
}

func handleCcQuery(h *pttHandler, e *gateway.MessageCreateEvent, score int, cc float64) {
	st := h.store.Bot.State()

	ptt := getPttFromCc(score, cc)
	formula := getPttFormula(score, cc, ptt)

	embed := discord.Embed{
		Fields: []discord.EmbedField{
			{
				Name:   "Score",
				Value:  strconv.Itoa(score),
				Inline: true,
			},
			{
				Name:   "Chart Constant",
				Value:  fmt.Sprintf("%.1f", cc),
				Inline: true,
			},
			{
				Name:  "Play Rating",
				Value: formula,
			},
		},
	}

	sendReply(st, embedbuilder.Info(embed), e)
}

func handleChartQuery(h *pttHandler, e *gateway.MessageCreateEvent, score int, songStr string, diffStr string) {
	st := h.store.Bot.State()

	matchSong := h.songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		sendSongQueryError(st, songStr, e)
		return
	}

	song := matchSong[0]

	diffKey, ok := parseDiffKey(diffStr)
	if !ok {
		sendInvalidDiffError(st, diffStr, e)
		return
	}

	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistError(st, diffKey, song.AltTitle, e)
		return
	}

	if chart.CC == 0.0 {
		sendCcUnknownError(st, diffKey, song.AltTitle, e)
		return
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

	sendReply(st, embedbuilder.Info(embed), e)
}

func getPttFromCc(score int, cc float64) float64 {
	if score >= 10000000 {
		return cc + 2.0
	} else if score >= 9800000 && score < 10000000 {
		return cc + 1.0 + ((float64(score) - 9800000) / 200000)
	} else {
		return cc + (float64(score)-9500000)/300000
	}
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
