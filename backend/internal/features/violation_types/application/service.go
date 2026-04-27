package application

import (
	"context"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return CreateOutput{}, err
	}

	violationType, err := violations.NewViolationType(
		input.Code,
		input.Title,
		input.Description,
		input.BaseFineAmount,
	)
	if err != nil {
		return CreateOutput{}, err
	}

	if err := s.repo.CreateViolationType(ctx, *violationType); err != nil {
		return CreateOutput{}, err
	}

	return CreateOutput{ViolationType: toOutput(*violationType)}, nil
}

func (s *Service) GetByID(ctx context.Context, input GetByIDInput) (GetByIDOutput, error) {
	if input.ID == uuid.Nil {
		return GetByIDOutput{}, coreerrors.ErrInvalidRequest
	}

	violationType, err := s.repo.GetViolationTypeByID(ctx, input.ID)
	if err != nil {
		return GetByIDOutput{}, err
	}

	return GetByIDOutput{ViolationType: toOutput(violationType)}, nil
}

func (s *Service) GetByCode(ctx context.Context, input GetByCodeInput) (GetByCodeOutput, error) {
	if strings.TrimSpace(input.Code) == "" {
		return GetByCodeOutput{}, coreerrors.ErrInvalidRequest
	}

	violationType, err := s.repo.GetViolationTypeByCode(ctx, strings.TrimSpace(input.Code))
	if err != nil {
		return GetByCodeOutput{}, err
	}

	return GetByCodeOutput{ViolationType: toOutput(violationType)}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	filter, err := normalizeListFilter(input)
	if err != nil {
		return ListOutput{}, err
	}

	total, err := s.repo.CountViolationTypes(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	violationTypes, err := s.repo.ListViolationTypes(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{
		ViolationTypes: toOutputs(violationTypes),
		Total:          total,
		Limit:          filter.Limit,
		Offset:         filter.Offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (UpdateOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return UpdateOutput{}, err
	}
	if input.ID == uuid.Nil {
		return UpdateOutput{}, coreerrors.ErrInvalidRequest
	}

	violationType, err := s.repo.GetViolationTypeByID(ctx, input.ID)
	if err != nil {
		return UpdateOutput{}, err
	}

	if input.Title != nil {
		if err := violationType.Rename(*input.Title); err != nil {
			return UpdateOutput{}, err
		}
	}
	if input.Description != nil {
		violationType.UpdateDescription(*input.Description)
	}
	if input.BaseFineAmount != nil {
		violationType.UpdateBaseFineAmount(*input.BaseFineAmount)
	}

	if err := s.repo.UpdateViolationType(ctx, violationType); err != nil {
		return UpdateOutput{}, err
	}

	return UpdateOutput{ViolationType: toOutput(violationType)}, nil
}

func (s *Service) Delete(ctx context.Context, input DeleteInput) error {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return err
	}
	if input.ID == uuid.Nil {
		return coreerrors.ErrInvalidRequest
	}

	return s.repo.DeleteViolationType(ctx, input.ID)
}

func (s *Service) Activate(ctx context.Context, input ActivateInput) (ActivateOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return ActivateOutput{}, err
	}
	if input.ID == uuid.Nil {
		return ActivateOutput{}, coreerrors.ErrInvalidRequest
	}

	violationType, err := s.repo.GetViolationTypeByID(ctx, input.ID)
	if err != nil {
		return ActivateOutput{}, err
	}

	violationType.Activate()
	if err := s.repo.UpdateViolationType(ctx, violationType); err != nil {
		return ActivateOutput{}, err
	}

	return ActivateOutput{ViolationType: toOutput(violationType)}, nil
}

func (s *Service) Deactivate(ctx context.Context, input DeactivateInput) (DeactivateOutput, error) {
	if err := requireAdmin(input.CurrentRole); err != nil {
		return DeactivateOutput{}, err
	}
	if input.ID == uuid.Nil {
		return DeactivateOutput{}, coreerrors.ErrInvalidRequest
	}

	violationType, err := s.repo.GetViolationTypeByID(ctx, input.ID)
	if err != nil {
		return DeactivateOutput{}, err
	}

	violationType.Deactivate()
	if err := s.repo.UpdateViolationType(ctx, violationType); err != nil {
		return DeactivateOutput{}, err
	}

	return DeactivateOutput{ViolationType: toOutput(violationType)}, nil
}

func normalizeListFilter(input ListInput) (ListViolationTypesFilter, error) {
	if input.Offset < 0 {
		return ListViolationTypesFilter{}, fmt.Errorf("offset must be non-negative: %w", coreerrors.ErrInvalidRequest)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return ListViolationTypesFilter{
		OnlyActive: input.OnlyActive,
		Search:     strings.TrimSpace(input.Search),
		Limit:      limit,
		Offset:     input.Offset,
	}, nil
}

func requireAdmin(role users.Role) error {
	if role != users.RoleAdmin {
		return coreerrors.ErrForbidden
	}

	return nil
}

func toOutput(violationType violations.ViolationType) ViolationTypeOutput {
	return ViolationTypeOutput{
		ID:             violationType.ID,
		Code:           violationType.Code,
		Title:          violationType.Title,
		Description:    violationType.Description,
		BaseFineAmount: violationType.BaseFineAmount,
		IsActive:       violationType.IsActive,
		CreatedAt:      violationType.CreatedAt,
		UpdatedAt:      violationType.UpdatedAt,
	}
}

func toOutputs(violationTypes []violations.ViolationType) []ViolationTypeOutput {
	outputs := make([]ViolationTypeOutput, 0, len(violationTypes))
	for _, violationType := range violationTypes {
		outputs = append(outputs, toOutput(violationType))
	}

	return outputs
}
