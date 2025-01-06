package commands

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/embedbuilder"
)

func sendFormatError(st *state.State, prefix string, handler cmd, e *gateway.MessageCreateEvent) {
	paramList := make([]string, 0, len(handler.params))
	for _, p := range handler.params {
		if !p.isOpt {
			paramList = append(paramList, fmt.Sprintf("[%s]", p.name))
		} else {
			paramList = append(paramList, fmt.Sprintf("(%s)", p.name))
		}
	}

	format := fmt.Sprintf("%s%s %s", prefix, strings.Join(handler.cmds, "/"), strings.Join(paramList, " "))

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%s`!", format)))
}

func sendSongQueryError(st *state.State, query string, e *gateway.MessageCreateEvent) {
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("No matching song found for query `%s`!", query)))
}

func sendInvalidDiffError(st *state.State, diffStr string, e *gateway.MessageCreateEvent) {
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid difficulty `%s`!", diffStr)))
}

func sendDiffNotExistError(st *state.State, diffKey string, songAltTitle string, e *gateway.MessageCreateEvent) {
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Difficulty %s does not exist for the song %s!", strings.ToUpper(diffKey), songAltTitle)))
}
