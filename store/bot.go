package store

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

type bot struct {
	botId  discord.UserID
	prefix string
	ctx    context.Context
	st     *state.State
}

func (b *bot) BotId() discord.UserID {
	return b.botId
}

func (b *bot) SetBotId(id discord.UserID) {
	b.botId = id
}

func (b *bot) Prefix() string {
	return b.prefix
}

func (b *bot) SetPrefix(p string) {
	b.prefix = p
}

func (b *bot) Context() context.Context {
	return b.ctx
}

func (b *bot) SetContext(c context.Context) {
	b.ctx = c
}

func (b *bot) State() *state.State {
	return b.st
}

func (b *bot) SetState(s *state.State) {
	b.st = s
}
