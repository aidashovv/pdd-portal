package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestServiceCreatePermissionsAndValidation(t *testing.T) {
	service, repo := newTestService()

	if _, err := service.Create(context.Background(), CreateInput{
		Code: "SPEEDING", Title: "Speeding", CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("non-admin Create() error = %v, want ErrForbidden", err)
	}

	output, err := service.Create(context.Background(), CreateInput{
		Code:           "SPEEDING",
		Title:          "Speeding",
		Description:    "Speed limit violation",
		BaseFineAmount: "500",
		CurrentRole:    users.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("admin Create() error = %v", err)
	}
	if output.ViolationType.Code != "SPEEDING" {
		t.Fatalf("created code = %q", output.ViolationType.Code)
	}

	if _, err := service.Create(context.Background(), CreateInput{
		Code: "", Title: "No code", CurrentRole: users.RoleAdmin,
	}); !errors.Is(err, coreerrors.ErrInvalidDomainValue) {
		t.Fatalf("invalid code Create() error = %v, want ErrInvalidDomainValue", err)
	}
	if _, err := service.Create(context.Background(), CreateInput{
		Code: "NO_TITLE", Title: "", CurrentRole: users.RoleAdmin,
	}); !errors.Is(err, coreerrors.ErrInvalidDomainValue) {
		t.Fatalf("invalid title Create() error = %v, want ErrInvalidDomainValue", err)
	}

	if _, err := service.Create(context.Background(), CreateInput{
		Code: "SPEEDING", Title: "Duplicate", CurrentRole: users.RoleAdmin,
	}); !errors.Is(err, coreerrors.ErrViolationTypeAlreadyExists) {
		t.Fatalf("duplicate code Create() error = %v, want ErrViolationTypeAlreadyExists", err)
	}

	if len(repo.items) != 1 {
		t.Fatalf("repo items = %d, want 1", len(repo.items))
	}
}

func TestServiceListAndGetByID(t *testing.T) {
	service, repo := newTestService()
	first := repo.addViolationType("SPEEDING", "Speeding")
	repo.addViolationType("PARKING", "Parking")

	list, err := service.List(context.Background(), ListInput{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if list.Total != 2 || len(list.ViolationTypes) != 2 {
		t.Fatalf("List() total=%d len=%d, want 2", list.Total, len(list.ViolationTypes))
	}

	got, err := service.GetByID(context.Background(), GetByIDInput{ID: first.ID})
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ViolationType.ID != first.ID {
		t.Fatalf("GetByID() id = %s, want %s", got.ViolationType.ID, first.ID)
	}
}

func TestServiceUpdateActivateDeactivate(t *testing.T) {
	service, repo := newTestService()
	violationType := repo.addViolationType("SPEEDING", "Speeding")

	title := "Updated title"
	description := "Updated description"
	baseFineAmount := "1000"
	updated, err := service.Update(context.Background(), UpdateInput{
		ID:             violationType.ID,
		Title:          &title,
		Description:    &description,
		BaseFineAmount: &baseFineAmount,
		CurrentRole:    users.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.ViolationType.Title != title || updated.ViolationType.Description != description || updated.ViolationType.BaseFineAmount != baseFineAmount {
		t.Fatalf("Update() output = %+v", updated.ViolationType)
	}

	deactivated, err := service.Deactivate(context.Background(), DeactivateInput{
		ID: violationType.ID, CurrentRole: users.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	if deactivated.ViolationType.IsActive {
		t.Fatal("Deactivate() left type active")
	}

	activated, err := service.Activate(context.Background(), ActivateInput{
		ID: violationType.ID, CurrentRole: users.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if !activated.ViolationType.IsActive {
		t.Fatal("Activate() left type inactive")
	}

	if _, err := service.Update(context.Background(), UpdateInput{
		ID: violationType.ID, Title: ptr(""), CurrentRole: users.RoleAdmin,
	}); !errors.Is(err, coreerrors.ErrInvalidDomainValue) {
		t.Fatalf("invalid title Update() error = %v, want ErrInvalidDomainValue", err)
	}
}

func ptr(value string) *string {
	return &value
}

func newTestService() (*Service, *violationTypesRepositoryStub) {
	repo := &violationTypesRepositoryStub{
		items:  map[uuid.UUID]violations.ViolationType{},
		byCode: map[string]uuid.UUID{},
	}
	return NewService(repo), repo
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

func (r *violationTypesRepositoryStub) ListViolationTypes(_ context.Context, filter ListViolationTypesFilter) ([]violations.ViolationType, error) {
	result := make([]violations.ViolationType, 0, len(r.items))
	for _, violationType := range r.items {
		if filter.OnlyActive != nil && violationType.IsActive != *filter.OnlyActive {
			continue
		}
		if filter.Search != "" && !strings.Contains(violationType.Code, filter.Search) && !strings.Contains(violationType.Title, filter.Search) {
			continue
		}
		result = append(result, violationType)
	}
	return result, nil
}

func (r *violationTypesRepositoryStub) CountViolationTypes(ctx context.Context, filter ListViolationTypesFilter) (int64, error) {
	items, err := r.ListViolationTypes(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (r *violationTypesRepositoryStub) UpdateViolationType(_ context.Context, violationType violations.ViolationType) error {
	if _, ok := r.items[violationType.ID]; !ok {
		return coreerrors.ErrViolationTypeNotFound
	}
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
