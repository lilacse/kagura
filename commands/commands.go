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

func ExtractParamForward(param string, wordCount int) (string, string, bool) {
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

func ExtractParamReverse(param string, wordCount int) (string, string, bool) {
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
