package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/google/uuid"
	"github.com/lilacse/kagura/commands"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type onMessageCreateHandler struct {
	store    *store.Store
	db       *database.Service
	datasvcs *dataservices.Provider
}

type commandHandler func(ctx context.Context, e *gateway.MessageCreateEvent) bool

func (h *onMessageCreateHandler) Handle(e *gateway.MessageCreateEvent) {
	if !e.Author.Bot {
		if isCommand(e, h.store.Bot.Prefix()) {
			handleCommand(e, h)
		}
	}
}

func isCommand(e *gateway.MessageCreateEvent, prefix string) bool {
	return strings.HasPrefix(e.Content, prefix)
}

func handleCommand(e *gateway.MessageCreateEvent, h *onMessageCreateHandler) {
	traceId := uuid.NewString()
	ctx := context.WithValue(h.store.Bot.Context(), logger.TraceId, traceId)

	handlers := []commandHandler{
		commands.NewSongHandler(h.store, h.datasvcs.SongData()).Handle,
		commands.NewPttHandler(h.store, h.datasvcs.SongData()).Handle,
		commands.NewStepHandler(h.store, h.datasvcs.SongData()).Handle,
		commands.NewSaveHandler(h.store, h.db, h.datasvcs.SongData()).Handle,
		commands.NewUnsaveHandler(h.store, h.db, h.datasvcs.SongData()).Handle,
		commands.NewB30Handler(h.store, h.db, h.datasvcs.SongData()).Handle,
		commands.NewScoresHandler(h.store, h.db, h.datasvcs.SongData()).Handle,
	}

	defer func() {
		r := recover()
		if r != nil {
			logger.Error(ctx, fmt.Sprintf("error handling command: %s\nstack trace: %s", r, debug.Stack()))

			st := h.store.Bot.State()

			d := api.SendMessageData{
				Embeds: []discord.Embed{
					embedbuilder.Error(ctx, fmt.Sprintf("%s", r)),
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
	}()

	for _, handler := range handlers {
		handled := handler(ctx, e)
		if handled {
			return
		}
	}
}
