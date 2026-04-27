package postgres

import (
	"time"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	FullName     string    `db:"full_name"`
	Role         int16     `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type SessionDTO struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	ExpiresAt        time.Time  `db:"expires_at"`
	RevokedAt        *time.Time `db:"revoked_at"`
	CreatedAt        time.Time  `db:"created_at"`
}
