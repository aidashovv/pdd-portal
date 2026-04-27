package application

import (
	"context"

	"pdd-service/internal/core/domain/violations"

	"github.com/google/uuid"
)

type Repository interface {
	CreateViolationType(ctx context.Context, violationType violations.ViolationType) error
	GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error)
	GetViolationTypeByCode(ctx context.Context, code string) (violations.ViolationType, error)
	ListViolationTypes(ctx context.Context, filter ListViolationTypesFilter) ([]violations.ViolationType, error)
	CountViolationTypes(ctx context.Context, filter ListViolationTypesFilter) (int64, error)
	UpdateViolationType(ctx context.Context, violationType violations.ViolationType) error
	DeleteViolationType(ctx context.Context, id uuid.UUID) error
}

type ListViolationTypesFilter struct {
	OnlyActive *bool
	Search     string
	Limit      int
	Offset     int
}
