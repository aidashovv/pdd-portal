package http

import "time"

type GetUserResponse struct {
	User UserResponse `json:"user"`
}

type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type UpdateRoleResponse struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
