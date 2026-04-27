package violations

import (
	"fmt"
	"strings"
	"time"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type Report struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ViolationTypeID uuid.UUID
	ModeratorID     uuid.UUID

	Title             string
	Description       string
	Location          string
	OccurredAt        time.Time
	Status            Status
	Video             Video
	ModerationComment string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewReport(
	userID uuid.UUID,
	violationTypeID uuid.UUID,
	title string,
	description string,
	location string,
	occurredAt time.Time,
	video Video,
) (*Report, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	location = strings.TrimSpace(location)

	if userID == uuid.Nil {
		return nil, fmt.Errorf("report user id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if violationTypeID == uuid.Nil {
		return nil, fmt.Errorf("report violation type id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if title == "" {
		return nil, fmt.Errorf("report title is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if description == "" {
		return nil, fmt.Errorf("report description is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if location == "" {
		return nil, fmt.Errorf("report location is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if occurredAt.IsZero() {
		return nil, fmt.Errorf("report occurred time is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if err := video.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Report{
		ID:              uuid.New(),
		UserID:          userID,
		ViolationTypeID: violationTypeID,
		Title:           title,
		Description:     description,
		Location:        location,
		OccurredAt:      occurredAt.UTC(),
		Status:          StatusDraft,
		Video:           video,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *Report) Submit() error {
	if r.Status != StatusDraft {
		return fmt.Errorf("submit report from %s: %w", r.Status, coreerrors.ErrInvalidTransition)
	}
	if err := r.ValidateVideoRequired(); err != nil {
		return err
	}

	r.Status = StatusSubmitted
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Report) UpdateDetails(
	violationTypeID uuid.UUID,
	title string,
	description string,
	location string,
	occurredAt time.Time,
	video Video,
) error {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	location = strings.TrimSpace(location)

	if violationTypeID == uuid.Nil {
		return fmt.Errorf("report violation type id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if title == "" {
		return fmt.Errorf("report title is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if description == "" {
		return fmt.Errorf("report description is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if location == "" {
		return fmt.Errorf("report location is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("report occurred time is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if err := video.Validate(); err != nil {
		return err
	}

	r.ViolationTypeID = violationTypeID
	r.Title = title
	r.Description = description
	r.Location = location
	r.OccurredAt = occurredAt.UTC()
	r.Video = video
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Report) StartReview(moderatorID uuid.UUID) error {
	if moderatorID == uuid.Nil {
		return fmt.Errorf("moderator id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if r.Status != StatusSubmitted {
		return fmt.Errorf("start review from %s: %w", r.Status, coreerrors.ErrInvalidTransition)
	}

	r.Status = StatusInReview
	r.ModeratorID = moderatorID
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Report) Approve(moderatorID uuid.UUID, comment string) error {
	if moderatorID == uuid.Nil {
		return fmt.Errorf("moderator id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if r.Status != StatusInReview {
		return fmt.Errorf("approve report from %s: %w", r.Status, coreerrors.ErrInvalidTransition)
	}

	r.Status = StatusApproved
	r.ModeratorID = moderatorID
	r.ModerationComment = strings.TrimSpace(comment)
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Report) Reject(moderatorID uuid.UUID, comment string) error {
	comment = strings.TrimSpace(comment)
	if moderatorID == uuid.Nil {
		return fmt.Errorf("moderator id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if comment == "" {
		return fmt.Errorf("moderation comment is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if r.Status != StatusInReview {
		return fmt.Errorf("reject report from %s: %w", r.Status, coreerrors.ErrInvalidTransition)
	}

	r.Status = StatusRejected
	r.ModeratorID = moderatorID
	r.ModerationComment = comment
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Report) MarkPaid() error {
	if r.Status != StatusApproved {
		return fmt.Errorf("mark report paid from %s: %w", r.Status, coreerrors.ErrInvalidTransition)
	}

	r.Status = StatusPaid
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r Report) IsOwner(userID uuid.UUID) bool {
	return r.UserID == userID
}

func (r Report) IsDraft() bool {
	return r.Status == StatusDraft
}

func (r Report) IsSubmitted() bool {
	return r.Status == StatusSubmitted
}

func (r Report) IsInReview() bool {
	return r.Status == StatusInReview
}

func (r Report) IsApproved() bool {
	return r.Status == StatusApproved
}

func (r Report) ValidateVideoRequired() error {
	if !r.Video.Source.IsValid() {
		return fmt.Errorf("report video is required: %w", coreerrors.ErrVideoRequired)
	}

	return r.Video.Validate()
}

func (r Report) CanBeViewedBy(userID uuid.UUID, role users.Role) bool {
	return role == users.RoleAdmin || role == users.RoleModerator || r.UserID == userID
}

func (r Report) CanBeEditedBy(userID uuid.UUID, role users.Role) bool {
	if role == users.RoleAdmin {
		return true
	}

	return role == users.RoleUser &&
		r.UserID == userID &&
		(r.Status == StatusDraft || r.Status == StatusSubmitted)
}

func (r Report) CanBeDeletedBy(userID uuid.UUID, role users.Role) bool {
	if role == users.RoleAdmin {
		return true
	}

	return role == users.RoleUser &&
		r.UserID == userID &&
		r.Status == StatusDraft
}
