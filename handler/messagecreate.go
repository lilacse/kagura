package handler

import (
	"context"
	"strings"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/google/uuid"
	"github.com/lilacse/kagura/commands/ptt"
	"github.com/lilacse/kagura/commands/song"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type CommandHandler func(ctx context.Context, e *gateway.MessageCreateEvent) (bool, error)

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
	}

	for _, handler := range handlers {
		isHandled, err := handler(ctx, e)
		if err != nil {
			logger.Error(ctx, "error handling command: "+err.Error())
			return
		}

		if isHandled {
			return
		}
	}
}
