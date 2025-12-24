package context_manager

import (
	"context"
	"strings"
)

// unexported key type prevents collisions
type nickKeyType struct{}

var nickKey = nickKeyType{}

// WithNick stores the nick in context (normalized to lowercase)
func WithNick(ctx context.Context, nick string) context.Context {
	return context.WithValue(ctx, nickKey, strings.ToLower(nick))
}

// GetNickContext returns nick from context, or "" if missing
func GetNickContext(ctx context.Context) string {
	v := ctx.Value(nickKey)
	nick, ok := v.(string)
	if !ok {
		return ""
	}
	return nick
}
