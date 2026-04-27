package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestServiceGetByIDPermissions(t *testing.T) {
	service, repo := newTestService()
	owner := repo.addUser(users.RoleUser)
	other := repo.addUser(users.RoleUser)
	moderator := repo.addUser(users.RoleModerator)
	admin := repo.addUser(users.RoleAdmin)

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: owner.ID, CurrentUserID: owner.ID, CurrentRole: users.RoleUser,
	}); err != nil {
		t.Fatalf("owner GetByID() error = %v", err)
	}

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: other.ID, CurrentUserID: owner.ID, CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("user GetByID(other) error = %v, want ErrForbidden", err)
	}

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: owner.ID, CurrentUserID: moderator.ID, CurrentRole: users.RoleModerator,
	}); err != nil {
		t.Fatalf("moderator GetByID(any) error = %v", err)
	}

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: owner.ID, CurrentUserID: admin.ID, CurrentRole: users.RoleAdmin,
	}); err != nil {
		t.Fatalf("admin GetByID(any) error = %v", err)
	}
}

func TestServiceUpdateRolePermissions(t *testing.T) {
	service, repo := newTestService()
	target := repo.addUser(users.RoleUser)

	if _, err := service.UpdateRole(context.Background(), UpdateRoleInput{
		ID: target.ID, Role: users.RoleModerator, CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("non-admin UpdateRole() error = %v, want ErrForbidden", err)
	}

	if _, err := service.UpdateRole(context.Background(), UpdateRoleInput{
		ID: target.ID, Role: users.Role(99), CurrentRole: users.RoleAdmin,
	}); !errors.Is(err, coreerrors.ErrInvalidRequest) {
		t.Fatalf("invalid role UpdateRole() error = %v, want ErrInvalidRequest", err)
	}

	output, err := service.UpdateRole(context.Background(), UpdateRoleInput{
		ID: target.ID, Role: users.RoleModerator, CurrentRole: users.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("admin UpdateRole() error = %v", err)
	}
	if output.User.Role != users.RoleModerator {
		t.Fatalf("updated role = %s, want %s", output.User.Role, users.RoleModerator)
	}
}

func newTestService() (*Service, *usersRepositoryStub) {
	repo := &usersRepositoryStub{users: map[uuid.UUID]users.User{}}
	return NewService(repo), repo
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
		return users.User{}, coreerrors.ErrUserNotFound
	}
	return user, nil
}

func (r *usersRepositoryStub) ListUsers(_ context.Context, filter ListUsersFilter) ([]users.User, error) {
	result := make([]users.User, 0, len(r.users))
	for _, user := range r.users {
		if filter.Role != nil && user.Role != *filter.Role {
			continue
		}
		result = append(result, user)
	}
	return result, nil
}

func (r *usersRepositoryStub) CountUsers(_ context.Context, filter ListUsersFilter) (int64, error) {
	users, err := r.ListUsers(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	return int64(len(users)), nil
}

func (r *usersRepositoryStub) UpdateUserRole(_ context.Context, id uuid.UUID, role users.Role) error {
	user, ok := r.users[id]
	if !ok {
		return coreerrors.ErrUserNotFound
	}
	user.Role = role
	user.UpdatedAt = time.Now().UTC()
	r.users[id] = user
	return nil
}
