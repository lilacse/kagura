package store

type Store struct {
	Bot bot
}

var store Store = Store{}

func GetStore() *Store {
	return &store
}
