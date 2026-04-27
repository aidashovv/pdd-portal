package postgres

import (
	"fmt"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
)

func payoutToDTO(payout payouts.Payout) PayoutDTO {
	return PayoutDTO{
		ID:        payout.ID,
		ReportID:  payout.ReportID,
		UserID:    payout.UserID,
		Amount:    payout.Amount,
		Status:    int16(payout.Status),
		CreatedAt: payout.CreatedAt,
		UpdatedAt: payout.UpdatedAt,
	}
}

func (dto PayoutDTO) toDomain() (payouts.Payout, error) {
	status := payouts.Status(dto.Status)
	if !status.IsValid() {
		return payouts.Payout{}, fmt.Errorf("invalid payout status from db: %w", coreerrors.ErrInvalidDomainValue)
	}
	return payouts.Payout{ID: dto.ID, ReportID: dto.ReportID, UserID: dto.UserID, Amount: dto.Amount, Status: status, CreatedAt: dto.CreatedAt, UpdatedAt: dto.UpdatedAt}, nil
}

func ruleToDTO(rule payouts.Rule) RuleDTO {
	return RuleDTO{ID: rule.ID, ViolationTypeID: rule.ViolationTypeID, Percent: rule.Percent, IsActive: rule.IsActive, CreatedAt: rule.CreatedAt, UpdatedAt: rule.UpdatedAt}
}

func (dto RuleDTO) toDomain() payouts.Rule {
	return payouts.Rule{ID: dto.ID, ViolationTypeID: dto.ViolationTypeID, Percent: dto.Percent, IsActive: dto.IsActive, CreatedAt: dto.CreatedAt, UpdatedAt: dto.UpdatedAt}
}

func (dto ReportDTO) toDomain() (violations.Report, error) {
	status := violations.Status(dto.Status)
	if !status.IsValid() {
		return violations.Report{}, fmt.Errorf("invalid report status from db: %w", coreerrors.ErrInvalidDomainValue)
	}
	source := violations.Source(dto.VideoSource)
	if !source.IsValid() {
		return violations.Report{}, fmt.Errorf("invalid video source from db: %w", coreerrors.ErrInvalidDomainValue)
	}
	return violations.Report{
		ID: dto.ID, UserID: dto.UserID, ViolationTypeID: dto.ViolationTypeID, ModeratorID: dto.ModeratorID,
		Title: dto.Title, Description: dto.Description, Location: dto.Location, OccurredAt: dto.OccurredAt,
		Status: status, Video: violations.Video{Source: source, URL: dto.VideoURL, ObjectKey: dto.VideoObjectKey, ContentType: dto.VideoContentType, Size: dto.VideoSize},
		ModerationComment: dto.ModerationComment, CreatedAt: dto.CreatedAt, UpdatedAt: dto.UpdatedAt,
	}, nil
}

func (dto ViolationTypeDTO) toDomain() violations.ViolationType {
	return violations.ViolationType{ID: dto.ID, Code: dto.Code, Title: dto.Title, Description: dto.Description, BaseFineAmount: dto.BaseFineAmount, IsActive: dto.IsActive, CreatedAt: dto.CreatedAt, UpdatedAt: dto.UpdatedAt}
}
