package handler

import (
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices"
	"github.com/lilacse/kagura/store"
)

type handlerFactory struct {
	store    *store.Store
	db       *database.DbService
	datasvcs *dataservices.Provider
}

func NewHandlerFactory(store *store.Store, db *database.DbService, datasvcs *dataservices.Provider) *handlerFactory {
	return &handlerFactory{
		store:    store,
		db:       db,
		datasvcs: datasvcs,
	}
}

func (f *handlerFactory) NewOnMessageCreateHandler() *onMessageCreateHandler {
	return &onMessageCreateHandler{
		store:    f.store,
		db:       f.db,
		datasvcs: f.datasvcs,
	}
}
