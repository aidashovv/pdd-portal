package application

import (
	"time"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type CreateInput struct {
	Code           string
	Title          string
	Description    string
	BaseFineAmount string
	CurrentRole    users.Role
}

type CreateOutput struct {
	ViolationType ViolationTypeOutput
}

type GetByIDInput struct {
	ID uuid.UUID
}

type GetByIDOutput struct {
	ViolationType ViolationTypeOutput
}

type GetByCodeInput struct {
	Code string
}

type GetByCodeOutput struct {
	ViolationType ViolationTypeOutput
}

type ListInput struct {
	OnlyActive *bool
	Search     string
	Limit      int
	Offset     int
}

type ListOutput struct {
	ViolationTypes []ViolationTypeOutput
	Total          int64
	Limit          int
	Offset         int
}

type UpdateInput struct {
	ID             uuid.UUID
	Title          *string
	Description    *string
	BaseFineAmount *string
	CurrentRole    users.Role
}

type UpdateOutput struct {
	ViolationType ViolationTypeOutput
}

type DeleteInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type ActivateInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type ActivateOutput struct {
	ViolationType ViolationTypeOutput
}

type DeactivateInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type DeactivateOutput struct {
	ViolationType ViolationTypeOutput
}

type ViolationTypeOutput struct {
	ID             uuid.UUID
	Code           string
	Title          string
	Description    string
	BaseFineAmount string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
