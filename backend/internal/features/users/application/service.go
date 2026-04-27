package application

import (
	"context"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/users"
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

func (s *Service) GetByID(ctx context.Context, input GetByIDInput) (GetByIDOutput, error) {
	if input.ID == uuid.Nil || input.CurrentUserID == uuid.Nil {
		return GetByIDOutput{}, coreerrors.ErrInvalidRequest
	}
	if !input.CurrentRole.IsValid() {
		return GetByIDOutput{}, coreerrors.ErrForbidden
	}
	if input.CurrentRole == users.RoleUser && input.CurrentUserID != input.ID {
		return GetByIDOutput{}, coreerrors.ErrForbidden
	}

	user, err := s.repo.GetUserByID(ctx, input.ID)
	if err != nil {
		return GetByIDOutput{}, err
	}

	return GetByIDOutput{User: toUserOutput(user)}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return ListOutput{}, coreerrors.ErrForbidden
	}
	if input.Role != nil && !input.Role.IsValid() {
		return ListOutput{}, fmt.Errorf("invalid role filter: %w", coreerrors.ErrInvalidRequest)
	}

	filter, err := normalizeListFilter(input)
	if err != nil {
		return ListOutput{}, err
	}

	total, err := s.repo.CountUsers(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	domainUsers, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{
		Users:  toUserOutputs(domainUsers),
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (s *Service) UpdateRole(ctx context.Context, input UpdateRoleInput) (UpdateRoleOutput, error) {
	if input.CurrentRole != users.RoleAdmin {
		return UpdateRoleOutput{}, coreerrors.ErrForbidden
	}
	if input.ID == uuid.Nil {
		return UpdateRoleOutput{}, coreerrors.ErrInvalidRequest
	}
	if !input.Role.IsValid() {
		return UpdateRoleOutput{}, fmt.Errorf("invalid user role: %w", coreerrors.ErrInvalidRequest)
	}

	if err := s.repo.UpdateUserRole(ctx, input.ID, input.Role); err != nil {
		return UpdateRoleOutput{}, err
	}

	user, err := s.repo.GetUserByID(ctx, input.ID)
	if err != nil {
		return UpdateRoleOutput{}, err
	}

	return UpdateRoleOutput{User: toUserOutput(user)}, nil
}

func normalizeListFilter(input ListInput) (ListUsersFilter, error) {
	if input.Offset < 0 {
		return ListUsersFilter{}, fmt.Errorf("offset must be non-negative: %w", coreerrors.ErrInvalidRequest)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return ListUsersFilter{
		Role:   input.Role,
		Search: strings.TrimSpace(input.Search),
		Limit:  limit,
		Offset: input.Offset,
	}, nil
}

func toUserOutput(user users.User) UserOutput {
	return UserOutput{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toUserOutputs(domainUsers []users.User) []UserOutput {
	outputs := make([]UserOutput, 0, len(domainUsers))
	for _, user := range domainUsers {
		outputs = append(outputs, toUserOutput(user))
	}

	return outputs
}
