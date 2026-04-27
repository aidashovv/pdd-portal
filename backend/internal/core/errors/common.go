package errors

import "errors"

var (
	ErrInvalidRequest = errors.New("INVALID_REQUEST")
	ErrUnauthorized   = errors.New("UNAUTHORIZED")
	ErrForbidden      = errors.New("FORBIDDEN")
	ErrNotFound       = errors.New("NOT_FOUND")
	ErrNotImplemented = errors.New("NOT_IMPLEMENTED")
	ErrInternalError  = errors.New("INTERNAL_ERROR")

	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrEmailAlreadyExists = errors.New("EMAIL_ALREADY_EXISTS")
	ErrSessionNotFound    = errors.New("SESSION_NOT_FOUND")
	ErrSessionRevoked     = errors.New("SESSION_REVOKED")

	ErrViolationTypeAlreadyExists = errors.New("VIOLATION_TYPE_ALREADY_EXISTS")
	ErrInvalidReportStatus        = errors.New("INVALID_REPORT_STATUS")
	ErrPayoutRuleNotFound         = errors.New("PAYOUT_RULE_NOT_FOUND")
	ErrPayoutRuleAlreadyExists    = errors.New("PAYOUT_RULE_ALREADY_EXISTS")
	ErrInvalidPayoutStatus        = errors.New("INVALID_PAYOUT_STATUS")
)
