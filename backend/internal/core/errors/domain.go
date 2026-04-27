package errors

import "errors"

var (
	ErrUserNotFound          = errors.New("USER_NOT_FOUND")
	ErrReportNotFound        = errors.New("REPORT_NOT_FOUND")
	ErrViolationTypeNotFound = errors.New("VIOLATION_TYPE_NOT_FOUND")
	ErrPayoutNotFound        = errors.New("PAYOUT_NOT_FOUND")

	ErrInvalidDomainValue = errors.New("INVALID_DOMAIN_VALUE")
	ErrInvalidTransition  = errors.New("INVALID_TRANSITION")
	ErrVideoRequired      = errors.New("VIDEO_REQUIRED")
)
