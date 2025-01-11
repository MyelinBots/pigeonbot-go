package context_manager

import "context"

type Nick struct{}

func SetNickContext(ctx context.Context, nick string) context.Context {
	return context.WithValue(ctx, Nick{}, nick)
}

func GetNickContext(ctx context.Context) string {
	return ctx.Value(Nick{}).(string)
}
