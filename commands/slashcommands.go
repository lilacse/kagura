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
	diffChoices := []discord.StringChoice{
		{Name: "Past", Value: "pst"},
		{Name: "Present", Value: "prs"},
		{Name: "Future", Value: "ftr"},
		{Name: "Beyond", Value: "byd"},
		{Name: "Eternal", Value: "etr"},
	}

	levelChoices := []discord.StringChoice{
		{Name: "Lv1", Value: "1"},
		{Name: "Lv2", Value: "2"},
		{Name: "Lv3", Value: "3"},
		{Name: "Lv4", Value: "4"},
		{Name: "Lv5", Value: "5"},
		{Name: "Lv6", Value: "6"},
		{Name: "Lv7", Value: "7"},
		{Name: "Lv7+", Value: "7+"},
		{Name: "Lv8", Value: "8"},
		{Name: "Lv8+", Value: "8+"},
		{Name: "Lv9", Value: "9"},
		{Name: "Lv9+", Value: "9+"},
		{Name: "Lv10", Value: "10"},
		{Name: "Lv10+", Value: "10+"},
		{Name: "Lv11", Value: "11"},
		{Name: "Lv11+", Value: "11+"},
		{Name: "Lv12", Value: "12"},
	}

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
					Choices:     diffChoices,
				},
				&discord.IntegerOption{
					OptionName:  "score",
					Description: "The score of the play, supports short score format (e.g. 980 instead of 9800000)",
					Required:    true,
				},
			},
		},
		{
			Name:        "save",
			Description: "Saves a score",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "song",
					Description: "Search term for the song",
					Required:    true,
				},
				&discord.StringOption{
					OptionName:  "diff",
					Description: "The difficulty of the chart",
					Required:    true,
					Choices:     diffChoices,
				},
				&discord.IntegerOption{
					OptionName:  "score",
					Description: "The score of the play",
					Required:    true,
				},
			},
		},
		{
			Name:        "unsave",
			Description: "Unsaves a score",
			Options: []discord.CommandOption{
				&discord.IntegerOption{
					OptionName:  "score_id",
					Description: "The ID of the score to unsave",
					Required:    true,
				},
			},
		},
		{
			Name:        "ptt",
			Description: "Calculates the rating of a play",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "song",
					Description: "Search term for the song",
					Required:    true,
				},
				&discord.StringOption{
					OptionName:  "diff",
					Description: "The difficulty of the chart",
					Required:    true,
					Choices:     diffChoices,
				},
				&discord.IntegerOption{
					OptionName:  "score",
					Description: "The score of the play, supports short score format (e.g. 980 instead of 9800000)",
					Required:    true,
				},
			},
		},
		{
			Name:        "random",
			Description: "Returns a random song",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "level",
					Description: "The level of the chart",
					Required:    false,
					Choices:     levelChoices,
				},
				&discord.StringOption{
					OptionName:  "diff",
					Description: "The difficulty of the chart",
					Required:    false,
					Choices:     diffChoices,
				},
			},
		},
	}

	err := cmdroute.OverwriteCommands(st, cmds)
	if err != nil {
		logger.Fatal(ctx, "failed to register slash commands")
	}
}
