package handler

import (
	"context"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
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
		// this branch is currently left as blank
	}
}
