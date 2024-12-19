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

type CommandHandler func(ctx context.Context, e *gateway.MessageCreateEvent) bool

func OnMessageCreate(e *gateway.MessageCreateEvent) {
	if !e.Author.Bot {
		if isCommand(e) {
			handleCommand(e)
		}
	}
}

func isCommand(e *gateway.MessageCreateEvent) bool {
	return strings.HasPrefix(e.Content, store.GetPrefix())
}

func handleCommand(e *gateway.MessageCreateEvent) {
	traceId := uuid.NewString()
	ctx := context.WithValue(store.GetContext(), logger.TraceId, traceId)

	handlers := []CommandHandler{
		song.Handle,
		ptt.Handle,
		step.Handle,
	}

	defer func() {
		r := recover()
		if r != nil {
			logger.Error(ctx, fmt.Sprintf("error handling command: %s\nstack trace: %s", r, debug.Stack()))

			st := store.GetState()
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
