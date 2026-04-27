package postgres

import (
	"fmt"

	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
)

func reportToDTO(report violations.Report) ReportDTO {
	return ReportDTO{
		ID:                report.ID,
		UserID:            report.UserID,
		ViolationTypeID:   report.ViolationTypeID,
		Title:             report.Title,
		Description:       report.Description,
		Location:          report.Location,
		OccurredAt:        report.OccurredAt,
		Status:            int16(report.Status),
		VideoSource:       int16(report.Video.Source),
		VideoURL:          report.Video.URL,
		VideoObjectKey:    report.Video.ObjectKey,
		VideoContentType:  report.Video.ContentType,
		VideoSize:         report.Video.Size,
		ModeratorID:       report.ModeratorID,
		ModerationComment: report.ModerationComment,
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
	}
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

	video := violations.Video{
		Source:      source,
		URL:         dto.VideoURL,
		ObjectKey:   dto.VideoObjectKey,
		ContentType: dto.VideoContentType,
		Size:        dto.VideoSize,
	}

	return violations.Report{
		ID:                dto.ID,
		UserID:            dto.UserID,
		ViolationTypeID:   dto.ViolationTypeID,
		ModeratorID:       dto.ModeratorID,
		Title:             dto.Title,
		Description:       dto.Description,
		Location:          dto.Location,
		OccurredAt:        dto.OccurredAt,
		Status:            status,
		Video:             video,
		ModerationComment: dto.ModerationComment,
		CreatedAt:         dto.CreatedAt,
		UpdatedAt:         dto.UpdatedAt,
	}, nil
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
