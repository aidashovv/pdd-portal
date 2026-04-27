package middlewares

import "net/http"

func RateLimitMiddleware(next http.Handler) http.Handler {
	return next
}
