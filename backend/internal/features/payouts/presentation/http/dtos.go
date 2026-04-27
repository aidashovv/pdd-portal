package http

import "time"

type CreatePayoutFromReportRequest struct {
	ReportID string `json:"report_id" validate:"required"`
}

type CreateRuleRequest struct {
	ViolationTypeID string `json:"violation_type_id" validate:"required"`
	Percent         string `json:"percent" validate:"required"`
}

type UpdateRuleRequest struct {
	Percent string `json:"percent" validate:"required"`
}

type PayoutResponse struct {
	ID        string    `json:"id"`
	ReportID  string    `json:"report_id"`
	UserID    string    `json:"user_id"`
	Amount    string    `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RuleResponse struct {
	ID              string    `json:"id"`
	ViolationTypeID string    `json:"violation_type_id"`
	Percent         string    `json:"percent"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListPayoutsResponse struct {
	Payouts []PayoutResponse `json:"payouts"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

type GetPayoutResponse struct {
	Payout PayoutResponse `json:"payout"`
}

type CreatePayoutFromReportResponse struct {
	Payout PayoutResponse `json:"payout"`
}

type MarkPayoutResponse struct {
	Payout PayoutResponse `json:"payout"`
}

type ListRulesResponse struct {
	Rules  []RuleResponse `json:"rules"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type CreateRuleResponse struct {
	Rule RuleResponse `json:"rule"`
}

type UpdateRuleResponse struct {
	Rule RuleResponse `json:"rule"`
}

type RuleStatusResponse struct {
	Rule RuleResponse `json:"rule"`
}
