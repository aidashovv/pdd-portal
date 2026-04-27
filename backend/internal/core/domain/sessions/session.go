package sessions

import (
	"fmt"
	"time"

	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID

	RefreshTokenHash string

	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewSession(userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) (*Session, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("session user id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if refreshTokenHash == "" {
		return nil, fmt.Errorf("session refresh token hash is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if expiresAt.IsZero() {
		return nil, fmt.Errorf("session expiration time is required: %w", coreerrors.ErrInvalidDomainValue)
	}

	return &Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt.UTC(),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func (s Session) IsExpired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

func (s Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) Revoke(now time.Time) error {
	if s.IsRevoked() {
		return fmt.Errorf("session already revoked: %w", coreerrors.ErrInvalidTransition)
	}

	revokedAt := now.UTC()
	s.RevokedAt = &revokedAt
	return nil
}

func (s Session) IsActive(now time.Time) bool {
	return !s.IsRevoked() && !s.IsExpired(now)
}
