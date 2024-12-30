package handler

import (
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/store"
)

type handlerFactory struct {
	store *store.Store
	db    *database.DbService
}

func NewHandlerFactory(store *store.Store, db *database.DbService) *handlerFactory {
	return &handlerFactory{
		store: store,
		db:    db,
	}
}

func (f *handlerFactory) NewOnMessageCreateHandler() *onMessageCreateHandler {
	return &onMessageCreateHandler{
		store: f.store,
		db:    f.db,
	}
}
