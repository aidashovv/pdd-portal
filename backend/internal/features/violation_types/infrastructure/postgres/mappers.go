package postgres

import "pdd-service/internal/core/domain/violations"

func violationTypeToDTO(violationType violations.ViolationType) ViolationTypeDTO {
	return ViolationTypeDTO{
		ID:             violationType.ID,
		Code:           violationType.Code,
		Title:          violationType.Title,
		Description:    violationType.Description,
		BaseFineAmount: violationType.BaseFineAmount,
		IsActive:       violationType.IsActive,
		CreatedAt:      violationType.CreatedAt,
		UpdatedAt:      violationType.UpdatedAt,
	}
}

func (dto ViolationTypeDTO) toDomain() violations.ViolationType {
	return violations.ViolationType{
		ID:             dto.ID,
		Code:           dto.Code,
		Title:          dto.Title,
		Description:    dto.Description,
		BaseFineAmount: dto.BaseFineAmount,
		IsActive:       dto.IsActive,
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
	}
}
