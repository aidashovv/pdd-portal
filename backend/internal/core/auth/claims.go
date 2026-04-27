package auth

import (
	"time"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID
	Email     string
	Role      users.Role
	TokenType TokenType
	ExpiresAt time.Time
	IssuedAt  time.Time
}
