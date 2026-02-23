package game

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var p = message.NewPrinter(language.English)

func fmtNum(n int) string {
	return p.Sprintf("%d", n)
}
