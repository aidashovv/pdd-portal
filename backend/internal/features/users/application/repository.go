package application

import (
	"context"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (users.User, error)
	ListUsers(ctx context.Context, filter ListUsersFilter) ([]users.User, error)
	CountUsers(ctx context.Context, filter ListUsersFilter) (int64, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role users.Role) error
}

type ListUsersFilter struct {
	Role   *users.Role
	Search string
	Limit  int
	Offset int
}
