package postgres

import (
	"time"

	"github.com/google/uuid"
)

type ReportDTO struct {
	ID                uuid.UUID `db:"id"`
	UserID            uuid.UUID `db:"user_id"`
	ViolationTypeID   uuid.UUID `db:"violation_type_id"`
	Title             string    `db:"title"`
	Description       string    `db:"description"`
	Location          string    `db:"location"`
	OccurredAt        time.Time `db:"occurred_at"`
	Status            int16     `db:"status"`
	VideoSource       int16     `db:"video_source"`
	VideoURL          string    `db:"video_url"`
	VideoObjectKey    string    `db:"video_object_key"`
	VideoContentType  string    `db:"video_content_type"`
	VideoSize         int64     `db:"video_size"`
	ModeratorID       uuid.UUID `db:"moderator_id"`
	ModerationComment string    `db:"moderation_comment"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type ViolationTypeDTO struct {
	ID             uuid.UUID `db:"id"`
	Code           string    `db:"code"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	BaseFineAmount string    `db:"base_fine_amount"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
