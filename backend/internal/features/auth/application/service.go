package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	coreauth "pdd-service/internal/core/auth"
	"pdd-service/internal/core/domain/sessions"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
)

type AuthService struct {
	usersRepo    UsersRepository
	sessionsRepo SessionsRepository
	tokens       *coreauth.TokenManager
}

func NewAuthService(usersRepo UsersRepository, sessionsRepo SessionsRepository, tokens *coreauth.TokenManager) *AuthService {
	return &AuthService{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
		tokens:       tokens,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	email := normalizeEmail(input.Email)

	exists, err := s.usersRepo.EmailExists(ctx, email)
	if err != nil {
		return RegisterOutput{}, err
	}
	if exists {
		return RegisterOutput{}, coreerrors.ErrEmailAlreadyExists
	}

	passwordHash, err := coreauth.HashPassword(input.Password)
	if err != nil {
		return RegisterOutput{}, err
	}

	user, err := users.NewUser(email, passwordHash, input.FullName, users.RoleUser)
	if err != nil {
		return RegisterOutput{}, err
	}

	if err := s.usersRepo.CreateUser(ctx, *user); err != nil {
		return RegisterOutput{}, err
	}

	tokens, err := s.issueTokens(ctx, *user)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{User: toUserOutput(*user), Tokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	user, err := s.usersRepo.GetUserByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		if errors.Is(err, coreerrors.ErrUserNotFound) {
			return LoginOutput{}, coreerrors.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if !coreauth.VerifyPassword(user.PasswordHash, input.Password) {
		return LoginOutput{}, coreerrors.ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{User: toUserOutput(user), Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	claims, err := s.tokens.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return RefreshOutput{}, err
	}

	refreshTokenHash := coreauth.HashRefreshToken(input.RefreshToken)
	session, err := s.sessionsRepo.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return RefreshOutput{}, err
	}
	if session.UserID != claims.UserID {
		return RefreshOutput{}, coreerrors.ErrUnauthorized
	}
	if session.IsRevoked() {
		return RefreshOutput{}, coreerrors.ErrSessionRevoked
	}
	if session.IsExpired(time.Now().UTC()) {
		return RefreshOutput{}, coreerrors.ErrUnauthorized
	}

	user, err := s.usersRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return RefreshOutput{}, err
	}

	if err := s.sessionsRepo.RevokeSession(ctx, session.ID); err != nil {
		return RefreshOutput{}, err
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{User: toUserOutput(user), Tokens: tokens}, nil
}

func (s *AuthService) Logout(ctx context.Context, input LogoutInput) error {
	claims, err := s.tokens.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return err
	}

	refreshTokenHash := coreauth.HashRefreshToken(input.RefreshToken)
	session, err := s.sessionsRepo.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return err
	}
	if session.UserID != claims.UserID {
		return coreerrors.ErrUnauthorized
	}
	if session.IsRevoked() {
		return coreerrors.ErrSessionRevoked
	}

	return s.sessionsRepo.RevokeSession(ctx, session.ID)
}

func (s *AuthService) Me(ctx context.Context, input MeInput) (MeOutput, error) {
	user, err := s.usersRepo.GetUserByID(ctx, input.UserID)
	if err != nil {
		return MeOutput{}, err
	}

	return MeOutput{User: toUserOutput(user)}, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user users.User) (TokensOutput, error) {
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return TokensOutput{}, err
	}
	refreshToken, refreshExpiresAt, err := s.tokens.GenerateRefreshToken(user)
	if err != nil {
		return TokensOutput{}, err
	}

	session, err := sessions.NewSession(user.ID, coreauth.HashRefreshToken(refreshToken), refreshExpiresAt)
	if err != nil {
		return TokensOutput{}, err
	}
	if err := s.sessionsRepo.CreateSession(ctx, *session); err != nil {
		return TokensOutput{}, fmt.Errorf("create session: %w", err)
	}

	return TokensOutput{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func toUserOutput(user users.User) UserOutput {
	return UserOutput{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}
}
