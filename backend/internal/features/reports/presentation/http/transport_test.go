package http

import (
	"bytes"
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/features/reports/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCreateReportAuthorized(t *testing.T) {
	handler, repo := newTestHandler()
	userID := uuid.New()
	violationTypeID := repo.addViolationType()

	body := `{
		"violation_type_id":"` + violationTypeID.String() + `",
		"title":"Broken parking",
		"description":"Car parked on sidewalk",
		"location":"Main street",
		"occurred_at":"2026-01-02T15:04:05Z",
		"video":{"source":"EXTERNAL_URL","url":"https://example.com/video.mp4"}
	}`
	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reports", bytes.NewBufferString(body))
	request = request.WithContext(httphelpers.WithUser(request.Context(), userID, "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != nethttp.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusCreated)
	}
}

func TestCreateReportUnauthorized(t *testing.T) {
	handler, _ := newTestHandler()
	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reports", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusUnauthorized)
	}
}

func TestApproveForbiddenForUser(t *testing.T) {
	handler, repo := newTestHandler()
	userID := uuid.New()
	report := repo.addReport(t, userID, repo.addViolationType())
	must(t, report.Submit())
	must(t, report.StartReview(uuid.New()))
	repo.reports[report.ID] = report

	router := chi.NewRouter()
	router.Post("/api/v1/reports/{id}/approve", handler.Approve)
	request := httptest.NewRequest(
		nethttp.MethodPost,
		"/api/v1/reports/"+report.ID.String()+"/approve",
		bytes.NewBufferString(`{"comment":"ok"}`),
	)
	request = request.WithContext(httphelpers.WithUser(request.Context(), userID, "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusForbidden)
	}
}

func TestListReportsWorksForUser(t *testing.T) {
	handler, repo := newTestHandler()
	userID := uuid.New()
	otherID := uuid.New()
	violationTypeID := repo.addViolationType()
	repo.addReport(t, userID, violationTypeID)
	repo.addReport(t, otherID, violationTypeID)

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/reports", nil)
	request = request.WithContext(httphelpers.WithUser(request.Context(), userID, "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusOK)
	}
}

func newTestHandler() (*ReportsHTTPHandler, *reportsRepositoryStub) {
	repo := &reportsRepositoryStub{
		reports:        map[uuid.UUID]violations.Report{},
		violationTypes: map[uuid.UUID]violations.ViolationType{},
	}
	return NewReportsHTTPHandler(application.NewService(repo, repo)), repo
}

type reportsRepositoryStub struct {
	reports        map[uuid.UUID]violations.Report
	violationTypes map[uuid.UUID]violations.ViolationType
}

func (r *reportsRepositoryStub) addViolationType() uuid.UUID {
	violationType, _ := violations.NewViolationType(uuid.NewString(), "Parking", "", "500")
	r.violationTypes[violationType.ID] = *violationType
	return violationType.ID
}

func (r *reportsRepositoryStub) addReport(t *testing.T, userID uuid.UUID, violationTypeID uuid.UUID) violations.Report {
	t.Helper()

	video, err := violations.NewExternalVideo("https://example.com/video.mp4")
	must(t, err)
	report, err := violations.NewReport(userID, violationTypeID, "Title", "Description", "Location", time.Now().UTC(), video)
	must(t, err)
	r.reports[report.ID] = *report
	return *report
}

func (r *reportsRepositoryStub) CreateReport(_ context.Context, report violations.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *reportsRepositoryStub) GetReportByID(_ context.Context, id uuid.UUID) (violations.Report, error) {
	report, ok := r.reports[id]
	if !ok {
		return violations.Report{}, coreerrors.ErrReportNotFound
	}
	return report, nil
}

func (r *reportsRepositoryStub) UpdateReport(_ context.Context, report violations.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *reportsRepositoryStub) DeleteReport(_ context.Context, id uuid.UUID) error {
	delete(r.reports, id)
	return nil
}

func (r *reportsRepositoryStub) ListReports(_ context.Context, filter application.ListReportsFilter) ([]violations.Report, error) {
	reports := make([]violations.Report, 0, len(r.reports))
	for _, report := range r.reports {
		if filter.UserID != nil && report.UserID != *filter.UserID {
			continue
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (r *reportsRepositoryStub) CountReports(ctx context.Context, filter application.ListReportsFilter) (int64, error) {
	reports, err := r.ListReports(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(reports)), nil
}

func (r *reportsRepositoryStub) ListReportsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]violations.Report, error) {
	return r.ListReports(ctx, application.ListReportsFilter{UserID: &userID, Limit: limit, Offset: offset})
}

func (r *reportsRepositoryStub) GetViolationTypeByID(_ context.Context, id uuid.UUID) (violations.ViolationType, error) {
	violationType, ok := r.violationTypes[id]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return violationType, nil
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
