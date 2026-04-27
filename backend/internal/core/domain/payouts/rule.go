package payouts

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type Rule struct {
	ID              uuid.UUID
	ViolationTypeID uuid.UUID

	Percent  string
	IsActive bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRule(violationTypeID uuid.UUID, percent string) (*Rule, error) {
	rule := &Rule{
		ID:              uuid.New(),
		ViolationTypeID: violationTypeID,
		Percent:         percent,
		IsActive:        true,
	}

	if err := rule.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	return rule, nil
}

func (r *Rule) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now().UTC()
}

func (r *Rule) Deactivate() {
	r.IsActive = false
	r.UpdatedAt = time.Now().UTC()
}

func (r *Rule) UpdatePercent(percent string) error {
	r.Percent = strings.TrimSpace(percent)
	if err := r.Validate(); err != nil {
		return err
	}

	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r Rule) Validate() error {
	if r.ViolationTypeID == uuid.Nil {
		return fmt.Errorf("payout rule violation type id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	rawPercent := strings.TrimSpace(r.Percent)
	if rawPercent == "" {
		return fmt.Errorf("payout rule percent is required: %w", coreerrors.ErrInvalidDomainValue)
	}

	percent, err := strconv.Atoi(rawPercent)
	if err != nil {
		return fmt.Errorf("payout rule percent is invalid: %w", coreerrors.ErrInvalidDomainValue)
	}
	if percent <= 0 || percent > 100 {
		return fmt.Errorf("payout rule percent must be between 1 and 100: %w", coreerrors.ErrInvalidDomainValue)
	}

	return nil
}
