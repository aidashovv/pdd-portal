package payouts

import (
	"fmt"
	"strings"
	"time"

	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

type Payout struct {
	ID       uuid.UUID
	ReportID uuid.UUID
	UserID   uuid.UUID

	Amount string
	Status Status

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPayout(reportID uuid.UUID, userID uuid.UUID, amount string) (*Payout, error) {
	amount = strings.TrimSpace(amount)
	if reportID == uuid.Nil {
		return nil, fmt.Errorf("payout report id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if userID == uuid.Nil {
		return nil, fmt.Errorf("payout user id is required: %w", coreerrors.ErrInvalidDomainValue)
	}
	if amount == "" {
		return nil, fmt.Errorf("payout amount is required: %w", coreerrors.ErrInvalidDomainValue)
	}

	now := time.Now().UTC()
	return &Payout{
		ID:        uuid.New(),
		ReportID:  reportID,
		UserID:    userID,
		Amount:    amount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Payout) MarkPending() error {
	if p.Status == StatusPaid {
		return fmt.Errorf("mark payout pending from %s: %w", p.Status, coreerrors.ErrInvalidTransition)
	}

	p.Status = StatusPending
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payout) MarkPaid() error {
	if !p.CanBePaid() {
		return fmt.Errorf("mark payout paid from %s: %w", p.Status, coreerrors.ErrInvalidTransition)
	}

	p.Status = StatusPaid
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payout) MarkFailed() error {
	if p.Status == StatusPaid {
		return fmt.Errorf("mark payout failed from %s: %w", p.Status, coreerrors.ErrInvalidTransition)
	}

	p.Status = StatusFailed
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p Payout) CanBePaid() bool {
	return p.Status == StatusPending || p.Status == StatusFailed
}
