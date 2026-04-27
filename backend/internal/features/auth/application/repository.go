package application

import (
	"context"

	"pdd-service/internal/core/domain/sessions"
	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, user users.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (users.User, error)
	GetUserByEmail(ctx context.Context, email string) (users.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

type SessionsRepository interface {
	CreateSession(ctx context.Context, session sessions.Session) error
	GetSessionByRefreshTokenHash(ctx context.Context, hash string) (sessions.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}
