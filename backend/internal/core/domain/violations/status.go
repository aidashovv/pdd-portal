package violations

import (
	"fmt"
	"strings"

	coreerrors "pdd-service/internal/core/errors"
)

type Status int16

const (
	StatusDraft Status = iota
	StatusSubmitted
	StatusInReview
	StatusApproved
	StatusRejected
	StatusPaid
)

func (s Status) String() string {
	switch s {
	case StatusDraft:
		return "DRAFT"
	case StatusSubmitted:
		return "SUBMITTED"
	case StatusInReview:
		return "IN_REVIEW"
	case StatusApproved:
		return "APPROVED"
	case StatusRejected:
		return "REJECTED"
	case StatusPaid:
		return "PAID"
	default:
		return "UNKNOWN"
	}
}

func (s Status) IsValid() bool {
	return s == StatusDraft ||
		s == StatusSubmitted ||
		s == StatusInReview ||
		s == StatusApproved ||
		s == StatusRejected ||
		s == StatusPaid
}

func ParseStatus(raw string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "DRAFT":
		return StatusDraft, nil
	case "SUBMITTED":
		return StatusSubmitted, nil
	case "IN_REVIEW":
		return StatusInReview, nil
	case "APPROVED":
		return StatusApproved, nil
	case "REJECTED":
		return StatusRejected, nil
	case "PAID":
		return StatusPaid, nil
	default:
		return StatusDraft, fmt.Errorf("parse report status %q: %w", raw, coreerrors.ErrInvalidDomainValue)
	}
}
