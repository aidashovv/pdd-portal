package http

import (
	"fmt"
	"net/http"
	"strconv"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/request"
	"pdd-service/internal/core/transport/http/response"
	"pdd-service/internal/features/users/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UsersHTTPHandler struct {
	service *application.Service
}

func NewUsersHTTPHandler(service *application.Service) *UsersHTTPHandler {
	return &UsersHTTPHandler{service: service}
}

func (h *UsersHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	currentRole, ok := httphelpers.GetUserRole(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	role, err := parseOptionalRoleQuery(r)
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
		currentRole,
		role,
		r.URL.Query().Get("search"),
		limit,
		offset,
	))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toListUsersResponse(output))
}

func (h *UsersHTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.service.GetByID(r.Context(), toGetByIDInput(id, currentUserID, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toGetUserResponse(output))
}

func (h *UsersHTTPHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateRoleRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	input, err := toUpdateRoleInput(id, req, currentRole)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.UpdateRole(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toUpdateRoleResponse(output))
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

func parseOptionalRoleQuery(r *http.Request) (*users.Role, error) {
	raw := r.URL.Query().Get("role")
	if raw == "" {
		return nil, nil
	}

	role, err := users.ParseRole(raw)
	if err != nil {
		return nil, fmt.Errorf("parse role query: %w", coreerrors.ErrInvalidRequest)
	}

	return &role, nil
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
