package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
)

type cmd struct {
	cmds   []string
	params [][]param
}

type param struct {
	name     string
	optional bool
}

func extractParamsString(cmd string, content string, prefix string) (string, bool) {
	cmd = prefix + cmd
	content, match := strings.CutPrefix(content, cmd)
	if match && len(content) > 0 {
		match = unicode.IsSpace(rune(content[0]))
	}

	if match {
		return strings.TrimSpace(content), true
	} else {
		return "", false
	}
}

func extractParamForward(param string, count int) (string, string, bool) {
	l := len(param)
	i := 0

	if l == 0 {
		return "", "", false
	}

	// skip spaces at the start of the string
	for i < l {
		r := rune(param[i])
		if !unicode.IsSpace(r) {
			break
		}
		i++
	}

	startIdx := i

	if startIdx == l {
		return "", "", false
	}

	for i < l {
		r := rune(param[i])
		if unicode.IsSpace(r) {
			count--
		}
		if count == 0 {
			break
		}
		i++
	}

	if i == l {
		count--
	}

	if count > 0 {
		return "", "", false
	}

	endIdx := i

	return param[endIdx:l], param[startIdx:endIdx], true
}

func extractParamBackwards(param string, count int) (string, string, bool) {
	i := len(param) - 1
	s := 0

	if i < 0 {
		return "", "", false
	}

	// find the stopping index to skip spaces at the start of the string
	for s <= i {
		r := rune(param[s])
		if !unicode.IsSpace(r) {
			break
		}
		s++
	}

	if s > i {
		return "", "", false
	}

	// skip trailing spaces
	for i >= s {
		r := rune(param[i])
		if !unicode.IsSpace(r) {
			break
		}
		i--
	}

	endIdx := i + 1

	if endIdx == s {
		return "", "", false
	}

	for i >= s {
		r := rune(param[i])
		if unicode.IsSpace(r) {
			count--
		}
		if count == 0 {
			break
		}
		i--
	}

	if i < s {
		count--
	}

	if count > 0 {
		return "", "", false
	}

	startIdx := i + 1

	return param[0:startIdx], param[startIdx:endIdx], true
}

func parseStep(s string) (int, bool) {
	step, err := strconv.Atoi(s)
	if err != nil || step > 1000 || step < 0 {
		return -1, false
	}

	return step, true
}

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

func parseCc(s string) (float64, bool) {
	cc, err := strconv.ParseFloat(s, 64)
	if err != nil || cc < 0 || cc > 15 || float64(int(cc*10))/10.0 != cc {
		return -1, false
	}

	return cc, true
}

func parseUserId(s string) (discord.UserID, bool) {
	if strings.HasPrefix(s, "<@") && strings.HasSuffix(s, ">") {
		s, _ = strings.CutPrefix(s, "<@")
		s, _ = strings.CutSuffix(s, ">")
	}

	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}

	userId := discord.UserID(id)
	if !userId.IsValid() {
		return 0, false
	}

	return userId, true
}

func parseScoreId(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return -1, false
	}

	return id, true
}

func parseDiffKey(diffStr string) (string, bool) {
	switch strings.ToLower(diffStr) {
	case "pst", "past":
		return "pst", true
	case "prs", "present":
		return "prs", true
	case "ftr", "future":
		return "ftr", true
	case "etr", "eternal":
		return "etr", true
	case "byd", "beyond":
		return "byd", true
	default:
		return "", false
	}
}

func logAndSendError(ctx context.Context, st *state.State, err error, e *gateway.MessageCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling text command: %s", err.Error()))
	sendReply(st, embedbuilder.Error(ctx, err.Error()), e)
}

func logAndSendInteractionError(ctx context.Context, st *state.State, err error, e *gateway.InteractionCreateEvent) {
	logger.Error(ctx, fmt.Sprintf("error when handling interaction: %s", err.Error()))
	sendInteractionReply(st, embedbuilder.Error(ctx, err.Error()), e)
}
