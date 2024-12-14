package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

var botId discord.UserID
var prefix string
var ctx context.Context
var st *state.State

func SetBotId(id discord.UserID) {
	botId = id
}

func GetBotId() discord.UserID {
	return botId
}

func SetPrefix(p string) {
	prefix = p
}

func GetPrefix() string {
	return prefix
}

func SetContext(c context.Context) {
	ctx = c
}

func GetContext() context.Context {
	return ctx
}

func SetState(s *state.State) {
	st = s
}

func GetState() *state.State {
	return st
}
