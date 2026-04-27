package http

import (
	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/features/violation_types/application"

	"github.com/google/uuid"
)

func toCreateInput(req CreateViolationTypeRequest, currentRole users.Role) application.CreateInput {
	return application.CreateInput{
		Code:           req.Code,
		Title:          req.Title,
		Description:    req.Description,
		BaseFineAmount: req.BaseFineAmount,
		CurrentRole:    currentRole,
	}
}

func toGetByIDInput(id uuid.UUID) application.GetByIDInput {
	return application.GetByIDInput{ID: id}
}

func toListInput(onlyActive *bool, search string, limit, offset int) application.ListInput {
	return application.ListInput{
		OnlyActive: onlyActive,
		Search:     search,
		Limit:      limit,
		Offset:     offset,
	}
}

func toUpdateInput(id uuid.UUID, req UpdateViolationTypeRequest, currentRole users.Role) application.UpdateInput {
	return application.UpdateInput{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		BaseFineAmount: req.BaseFineAmount,
		CurrentRole:    currentRole,
	}
}

func toDeleteInput(id uuid.UUID, currentRole users.Role) application.DeleteInput {
	return application.DeleteInput{ID: id, CurrentRole: currentRole}
}

func toActivateInput(id uuid.UUID, currentRole users.Role) application.ActivateInput {
	return application.ActivateInput{ID: id, CurrentRole: currentRole}
}

func toDeactivateInput(id uuid.UUID, currentRole users.Role) application.DeactivateInput {
	return application.DeactivateInput{ID: id, CurrentRole: currentRole}
}

func toCreateResponse(output application.CreateOutput) CreateViolationTypeResponse {
	return CreateViolationTypeResponse{ViolationType: toViolationTypeResponse(output.ViolationType)}
}

func toGetResponse(output application.GetByIDOutput) GetViolationTypeResponse {
	return GetViolationTypeResponse{ViolationType: toViolationTypeResponse(output.ViolationType)}
}

func toListResponse(output application.ListOutput) ListViolationTypesResponse {
	violationTypes := make([]ViolationTypeResponse, 0, len(output.ViolationTypes))
	for _, violationType := range output.ViolationTypes {
		violationTypes = append(violationTypes, toViolationTypeResponse(violationType))
	}

	return ListViolationTypesResponse{
		ViolationTypes: violationTypes,
		Total:          output.Total,
		Limit:          output.Limit,
		Offset:         output.Offset,
	}
}

func toUpdateResponse(output application.UpdateOutput) UpdateViolationTypeResponse {
	return UpdateViolationTypeResponse{ViolationType: toViolationTypeResponse(output.ViolationType)}
}

func toActivateResponse(output application.ActivateOutput) ActivateViolationTypeResponse {
	return ActivateViolationTypeResponse{ViolationType: toViolationTypeResponse(output.ViolationType)}
}

func toDeactivateResponse(output application.DeactivateOutput) DeactivateViolationTypeResponse {
	return DeactivateViolationTypeResponse{ViolationType: toViolationTypeResponse(output.ViolationType)}
}

func toViolationTypeResponse(output application.ViolationTypeOutput) ViolationTypeResponse {
	return ViolationTypeResponse{
		ID:             output.ID.String(),
		Code:           output.Code,
		Title:          output.Title,
		Description:    output.Description,
		BaseFineAmount: output.BaseFineAmount,
		IsActive:       output.IsActive,
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	}
}
