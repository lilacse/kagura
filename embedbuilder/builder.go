package embedbuilder

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/lilacse/kagura/logger"
)

var errorFooter = discord.EmbedFooter{
	Text: "Please consider reporting this to the developers!",
}

func Info(embed discord.Embed) discord.Embed {
	embed.Color = 0x3399ff
	return embed
}

func UserError(msg string) discord.Embed {
	embed := discord.Embed{
		Title:       "Oops",
		Color:       0xff5050,
		Description: msg,
	}

	return embed
}

func Error(ctx context.Context, msg string) discord.Embed {
	embed := discord.Embed{
		Title:       "Something went wrong",
		Color:       0x990033,
		Description: msg,
		Fields: []discord.EmbedField{
			{
				Name:  "Trace ID",
				Value: ctx.Value(logger.TraceId).(string),
			},
		},
		Footer: &errorFooter,
	}

	return embed
}
