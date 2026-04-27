package violations

import (
	"fmt"
	"strings"
	"time"

	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type ViolationType struct {
	ID uuid.UUID

	Code           string
	Title          string
	Description    string
	BaseFineAmount string
	IsActive       bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewViolationType(code, title, description string, baseFineAmount string) (*ViolationType, error) {
	code = strings.TrimSpace(code)
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	baseFineAmount = strings.TrimSpace(baseFineAmount)

	if code == "" {
		return nil, fmt.Errorf("violation type code is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if title == "" {
		return nil, fmt.Errorf("violation type title is required: %w", coreerrors.ErrInvalidDomainValue)
	}

	now := time.Now().UTC()
	return &ViolationType{
		ID:             uuid.New(),
		Code:           code,
		Title:          title,
		Description:    description,
		BaseFineAmount: baseFineAmount,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (t *ViolationType) Activate() {
	t.IsActive = true
	t.UpdatedAt = time.Now().UTC()
}

func (t *ViolationType) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now().UTC()
}

func (t *ViolationType) Rename(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("violation type title is required: %w", coreerrors.ErrInvalidDomainValue)
	}

	t.Title = title
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *ViolationType) UpdateDescription(description string) {
	t.Description = strings.TrimSpace(description)
	t.UpdatedAt = time.Now().UTC()
}

func (t *ViolationType) UpdateBaseFineAmount(baseFineAmount string) {
	t.BaseFineAmount = strings.TrimSpace(baseFineAmount)
	t.UpdatedAt = time.Now().UTC()
}

func (t ViolationType) IsAvailableForReports() bool {
	return t.IsActive
}
