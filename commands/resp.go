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

func sendInteractionReply(st *state.State, em discord.Embed, e *gateway.InteractionCreateEvent) {
	sendReplyWithComponents(st, em, []discord.ContainerComponent{}, e.ChannelID, e.Message.ID)
}

func sendCommandReply(st *state.State, em discord.Embed, e *gateway.InteractionCreateEvent) {
	sendInteractionResponse(st, em, []discord.ContainerComponent{}, e)
}

func sendCommandErrorReply(st *state.State, msg string, e *gateway.InteractionCreateEvent) {
	sendCommandReply(st, embedbuilder.UserError(msg), e)
}

func sendReplyWithComponents(st *state.State, em discord.Embed, cc []discord.ContainerComponent, channelId discord.ChannelID, replyId discord.MessageID) {
	d := api.SendMessageData{
		Embeds: []discord.Embed{
			em,
		},
		Components: cc,
		Reference: &discord.MessageReference{
			MessageID: replyId,
		},
		AllowedMentions: &api.AllowedMentions{
			RepliedUser: option.False,
		},
	}

	st.SendMessageComplex(channelId, d)
}

func sendInteractionResponse(st *state.State, em discord.Embed, cc []discord.ContainerComponent, e *gateway.InteractionCreateEvent) {
	ccs := discord.ContainerComponents{}
	for _, c := range cc {
		ccs = append(ccs, c)
	}

	d := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				em,
			},
			Components: &ccs,
			AllowedMentions: &api.AllowedMentions{
				RepliedUser: option.False,
			},
		},
	}

	st.RespondInteraction(e.InteractionEvent.ID, e.InteractionEvent.Token, d)
}

func sendSongQueryCommandError(st *state.State, query string, e *gateway.InteractionCreateEvent) {
	sendCommandErrorReply(st, fmt.Sprintf("No matching song found for query `%s`!", query), e)
}

func sendDiffNotExistCommandError(st *state.State, diffKey string, songAltTitle string, e *gateway.InteractionCreateEvent) {
	sendCommandErrorReply(st, fmt.Sprintf("Difficulty %s does not exist for the song %s!", strings.ToUpper(diffKey), songAltTitle), e)
}

func sendCcUnknownCommandError(st *state.State, diffKey string, songAltTitle string, e *gateway.InteractionCreateEvent) {
	sendCommandErrorReply(st, fmt.Sprintf("Chart constant is unknown for the difficulty %s for the song %s!", strings.ToUpper(diffKey), songAltTitle), e)
}
