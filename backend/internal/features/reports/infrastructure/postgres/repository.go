package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	pgpool "pdd-service/internal/core/repository/postgres/pool"
	"pdd-service/internal/features/reports/application"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	pool *pgpool.ConnectionPool
}

func NewRepository(pool *pgpool.ConnectionPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateReport(ctx context.Context, report violations.Report) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := reportToDTO(report)
	_, err := r.pool.Pool().Exec(ctx, `
		INSERT INTO violation_reports (
			id, user_id, violation_type_id, title, description, location, occurred_at,
			status, video_source, video_url, video_object_key, video_content_type,
			video_size, moderator_id, moderation_comment, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, dto.ID, dto.UserID, dto.ViolationTypeID, dto.Title, dto.Description, dto.Location, dto.OccurredAt,
		dto.Status, dto.VideoSource, nullableString(dto.VideoURL), nullableString(dto.VideoObjectKey),
		nullableString(dto.VideoContentType), nullableInt64(dto.VideoSize), nullableUUID(dto.ModeratorID),
		nullableString(dto.ModerationComment), dto.CreatedAt, dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}

	return nil
}

func (r *Repository) GetReportByID(ctx context.Context, id uuid.UUID) (violations.Report, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	return r.scanReport(r.pool.Pool().QueryRow(ctx, selectReportSQL()+` WHERE id = $1`, id))
}

func (r *Repository) UpdateReport(ctx context.Context, report violations.Report) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := reportToDTO(report)
	tag, err := r.pool.Pool().Exec(ctx, `
		UPDATE violation_reports
		SET violation_type_id = $2,
			title = $3,
			description = $4,
			location = $5,
			occurred_at = $6,
			status = $7,
			video_source = $8,
			video_url = $9,
			video_object_key = $10,
			video_content_type = $11,
			video_size = $12,
			moderator_id = $13,
			moderation_comment = $14,
			updated_at = $15
		WHERE id = $1
	`, dto.ID, dto.ViolationTypeID, dto.Title, dto.Description, dto.Location, dto.OccurredAt,
		dto.Status, dto.VideoSource, nullableString(dto.VideoURL), nullableString(dto.VideoObjectKey),
		nullableString(dto.VideoContentType), nullableInt64(dto.VideoSize), nullableUUID(dto.ModeratorID),
		nullableString(dto.ModerationComment), dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return reportNotFoundError()
	}

	return nil
}

func (r *Repository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	tag, err := r.pool.Pool().Exec(ctx, `DELETE FROM violation_reports WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return reportNotFoundError()
	}

	return nil
}

func (r *Repository) ListReports(ctx context.Context, filter application.ListReportsFilter) ([]violations.Report, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`
		%s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, selectReportSQL(), whereSQL, len(args)-1, len(args))

	rows, err := r.pool.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	reports := make([]violations.Report, 0, filter.Limit)
	for rows.Next() {
		report, err := scanReportRow(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports rows: %w", err)
	}

	return reports, nil
}

func (r *Repository) CountReports(ctx context.Context, filter application.ListReportsFilter) (int64, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM violation_reports %s`, whereSQL)

	var total int64
	if err := r.pool.Pool().QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count reports: %w", err)
	}

	return total, nil
}

func (r *Repository) ListReportsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]violations.Report, error) {
	return r.ListReports(ctx, application.ListReportsFilter{
		UserID: &userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *Repository) GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	var dto ViolationTypeDTO
	err := r.pool.Pool().QueryRow(ctx, `
		SELECT id, code, title, COALESCE(description, ''), COALESCE(base_fine_amount::text, ''), is_active, created_at, updated_at
		FROM violation_types
		WHERE id = $1
	`, id).Scan(
		&dto.ID,
		&dto.Code,
		&dto.Title,
		&dto.Description,
		&dto.BaseFineAmount,
		&dto.IsActive,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return violations.ViolationType{}, violationTypeNotFoundError()
		}
		return violations.ViolationType{}, fmt.Errorf("get violation type by id: %w", err)
	}

	return dto.toDomain(), nil
}

func (r *Repository) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.QueryTimeout())
}

func (r *Repository) scanReport(row pgx.Row) (violations.Report, error) {
	report, err := scanReportRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return violations.Report{}, reportNotFoundError()
		}
		return violations.Report{}, err
	}

	return report, nil
}

type reportScanner interface {
	Scan(dest ...any) error
}

func scanReportRow(row reportScanner) (violations.Report, error) {
	var dto ReportDTO
	var videoURL *string
	var videoObjectKey *string
	var videoContentType *string
	var videoSize *int64
	var moderatorID *uuid.UUID
	var moderationComment *string

	err := row.Scan(
		&dto.ID,
		&dto.UserID,
		&dto.ViolationTypeID,
		&dto.Title,
		&dto.Description,
		&dto.Location,
		&dto.OccurredAt,
		&dto.Status,
		&dto.VideoSource,
		&videoURL,
		&videoObjectKey,
		&videoContentType,
		&videoSize,
		&moderatorID,
		&moderationComment,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		return violations.Report{}, fmt.Errorf("scan report: %w", err)
	}

	dto.VideoURL = derefString(videoURL)
	dto.VideoObjectKey = derefString(videoObjectKey)
	dto.VideoContentType = derefString(videoContentType)
	dto.VideoSize = derefInt64(videoSize)
	dto.ModeratorID = derefUUID(moderatorID)
	dto.ModerationComment = derefString(moderationComment)

	return dto.toDomain()
}

func buildListWhere(filter application.ListReportsFilter) (string, []any) {
	conditions := make([]string, 0, 5)
	args := make([]any, 0, 5)

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if filter.Status != nil {
		args = append(args, int16(*filter.Status))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if filter.ViolationTypeID != nil {
		args = append(args, *filter.ViolationTypeID)
		conditions = append(conditions, fmt.Sprintf("violation_type_id = $%d", len(args)))
	}
	if filter.CreatedFrom != nil {
		args = append(args, *filter.CreatedFrom)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if filter.CreatedTo != nil {
		args = append(args, *filter.CreatedTo)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func selectReportSQL() string {
	return `
		SELECT id, user_id, violation_type_id, title, description, location, occurred_at,
			status, video_source, video_url, video_object_key, video_content_type,
			video_size, moderator_id, moderation_comment, created_at, updated_at
		FROM violation_reports
	`
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableInt64(value int64) any {
	if value == 0 {
		return nil
	}

	return value
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}

	return value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func derefInt64(value *int64) int64 {
	if value == nil {
		return 0
	}

	return *value
}

func derefUUID(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}

	return *value
}

func reportNotFoundError() error {
	return fmt.Errorf("%w: %w", coreerrors.ErrReportNotFound, coreerrors.ErrNotFound)
}

func violationTypeNotFoundError() error {
	return fmt.Errorf("%w: %w", coreerrors.ErrViolationTypeNotFound, coreerrors.ErrNotFound)
}
