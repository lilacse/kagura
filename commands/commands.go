package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
)

type cmd struct {
	cmds   []string
	params []param
}

type param struct {
	name  string
	isOpt bool
}

func extractParamsString(cmd string, content string, prefix string) (string, bool) {
	cmd = prefix + cmd
	content, isMatch := strings.CutPrefix(content, cmd)
	if isMatch && len(content) > 0 {
		isMatch = unicode.IsSpace(rune(content[0]))
	}

	if isMatch {
		return strings.TrimSpace(content), true
	} else {
		return "", false
	}
}

func extractParamForward(param string, wordCount int) (string, string, bool) {
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
			wordCount--
		}
		if wordCount == 0 {
			break
		}
		i++
	}

	if i == l {
		wordCount--
	}

	if wordCount > 0 {
		return "", "", false
	}

	endIdx := i

	return param[endIdx:l], param[startIdx:endIdx], true
}

func extractParamReverse(param string, wordCount int) (string, string, bool) {
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
			wordCount--
		}
		if wordCount == 0 {
			break
		}
		i--
	}

	if i < s {
		wordCount--
	}

	if wordCount > 0 {
		return "", "", false
	}

	startIdx := i + 1

	return param[0:startIdx], param[startIdx:endIdx], true
}

func parseScore(scoreStr string) (int, string, bool) {
	score, err := strconv.Atoi(scoreStr)
	if err != nil || score > 10009999 {
		return -1, fmt.Sprintf("Invalid score `%s`!", scoreStr), false
	} else if score < 100 {
		return -1, fmt.Sprintf("Invalid score `%s`, expecting at least 3 digits!", scoreStr), false
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

func getDiffKey(diffStr string) (string, bool) {
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
	logger.Error(ctx, fmt.Sprintf("error when handling command: %s", err.Error()))
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Error(ctx, err.Error()))
}

func sendFormatError(st *state.State, prefix string, handler cmd, e *gateway.MessageCreateEvent) {
	paramList := make([]string, 0, len(handler.params))
	for _, p := range handler.params {
		if !p.isOpt {
			paramList = append(paramList, fmt.Sprintf("[%s]", p.name))
		} else {
			paramList = append(paramList, fmt.Sprintf("(%s)", p.name))
		}
	}

	format := fmt.Sprintf("%s%s %s", prefix, strings.Join(handler.cmds, "/"), strings.Join(paramList, " "))

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%s`!", format)))
}
