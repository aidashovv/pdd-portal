package http

import (
	"fmt"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/features/payouts/application"

	"github.com/google/uuid"
)

func toListInput(userID *uuid.UUID, status *payouts.Status, limit, offset int, currentUserID uuid.UUID, currentRole users.Role) application.ListInput {
	return application.ListInput{UserID: userID, Status: status, Limit: limit, Offset: offset, CurrentUserID: currentUserID, CurrentRole: currentRole}
}

func toListByUserIDInput(userID uuid.UUID, limit, offset int, currentUserID uuid.UUID, currentRole users.Role) application.ListByUserIDInput {
	return application.ListByUserIDInput{UserID: userID, Limit: limit, Offset: offset, CurrentUserID: currentUserID, CurrentRole: currentRole}
}

func toCreateRuleInput(req CreateRuleRequest, currentRole users.Role) (application.CreateRuleInput, error) {
	violationTypeID, err := uuid.Parse(req.ViolationTypeID)
	if err != nil {
		return application.CreateRuleInput{}, fmt.Errorf("parse violation_type_id: %w", coreerrors.ErrInvalidRequest)
	}
	return application.CreateRuleInput{ViolationTypeID: violationTypeID, Percent: req.Percent, CurrentRole: currentRole}, nil
}

func toCreatePayoutInput(req CreatePayoutFromReportRequest, currentRole users.Role) (application.CreatePayoutForApprovedReportInput, error) {
	reportID, err := uuid.Parse(req.ReportID)
	if err != nil {
		return application.CreatePayoutForApprovedReportInput{}, fmt.Errorf("parse report_id: %w", coreerrors.ErrInvalidRequest)
	}
	return application.CreatePayoutForApprovedReportInput{ReportID: reportID, CurrentRole: currentRole}, nil
}

func toPayoutResponse(output application.PayoutOutput) PayoutResponse {
	return PayoutResponse{ID: output.ID.String(), ReportID: output.ReportID.String(), UserID: output.UserID.String(), Amount: output.Amount, Status: output.Status.String(), CreatedAt: output.CreatedAt, UpdatedAt: output.UpdatedAt}
}

func toPayoutResponses(outputs []application.PayoutOutput) []PayoutResponse {
	responses := make([]PayoutResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toPayoutResponse(output))
	}
	return responses
}

func toRuleResponse(output application.RuleOutput) RuleResponse {
	return RuleResponse{ID: output.ID.String(), ViolationTypeID: output.ViolationTypeID.String(), Percent: output.Percent, IsActive: output.IsActive, CreatedAt: output.CreatedAt, UpdatedAt: output.UpdatedAt}
}

func toRuleResponses(outputs []application.RuleOutput) []RuleResponse {
	responses := make([]RuleResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toRuleResponse(output))
	}
	return responses
}
