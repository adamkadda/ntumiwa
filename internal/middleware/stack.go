package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

/*
	Thank you DreamsOfCode for this idea.
	I came across a couple of other methods of
	chaining middleware but I found yours quite
	elegant.

	https://github.com/dreamsofcode-io/nethttp/blob/main/middleware/middleware.go
*/

func NewStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		// iterate from the outer-most middleware first
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			// assign next to be the current middleware, linked to the rest via their own .next
			next = x(next)
		}

		return next
	}
}
