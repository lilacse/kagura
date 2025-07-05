package handler

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type factory struct {
	store    *store.Store
	db       *database.Service
	datasvcs *dataservices.Provider
}

func NewFactory(store *store.Store, db *database.Service, datasvcs *dataservices.Provider) *factory {
	return &factory{
		store:    store,
		db:       db,
		datasvcs: datasvcs,
	}
}

func (f *factory) NewOnMessageCreateHandler() *onMessageCreateHandler {
	return &onMessageCreateHandler{
		store:    f.store,
		db:       f.db,
		datasvcs: f.datasvcs,
	}
}

func (f *factory) NewOnInteractionCreateHandler() *onInteractionCreateHandler {
	return &onInteractionCreateHandler{
		store:    f.store,
		db:       f.db,
		datasvcs: f.datasvcs,
	}
}

func sendHandleError(ctx context.Context, r any, st *state.State, messageId discord.MessageID, channelId discord.ChannelID) {
	d := api.SendMessageData{
		Embeds: []discord.Embed{
			embedbuilder.Error(ctx, fmt.Sprintf("%s", r)),
		},
		Reference: &discord.MessageReference{
			MessageID: messageId,
		},
		AllowedMentions: &api.AllowedMentions{
			RepliedUser: option.False,
		},
	}

	st.SendMessageComplex(channelId, d)
}
