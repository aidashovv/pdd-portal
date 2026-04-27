package http

import (
	"bytes"
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/features/users/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestListUsersForbiddenForUser(t *testing.T) {
	handler, repo := newTestHandler()
	current := repo.addUser(users.RoleUser)

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/users", nil)
	request = request.WithContext(httphelpers.WithUser(request.Context(), current.ID, current.Email, current.Role))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusForbidden)
	}
}

func TestListUsersAllowedForModeratorAndAdmin(t *testing.T) {
	for _, role := range []users.Role{users.RoleModerator, users.RoleAdmin} {
		handler, repo := newTestHandler()
		current := repo.addUser(role)
		repo.addUser(users.RoleUser)

		request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/users", nil)
		request = request.WithContext(httphelpers.WithUser(request.Context(), current.ID, current.Email, current.Role))
		recorder := httptest.NewRecorder()

		handler.List(recorder, request)

		if recorder.Code != nethttp.StatusOK {
			t.Fatalf("role %s status = %d, want %d", role, recorder.Code, nethttp.StatusOK)
		}
	}
}

func TestUpdateRoleForbiddenForUser(t *testing.T) {
	handler, repo := newTestHandler()
	current := repo.addUser(users.RoleUser)
	target := repo.addUser(users.RoleUser)

	router := chi.NewRouter()
	router.Patch("/api/v1/users/{id}/role", handler.UpdateRole)

	request := httptest.NewRequest(
		nethttp.MethodPatch,
		"/api/v1/users/"+target.ID.String()+"/role",
		bytes.NewBufferString(`{"role":"MODERATOR"}`),
	)
	request = request.WithContext(httphelpers.WithUser(request.Context(), current.ID, current.Email, current.Role))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, nethttp.StatusForbidden)
	}
}

func newTestHandler() (*UsersHTTPHandler, *usersRepositoryStub) {
	repo := &usersRepositoryStub{users: map[uuid.UUID]users.User{}}
	return NewUsersHTTPHandler(application.NewService(repo)), repo
}

type usersRepositoryStub struct {
	users map[uuid.UUID]users.User
}

func (r *usersRepositoryStub) addUser(role users.Role) users.User {
	now := time.Now().UTC()
	user := users.User{
		ID:           uuid.New(),
		Email:        uuid.NewString() + "@example.com",
		PasswordHash: "hash",
		FullName:     "Test User",
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	r.users[user.ID] = user
	return user
}

func (r *usersRepositoryStub) GetUserByID(_ context.Context, id uuid.UUID) (users.User, error) {
	user, ok := r.users[id]
	if !ok {
		return users.User{}, nil
	}
	return user, nil
}

func (r *usersRepositoryStub) ListUsers(_ context.Context, filter application.ListUsersFilter) ([]users.User, error) {
	result := make([]users.User, 0, len(r.users))
	for _, user := range r.users {
		if filter.Role != nil && user.Role != *filter.Role {
			continue
		}
		result = append(result, user)
	}
	return result, nil
}

func (r *usersRepositoryStub) CountUsers(_ context.Context, filter application.ListUsersFilter) (int64, error) {
	users, err := r.ListUsers(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	return int64(len(users)), nil
}

func (r *usersRepositoryStub) UpdateUserRole(_ context.Context, id uuid.UUID, role users.Role) error {
	user := r.users[id]
	user.Role = role
	user.UpdatedAt = time.Now().UTC()
	r.users[id] = user
	return nil
}
