package commands

import (
	"strings"
	"unicode"

	"github.com/lilacse/kagura/store"
)

func ExtractParamsString(cmd string, content string) (string, bool) {
	cmd = store.GetPrefix() + cmd
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

func ExtractParamReverse(param string, wordCount int) (string, string, bool) {
	i := len(param) - 1

	if i < 0 {
		return "", "", false
	}

	// skips over trailing spaces
	for i >= 0 {
		r := rune(param[i])
		if !unicode.IsSpace(r) {
			break
		}
		i--
	}

	endIdx := i + 1

	if endIdx == 0 {
		return "", "", false
	}

	for i >= 0 {
		r := rune(param[i])
		if unicode.IsSpace(r) {
			wordCount--
		}
		if wordCount == 0 {
			break
		}
		i--
	}

	if i < 0 {
		wordCount--
	}

	if wordCount > 0 {
		return "", "", false
	}

	startIdx := i + 1

	return param[0:startIdx], param[startIdx:endIdx], true
}
