package application

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	payoutsRepo        PayoutsRepository
	rulesRepo          RulesRepository
	reportsRepo        ReportsRepository
	violationTypesRepo ViolationTypesRepository
}

func NewService(
	payoutsRepo PayoutsRepository,
	rulesRepo RulesRepository,
	reportsRepo ReportsRepository,
	violationTypesRepo ViolationTypesRepository,
) *Service {
	return &Service{
		payoutsRepo:        payoutsRepo,
		rulesRepo:          rulesRepo,
		reportsRepo:        reportsRepo,
		violationTypesRepo: violationTypesRepo,
	}
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return ListOutput{}, coreerrors.ErrForbidden
	}

	filter, err := normalizePayoutsFilter(input.UserID, input.Status, input.Limit, input.Offset)
	if err != nil {
		return ListOutput{}, err
	}
	total, err := s.payoutsRepo.CountPayouts(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	domainPayouts, err := s.payoutsRepo.ListPayouts(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{Payouts: toPayoutOutputs(domainPayouts), Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (s *Service) GetByID(ctx context.Context, input GetByIDInput) (GetByIDOutput, error) {
	payout, err := s.payoutsRepo.GetPayoutByID(ctx, input.ID)
	if err != nil {
		return GetByIDOutput{}, err
	}
	if !canViewPayout(payout, input.CurrentUserID, input.CurrentRole) {
		return GetByIDOutput{}, coreerrors.ErrForbidden
	}

	return GetByIDOutput{Payout: toPayoutOutput(payout)}, nil
}

func (s *Service) ListByUserID(ctx context.Context, input ListByUserIDInput) (ListByUserIDOutput, error) {
	if input.UserID == uuid.Nil {
		return ListByUserIDOutput{}, coreerrors.ErrInvalidRequest
	}
	if input.CurrentRole == users.RoleUser && input.UserID != input.CurrentUserID {
		return ListByUserIDOutput{}, coreerrors.ErrForbidden
	}
	if !input.CurrentRole.IsValid() || input.CurrentUserID == uuid.Nil {
		return ListByUserIDOutput{}, coreerrors.ErrUnauthorized
	}

	filter, err := normalizePayoutsFilter(&input.UserID, nil, input.Limit, input.Offset)
	if err != nil {
		return ListByUserIDOutput{}, err
	}
	total, err := s.payoutsRepo.CountPayouts(ctx, filter)
	if err != nil {
		return ListByUserIDOutput{}, err
	}
	domainPayouts, err := s.payoutsRepo.ListPayoutsByUserID(ctx, input.UserID, filter.Limit, filter.Offset)
	if err != nil {
		return ListByUserIDOutput{}, err
	}

	return ListByUserIDOutput{Payouts: toPayoutOutputs(domainPayouts), Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (CreateRuleOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return CreateRuleOutput{}, err
	}
	if _, err := s.violationTypesRepo.GetViolationTypeByID(ctx, input.ViolationTypeID); err != nil {
		return CreateRuleOutput{}, err
	}

	rule, err := payouts.NewRule(input.ViolationTypeID, input.Percent)
	if err != nil {
		return CreateRuleOutput{}, err
	}
	if err := s.rulesRepo.CreateRule(ctx, *rule); err != nil {
		return CreateRuleOutput{}, err
	}

	return CreateRuleOutput{Rule: toRuleOutput(*rule)}, nil
}

func (s *Service) GetActiveRuleByViolationTypeID(ctx context.Context, input GetActiveRuleByViolationTypeIDInput) (GetActiveRuleByViolationTypeIDOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return GetActiveRuleByViolationTypeIDOutput{}, coreerrors.ErrForbidden
	}

	rule, err := s.rulesRepo.GetActiveRuleByViolationTypeID(ctx, input.ViolationTypeID)
	if err != nil {
		return GetActiveRuleByViolationTypeIDOutput{}, err
	}

	return GetActiveRuleByViolationTypeIDOutput{Rule: toRuleOutput(rule)}, nil
}

func (s *Service) ListRules(ctx context.Context, input ListRulesInput) (ListRulesOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return ListRulesOutput{}, coreerrors.ErrForbidden
	}

	filter, err := normalizeRulesFilter(input.ViolationTypeID, input.OnlyActive, input.Limit, input.Offset)
	if err != nil {
		return ListRulesOutput{}, err
	}
	total, err := s.rulesRepo.CountRules(ctx, filter)
	if err != nil {
		return ListRulesOutput{}, err
	}
	rules, err := s.rulesRepo.ListRules(ctx, filter)
	if err != nil {
		return ListRulesOutput{}, err
	}

	return ListRulesOutput{Rules: toRuleOutputs(rules), Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (s *Service) UpdateRule(ctx context.Context, input UpdateRuleInput) (UpdateRuleOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return UpdateRuleOutput{}, err
	}

	rule, err := s.rulesRepo.GetRuleByID(ctx, input.ID)
	if err != nil {
		return UpdateRuleOutput{}, err
	}
	if err := rule.UpdatePercent(input.Percent); err != nil {
		return UpdateRuleOutput{}, err
	}
	if err := s.rulesRepo.UpdateRule(ctx, rule); err != nil {
		return UpdateRuleOutput{}, err
	}

	return UpdateRuleOutput{Rule: toRuleOutput(rule)}, nil
}

func (s *Service) ActivateRule(ctx context.Context, input ActivateRuleInput) (ActivateRuleOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return ActivateRuleOutput{}, err
	}

	rule, err := s.rulesRepo.GetRuleByID(ctx, input.ID)
	if err != nil {
		return ActivateRuleOutput{}, err
	}
	rule.Activate()
	if err := s.rulesRepo.UpdateRule(ctx, rule); err != nil {
		return ActivateRuleOutput{}, err
	}

	return ActivateRuleOutput{Rule: toRuleOutput(rule)}, nil
}

func (s *Service) DeactivateRule(ctx context.Context, input DeactivateRuleInput) (DeactivateRuleOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return DeactivateRuleOutput{}, err
	}

	rule, err := s.rulesRepo.GetRuleByID(ctx, input.ID)
	if err != nil {
		return DeactivateRuleOutput{}, err
	}
	rule.Deactivate()
	if err := s.rulesRepo.UpdateRule(ctx, rule); err != nil {
		return DeactivateRuleOutput{}, err
	}

	return DeactivateRuleOutput{Rule: toRuleOutput(rule)}, nil
}

func (s *Service) CreatePayoutForApprovedReport(ctx context.Context, input CreatePayoutForApprovedReportInput) (CreatePayoutForApprovedReportOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}

	report, err := s.reportsRepo.GetReportByID(ctx, input.ReportID)
	if err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}
	if !report.IsApproved() {
		return CreatePayoutForApprovedReportOutput{}, fmt.Errorf("report must be approved: %w", coreerrors.ErrInvalidRequest)
	}

	violationType, err := s.violationTypesRepo.GetViolationTypeByID(ctx, report.ViolationTypeID)
	if err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}
	rule, err := s.rulesRepo.GetActiveRuleByViolationTypeID(ctx, report.ViolationTypeID)
	if err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}

	amount, err := calculateAmount(violationType.BaseFineAmount, rule.Percent)
	if err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}
	payout, err := payouts.NewPayout(report.ID, report.UserID, amount)
	if err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}
	if err := s.payoutsRepo.CreatePayout(ctx, *payout); err != nil {
		return CreatePayoutForApprovedReportOutput{}, err
	}

	return CreatePayoutForApprovedReportOutput{Payout: toPayoutOutput(*payout)}, nil
}

func (s *Service) MarkPaid(ctx context.Context, input MarkPaidInput) (MarkPaidOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return MarkPaidOutput{}, err
	}

	payout, err := s.payoutsRepo.GetPayoutByID(ctx, input.ID)
	if err != nil {
		return MarkPaidOutput{}, err
	}
	if err := payout.MarkPaid(); err != nil {
		return MarkPaidOutput{}, mapPayoutTransitionError(err)
	}
	if err := s.payoutsRepo.UpdatePayout(ctx, payout); err != nil {
		return MarkPaidOutput{}, err
	}

	return MarkPaidOutput{Payout: toPayoutOutput(payout)}, nil
}

func (s *Service) MarkFailed(ctx context.Context, input MarkFailedInput) (MarkFailedOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return MarkFailedOutput{}, err
	}

	payout, err := s.payoutsRepo.GetPayoutByID(ctx, input.ID)
	if err != nil {
		return MarkFailedOutput{}, err
	}
	if err := payout.MarkFailed(); err != nil {
		return MarkFailedOutput{}, mapPayoutTransitionError(err)
	}
	if err := s.payoutsRepo.UpdatePayout(ctx, payout); err != nil {
		return MarkFailedOutput{}, err
	}

	return MarkFailedOutput{Payout: toPayoutOutput(payout)}, nil
}

func normalizePayoutsFilter(userID *uuid.UUID, status *payouts.Status, limit int, offset int) (ListPayoutsFilter, error) {
	if offset < 0 {
		return ListPayoutsFilter{}, fmt.Errorf("offset must be non-negative: %w", coreerrors.ErrInvalidRequest)
	}
	if status != nil && !status.IsValid() {
		return ListPayoutsFilter{}, coreerrors.ErrInvalidRequest
	}
	limit = normalizeLimit(limit)
	return ListPayoutsFilter{UserID: userID, Status: status, Limit: limit, Offset: offset}, nil
}

func normalizeRulesFilter(violationTypeID *uuid.UUID, onlyActive *bool, limit int, offset int) (ListRulesFilter, error) {
	if offset < 0 {
		return ListRulesFilter{}, fmt.Errorf("offset must be non-negative: %w", coreerrors.ErrInvalidRequest)
	}
	limit = normalizeLimit(limit)
	return ListRulesFilter{ViolationTypeID: violationTypeID, OnlyActive: onlyActive, Limit: limit, Offset: offset}, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func canViewPayout(payout payouts.Payout, currentUserID uuid.UUID, currentRole users.Role) bool {
	return currentRole == users.RoleAdmin || currentRole == users.RoleModerator || payout.UserID == currentUserID
}

func requireAdmin(role users.Role) error {
	if role != users.RoleAdmin {
		return coreerrors.ErrForbidden
	}
	return nil
}

func calculateAmount(baseFineAmount string, percent string) (string, error) {
	base, ok := new(big.Rat).SetString(strings.TrimSpace(baseFineAmount))
	if !ok || base.Sign() <= 0 {
		return "", fmt.Errorf("base fine amount is required: %w", coreerrors.ErrInvalidRequest)
	}
	pct, ok := new(big.Rat).SetString(strings.TrimSpace(percent))
	if !ok || pct.Sign() <= 0 {
		return "", fmt.Errorf("payout rule percent is invalid: %w", coreerrors.ErrInvalidRequest)
	}
	amount := new(big.Rat).Mul(base, pct)
	amount.Quo(amount, big.NewRat(100, 1))
	return amount.FloatString(2), nil
}

func mapPayoutTransitionError(err error) error {
	if errors.Is(err, coreerrors.ErrInvalidTransition) {
		return fmt.Errorf("%w: %w: %w", err, coreerrors.ErrInvalidPayoutStatus, coreerrors.ErrInvalidRequest)
	}
	return err
}

func toPayoutOutput(payout payouts.Payout) PayoutOutput {
	return PayoutOutput{
		ID:        payout.ID,
		ReportID:  payout.ReportID,
		UserID:    payout.UserID,
		Amount:    payout.Amount,
		Status:    payout.Status,
		CreatedAt: payout.CreatedAt,
		UpdatedAt: payout.UpdatedAt,
	}
}

func toPayoutOutputs(domainPayouts []payouts.Payout) []PayoutOutput {
	outputs := make([]PayoutOutput, 0, len(domainPayouts))
	for _, payout := range domainPayouts {
		outputs = append(outputs, toPayoutOutput(payout))
	}
	return outputs
}

func toRuleOutput(rule payouts.Rule) RuleOutput {
	return RuleOutput{
		ID:              rule.ID,
		ViolationTypeID: rule.ViolationTypeID,
		Percent:         rule.Percent,
		IsActive:        rule.IsActive,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       rule.UpdatedAt,
	}
}

func toRuleOutputs(rules []payouts.Rule) []RuleOutput {
	outputs := make([]RuleOutput, 0, len(rules))
	for _, rule := range rules {
		outputs = append(outputs, toRuleOutput(rule))
	}
	return outputs
}
