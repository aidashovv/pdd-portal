package application

import (
	"time"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type ListInput struct {
	UserID        *uuid.UUID
	Status        *payouts.Status
	Limit         int
	Offset        int
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type ListOutput struct {
	Payouts []PayoutOutput
	Total   int64
	Limit   int
	Offset  int
}

type GetByIDInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type GetByIDOutput struct {
	Payout PayoutOutput
}

type ListByUserIDInput struct {
	UserID        uuid.UUID
	Limit         int
	Offset        int
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type ListByUserIDOutput struct {
	Payouts []PayoutOutput
	Total   int64
	Limit   int
	Offset  int
}

type CreateRuleInput struct {
	ViolationTypeID uuid.UUID
	Percent         string
	CurrentRole     users.Role
}

type CreateRuleOutput struct {
	Rule RuleOutput
}

type GetActiveRuleByViolationTypeIDInput struct {
	ViolationTypeID uuid.UUID
	CurrentRole     users.Role
}

type GetActiveRuleByViolationTypeIDOutput struct {
	Rule RuleOutput
}

type ListRulesInput struct {
	ViolationTypeID *uuid.UUID
	OnlyActive      *bool
	Limit           int
	Offset          int
	CurrentRole     users.Role
}

type ListRulesOutput struct {
	Rules  []RuleOutput
	Total  int64
	Limit  int
	Offset int
}

type UpdateRuleInput struct {
	ID          uuid.UUID
	Percent     string
	CurrentRole users.Role
}

type UpdateRuleOutput struct {
	Rule RuleOutput
}

type ActivateRuleInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type ActivateRuleOutput struct {
	Rule RuleOutput
}

type DeactivateRuleInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type DeactivateRuleOutput struct {
	Rule RuleOutput
}

type CreatePayoutForApprovedReportInput struct {
	ReportID    uuid.UUID
	CurrentRole users.Role
}

type CreatePayoutForApprovedReportOutput struct {
	Payout PayoutOutput
}

type MarkPaidInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type MarkPaidOutput struct {
	Payout PayoutOutput
}

type MarkFailedInput struct {
	ID          uuid.UUID
	CurrentRole users.Role
}

type MarkFailedOutput struct {
	Payout PayoutOutput
}

type PayoutOutput struct {
	ID        uuid.UUID
	ReportID  uuid.UUID
	UserID    uuid.UUID
	Amount    string
	Status    payouts.Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RuleOutput struct {
	ID              uuid.UUID
	ViolationTypeID uuid.UUID
	Percent         string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
