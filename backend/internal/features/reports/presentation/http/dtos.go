package http

import "time"

type CreateReportRequest struct {
	UserID          string       `json:"user_id"`
	ViolationTypeID string       `json:"violation_type_id" validate:"required"`
	Title           string       `json:"title" validate:"required"`
	Description     string       `json:"description" validate:"required"`
	Location        string       `json:"location" validate:"required"`
	OccurredAt      time.Time    `json:"occurred_at" validate:"required"`
	Video           VideoRequest `json:"video" validate:"required"`
}

type UpdateReportRequest struct {
	ViolationTypeID *string       `json:"violation_type_id"`
	Title           *string       `json:"title"`
	Description     *string       `json:"description"`
	Location        *string       `json:"location"`
	OccurredAt      *time.Time    `json:"occurred_at"`
	Video           *VideoRequest `json:"video"`
}

type VideoRequest struct {
	Source      string `json:"source" validate:"required"`
	URL         string `json:"url"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type ModerationRequest struct {
	Comment string `json:"comment"`
}

type ModerateRequest struct {
	Action  string `json:"action" validate:"required"`
	Comment string `json:"comment"`
}

type ReportResponse struct {
	ID                string        `json:"id"`
	UserID            string        `json:"user_id"`
	ViolationTypeID   string        `json:"violation_type_id"`
	ModeratorID       string        `json:"moderator_id,omitempty"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Location          string        `json:"location"`
	OccurredAt        time.Time     `json:"occurred_at"`
	Status            string        `json:"status"`
	Video             VideoResponse `json:"video"`
	ModerationComment string        `json:"moderation_comment,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type VideoResponse struct {
	Source      string `json:"source"`
	URL         string `json:"url,omitempty"`
	ObjectKey   string `json:"object_key,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

type CreateReportResponse struct {
	Report ReportResponse `json:"report"`
}

type GetReportResponse struct {
	Report ReportResponse `json:"report"`
}

type ListReportsResponse struct {
	Reports []ReportResponse `json:"reports"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

type UpdateReportResponse struct {
	Report ReportResponse `json:"report"`
}

type DeleteReportResponse struct {
	OK bool `json:"ok"`
}

type SubmitReportResponse struct {
	Report ReportResponse `json:"report"`
}

type StartReviewReportResponse struct {
	Report ReportResponse `json:"report"`
}

type ApproveReportResponse struct {
	Report ReportResponse `json:"report"`
}

type RejectReportResponse struct {
	Report ReportResponse `json:"report"`
}
