package game

import "strings"

func canonicalPlayerName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
