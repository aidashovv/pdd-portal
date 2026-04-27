package response

import (
	"encoding/json"
	"errors"
	"net/http"

	apperr "pdd-service/internal/core/errors"
)

type ResponseHandler struct {
	w http.ResponseWriter
}

func NewResponseHandler(w http.ResponseWriter) *ResponseHandler {
	return &ResponseHandler{w: w}
}

func (h *ResponseHandler) JSON(status int, data any) {
	h.w.Header().Set("Content-Type", "application/json")
	h.w.WriteHeader(status)
	_ = json.NewEncoder(h.w).Encode(data)
}

func (h *ResponseHandler) HandleError(err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"

	switch {
	case errors.Is(err, apperr.ErrInvalidRequest):
		status = http.StatusBadRequest
		code = "INVALID_REQUEST"
	case errors.Is(err, apperr.ErrInvalidDomainValue):
		status = http.StatusBadRequest
		code = "INVALID_DOMAIN_VALUE"
	case errors.Is(err, apperr.ErrEmailAlreadyExists):
		status = http.StatusConflict
		code = "EMAIL_ALREADY_EXISTS"
	case errors.Is(err, apperr.ErrPayoutRuleAlreadyExists):
		status = http.StatusConflict
		code = "PAYOUT_RULE_ALREADY_EXISTS"
	case errors.Is(err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "UNAUTHORIZED"
	case errors.Is(err, apperr.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		code = "INVALID_CREDENTIALS"
	case errors.Is(err, apperr.ErrSessionRevoked):
		status = http.StatusUnauthorized
		code = "SESSION_REVOKED"
	case errors.Is(err, apperr.ErrForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"
	case errors.Is(err, apperr.ErrInvalidReportStatus), errors.Is(err, apperr.ErrInvalidPayoutStatus):
		status = http.StatusBadRequest
		code = "INVALID_STATUS"
	case errors.Is(err, apperr.ErrNotFound), errors.Is(err, apperr.ErrUserNotFound), errors.Is(err, apperr.ErrSessionNotFound), errors.Is(err, apperr.ErrPayoutNotFound), errors.Is(err, apperr.ErrPayoutRuleNotFound):
		status = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, apperr.ErrNotImplemented):
		status = http.StatusNotImplemented
		code = "NOT_IMPLEMENTED"
	}

	h.JSON(status, ErrorResponse{
		Error: ErrorData{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func (h *ResponseHandler) OK(data any) {
	h.JSON(http.StatusOK, data)
}

func (h *ResponseHandler) Created(data any) {
	h.JSON(http.StatusCreated, data)
}
