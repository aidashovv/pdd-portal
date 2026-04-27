package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestPayoutAccessRules(t *testing.T) {
	service, repo := newTestService()
	ownerID := uuid.New()
	otherID := uuid.New()
	own := repo.addPayout(ownerID, payouts.StatusPending)
	other := repo.addPayout(otherID, payouts.StatusPending)

	list, err := service.ListByUserID(context.Background(), ListByUserIDInput{UserID: ownerID, CurrentUserID: ownerID, CurrentRole: users.RoleUser})
	if err != nil || list.Total != 1 || list.Payouts[0].ID != own.ID {
		t.Fatalf("ListByUserID own = %+v, err=%v", list, err)
	}
	if _, err := service.ListByUserID(context.Background(), ListByUserIDInput{UserID: otherID, CurrentUserID: ownerID, CurrentRole: users.RoleUser}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("ListByUserID other error = %v, want ErrForbidden", err)
	}
	if _, err := service.List(context.Background(), ListInput{CurrentUserID: ownerID, CurrentRole: users.RoleModerator}); err != nil {
		t.Fatalf("moderator List() error = %v", err)
	}
	if _, err := service.List(context.Background(), ListInput{CurrentUserID: ownerID, CurrentRole: users.RoleAdmin}); err != nil {
		t.Fatalf("admin List() error = %v", err)
	}
	if _, err := service.GetByID(context.Background(), GetByIDInput{ID: own.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser}); err != nil {
		t.Fatalf("GetByID own error = %v", err)
	}
	if _, err := service.GetByID(context.Background(), GetByIDInput{ID: other.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("GetByID other error = %v, want ErrForbidden", err)
	}
}

func TestPayoutRules(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType("1000")

	if _, err := service.CreateRule(context.Background(), CreateRuleInput{ViolationTypeID: violationTypeID, Percent: "10", CurrentRole: users.RoleUser}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("non-admin CreateRule error = %v, want ErrForbidden", err)
	}
	if _, err := service.CreateRule(context.Background(), CreateRuleInput{ViolationTypeID: violationTypeID, Percent: "101", CurrentRole: users.RoleAdmin}); !errors.Is(err, coreerrors.ErrInvalidDomainValue) {
		t.Fatalf("invalid percent error = %v, want ErrInvalidDomainValue", err)
	}
	created, err := service.CreateRule(context.Background(), CreateRuleInput{ViolationTypeID: violationTypeID, Percent: "10", CurrentRole: users.RoleAdmin})
	if err != nil {
		t.Fatalf("CreateRule error = %v", err)
	}
	deactivated, err := service.DeactivateRule(context.Background(), DeactivateRuleInput{ID: created.Rule.ID, CurrentRole: users.RoleAdmin})
	if err != nil || deactivated.Rule.IsActive {
		t.Fatalf("DeactivateRule = %+v, err=%v", deactivated, err)
	}
	activated, err := service.ActivateRule(context.Background(), ActivateRuleInput{ID: created.Rule.ID, CurrentRole: users.RoleAdmin})
	if err != nil || !activated.Rule.IsActive {
		t.Fatalf("ActivateRule = %+v, err=%v", activated, err)
	}
}

func TestCreatePayoutFromReportAndTransitions(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType("1000")
	rule := repo.addRule(violationTypeID, "10")
	report := repo.addReport(t, uuid.New(), violationTypeID, violations.StatusApproved)

	created, err := service.CreatePayoutForApprovedReport(context.Background(), CreatePayoutForApprovedReportInput{ReportID: report.ID, CurrentRole: users.RoleAdmin})
	if err != nil {
		t.Fatalf("CreatePayoutForApprovedReport error = %v", err)
	}
	if created.Payout.Amount != "100.00" || created.Payout.UserID != report.UserID || rule.ID == uuid.Nil {
		t.Fatalf("created payout = %+v", created.Payout)
	}

	nonApproved := repo.addReport(t, uuid.New(), violationTypeID, violations.StatusSubmitted)
	if _, err := service.CreatePayoutForApprovedReport(context.Background(), CreatePayoutForApprovedReportInput{ReportID: nonApproved.ID, CurrentRole: users.RoleAdmin}); !errors.Is(err, coreerrors.ErrInvalidRequest) {
		t.Fatalf("non-approved report error = %v, want ErrInvalidRequest", err)
	}

	paid, err := service.MarkPaid(context.Background(), MarkPaidInput{ID: created.Payout.ID, CurrentRole: users.RoleAdmin})
	if err != nil || paid.Payout.Status != payouts.StatusPaid {
		t.Fatalf("MarkPaid = %+v, err=%v", paid, err)
	}
	if _, err := service.MarkFailed(context.Background(), MarkFailedInput{ID: created.Payout.ID, CurrentRole: users.RoleAdmin}); !errors.Is(err, coreerrors.ErrInvalidPayoutStatus) {
		t.Fatalf("MarkFailed paid error = %v, want ErrInvalidPayoutStatus", err)
	}

	failedPayout := repo.addPayout(uuid.New(), payouts.StatusPending)
	failed, err := service.MarkFailed(context.Background(), MarkFailedInput{ID: failedPayout.ID, CurrentRole: users.RoleAdmin})
	if err != nil || failed.Payout.Status != payouts.StatusFailed {
		t.Fatalf("MarkFailed = %+v, err=%v", failed, err)
	}
}

func newTestService() (*Service, *repoStub) {
	repo := &repoStub{
		payouts:        map[uuid.UUID]payouts.Payout{},
		rules:          map[uuid.UUID]payouts.Rule{},
		reports:        map[uuid.UUID]violations.Report{},
		violationTypes: map[uuid.UUID]violations.ViolationType{},
	}
	return NewService(repo, repo, repo, repo), repo
}

type repoStub struct {
	payouts        map[uuid.UUID]payouts.Payout
	rules          map[uuid.UUID]payouts.Rule
	reports        map[uuid.UUID]violations.Report
	violationTypes map[uuid.UUID]violations.ViolationType
}

func (r *repoStub) addPayout(userID uuid.UUID, status payouts.Status) payouts.Payout {
	now := time.Now().UTC()
	payout := payouts.Payout{ID: uuid.New(), ReportID: uuid.New(), UserID: userID, Amount: "100", Status: status, CreatedAt: now, UpdatedAt: now}
	r.payouts[payout.ID] = payout
	return payout
}

func (r *repoStub) addViolationType(baseFine string) uuid.UUID {
	vt, _ := violations.NewViolationType(uuid.NewString(), "Violation", "", baseFine)
	r.violationTypes[vt.ID] = *vt
	return vt.ID
}

func (r *repoStub) addRule(violationTypeID uuid.UUID, percent string) payouts.Rule {
	rule, _ := payouts.NewRule(violationTypeID, percent)
	r.rules[rule.ID] = *rule
	return *rule
}

func (r *repoStub) addReport(t *testing.T, userID uuid.UUID, violationTypeID uuid.UUID, status violations.Status) violations.Report {
	t.Helper()
	video, err := violations.NewExternalVideo("https://example.com/video.mp4")
	if err != nil {
		t.Fatal(err)
	}
	report, err := violations.NewReport(userID, violationTypeID, "Title", "Description", "Location", time.Now().UTC(), video)
	if err != nil {
		t.Fatal(err)
	}
	report.Status = status
	r.reports[report.ID] = *report
	return *report
}

func (r *repoStub) CreatePayout(_ context.Context, payout payouts.Payout) error {
	r.payouts[payout.ID] = payout
	return nil
}
func (r *repoStub) GetPayoutByID(_ context.Context, id uuid.UUID) (payouts.Payout, error) {
	p, ok := r.payouts[id]
	if !ok {
		return payouts.Payout{}, coreerrors.ErrPayoutNotFound
	}
	return p, nil
}
func (r *repoStub) ListPayouts(_ context.Context, filter ListPayoutsFilter) ([]payouts.Payout, error) {
	out := []payouts.Payout{}
	for _, p := range r.payouts {
		if filter.UserID != nil && p.UserID != *filter.UserID {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}
func (r *repoStub) CountPayouts(ctx context.Context, filter ListPayoutsFilter) (int64, error) {
	items, _ := r.ListPayouts(ctx, filter)
	return int64(len(items)), nil
}
func (r *repoStub) ListPayoutsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]payouts.Payout, error) {
	return r.ListPayouts(ctx, ListPayoutsFilter{UserID: &userID, Limit: limit, Offset: offset})
}
func (r *repoStub) UpdatePayout(_ context.Context, payout payouts.Payout) error {
	r.payouts[payout.ID] = payout
	return nil
}
func (r *repoStub) CreateRule(_ context.Context, rule payouts.Rule) error {
	r.rules[rule.ID] = rule
	return nil
}
func (r *repoStub) GetRuleByID(_ context.Context, id uuid.UUID) (payouts.Rule, error) {
	rule, ok := r.rules[id]
	if !ok {
		return payouts.Rule{}, coreerrors.ErrPayoutRuleNotFound
	}
	return rule, nil
}
func (r *repoStub) GetActiveRuleByViolationTypeID(_ context.Context, violationTypeID uuid.UUID) (payouts.Rule, error) {
	for _, rule := range r.rules {
		if rule.ViolationTypeID == violationTypeID && rule.IsActive {
			return rule, nil
		}
	}
	return payouts.Rule{}, coreerrors.ErrPayoutRuleNotFound
}
func (r *repoStub) ListRules(_ context.Context, filter ListRulesFilter) ([]payouts.Rule, error) {
	out := []payouts.Rule{}
	for _, rule := range r.rules {
		out = append(out, rule)
	}
	return out, nil
}
func (r *repoStub) CountRules(ctx context.Context, filter ListRulesFilter) (int64, error) {
	items, _ := r.ListRules(ctx, filter)
	return int64(len(items)), nil
}
func (r *repoStub) UpdateRule(_ context.Context, rule payouts.Rule) error {
	r.rules[rule.ID] = rule
	return nil
}
func (r *repoStub) GetReportByID(_ context.Context, id uuid.UUID) (violations.Report, error) {
	report, ok := r.reports[id]
	if !ok {
		return violations.Report{}, coreerrors.ErrReportNotFound
	}
	return report, nil
}
func (r *repoStub) GetViolationTypeByID(_ context.Context, id uuid.UUID) (violations.ViolationType, error) {
	vt, ok := r.violationTypes[id]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return vt, nil
}
