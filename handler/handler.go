package handler

import (
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
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
