package http

import "time"

type CreateViolationTypeRequest struct {
	Code           string `json:"code" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Description    string `json:"description"`
	BaseFineAmount string `json:"base_fine_amount"`
}

type CreateViolationTypeResponse struct {
	ViolationType ViolationTypeResponse `json:"violation_type"`
}

type GetViolationTypeResponse struct {
	ViolationType ViolationTypeResponse `json:"violation_type"`
}

type ListViolationTypesResponse struct {
	ViolationTypes []ViolationTypeResponse `json:"violation_types"`
	Total          int64                   `json:"total"`
	Limit          int                     `json:"limit"`
	Offset         int                     `json:"offset"`
}

type UpdateViolationTypeRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	BaseFineAmount *string `json:"base_fine_amount"`
}

type UpdateViolationTypeResponse struct {
	ViolationType ViolationTypeResponse `json:"violation_type"`
}

type ActivateViolationTypeResponse struct {
	ViolationType ViolationTypeResponse `json:"violation_type"`
}

type DeactivateViolationTypeResponse struct {
	ViolationType ViolationTypeResponse `json:"violation_type"`
}

type DeleteViolationTypeResponse struct {
	OK bool `json:"ok"`
}

type ViolationTypeResponse struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	BaseFineAmount string    `json:"base_fine_amount"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
