package payouts

import (
	"fmt"
	"strings"

	coreerrors "pdd-service/internal/core/errors"
)

type Status int16

const (
	StatusPending Status = iota
	StatusPaid
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusPaid:
		return "PAID"
	case StatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

func (s Status) IsValid() bool {
	return s == StatusPending || s == StatusPaid || s == StatusFailed
}

func ParseStatus(raw string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "PENDING":
		return StatusPending, nil
	case "PAID":
		return StatusPaid, nil
	case "FAILED":
		return StatusFailed, nil
	default:
		return StatusPending, fmt.Errorf("parse payout status %q: %w", raw, coreerrors.ErrInvalidDomainValue)
	}
}
