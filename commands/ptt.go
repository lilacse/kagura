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

	score, errMsg, ok := parseShortScore(scoreStr)
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

	diffKey, ok := parseDiffKey(diffStr)
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

	var formula string
	ptt := chart.GetScoreRating(score)

	if score >= 10000000 {
		formula = fmt.Sprintf("%.1f + 2.0 = **%.4f**", chart.CC, ptt)
	} else if score >= 9800000 && score < 10000000 {
		formula = fmt.Sprintf("%.1f + 1.0 + ((%v - 9800000) / 200000) = **%.4f**", chart.CC, score, ptt)
	} else {
		if ptt >= 0.0 {
			formula = fmt.Sprintf("%.1f + (%v - 9500000) / 300000 = **%.4f**", chart.CC, score, ptt)
		} else {
			formula = fmt.Sprintf("%.1f + (%v - 9500000) / 300000 = %.4f (considered as **0.0**)", chart.CC, score, ptt)
		}

	}

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

	return true
}
