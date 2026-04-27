package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/request"
	"pdd-service/internal/core/transport/http/response"
	"pdd-service/internal/features/reports/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportsHTTPHandler struct {
	service *application.Service
}

func NewReportsHTTPHandler(service *application.Service) *ReportsHTTPHandler {
	return &ReportsHTTPHandler{service: service}
}

func (h *ReportsHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	userID, err := parseOptionalUUIDQuery(r, "user_id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	status, err := parseOptionalStatusQuery(r)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	violationTypeID, err := parseOptionalUUIDQuery(r, "violation_type_id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	input := toListInput(userID, status, violationTypeID, currentUserID, currentRole, limit, offset)
	createdFrom, err := parseOptionalTimeQuery(r, "created_from")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	createdTo, err := parseOptionalTimeQuery(r, "created_to")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	input.CreatedFrom = createdFrom
	input.CreatedTo = createdTo

	output, err := h.service.List(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toListResponse(output))
}

func (h *ReportsHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	var req CreateReportRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	input, err := toCreateInput(req, currentUserID, currentRole)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Create(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).Created(toCreateResponse(output))
}

func (h *ReportsHTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.GetByID(r.Context(), application.GetByIDInput{
		ID:            id,
		CurrentUserID: currentUserID,
		CurrentRole:   currentRole,
	})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toGetResponse(output))
}

func (h *ReportsHTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	var req UpdateReportRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	input, err := toUpdateInput(id, req, currentUserID, currentRole)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Update(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toUpdateResponse(output))
}

func (h *ReportsHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	if err := h.service.Delete(r.Context(), application.DeleteInput{
		ID:            id,
		CurrentUserID: currentUserID,
		CurrentRole:   currentRole,
	}); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(DeleteReportResponse{OK: true})
}

func (h *ReportsHTTPHandler) Submit(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Submit(r.Context(), application.SubmitInput{
		ID:            id,
		CurrentUserID: currentUserID,
		CurrentRole:   currentRole,
	})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toSubmitResponse(output))
}

func (h *ReportsHTTPHandler) StartReview(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.StartReview(r.Context(), application.StartReviewInput{
		ID:            id,
		CurrentUserID: currentUserID,
		CurrentRole:   currentRole,
	})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toStartReviewResponse(output))
}

func (h *ReportsHTTPHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.moderate(w, r, "approve")
}

func (h *ReportsHTTPHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.moderate(w, r, "reject")
}

func (h *ReportsHTTPHandler) Moderate(w http.ResponseWriter, r *http.Request) {
	var req ModerateRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	switch req.Action {
	case "approve":
		h.moderateWithComment(w, r, "approve", req.Comment)
	case "reject":
		h.moderateWithComment(w, r, "reject", req.Comment)
	default:
		response.NewResponseHandler(w).HandleError(coreerrors.ErrInvalidRequest)
	}
}

func (h *ReportsHTTPHandler) moderate(w http.ResponseWriter, r *http.Request, action string) {
	var req ModerationRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	h.moderateWithComment(w, r, action, req.Comment)
}

func (h *ReportsHTTPHandler) moderateWithComment(w http.ResponseWriter, r *http.Request, action string, comment string) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	switch action {
	case "approve":
		output, err := h.service.Approve(r.Context(), application.ApproveInput{
			ID:            id,
			Comment:       comment,
			CurrentUserID: currentUserID,
			CurrentRole:   currentRole,
		})
		if err != nil {
			response.NewResponseHandler(w).HandleError(err)
			return
		}
		response.NewResponseHandler(w).OK(toApproveResponse(output))
	case "reject":
		output, err := h.service.Reject(r.Context(), application.RejectInput{
			ID:            id,
			Comment:       comment,
			CurrentUserID: currentUserID,
			CurrentRole:   currentRole,
		})
		if err != nil {
			response.NewResponseHandler(w).HandleError(err)
			return
		}
		response.NewResponseHandler(w).OK(toRejectResponse(output))
	default:
		response.NewResponseHandler(w).HandleError(coreerrors.ErrInvalidRequest)
	}
}

func currentUser(r *http.Request) (uuid.UUID, users.Role, bool) {
	userID, ok := httphelpers.GetUserID(r.Context())
	if !ok {
		return uuid.Nil, users.RoleUser, false
	}
	role, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		return uuid.Nil, users.RoleUser, false
	}

	return userID, role, true
}

func parseUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	value, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid path param %q: %w", key, coreerrors.ErrInvalidRequest)
	}

	return value, nil
}

func parseOptionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := uuid.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse uuid query %q: %w", key, coreerrors.ErrInvalidRequest)
	}

	return &value, nil
}

func parseOptionalStatusQuery(r *http.Request) (*violations.Status, error) {
	raw := r.URL.Query().Get("status")
	if raw == "" {
		return nil, nil
	}

	status, err := violations.ParseStatus(raw)
	if err != nil {
		return nil, fmt.Errorf("parse status query: %w", coreerrors.ErrInvalidRequest)
	}

	return &status, nil
}

func parseOptionalTimeQuery(r *http.Request, key string) (*time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("parse time query %q: %w", key, coreerrors.ErrInvalidRequest)
	}

	return &value, nil
}

func parsePagination(r *http.Request) (int, int, error) {
	limit, err := parseIntQuery(r, "limit", 0)
	if err != nil {
		return 0, 0, err
	}
	offset, err := parseIntQuery(r, "offset", 0)
	if err != nil {
		return 0, 0, err
	}

	return limit, offset, nil
}

func parseIntQuery(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse int query %q: %w", key, coreerrors.ErrInvalidRequest)
	}

	return value, nil
}
