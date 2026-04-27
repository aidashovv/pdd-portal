package application

import (
	"context"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/violations"

	"github.com/google/uuid"
)

type PayoutsRepository interface {
	CreatePayout(ctx context.Context, payout payouts.Payout) error
	GetPayoutByID(ctx context.Context, id uuid.UUID) (payouts.Payout, error)
	ListPayouts(ctx context.Context, filter ListPayoutsFilter) ([]payouts.Payout, error)
	CountPayouts(ctx context.Context, filter ListPayoutsFilter) (int64, error)
	ListPayoutsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]payouts.Payout, error)
	UpdatePayout(ctx context.Context, payout payouts.Payout) error
}

type RulesRepository interface {
	CreateRule(ctx context.Context, rule payouts.Rule) error
	GetRuleByID(ctx context.Context, id uuid.UUID) (payouts.Rule, error)
	GetActiveRuleByViolationTypeID(ctx context.Context, violationTypeID uuid.UUID) (payouts.Rule, error)
	ListRules(ctx context.Context, filter ListRulesFilter) ([]payouts.Rule, error)
	CountRules(ctx context.Context, filter ListRulesFilter) (int64, error)
	UpdateRule(ctx context.Context, rule payouts.Rule) error
}

type ReportsRepository interface {
	GetReportByID(ctx context.Context, id uuid.UUID) (violations.Report, error)
}

type ViolationTypesRepository interface {
	GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error)
}

type ListPayoutsFilter struct {
	UserID *uuid.UUID
	Status *payouts.Status
	Limit  int
	Offset int
}

type ListRulesFilter struct {
	ViolationTypeID *uuid.UUID
	OnlyActive      *bool
	Limit           int
	Offset          int
}
