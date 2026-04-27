package application

import (
	"context"
	"errors"
	"fmt"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	reportsRepo        ReportsRepository
	violationTypesRepo ViolationTypesRepository
}

func NewService(reportsRepo ReportsRepository, violationTypesRepo ViolationTypesRepository) *Service {
	return &Service{
		reportsRepo:        reportsRepo,
		violationTypesRepo: violationTypesRepo,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	userID := input.UserID
	if userID == uuid.Nil {
		userID = input.CurrentUserID
	}
	if input.CurrentRole == users.RoleUser && userID != input.CurrentUserID {
		return CreateOutput{}, coreerrors.ErrForbidden
	}
	if !input.CurrentRole.IsValid() || input.CurrentUserID == uuid.Nil {
		return CreateOutput{}, coreerrors.ErrUnauthorized
	}
	if _, err := s.violationTypesRepo.GetViolationTypeByID(ctx, input.ViolationTypeID); err != nil {
		return CreateOutput{}, err
	}

	video, err := buildVideo(input.Video)
	if err != nil {
		return CreateOutput{}, err
	}

	report, err := violations.NewReport(
		userID,
		input.ViolationTypeID,
		input.Title,
		input.Description,
		input.Location,
		input.OccurredAt,
		video,
	)
	if err != nil {
		return CreateOutput{}, err
	}

	if err := s.reportsRepo.CreateReport(ctx, *report); err != nil {
		return CreateOutput{}, err
	}

	return CreateOutput{Report: toReportOutput(*report)}, nil
}

func (s *Service) GetByID(ctx context.Context, input GetByIDInput) (GetByIDOutput, error) {
	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return GetByIDOutput{}, err
	}
	if !report.CanBeViewedBy(input.CurrentUserID, input.CurrentRole) {
		return GetByIDOutput{}, coreerrors.ErrForbidden
	}

	return GetByIDOutput{Report: toReportOutput(report)}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	filter, err := normalizeListFilter(input)
	if err != nil {
		return ListOutput{}, err
	}

	total, err := s.reportsRepo.CountReports(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	reports, err := s.reportsRepo.ListReports(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{
		Reports: toReportOutputs(reports),
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (UpdateOutput, error) {
	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return UpdateOutput{}, err
	}
	if !report.CanBeEditedBy(input.CurrentUserID, input.CurrentRole) {
		return UpdateOutput{}, coreerrors.ErrForbidden
	}

	violationTypeID := report.ViolationTypeID
	if input.ViolationTypeID != nil {
		violationTypeID = *input.ViolationTypeID
	}
	if _, err := s.violationTypesRepo.GetViolationTypeByID(ctx, violationTypeID); err != nil {
		return UpdateOutput{}, err
	}

	title := report.Title
	if input.Title != nil {
		title = *input.Title
	}
	description := report.Description
	if input.Description != nil {
		description = *input.Description
	}
	location := report.Location
	if input.Location != nil {
		location = *input.Location
	}
	occurredAt := report.OccurredAt
	if input.OccurredAt != nil {
		occurredAt = *input.OccurredAt
	}
	video := report.Video
	if input.Video != nil {
		video, err = buildVideo(*input.Video)
		if err != nil {
			return UpdateOutput{}, err
		}
	}

	if err := report.UpdateDetails(violationTypeID, title, description, location, occurredAt, video); err != nil {
		return UpdateOutput{}, err
	}
	if err := s.reportsRepo.UpdateReport(ctx, report); err != nil {
		return UpdateOutput{}, err
	}

	return UpdateOutput{Report: toReportOutput(report)}, nil
}

func (s *Service) Delete(ctx context.Context, input DeleteInput) error {
	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if !report.CanBeDeletedBy(input.CurrentUserID, input.CurrentRole) {
		return coreerrors.ErrForbidden
	}

	return s.reportsRepo.DeleteReport(ctx, input.ID)
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (SubmitOutput, error) {
	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return SubmitOutput{}, err
	}
	if input.CurrentRole != users.RoleUser || !report.IsOwner(input.CurrentUserID) {
		return SubmitOutput{}, coreerrors.ErrForbidden
	}
	if err := report.Submit(); err != nil {
		return SubmitOutput{}, mapTransitionError(err)
	}
	if err := s.reportsRepo.UpdateReport(ctx, report); err != nil {
		return SubmitOutput{}, err
	}

	return SubmitOutput{Report: toReportOutput(report)}, nil
}

func (s *Service) StartReview(ctx context.Context, input StartReviewInput) (StartReviewOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return StartReviewOutput{}, coreerrors.ErrForbidden
	}

	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return StartReviewOutput{}, err
	}
	if err := report.StartReview(input.CurrentUserID); err != nil {
		return StartReviewOutput{}, mapTransitionError(err)
	}
	if err := s.reportsRepo.UpdateReport(ctx, report); err != nil {
		return StartReviewOutput{}, err
	}

	return StartReviewOutput{Report: toReportOutput(report)}, nil
}

func (s *Service) Approve(ctx context.Context, input ApproveInput) (ApproveOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return ApproveOutput{}, coreerrors.ErrForbidden
	}

	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return ApproveOutput{}, err
	}
	if err := report.Approve(input.CurrentUserID, input.Comment); err != nil {
		return ApproveOutput{}, mapTransitionError(err)
	}
	if err := s.reportsRepo.UpdateReport(ctx, report); err != nil {
		return ApproveOutput{}, err
	}

	return ApproveOutput{Report: toReportOutput(report)}, nil
}

func (s *Service) Reject(ctx context.Context, input RejectInput) (RejectOutput, error) {
	if !input.CurrentRole.CanModerateReports() {
		return RejectOutput{}, coreerrors.ErrForbidden
	}

	report, err := s.reportsRepo.GetReportByID(ctx, input.ID)
	if err != nil {
		return RejectOutput{}, err
	}
	if err := report.Reject(input.CurrentUserID, input.Comment); err != nil {
		return RejectOutput{}, mapTransitionError(err)
	}
	if err := s.reportsRepo.UpdateReport(ctx, report); err != nil {
		return RejectOutput{}, err
	}

	return RejectOutput{Report: toReportOutput(report)}, nil
}

func normalizeListFilter(input ListInput) (ListReportsFilter, error) {
	if !input.CurrentRole.IsValid() || input.CurrentUserID == uuid.Nil {
		return ListReportsFilter{}, coreerrors.ErrUnauthorized
	}
	if input.Offset < 0 {
		return ListReportsFilter{}, fmt.Errorf("offset must be non-negative: %w", coreerrors.ErrInvalidRequest)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	userID := input.UserID
	if input.CurrentRole == users.RoleUser {
		userID = &input.CurrentUserID
	}

	return ListReportsFilter{
		UserID:          userID,
		Status:          input.Status,
		ViolationTypeID: input.ViolationTypeID,
		CreatedFrom:     input.CreatedFrom,
		CreatedTo:       input.CreatedTo,
		Limit:           limit,
		Offset:          input.Offset,
	}, nil
}

func buildVideo(input VideoInput) (violations.Video, error) {
	switch input.Source {
	case violations.SourceExternalURL:
		return violations.NewExternalVideo(input.URL)
	case violations.SourceS3Upload:
		return violations.NewS3Video(input.ObjectKey, input.URL, input.ContentType, input.Size)
	default:
		return violations.Video{}, fmt.Errorf("report video is required: %w", coreerrors.ErrVideoRequired)
	}
}

func mapTransitionError(err error) error {
	if errors.Is(err, coreerrors.ErrInvalidTransition) {
		return fmt.Errorf("%w: %w: %w", err, coreerrors.ErrInvalidReportStatus, coreerrors.ErrInvalidRequest)
	}

	return err
}

func toReportOutput(report violations.Report) ReportOutput {
	return ReportOutput{
		ID:                report.ID,
		UserID:            report.UserID,
		ViolationTypeID:   report.ViolationTypeID,
		ModeratorID:       report.ModeratorID,
		Title:             report.Title,
		Description:       report.Description,
		Location:          report.Location,
		OccurredAt:        report.OccurredAt,
		Status:            report.Status,
		Video:             toVideoOutput(report.Video),
		ModerationComment: report.ModerationComment,
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
	}
}

func toReportOutputs(reports []violations.Report) []ReportOutput {
	outputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		outputs = append(outputs, toReportOutput(report))
	}

	return outputs
}

func toVideoOutput(video violations.Video) VideoOutput {
	return VideoOutput{
		Source:      video.Source,
		URL:         video.URL,
		ObjectKey:   video.ObjectKey,
		ContentType: video.ContentType,
		Size:        video.Size,
	}
}
