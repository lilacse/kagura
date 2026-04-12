package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
)

func parseShortScore(s string) (int, string, bool) {
	score, err := strconv.Atoi(s)
	if err != nil || score > 10009999 || score < 0 {
		return -1, fmt.Sprintf("Invalid score `%s`!", s), false
	} else if score < 100 {
		return -1, fmt.Sprintf("Invalid score `%s`, expecting at least 3 digits!", s), false
	}

	// treat scores submitted with 3 digits to 6 digits, we append zeroes to them until it reaches 7 digits
	if score == 100 {
		score = 10000000
	} else {
		for score < 1000000 {
			score *= 10
		}
	}

	return score, "", true
}

func parseFullScore(s string) (int, string, bool) {
	score, err := strconv.Atoi(s)
	if err != nil || score > 10009999 {
		return -1, fmt.Sprintf("Invalid full score `%s`!", s), false
	} else if score < 1000000 {
		return -1, fmt.Sprintf("Invalid full score `%s`, expecting at least 7 digits!", s), false
	}

	return score, "", true
}

func parseScoreId(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return -1, false
	}

	return id, true
}

func getFullDiffName(diffKey string) string {
	switch diffKey {
	case "pst":
		return "Past"
	case "prs":
		return "Present"
	case "ftr":
		return "Future"
	case "etr":
		return "Eternal"
	case "byd":
		return "Beyond"
	default:
		return ""
	}
}

func logAndSendCommandError(ctx context.Context, st *state.State, err error, e *gateway.InteractionCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling slash command: %s", err.Error()))
	sendCommandErrorReply(st, err.Error(), e)
}

func logAndSendInteractionError(ctx context.Context, st *state.State, err error, e *gateway.InteractionCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling interaction: %s", err.Error()))
	sendInteractionReply(st, embedbuilder.Error(ctx, err.Error()), e)
}
