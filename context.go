package main

import "golang.org/x/net/context"

type ctxKey int

const (
	ctxConfigKey ctxKey = iota
	ctxUserKey
	ctxUserStoreKey
	ctxSessionKey
)

func ConfigFromContext(ctx context.Context) Config {
	return ctx.Value(ctxConfigKey).(Config)
}
func ContextWithConfig(parent context.Context, c Config) context.Context {
	return context.WithValue(parent, ctxConfigKey, c)
}

func UserFromContext(ctx context.Context) User {
	return ctx.Value(ctxUserKey).(User)
}
func ContextWithUser(parent context.Context, u User) context.Context {
	return context.WithValue(parent, ctxUserKey, u)
}

func UserStoreFromContext(ctx context.Context) UserStore {
	return ctx.Value(ctxUserStoreKey).(UserStore)
}
func ContextWithUserStore(parent context.Context, us UserStore) context.Context {
	return context.WithValue(parent, ctxUserStoreKey, us)
}

func SessionFromContext(ctx context.Context) Session {
	return ctx.Value(ctxSessionKey).(Session)
}
func ContextWithSession(parent context.Context, s Session) context.Context {
	return context.WithValue(parent, ctxSessionKey, s)
}
