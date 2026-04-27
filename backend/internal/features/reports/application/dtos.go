package application

import (
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"

	"github.com/google/uuid"
)

type VideoInput struct {
	Source      violations.Source
	URL         string
	ObjectKey   string
	ContentType string
	Size        int64
}

type CreateInput struct {
	UserID          uuid.UUID
	ViolationTypeID uuid.UUID
	Title           string
	Description     string
	Location        string
	OccurredAt      time.Time
	Video           VideoInput
	CurrentUserID   uuid.UUID
	CurrentRole     users.Role
}

type CreateOutput struct {
	Report ReportOutput
}

type GetByIDInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type GetByIDOutput struct {
	Report ReportOutput
}

type ListInput struct {
	UserID          *uuid.UUID
	Status          *violations.Status
	ViolationTypeID *uuid.UUID
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Limit           int
	Offset          int
	CurrentUserID   uuid.UUID
	CurrentRole     users.Role
}

type ListOutput struct {
	Reports []ReportOutput
	Total   int64
	Limit   int
	Offset  int
}

type UpdateInput struct {
	ID              uuid.UUID
	ViolationTypeID *uuid.UUID
	Title           *string
	Description     *string
	Location        *string
	OccurredAt      *time.Time
	Video           *VideoInput
	CurrentUserID   uuid.UUID
	CurrentRole     users.Role
}

type UpdateOutput struct {
	Report ReportOutput
}

type DeleteInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type SubmitInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type SubmitOutput struct {
	Report ReportOutput
}

type StartReviewInput struct {
	ID            uuid.UUID
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type StartReviewOutput struct {
	Report ReportOutput
}

type ApproveInput struct {
	ID            uuid.UUID
	Comment       string
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type ApproveOutput struct {
	Report ReportOutput
}

type RejectInput struct {
	ID            uuid.UUID
	Comment       string
	CurrentUserID uuid.UUID
	CurrentRole   users.Role
}

type RejectOutput struct {
	Report ReportOutput
}

type ReportOutput struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	ViolationTypeID   uuid.UUID
	ModeratorID       uuid.UUID
	Title             string
	Description       string
	Location          string
	OccurredAt        time.Time
	Status            violations.Status
	Video             VideoOutput
	ModerationComment string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type VideoOutput struct {
	Source      violations.Source
	URL         string
	ObjectKey   string
	ContentType string
	Size        int64
}
