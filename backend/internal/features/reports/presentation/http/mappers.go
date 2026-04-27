package http

import (
	"fmt"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	"pdd-service/internal/features/reports/application"

	"github.com/google/uuid"
)

func toCreateInput(req CreateReportRequest, currentUserID uuid.UUID, currentRole users.Role) (application.CreateInput, error) {
	userID := uuid.Nil
	if req.UserID != "" {
		parsed, err := uuid.Parse(req.UserID)
		if err != nil {
			return application.CreateInput{}, fmt.Errorf("parse user_id: %w", coreerrors.ErrInvalidRequest)
		}
		userID = parsed
	}

	violationTypeID, err := uuid.Parse(req.ViolationTypeID)
	if err != nil {
		return application.CreateInput{}, fmt.Errorf("parse violation_type_id: %w", coreerrors.ErrInvalidRequest)
	}

	video, err := toVideoInput(req.Video)
	if err != nil {
		return application.CreateInput{}, err
	}

	return application.CreateInput{
		UserID:          userID,
		ViolationTypeID: violationTypeID,
		Title:           req.Title,
		Description:     req.Description,
		Location:        req.Location,
		OccurredAt:      req.OccurredAt,
		Video:           video,
		CurrentUserID:   currentUserID,
		CurrentRole:     currentRole,
	}, nil
}

func toUpdateInput(id uuid.UUID, req UpdateReportRequest, currentUserID uuid.UUID, currentRole users.Role) (application.UpdateInput, error) {
	var violationTypeID *uuid.UUID
	if req.ViolationTypeID != nil {
		parsed, err := uuid.Parse(*req.ViolationTypeID)
		if err != nil {
			return application.UpdateInput{}, fmt.Errorf("parse violation_type_id: %w", coreerrors.ErrInvalidRequest)
		}
		violationTypeID = &parsed
	}

	var video *application.VideoInput
	if req.Video != nil {
		parsed, err := toVideoInput(*req.Video)
		if err != nil {
			return application.UpdateInput{}, err
		}
		video = &parsed
	}

	return application.UpdateInput{
		ID:              id,
		ViolationTypeID: violationTypeID,
		Title:           req.Title,
		Description:     req.Description,
		Location:        req.Location,
		OccurredAt:      req.OccurredAt,
		Video:           video,
		CurrentUserID:   currentUserID,
		CurrentRole:     currentRole,
	}, nil
}

func toVideoInput(req VideoRequest) (application.VideoInput, error) {
	source, err := violations.ParseSource(req.Source)
	if err != nil {
		return application.VideoInput{}, fmt.Errorf("parse video source: %w", coreerrors.ErrInvalidRequest)
	}

	return application.VideoInput{
		Source:      source,
		URL:         req.URL,
		ObjectKey:   req.ObjectKey,
		ContentType: req.ContentType,
		Size:        req.Size,
	}, nil
}

func toListInput(
	userID *uuid.UUID,
	status *violations.Status,
	violationTypeID *uuid.UUID,
	currentUserID uuid.UUID,
	currentRole users.Role,
	limit int,
	offset int,
) application.ListInput {
	return application.ListInput{
		UserID:          userID,
		Status:          status,
		ViolationTypeID: violationTypeID,
		Limit:           limit,
		Offset:          offset,
		CurrentUserID:   currentUserID,
		CurrentRole:     currentRole,
	}
}

func toReportResponse(output application.ReportOutput) ReportResponse {
	moderatorID := ""
	if output.ModeratorID != uuid.Nil {
		moderatorID = output.ModeratorID.String()
	}

	return ReportResponse{
		ID:                output.ID.String(),
		UserID:            output.UserID.String(),
		ViolationTypeID:   output.ViolationTypeID.String(),
		ModeratorID:       moderatorID,
		Title:             output.Title,
		Description:       output.Description,
		Location:          output.Location,
		OccurredAt:        output.OccurredAt,
		Status:            output.Status.String(),
		Video:             toVideoResponse(output.Video),
		ModerationComment: output.ModerationComment,
		CreatedAt:         output.CreatedAt,
		UpdatedAt:         output.UpdatedAt,
	}
}

func toVideoResponse(output application.VideoOutput) VideoResponse {
	return VideoResponse{
		Source:      output.Source.String(),
		URL:         output.URL,
		ObjectKey:   output.ObjectKey,
		ContentType: output.ContentType,
		Size:        output.Size,
	}
}

func toCreateResponse(output application.CreateOutput) CreateReportResponse {
	return CreateReportResponse{Report: toReportResponse(output.Report)}
}

func toGetResponse(output application.GetByIDOutput) GetReportResponse {
	return GetReportResponse{Report: toReportResponse(output.Report)}
}

func toListResponse(output application.ListOutput) ListReportsResponse {
	reports := make([]ReportResponse, 0, len(output.Reports))
	for _, report := range output.Reports {
		reports = append(reports, toReportResponse(report))
	}

	return ListReportsResponse{
		Reports: reports,
		Total:   output.Total,
		Limit:   output.Limit,
		Offset:  output.Offset,
	}
}

func toUpdateResponse(output application.UpdateOutput) UpdateReportResponse {
	return UpdateReportResponse{Report: toReportResponse(output.Report)}
}

func toSubmitResponse(output application.SubmitOutput) SubmitReportResponse {
	return SubmitReportResponse{Report: toReportResponse(output.Report)}
}

func toStartReviewResponse(output application.StartReviewOutput) StartReviewReportResponse {
	return StartReviewReportResponse{Report: toReportResponse(output.Report)}
}

func toApproveResponse(output application.ApproveOutput) ApproveReportResponse {
	return ApproveReportResponse{Report: toReportResponse(output.Report)}
}

func toRejectResponse(output application.RejectOutput) RejectReportResponse {
	return RejectReportResponse{Report: toReportResponse(output.Report)}
}
