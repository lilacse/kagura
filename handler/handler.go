package handler

import (
	"github.com/lilacse/kagura/store"
)

type handlerFactory struct {
	store *store.Store
}

func NewHandlerFactory(store *store.Store) *handlerFactory {
	return &handlerFactory{store: store}
}

func (f *handlerFactory) NewOnMessageCreateHandler() *onMessageCreateHandler {
	return &onMessageCreateHandler{store: f.store}
}
