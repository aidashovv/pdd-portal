package http

import (
	"fmt"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/features/users/application"

	"github.com/google/uuid"
)

func toGetByIDInput(id, currentUserID uuid.UUID, currentRole users.Role) application.GetByIDInput {
	return application.GetByIDInput{
		ID:            id,
		CurrentUserID: currentUserID,
		CurrentRole:   currentRole,
	}
}

func toListInput(currentRole users.Role, role *users.Role, search string, limit, offset int) application.ListInput {
	return application.ListInput{
		CurrentRole: currentRole,
		Role:        role,
		Search:      search,
		Limit:       limit,
		Offset:      offset,
	}
}

func toUpdateRoleInput(id uuid.UUID, req UpdateRoleRequest, currentRole users.Role) (application.UpdateRoleInput, error) {
	role, err := users.ParseRole(req.Role)
	if err != nil {
		return application.UpdateRoleInput{}, fmt.Errorf("parse role: %w", coreerrors.ErrInvalidRequest)
	}

	return application.UpdateRoleInput{
		ID:          id,
		Role:        role,
		CurrentRole: currentRole,
	}, nil
}

func toGetUserResponse(output application.GetByIDOutput) GetUserResponse {
	return GetUserResponse{User: toUserResponse(output.User)}
}

func toListUsersResponse(output application.ListOutput) ListUsersResponse {
	users := make([]UserResponse, 0, len(output.Users))
	for _, user := range output.Users {
		users = append(users, toUserResponse(user))
	}

	return ListUsersResponse{
		Users:  users,
		Total:  output.Total,
		Limit:  output.Limit,
		Offset: output.Offset,
	}
}

func toUpdateRoleResponse(output application.UpdateRoleOutput) UpdateRoleResponse {
	return UpdateRoleResponse{User: toUserResponse(output.User)}
}

func toUserResponse(output application.UserOutput) UserResponse {
	return UserResponse{
		ID:        output.ID.String(),
		Email:     output.Email,
		FullName:  output.FullName,
		Role:      output.Role.String(),
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}
}
