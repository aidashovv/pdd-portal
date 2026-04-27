package middlewares

import (
	"net/http"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/response"
)

func RequireRoles(allowedRoles ...users.Role) func(http.Handler) http.Handler {
	allowed := make(map[users.Role]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := helpers.GetUserRole(r.Context())
			if !ok {
				response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
				return
			}
			if _, ok := allowed[role]; !ok {
				response.NewResponseHandler(w).HandleError(coreerrors.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
