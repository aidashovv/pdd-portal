package middlewares

import (
	"net/http"
	"strings"

	coreauth "pdd-service/internal/core/auth"
	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/response"
)

type AccessTokenParser interface {
	ParseAccessToken(raw string) (coreauth.Claims, error)
}

func AuthMiddleware(parser AccessTokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
				return
			}

			claims, err := parser.ParseAccessToken(strings.TrimPrefix(authHeader, "Bearer "))
			if err != nil {
				response.NewResponseHandler(w).HandleError(err)
				return
			}

			ctx := helpers.WithUser(r.Context(), claims.UserID, claims.Email, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
