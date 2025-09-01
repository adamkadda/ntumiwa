package cookies

import "context"

type cookieContextKey struct{}

var cookieKey = cookieContextKey{}

func WithCookieWritten(ctx context.Context) context.Context {
	return context.WithValue(ctx, cookieKey, true)
}

func CookieWritten(ctx context.Context) bool {
	// type assertion even though nobody's gonna touch this
	val, ok := ctx.Value(cookieKey).(bool)
	return ok && val
}
