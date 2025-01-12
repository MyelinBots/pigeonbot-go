package context_manager

import (
	"context"
	"strings"
)

type Nick struct{}

func SetNickContext(ctx context.Context, nick string) context.Context {
	return context.WithValue(ctx, Nick{}, strings.ToLower(nick))
}

func GetNickContext(ctx context.Context) string {
	return ctx.Value(Nick{}).(string)
}
