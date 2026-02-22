package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/google/uuid"
	"github.com/lilacse/kagura/commands"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type onInteractionCreateHandler struct {
	store    *store.Store
	db       *database.Service
	datasvcs *dataservices.Provider
}

type interactionType int

const (
	componentInteraction interactionType = iota
	commandInteraction
)

type interactionHandler func(ctx context.Context, e *gateway.InteractionCreateEvent) bool

func (h *onInteractionCreateHandler) Handle(e *gateway.InteractionCreateEvent) {
	if e.Data.InteractionType() == discord.ComponentInteractionType {
		if !e.Message.Timestamp.Time().After(time.Now().Add(-10 * time.Minute)) {
			return
		}

		switch e.Data.(type) {
		case *discord.ButtonInteraction:
			params := strings.Split(string(e.Data.(*discord.ButtonInteraction).CustomID), ",")
			if params[0] == e.SenderID().String() {
				handleInteraction(e, h, componentInteraction)
			}
		}
	} else if e.Data.InteractionType() == discord.CommandInteractionType {
		switch e.Data.(type) {
		case *discord.CommandInteraction:
			handleInteraction(e, h, commandInteraction)
		}
	}
}

func handleInteraction(e *gateway.InteractionCreateEvent, h *onInteractionCreateHandler, t interactionType) {
	traceId := uuid.NewString()
	ctx := context.WithValue(h.store.Bot.Context(), logger.TraceId, traceId)

	componentHandlers := []interactionHandler{
		commands.NewScoresHandler(h.store, h.db, h.datasvcs.SongData()).HandleScorePageSelect,
		commands.NewB30Handler(h.store, h.db, h.datasvcs.SongData()).HandleB30PageSelect,
	}

	commandHandlers := []interactionHandler{
		commands.NewSongHandler(h.store, h.datasvcs.SongData()).HandleSlashCommand,
	}

	defer func() {
		r := recover()
		if r != nil {
			logger.Error(ctx, fmt.Sprintf("error handling interaction: %s\nstack trace: %s", r, debug.Stack()))
			switch t {
			case componentInteraction:
				sendHandleError(ctx, r, h.store.Bot.State(), e.Message.ID, e.ChannelID)
			case commandInteraction:
				sendCommandError(ctx, r, h.store.Bot.State(), e)
			}
		}
	}()

	var hs []interactionHandler
	switch t {
	case componentInteraction:
		hs = componentHandlers
	case commandInteraction:
		hs = commandHandlers
	}

	for _, h := range hs {
		handled := h(ctx, e)
		if handled {
			return
		}
	}
}
