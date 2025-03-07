package commands

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type songHandler struct {
	cmd
	store    *store.Store
	songdata *songdata.Service
}

func NewSongHandler(store *store.Store, songdata *songdata.Service) *songHandler {
	return &songHandler{
		cmd: cmd{
			cmds: []string{"song"},
			params: []param{
				{
					name: "query",
				},
			},
		},
		store:    store,
		songdata: songdata,
	}
}

func (h *songHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	if params == "" {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	matched := h.songdata.Search(params, 1)
	if len(matched) == 0 {
		sendSongQueryError(st, params, e)
		return true
	}

	song := matched[0]
	charts := song.Charts

	diffIndex := map[string]int{
		"pst": 1,
		"prs": 2,
		"ftr": 3,
		"etr": 4,
		"byd": 5,
	}

	slices.SortFunc(charts, func(a, b songdata.Chart) int {
		return diffIndex[a.Diff] - diffIndex[b.Diff]
	})

	chartEmbeds := []discord.EmbedField{}
	for _, chart := range song.Charts {
		chartEmbeds = append(chartEmbeds, discord.EmbedField{
			Name:   chart.GetDiffDisplayName(),
			Value:  fmt.Sprintf("Lv%s (%s) (v%s)", chart.Level, chart.GetCCString(), chart.Ver),
			Inline: true,
		})
	}

	songEmbed := discord.Embed{
		Fields: []discord.EmbedField{
			{
				Name:  "Title",
				Value: song.Title,
			},
			{
				Name:  "Artist",
				Value: song.Artist,
			},
			{
				Name: "",
			},
			{
				Name:  "Charts",
				Value: "",
			},
		},
	}

	songEmbed.Fields = append(songEmbed.Fields, chartEmbeds...)

	ytQuery := url.QueryEscape(fmt.Sprintf("Arcaea %s Chart View", song.Title))
	chartViewLink := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", ytQuery)

	message := api.SendMessageData{
		Embeds: []discord.Embed{
			embedbuilder.Info(songEmbed),
		},
		Components: discord.Components(
			&discord.ButtonComponent{
				Label: "Find Chart View on YouTube",
				Style: discord.LinkButtonStyle(chartViewLink),
			},
			&discord.ButtonComponent{
				Label: "Fandom page",
				Style: discord.LinkButtonStyle(song.Url),
			},
		),
		Reference: &discord.MessageReference{
			MessageID: e.ID,
			ChannelID: e.ChannelID,
			GuildID:   e.GuildID,
		},
	}

	st.SendMessageComplex(e.ChannelID, message)

	return true
}
