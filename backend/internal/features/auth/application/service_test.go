package application

import (
	"context"
	"errors"
	"testing"
	"time"

	coreauth "pdd-service/internal/core/auth"
	"pdd-service/internal/core/domain/sessions"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestAuthServiceRegister(t *testing.T) {
	service, repo := newTestAuthService(t)

	output, err := service.Register(context.Background(), RegisterInput{
		Email:    "USER@Example.COM",
		Password: "secret-password",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if output.User.Email != "user@example.com" {
		t.Fatalf("Register() email = %q", output.User.Email)
	}
	if output.Tokens.AccessToken == "" || output.Tokens.RefreshToken == "" {
		t.Fatal("Register() returned empty tokens")
	}
	if len(repo.sessionsByHash) != 1 {
		t.Fatalf("sessions stored = %d, want 1", len(repo.sessionsByHash))
	}
	for _, session := range repo.sessionsByHash {
		if session.RefreshTokenHash == output.Tokens.RefreshToken {
			t.Fatal("refresh token plaintext was stored")
		}
	}
}

func TestAuthServiceLogin(t *testing.T) {
	service, _ := newTestAuthService(t)
	if _, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "secret-password",
		FullName: "Test User",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	output, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "secret-password",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if output.Tokens.AccessToken == "" || output.Tokens.RefreshToken == "" {
		t.Fatal("Login() returned empty tokens")
	}

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, coreerrors.ErrInvalidCredentials) {
		t.Fatalf("Login() wrong password error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthServiceRefresh(t *testing.T) {
	service, repo := newTestAuthService(t)
	registered, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "secret-password",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	oldHash := coreauth.HashRefreshToken(registered.Tokens.RefreshToken)

	refreshed, err := service.Refresh(context.Background(), RefreshInput{
		RefreshToken: registered.Tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.Tokens.RefreshToken == "" || refreshed.Tokens.RefreshToken == registered.Tokens.RefreshToken {
		t.Fatal("Refresh() did not rotate refresh token")
	}
	if !repo.sessionsByHash[oldHash].IsRevoked() {
		t.Fatal("Refresh() did not revoke old session")
	}
}

func TestAuthServiceLogout(t *testing.T) {
	service, repo := newTestAuthService(t)
	registered, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "secret-password",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := service.Logout(context.Background(), LogoutInput{RefreshToken: registered.Tokens.RefreshToken}); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	hash := coreauth.HashRefreshToken(registered.Tokens.RefreshToken)
	if !repo.sessionsByHash[hash].IsRevoked() {
		t.Fatal("Logout() did not revoke session")
	}
}

func newTestAuthService(t *testing.T) (*AuthService, *authRepositoryStub) {
	t.Helper()

	manager, err := coreauth.NewTokenManager(coreauth.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     time.Hour,
		RefreshTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	repo := &authRepositoryStub{
		usersByID:      map[uuid.UUID]users.User{},
		usersByEmail:   map[string]users.User{},
		sessionsByID:   map[uuid.UUID]sessions.Session{},
		sessionsByHash: map[string]sessions.Session{},
	}

	return NewAuthService(repo, repo, manager), repo
}

type authRepositoryStub struct {
	usersByID      map[uuid.UUID]users.User
	usersByEmail   map[string]users.User
	sessionsByID   map[uuid.UUID]sessions.Session
	sessionsByHash map[string]sessions.Session
}

func (r *authRepositoryStub) CreateUser(_ context.Context, user users.User) error {
	if _, ok := r.usersByEmail[user.Email]; ok {
		return coreerrors.ErrEmailAlreadyExists
	}
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *authRepositoryStub) GetUserByID(_ context.Context, id uuid.UUID) (users.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return users.User{}, coreerrors.ErrUserNotFound
	}
	return user, nil
}

func (r *authRepositoryStub) GetUserByEmail(_ context.Context, email string) (users.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return users.User{}, coreerrors.ErrUserNotFound
	}
	return user, nil
}

func (r *authRepositoryStub) EmailExists(_ context.Context, email string) (bool, error) {
	_, ok := r.usersByEmail[email]
	return ok, nil
}

func (r *authRepositoryStub) CreateSession(_ context.Context, session sessions.Session) error {
	r.sessionsByID[session.ID] = session
	r.sessionsByHash[session.RefreshTokenHash] = session
	return nil
}

func (r *authRepositoryStub) GetSessionByRefreshTokenHash(_ context.Context, hash string) (sessions.Session, error) {
	session, ok := r.sessionsByHash[hash]
	if !ok {
		return sessions.Session{}, coreerrors.ErrSessionNotFound
	}
	return session, nil
}

func (r *authRepositoryStub) RevokeSession(_ context.Context, sessionID uuid.UUID) error {
	session, ok := r.sessionsByID[sessionID]
	if !ok {
		return coreerrors.ErrSessionNotFound
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	r.sessionsByID[sessionID] = session
	r.sessionsByHash[session.RefreshTokenHash] = session
	return nil
}

func (r *authRepositoryStub) RevokeAllUserSessions(_ context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	for id, session := range r.sessionsByID {
		if session.UserID != userID {
			continue
		}
		session.RevokedAt = &now
		r.sessionsByID[id] = session
		r.sessionsByHash[session.RefreshTokenHash] = session
	}
	return nil
}
