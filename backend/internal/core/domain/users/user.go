package users

import (
	"fmt"
	"strings"
	"time"

	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID

	Email        string
	PasswordHash string
	FullName     string
	Role         Role

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, passwordHash, fullName string, role Role) (*User, error) {
	email = strings.TrimSpace(email)
	passwordHash = strings.TrimSpace(passwordHash)
	fullName = strings.TrimSpace(fullName)

	if email == "" {
		return nil, fmt.Errorf("email is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("password hash is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if fullName == "" {
		return nil, fmt.Errorf("full name is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if !role.IsValid() {
		return nil, fmt.Errorf("invalid role: %w", coreerrors.ErrInvalidDomainValue)
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u User) IsAdmin() bool {
	return u.Role.IsAdmin()
}

func (u User) IsModerator() bool {
	return u.Role.IsModerator()
}

func (u User) CanModerateReports() bool {
	return u.Role.CanModerateReports()
}

func (u User) CanManageViolationTypes() bool {
	return u.Role.CanManageViolationTypes()
}

func (u User) CanCreateReport() bool {
	return u.Role.CanCreateReport()
}
