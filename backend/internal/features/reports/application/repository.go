package application

import (
	"context"
	"time"

	"pdd-service/internal/core/domain/violations"

	"github.com/google/uuid"
)

type ReportsRepository interface {
	CreateReport(ctx context.Context, report violations.Report) error
	GetReportByID(ctx context.Context, id uuid.UUID) (violations.Report, error)
	UpdateReport(ctx context.Context, report violations.Report) error
	DeleteReport(ctx context.Context, id uuid.UUID) error
	ListReports(ctx context.Context, filter ListReportsFilter) ([]violations.Report, error)
	CountReports(ctx context.Context, filter ListReportsFilter) (int64, error)
	ListReportsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]violations.Report, error)
}

type ViolationTypesRepository interface {
	GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error)
}

type ListReportsFilter struct {
	UserID          *uuid.UUID
	Status          *violations.Status
	ViolationTypeID *uuid.UUID
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Limit           int
	Offset          int
}
