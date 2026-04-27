package http

import (
	"bytes"
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/features/payouts/application"

	"github.com/google/uuid"
)

func TestListPayoutsForbiddenForUser(t *testing.T) {
	handler, _ := newTestHandler()
	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/payouts", nil)
	request = request.WithContext(httphelpers.WithUser(request.Context(), uuid.New(), "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusForbidden)
	}
}

func TestRuleMutationForbiddenForUserAndAllowedForAdmin(t *testing.T) {
	handler, repo := newTestHandler()
	violationTypeID := repo.addViolationType("1000")
	body := bytes.NewBufferString(`{"violation_type_id":"` + violationTypeID.String() + `","percent":"10"}`)

	userRequest := httptest.NewRequest(nethttp.MethodPost, "/api/v1/payout-rules", bytes.NewBuffer(body.Bytes()))
	userRequest = userRequest.WithContext(httphelpers.WithUser(userRequest.Context(), uuid.New(), "user@example.com", users.RoleUser))
	userRecorder := httptest.NewRecorder()
	handler.CreateRule(userRecorder, userRequest)
	if userRecorder.Code != nethttp.StatusForbidden {
		t.Fatalf("user status = %d, want %d", userRecorder.Code, nethttp.StatusForbidden)
	}

	adminRequest := httptest.NewRequest(nethttp.MethodPost, "/api/v1/payout-rules", bytes.NewBuffer(body.Bytes()))
	adminRequest = adminRequest.WithContext(httphelpers.WithUser(adminRequest.Context(), uuid.New(), "admin@example.com", users.RoleAdmin))
	adminRecorder := httptest.NewRecorder()
	handler.CreateRule(adminRecorder, adminRequest)
	if adminRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("admin status = %d, want %d", adminRecorder.Code, nethttp.StatusCreated)
	}
}

func newTestHandler() (*PayoutsHTTPHandler, *repoStub) {
	repo := &repoStub{payouts: map[uuid.UUID]payouts.Payout{}, rules: map[uuid.UUID]payouts.Rule{}, violationTypes: map[uuid.UUID]violations.ViolationType{}, reports: map[uuid.UUID]violations.Report{}}
	return NewPayoutsHTTPHandler(application.NewService(repo, repo, repo, repo)), repo
}

type repoStub struct {
	payouts        map[uuid.UUID]payouts.Payout
	rules          map[uuid.UUID]payouts.Rule
	reports        map[uuid.UUID]violations.Report
	violationTypes map[uuid.UUID]violations.ViolationType
}

func (r *repoStub) addViolationType(baseFine string) uuid.UUID {
	vt, _ := violations.NewViolationType(uuid.NewString(), "Violation", "", baseFine)
	r.violationTypes[vt.ID] = *vt
	return vt.ID
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
func (r *repoStub) ListPayouts(_ context.Context, filter application.ListPayoutsFilter) ([]payouts.Payout, error) {
	return []payouts.Payout{}, nil
}
func (r *repoStub) CountPayouts(_ context.Context, filter application.ListPayoutsFilter) (int64, error) {
	return 0, nil
}
func (r *repoStub) ListPayoutsByUserID(_ context.Context, userID uuid.UUID, limit, offset int) ([]payouts.Payout, error) {
	return []payouts.Payout{}, nil
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
	return payouts.Rule{}, coreerrors.ErrPayoutRuleNotFound
}
func (r *repoStub) ListRules(_ context.Context, filter application.ListRulesFilter) ([]payouts.Rule, error) {
	return []payouts.Rule{}, nil
}
func (r *repoStub) CountRules(_ context.Context, filter application.ListRulesFilter) (int64, error) {
	return 0, nil
}
func (r *repoStub) UpdateRule(_ context.Context, rule payouts.Rule) error {
	r.rules[rule.ID] = rule
	return nil
}
func (r *repoStub) GetReportByID(_ context.Context, id uuid.UUID) (violations.Report, error) {
	return violations.Report{}, coreerrors.ErrReportNotFound
}
func (r *repoStub) GetViolationTypeByID(_ context.Context, id uuid.UUID) (violations.ViolationType, error) {
	vt, ok := r.violationTypes[id]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return vt, nil
}

var _ = time.Now
