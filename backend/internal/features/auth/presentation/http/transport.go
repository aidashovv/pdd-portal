package http

import (
	"net/http"

	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/request"
	"pdd-service/internal/core/transport/http/response"
	"pdd-service/internal/features/auth/application"
)

type AuthHTTPHandler struct {
	service *application.AuthService
}

func NewAuthHTTPHandler(service *application.AuthService) *AuthHTTPHandler {
	return &AuthHTTPHandler{service: service}
}

func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Register(r.Context(), toRegisterInput(req))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).Created(toRegisterResponse(output))
}

func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Login(r.Context(), toLoginInput(req))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toLoginResponse(output))
}

func (h *AuthHTTPHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	output, err := h.service.Refresh(r.Context(), toRefreshInput(req))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toRefreshResponse(output))
}

func (h *AuthHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	if err := h.service.Logout(r.Context(), toLogoutInput(req)); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(LogoutResponse{OK: true})
}

func (h *AuthHTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := httphelpers.GetUserID(r.Context())
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}

	output, err := h.service.Me(r.Context(), application.MeInput{UserID: userID})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}

	response.NewResponseHandler(w).OK(toMeResponse(output))
}
