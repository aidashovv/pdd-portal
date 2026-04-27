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
	"pdd-service/internal/features/violation_types/application"

	"github.com/google/uuid"
)

func TestCreateForbiddenForUser(t *testing.T) {
	handler, _ := newTestHandler()
	currentUserID := uuid.New()

	request := httptest.NewRequest(
		nethttp.MethodPost,
		"/api/v1/violation-types",
		bytes.NewBufferString(`{"code":"SPEEDING","title":"Speeding"}`),
	)
	request = request.WithContext(httphelpers.WithUser(request.Context(), currentUserID, "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusForbidden)
	}
}

func TestCreateAllowedForAdmin(t *testing.T) {
	handler, repo := newTestHandler()
	currentUserID := uuid.New()

	request := httptest.NewRequest(
		nethttp.MethodPost,
		"/api/v1/violation-types",
		bytes.NewBufferString(`{"code":"SPEEDING","title":"Speeding","description":"desc","base_fine_amount":"500"}`),
	)
	request = request.WithContext(httphelpers.WithUser(request.Context(), currentUserID, "admin@example.com", users.RoleAdmin))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != nethttp.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusCreated)
	}
	if len(repo.items) != 1 {
		t.Fatalf("repo items = %d, want 1", len(repo.items))
	}
}

func TestListAllowedForAuthorizedUser(t *testing.T) {
	handler, repo := newTestHandler()
	repo.addViolationType("SPEEDING", "Speeding")
	currentUserID := uuid.New()

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/violation-types", nil)
	request = request.WithContext(httphelpers.WithUser(request.Context(), currentUserID, "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusOK)
	}
}

func newTestHandler() (*ViolationTypesHTTPHandler, *violationTypesRepositoryStub) {
	repo := &violationTypesRepositoryStub{
		items:  map[uuid.UUID]violations.ViolationType{},
		byCode: map[string]uuid.UUID{},
	}
	return NewViolationTypesHTTPHandler(application.NewService(repo)), repo
}

type violationTypesRepositoryStub struct {
	items  map[uuid.UUID]violations.ViolationType
	byCode map[string]uuid.UUID
}

func (r *violationTypesRepositoryStub) addViolationType(code, title string) violations.ViolationType {
	now := time.Now().UTC()
	violationType := violations.ViolationType{
		ID:             uuid.New(),
		Code:           code,
		Title:          title,
		Description:    "description",
		BaseFineAmount: "500",
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.items[violationType.ID] = violationType
	r.byCode[violationType.Code] = violationType.ID
	return violationType
}

func (r *violationTypesRepositoryStub) CreateViolationType(_ context.Context, violationType violations.ViolationType) error {
	if _, ok := r.byCode[violationType.Code]; ok {
		return coreerrors.ErrViolationTypeAlreadyExists
	}
	r.items[violationType.ID] = violationType
	r.byCode[violationType.Code] = violationType.ID
	return nil
}

func (r *violationTypesRepositoryStub) GetViolationTypeByID(_ context.Context, id uuid.UUID) (violations.ViolationType, error) {
	violationType, ok := r.items[id]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return violationType, nil
}

func (r *violationTypesRepositoryStub) GetViolationTypeByCode(_ context.Context, code string) (violations.ViolationType, error) {
	id, ok := r.byCode[code]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return r.items[id], nil
}

func (r *violationTypesRepositoryStub) ListViolationTypes(_ context.Context, filter application.ListViolationTypesFilter) ([]violations.ViolationType, error) {
	result := make([]violations.ViolationType, 0, len(r.items))
	for _, violationType := range r.items {
		if filter.OnlyActive != nil && violationType.IsActive != *filter.OnlyActive {
			continue
		}
		result = append(result, violationType)
	}
	return result, nil
}

func (r *violationTypesRepositoryStub) CountViolationTypes(ctx context.Context, filter application.ListViolationTypesFilter) (int64, error) {
	items, err := r.ListViolationTypes(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (r *violationTypesRepositoryStub) UpdateViolationType(_ context.Context, violationType violations.ViolationType) error {
	r.items[violationType.ID] = violationType
	r.byCode[violationType.Code] = violationType.ID
	return nil
}

func (r *violationTypesRepositoryStub) DeleteViolationType(_ context.Context, id uuid.UUID) error {
	if _, ok := r.items[id]; !ok {
		return coreerrors.ErrViolationTypeNotFound
	}
	delete(r.byCode, r.items[id].Code)
	delete(r.items, id)
	return nil
}
