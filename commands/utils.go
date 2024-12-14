package commands

import (
	"strings"
	"unicode"

	"github.com/lilacse/kagura/store"
)

func ExtractParams(cmd string, content string) (string, bool) {
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
