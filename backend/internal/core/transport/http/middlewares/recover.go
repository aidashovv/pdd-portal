package middlewares

import (
	"net/http"

	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/core/transport/http/response"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				response.NewResponseHandler(w).HandleError(coreerrors.ErrInternalError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
