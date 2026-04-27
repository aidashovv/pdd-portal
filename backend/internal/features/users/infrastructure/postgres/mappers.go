package postgres

import (
	"fmt"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
)

func (dto UserDTO) toDomain() (users.User, error) {
	role := users.Role(dto.Role)
	if !role.IsValid() {
		return users.User{}, fmt.Errorf("invalid user role from db: %w", coreerrors.ErrInvalidDomainValue)
	}

	return users.User{
		ID:           dto.ID,
		Email:        dto.Email,
		PasswordHash: dto.PasswordHash,
		FullName:     dto.FullName,
		Role:         role,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}, nil
}
