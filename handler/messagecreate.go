package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/google/uuid"
	"github.com/lilacse/kagura/commands/ptt"
	"github.com/lilacse/kagura/commands/song"
	"github.com/lilacse/kagura/commands/step"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type onMessageCreateHandler struct {
	store *store.Store
}

type commandHandler func(ctx context.Context, e *gateway.MessageCreateEvent) bool

func (h *onMessageCreateHandler) Handle(e *gateway.MessageCreateEvent) {
	if !e.Author.Bot {
		if isCommand(e, h.store.Bot.Prefix()) {
			handleCommand(e, h.store)
		}
	}
}

func isCommand(e *gateway.MessageCreateEvent, prefix string) bool {
	return strings.HasPrefix(e.Content, prefix)
}

func handleCommand(e *gateway.MessageCreateEvent, store *store.Store) {
	traceId := uuid.NewString()
	ctx := context.WithValue(store.Bot.Context(), logger.TraceId, traceId)

	handlers := []commandHandler{
		song.NewHandler(store).Handle,
		ptt.NewHandler(store).Handle,
		step.NewHandler(store).Handle,
	}

	defer func() {
		r := recover()
		if r != nil {
			logger.Error(ctx, fmt.Sprintf("error handling command: %s\nstack trace: %s", r, debug.Stack()))

			st := store.Bot.State()
			st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Error(ctx, fmt.Sprintf("%s", r)))
		}
	}()

	for _, handler := range handlers {
		isHandled := handler(ctx, e)
		if isHandled {
			return
		}
	}
}
