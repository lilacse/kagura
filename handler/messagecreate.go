package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/google/uuid"
	"github.com/lilacse/kagura/commands/ptt"
	"github.com/lilacse/kagura/commands/save"
	"github.com/lilacse/kagura/commands/song"
	"github.com/lilacse/kagura/commands/step"
	"github.com/lilacse/kagura/commands/unsave"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type onMessageCreateHandler struct {
	store *store.Store
	db    *database.DbService
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
		song.NewHandler(h.store).Handle,
		ptt.NewHandler(h.store).Handle,
		step.NewHandler(h.store).Handle,
		save.NewHandler(h.store, h.db).Handle,
		unsave.NewHandler(h.store, h.db).Handle,
	}

	defer func() {
		r := recover()
		if r != nil {
			logger.Error(ctx, fmt.Sprintf("error handling command: %s\nstack trace: %s", r, debug.Stack()))

			st := h.store.Bot.State()
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
