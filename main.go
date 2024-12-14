package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/handler"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	store.SetContext(ctx)
	defer stop()

	logger.Info(ctx, "starting up...")

	token := os.Getenv("KAGURA_TOKEN")
	if token == "" {
		logger.Fatal(ctx, "environment variable KAGURA_TOKEN is not set")
	}

	s := state.New("Bot " + token)
	store.SetState(s)

	s.AddIntents(gateway.IntentDirectMessages)
	s.AddIntents(gateway.IntentGuildMessages)

	s.AddHandler(handler.OnMessageCreate)

	u, err := s.Me()
	if err != nil {
		logger.Fatal(ctx, "failed to get bot user with error "+err.Error())
	}

	logger.Info(ctx, fmt.Sprintf("bot user is: %s#%s (%v)", u.Username, u.Discriminator, u.ID))
	store.SetBotId(u.ID)

	prefix := os.Getenv("KAGURA_PREFIX")
	if prefix == "" {
		prefix = "~"
	}
	store.SetPrefix(prefix)

	songdata.Init(ctx)

	logger.Info(ctx, "starting connection to Discord. bot should be ready!")
	err = s.Connect(ctx)
	if err != nil {
		logger.Fatal(ctx, "connection to Discord is broken with error "+err.Error())
	}

	logger.Info(ctx, "received stopping signal, bot exiting")
}
