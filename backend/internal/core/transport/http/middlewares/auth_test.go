package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coreauth "pdd-service/internal/core/auth"
	"pdd-service/internal/core/domain/users"
	httphelpers "pdd-service/internal/core/transport/http/helpers"

	"github.com/google/uuid"
)

func TestAuthMiddlewareUnauthorizedRequest(t *testing.T) {
	manager := newTestTokenManager(t)
	handler := AuthMiddleware(manager)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareSetsUserContext(t *testing.T) {
	manager := newTestTokenManager(t)
	user := users.User{ID: uuid.New(), Email: "user@example.com", Role: users.RoleUser}
	token, _, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	handler := AuthMiddleware(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := httphelpers.GetUserID(r.Context())
		if !ok || userID != user.ID {
			t.Fatalf("context user id = %s, ok = %v", userID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestRequireRolesForbidden(t *testing.T) {
	handler := RequireRoles(users.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request = request.WithContext(httphelpers.WithUser(request.Context(), uuid.New(), "user@example.com", users.RoleUser))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func newTestTokenManager(t *testing.T) *coreauth.TokenManager {
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

	return manager
}
