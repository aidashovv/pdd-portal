package auth

import (
	"errors"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestTokenManagerGenerateAndParse(t *testing.T) {
	manager := newTestTokenManager(t)
	user := users.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Role:  users.RoleUser,
	}

	accessToken, _, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	accessClaims, err := manager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if accessClaims.UserID != user.ID || accessClaims.Email != user.Email || accessClaims.Role != user.Role {
		t.Fatalf("ParseAccessToken() claims = %+v, want user %+v", accessClaims, user)
	}

	refreshToken, _, err := manager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	refreshClaims, err := manager.ParseRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Fatalf("ParseRefreshToken() type = %s, want %s", refreshClaims.TokenType, TokenTypeRefresh)
	}
	if _, err := manager.ParseAccessToken(refreshToken); !errors.Is(err, coreerrors.ErrUnauthorized) {
		t.Fatalf("ParseAccessToken(refresh) error = %v, want ErrUnauthorized", err)
	}
}

func TestHashRefreshToken(t *testing.T) {
	raw := "refresh-token"
	hash := HashRefreshToken(raw)
	if hash == "" || hash == raw {
		t.Fatalf("HashRefreshToken() = %q", hash)
	}
	if HashRefreshToken(raw) != hash {
		t.Fatal("HashRefreshToken() is not deterministic")
	}
}

func newTestTokenManager(t *testing.T) *TokenManager {
	t.Helper()

	manager, err := NewTokenManager(Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     time.Hour,
		RefreshTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	return manager
}
