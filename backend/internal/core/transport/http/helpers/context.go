package helpers

import (
	"context"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

func WithUser(ctx context.Context, userID uuid.UUID, email string, role users.Role) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UserEmailKey, email)
	ctx = context.WithValue(ctx, UserRoleKey, role)
	return ctx
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return value, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(UserEmailKey).(string)
	return value, ok
}

func GetUserRole(ctx context.Context) (users.Role, bool) {
	value, ok := ctx.Value(UserRoleKey).(users.Role)
	return value, ok
}
