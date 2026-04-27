package http

import (
	"fmt"
	"net/http"
	"strconv"

	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/request"
	"pdd-service/internal/core/transport/http/response"
	"pdd-service/internal/features/violation_types/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ViolationTypesHTTPHandler struct {
	service *application.Service
}

func NewViolationTypesHTTPHandler(service *application.Service) *ViolationTypesHTTPHandler {
	return &ViolationTypesHTTPHandler{service: service}
}

func (h *ViolationTypesHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	onlyActive, err := parseOptionalBoolQuery(r, "only_active")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.List(r.Context(), toListInput(
		onlyActive,
		r.URL.Query().Get("search"),
		limit,
		offset,
	))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toListResponse(output))
}

func (h *ViolationTypesHTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.GetByID(r.Context(), toGetByIDInput(id))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toGetResponse(output))
}

func (h *ViolationTypesHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	var req CreateViolationTypeRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Create(r.Context(), toCreateInput(req, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).Created(toCreateResponse(output))
}

func (h *ViolationTypesHTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	var req UpdateViolationTypeRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Update(r.Context(), toUpdateInput(id, req, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toUpdateResponse(output))
}

func (h *ViolationTypesHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	if err := h.service.Delete(r.Context(), toDeleteInput(id, currentRole)); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(DeleteViolationTypeResponse{OK: true})
}

func (h *ViolationTypesHTTPHandler) Activate(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Activate(r.Context(), toActivateInput(id, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toActivateResponse(output))
}

func (h *ViolationTypesHTTPHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Deactivate(r.Context(), toDeactivateInput(id, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toDeactivateResponse(output))
}

func parseUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	value, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid path param %q: %w", key, coreerrors.ErrInvalidRequest)
	}

	return value, nil
}

func parseOptionalBoolQuery(r *http.Request, key string) (*bool, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("parse bool query %q: %w", key, coreerrors.ErrInvalidRequest)
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
