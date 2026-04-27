package application

import (
	"time"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type GetByIDInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type GetByIDOutput struct {
	User UserOutput
}

type ListInput struct {
	CurrentRole users.Role
	Role        *users.Role
	Search      string
	Limit       int
	Offset      int
}

type ListOutput struct {
	Users  []UserOutput
	Total  int64
	Limit  int
	Offset int
}

type UpdateRoleInput struct {
	ID          uuid.UUID
	Role        users.Role
	CurrentRole users.Role
}

type UpdateRoleOutput struct {
	User UserOutput
}

type UserOutput struct {
	ID        uuid.UUID
	Email     string
	FullName  string
	Role      users.Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
