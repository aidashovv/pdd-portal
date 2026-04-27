package postgres

import (
	"fmt"

	"pdd-service/internal/core/domain/sessions"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
)

func userToDTO(user users.User) UserDTO {
	return UserDTO{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		Role:         int16(user.Role),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

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

func sessionToDTO(session sessions.Session) SessionDTO {
	return SessionDTO{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
		RevokedAt:        session.RevokedAt,
		CreatedAt:        session.CreatedAt,
	}
}

func (dto SessionDTO) toDomain() sessions.Session {
	return sessions.Session{
		ID:               dto.ID,
		UserID:           dto.UserID,
		RefreshTokenHash: dto.RefreshTokenHash,
		ExpiresAt:        dto.ExpiresAt,
		RevokedAt:        dto.RevokedAt,
		CreatedAt:        dto.CreatedAt,
	}
}
