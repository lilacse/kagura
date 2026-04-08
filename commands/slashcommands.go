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
			Name:        "song",
			Description: "Queries for a song",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "query",
					Description: "Search term for the song",
					Required:    true,
				},
			},
		},
		{
			Name:        "step",
			Description: "Calculates the amount of steps a play gives in World Mode",
			Options: []discord.CommandOption{
				&discord.NumberOption{
					OptionName:  "stat",
					Description: "The STEP stat of the partner",
					Required:    true,
				},
				&discord.StringOption{
					OptionName:  "song",
					Description: "Search term for the song",
					Required:    true,
				},
				&discord.StringOption{
					OptionName:  "diff",
					Description: "The difficulty of the chart",
					Required:    true,
					Choices: []discord.StringChoice{
						{Name: "Past", Value: "pst"},
						{Name: "Present", Value: "prs"},
						{Name: "Future", Value: "ftr"},
						{Name: "Beyond", Value: "byd"},
						{Name: "Eternal", Value: "etr"},
					},
				},
				&discord.IntegerOption{
					OptionName:  "score",
					Description: "The score of the play, supports short score format (e.g. 980 instead of 9800000)",
					Required:    true,
				},
			},
		},
	}

	err := cmdroute.OverwriteCommands(st, cmds)
	if err != nil {
		logger.Fatal(ctx, "failed to register slash commands")
	}
}
