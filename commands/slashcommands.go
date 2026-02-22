package commands

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/logger"
)

func RegisterCommands(ctx context.Context, st *state.State) {
	cmds := []api.CreateCommandData{
		{
			Name: "song",
			Description: "Queries for a song",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName: "query",
					Description: "Search term for the song",
					Required: true,
				},
			},
		},
	}

	err := cmdroute.OverwriteCommands(st, cmds)
	if err != nil {
		logger.Fatal(ctx, "failed to register slash commands")
	}
}
