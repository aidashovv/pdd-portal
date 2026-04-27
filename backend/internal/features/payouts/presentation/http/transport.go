package http

import (
	"fmt"
	"net/http"
	"strconv"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	httphelpers "pdd-service/internal/core/transport/http/helpers"
	"pdd-service/internal/core/transport/http/request"
	"pdd-service/internal/core/transport/http/response"
	"pdd-service/internal/features/payouts/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PayoutsHTTPHandler struct {
	service *application.Service
}

func NewPayoutsHTTPHandler(service *application.Service) *PayoutsHTTPHandler {
	return &PayoutsHTTPHandler{service: service}
}

func (h *PayoutsHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
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
	limit, offset, err := parsePagination(r)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	output, err := h.service.List(r.Context(), toListInput(userID, status, limit, offset, currentUserID, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(ListPayoutsResponse{Payouts: toPayoutResponses(output.Payouts), Total: output.Total, Limit: output.Limit, Offset: output.Offset})
}

func (h *PayoutsHTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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
	output, err := h.service.GetByID(r.Context(), application.GetByIDInput{ID: id, CurrentUserID: currentUserID, CurrentRole: currentRole})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(GetPayoutResponse{Payout: toPayoutResponse(output.Payout)})
}

func (h *PayoutsHTTPHandler) ListByUserID(w http.ResponseWriter, r *http.Request) {
	currentUserID, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	userID, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	output, err := h.service.ListByUserID(r.Context(), toListByUserIDInput(userID, limit, offset, currentUserID, currentRole))
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(ListPayoutsResponse{Payouts: toPayoutResponses(output.Payouts), Total: output.Total, Limit: output.Limit, Offset: output.Offset})
}

func (h *PayoutsHTTPHandler) CreateFromReport(w http.ResponseWriter, r *http.Request) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	var req CreatePayoutFromReportRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	input, err := toCreatePayoutInput(req, currentRole)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	output, err := h.service.CreatePayoutForApprovedReport(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).Created(CreatePayoutFromReportResponse{Payout: toPayoutResponse(output.Payout)})
}

func (h *PayoutsHTTPHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	h.mark(w, r, true)
}

func (h *PayoutsHTTPHandler) MarkFailed(w http.ResponseWriter, r *http.Request) {
	h.mark(w, r, false)
}

func (h *PayoutsHTTPHandler) mark(w http.ResponseWriter, r *http.Request, paid bool) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	if paid {
		output, err := h.service.MarkPaid(r.Context(), application.MarkPaidInput{ID: id, CurrentRole: currentRole})
		if err != nil {
			response.NewResponseHandler(w).HandleError(err)
			return
		}
		response.NewResponseHandler(w).OK(MarkPayoutResponse{Payout: toPayoutResponse(output.Payout)})
		return
	}
	output, err := h.service.MarkFailed(r.Context(), application.MarkFailedInput{ID: id, CurrentRole: currentRole})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(MarkPayoutResponse{Payout: toPayoutResponse(output.Payout)})
}

func (h *PayoutsHTTPHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	violationTypeID, err := parseOptionalUUIDQuery(r, "violation_type_id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
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
	output, err := h.service.ListRules(r.Context(), application.ListRulesInput{ViolationTypeID: violationTypeID, OnlyActive: onlyActive, Limit: limit, Offset: offset, CurrentRole: currentRole})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(ListRulesResponse{Rules: toRuleResponses(output.Rules), Total: output.Total, Limit: output.Limit, Offset: output.Offset})
}

func (h *PayoutsHTTPHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	var req CreateRuleRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	input, err := toCreateRuleInput(req, currentRole)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	output, err := h.service.CreateRule(r.Context(), input)
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).Created(CreateRuleResponse{Rule: toRuleResponse(output.Rule)})
}

func (h *PayoutsHTTPHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	var req UpdateRuleRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	output, err := h.service.UpdateRule(r.Context(), application.UpdateRuleInput{ID: id, Percent: req.Percent, CurrentRole: currentRole})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(UpdateRuleResponse{Rule: toRuleResponse(output.Rule)})
}

func (h *PayoutsHTTPHandler) ActivateRule(w http.ResponseWriter, r *http.Request) {
	h.ruleStatus(w, r, true)
}

func (h *PayoutsHTTPHandler) DeactivateRule(w http.ResponseWriter, r *http.Request) {
	h.ruleStatus(w, r, false)
}

func (h *PayoutsHTTPHandler) ruleStatus(w http.ResponseWriter, r *http.Request, active bool) {
	_, currentRole, ok := currentUser(r)
	if !ok {
		response.NewResponseHandler(w).HandleError(coreerrors.ErrUnauthorized)
		return
	}
	id, err := parseUUIDPathParam(r, "id")
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	if active {
		output, err := h.service.ActivateRule(r.Context(), application.ActivateRuleInput{ID: id, CurrentRole: currentRole})
		if err != nil {
			response.NewResponseHandler(w).HandleError(err)
			return
		}
		response.NewResponseHandler(w).OK(RuleStatusResponse{Rule: toRuleResponse(output.Rule)})
		return
	}
	output, err := h.service.DeactivateRule(r.Context(), application.DeactivateRuleInput{ID: id, CurrentRole: currentRole})
	if err != nil {
		response.NewResponseHandler(w).HandleError(err)
		return
	}
	response.NewResponseHandler(w).OK(RuleStatusResponse{Rule: toRuleResponse(output.Rule)})
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

func parseOptionalStatusQuery(r *http.Request) (*payouts.Status, error) {
	raw := r.URL.Query().Get("status")
	if raw == "" {
		return nil, nil
	}
	status, err := payouts.ParseStatus(raw)
	if err != nil {
		return nil, fmt.Errorf("parse status query: %w", coreerrors.ErrInvalidRequest)
	}
	return &status, nil
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
