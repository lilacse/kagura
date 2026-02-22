package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

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

var noMatchError = errors.New("no matching song found")

func NewSongHandler(store *store.Store, songdata *songdata.Service) *songHandler {
	return &songHandler{
		cmd: cmd{
			cmds: []string{"song"},
			params: [][]param{
				{
					{
						name: "query",
					},
				},
			},
		},
		store:    store,
		songdata: songdata,
	}
}

func (h *songHandler) HandleTextCommand(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	if params == "" {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	res, err := getSongCmdResultEmbed(h, params)
	if err == noMatchError {
		sendSongQueryError(st, params, e)
		return true
	}

	sendReply(st, res, e)

	return true
}

func (h *songHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "song" {
		return false
	}

	st := h.store.Bot.State()

	query := data.Options.Find("query").String()
	res, err := getSongCmdResultEmbed(h, query)

	if err == noMatchError {
		sendSongQueryCommandError(st, query, e)
		return true
	}

	sendInteractionResponse(st, res, []discord.ContainerComponent{}, e)

	return true
}

func getSongCmdResultEmbed(h *songHandler, query string) (discord.Embed, error) {
	matched := h.songdata.Search(query, 1)
	if len(matched) == 0 {
		return discord.Embed{}, noMatchError
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

	ytQuery := url.QueryEscape(fmt.Sprintf("Arcaea %s Chart View", song.Title))
	chartViewLink := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", ytQuery)

	linksText := fmt.Sprintf("\u2002▹\u2002[Find Chart View on YouTube](%s)", chartViewLink)

	if song.Urls["fandom"] != "" {
		linksText = fmt.Sprintf("%s\n▹\u2002[Fandom](%s)", linksText, song.Urls["fandom"])
	}

	if song.Urls["mcd.blue"] != "" {
		linksText = fmt.Sprintf("%s\n▹\u2002[Arcaea中文维基](%s)", linksText, song.Urls["mcd.blue"])
	}

	embedFields := []discord.EmbedField{
		{
			Name:  "Title",
			Value: song.EscapedTitle(),
		},
		{
			Name:  "Artist",
			Value: song.EscapedArtist(),
		},
		{
			Name:  "",
			Value: "**Charts**",
		},
	}

	for _, chart := range song.Charts {
		embedFields = append(embedFields, discord.EmbedField{
			Name:   chart.GetDiffDisplayName(),
			Value:  fmt.Sprintf("Lv%s (%s) (v%s)", chart.Level, chart.GetCCString(), chart.Ver),
			Inline: true,
		})
	}

	embedFields = append(embedFields, discord.EmbedField{
		Name:  "",
		Value: linksText,
	})

	songEmbed := discord.Embed{
		Fields: embedFields,
	}

	return embedbuilder.Info(songEmbed), nil
}
