package application

import (
	"time"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type RegisterOutput struct {
	User   UserOutput
	Tokens TokensOutput
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User   UserOutput
	Tokens TokensOutput
}

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	User   UserOutput
	Tokens TokensOutput
}

type LogoutInput struct {
	RefreshToken string
}

type MeInput struct {
	UserID uuid.UUID
}

type MeOutput struct {
	User UserOutput
}

type UserOutput struct {
	ID       uuid.UUID
	Email    string
	FullName string
	Role     users.Role
}

type TokensOutput struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
