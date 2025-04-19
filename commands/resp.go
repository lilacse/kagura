package commands

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/lilacse/kagura/embedbuilder"
)

func sendReply(st *state.State, em discord.Embed, e *gateway.MessageCreateEvent) {
	d := api.SendMessageData{
		Embeds: []discord.Embed{
			em,
		},
		Reference: &discord.MessageReference{
			MessageID: e.ID,
		},
		AllowedMentions: &api.AllowedMentions{
			RepliedUser: option.False,
		},
	}

	st.SendMessageComplex(e.ChannelID, d)
}

func sendFormatError(st *state.State, prefix string, handler cmd, e *gateway.MessageCreateEvent) {
	formatList := make([]string, 0)

	for _, f := range handler.params {
		paramList := make([]string, 0, len(handler.params))
		for _, p := range f {
			if !p.optional {
				paramList = append(paramList, fmt.Sprintf("[%s]", p.name))
			} else {
				paramList = append(paramList, fmt.Sprintf("(%s)", p.name))
			}
		}

		format := fmt.Sprintf("%s%s %s", prefix, strings.Join(handler.cmds, "/"), strings.Join(paramList, " "))
		formatList = append(formatList, format)
	}

	if len(formatList) == 1 {
		sendReply(st, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%s`!", formatList[0])), e)
	} else {
		formatListStr := ""
		for _, f := range formatList {
			formatListStr += fmt.Sprintf("- `%s`\n", f)
		}
		sendReply(st, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting one of these formats!\n%s", formatListStr)), e)
	}
}

func sendSongQueryError(st *state.State, query string, e *gateway.MessageCreateEvent) {
	sendReply(st, embedbuilder.UserError(fmt.Sprintf("No matching song found for query `%s`!", query)), e)
}

func sendInvalidDiffError(st *state.State, diffStr string, e *gateway.MessageCreateEvent) {
	sendReply(st, embedbuilder.UserError(fmt.Sprintf("Invalid difficulty `%s`!", diffStr)), e)
}

func sendDiffNotExistError(st *state.State, diffKey string, songAltTitle string, e *gateway.MessageCreateEvent) {
	sendReply(st, embedbuilder.UserError(fmt.Sprintf("Difficulty %s does not exist for the song %s!", strings.ToUpper(diffKey), songAltTitle)), e)
}

func sendCcUnknownError(st *state.State, diffKey string, songAltTitle string, e *gateway.MessageCreateEvent) {
	sendReply(st, embedbuilder.UserError(fmt.Sprintf("Chart constant is unknown for the difficulty %s for the song %s!", strings.ToUpper(diffKey), songAltTitle)), e)
}
