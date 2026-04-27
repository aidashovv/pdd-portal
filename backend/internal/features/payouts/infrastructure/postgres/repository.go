package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/payouts"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	pgpool "pdd-service/internal/core/repository/postgres/pool"
	"pdd-service/internal/features/payouts/application"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	pool *pgpool.ConnectionPool
}

func NewRepository(pool *pgpool.ConnectionPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreatePayout(ctx context.Context, payout payouts.Payout) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	dto := payoutToDTO(payout)
	_, err := r.pool.Pool().Exec(ctx, `INSERT INTO payouts (id, report_id, user_id, amount, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		dto.ID, dto.ReportID, dto.UserID, dto.Amount, dto.Status, dto.CreatedAt, dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create payout: %w", err)
	}
	return nil
}

func (r *Repository) GetPayoutByID(ctx context.Context, id uuid.UUID) (payouts.Payout, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	return r.scanPayout(r.pool.Pool().QueryRow(ctx, selectPayoutSQL()+` WHERE id = $1`, id))
}

func (r *Repository) ListPayouts(ctx context.Context, filter application.ListPayoutsFilter) ([]payouts.Payout, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	whereSQL, args := buildPayoutsWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`%s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, selectPayoutSQL(), whereSQL, len(args)-1, len(args))
	rows, err := r.pool.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()
	result := make([]payouts.Payout, 0, filter.Limit)
	for rows.Next() {
		payout, err := scanPayoutRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, payout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payout rows: %w", err)
	}
	return result, nil
}

func (r *Repository) CountPayouts(ctx context.Context, filter application.ListPayoutsFilter) (int64, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	whereSQL, args := buildPayoutsWhere(filter)
	var total int64
	if err := r.pool.Pool().QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM payouts %s`, whereSQL), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count payouts: %w", err)
	}
	return total, nil
}

func (r *Repository) ListPayoutsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]payouts.Payout, error) {
	return r.ListPayouts(ctx, application.ListPayoutsFilter{UserID: &userID, Limit: limit, Offset: offset})
}

func (r *Repository) UpdatePayout(ctx context.Context, payout payouts.Payout) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	dto := payoutToDTO(payout)
	tag, err := r.pool.Pool().Exec(ctx, `UPDATE payouts SET amount=$2, status=$3, updated_at=$4 WHERE id=$1`, dto.ID, dto.Amount, dto.Status, dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update payout: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return payoutNotFoundError()
	}
	return nil
}

func (r *Repository) CreateRule(ctx context.Context, rule payouts.Rule) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	dto := ruleToDTO(rule)
	_, err := r.pool.Pool().Exec(ctx, `INSERT INTO payout_rules (id, violation_type_id, percent, is_active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		dto.ID, dto.ViolationTypeID, dto.Percent, dto.IsActive, dto.CreatedAt, dto.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return coreerrors.ErrPayoutRuleAlreadyExists
		}
		return fmt.Errorf("create payout rule: %w", err)
	}
	return nil
}

func (r *Repository) GetRuleByID(ctx context.Context, id uuid.UUID) (payouts.Rule, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	return r.scanRule(r.pool.Pool().QueryRow(ctx, selectRuleSQL()+` WHERE id = $1`, id))
}

func (r *Repository) GetActiveRuleByViolationTypeID(ctx context.Context, violationTypeID uuid.UUID) (payouts.Rule, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	return r.scanRule(r.pool.Pool().QueryRow(ctx, selectRuleSQL()+` WHERE violation_type_id = $1 AND is_active = TRUE ORDER BY updated_at DESC LIMIT 1`, violationTypeID))
}

func (r *Repository) ListRules(ctx context.Context, filter application.ListRulesFilter) ([]payouts.Rule, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	whereSQL, args := buildRulesWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`%s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, selectRuleSQL(), whereSQL, len(args)-1, len(args))
	rows, err := r.pool.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payout rules: %w", err)
	}
	defer rows.Close()
	result := make([]payouts.Rule, 0, filter.Limit)
	for rows.Next() {
		rule, err := scanRuleRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payout rule rows: %w", err)
	}
	return result, nil
}

func (r *Repository) CountRules(ctx context.Context, filter application.ListRulesFilter) (int64, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	whereSQL, args := buildRulesWhere(filter)
	var total int64
	if err := r.pool.Pool().QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM payout_rules %s`, whereSQL), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count payout rules: %w", err)
	}
	return total, nil
}

func (r *Repository) UpdateRule(ctx context.Context, rule payouts.Rule) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	dto := ruleToDTO(rule)
	tag, err := r.pool.Pool().Exec(ctx, `UPDATE payout_rules SET percent=$2, is_active=$3, updated_at=$4 WHERE id=$1`, dto.ID, dto.Percent, dto.IsActive, dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update payout rule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ruleNotFoundError()
	}
	return nil
}

func (r *Repository) GetReportByID(ctx context.Context, id uuid.UUID) (violations.Report, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	var dto ReportDTO
	var videoURL, videoObjectKey, videoContentType, moderationComment *string
	var videoSize *int64
	var moderatorID *uuid.UUID
	err := r.pool.Pool().QueryRow(ctx, `
		SELECT id, user_id, violation_type_id, title, description, location, occurred_at, status, video_source,
			video_url, video_object_key, video_content_type, video_size, moderator_id, moderation_comment, created_at, updated_at
		FROM violation_reports WHERE id = $1
	`, id).Scan(&dto.ID, &dto.UserID, &dto.ViolationTypeID, &dto.Title, &dto.Description, &dto.Location, &dto.OccurredAt, &dto.Status, &dto.VideoSource,
		&videoURL, &videoObjectKey, &videoContentType, &videoSize, &moderatorID, &moderationComment, &dto.CreatedAt, &dto.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return violations.Report{}, fmt.Errorf("%w: %w", coreerrors.ErrReportNotFound, coreerrors.ErrNotFound)
		}
		return violations.Report{}, fmt.Errorf("get report by id: %w", err)
	}
	dto.VideoURL, dto.VideoObjectKey, dto.VideoContentType = derefString(videoURL), derefString(videoObjectKey), derefString(videoContentType)
	dto.VideoSize, dto.ModeratorID, dto.ModerationComment = derefInt64(videoSize), derefUUID(moderatorID), derefString(moderationComment)
	return dto.toDomain()
}

func (r *Repository) GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()
	var dto ViolationTypeDTO
	err := r.pool.Pool().QueryRow(ctx, `SELECT id, code, title, COALESCE(description,''), COALESCE(base_fine_amount::text,''), is_active, created_at, updated_at FROM violation_types WHERE id=$1`, id).
		Scan(&dto.ID, &dto.Code, &dto.Title, &dto.Description, &dto.BaseFineAmount, &dto.IsActive, &dto.CreatedAt, &dto.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return violations.ViolationType{}, fmt.Errorf("%w: %w", coreerrors.ErrViolationTypeNotFound, coreerrors.ErrNotFound)
		}
		return violations.ViolationType{}, fmt.Errorf("get violation type by id: %w", err)
	}
	return dto.toDomain(), nil
}

func (r *Repository) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.QueryTimeout())
}

func (r *Repository) scanPayout(row pgx.Row) (payouts.Payout, error) {
	payout, err := scanPayoutRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return payouts.Payout{}, payoutNotFoundError()
		}
		return payouts.Payout{}, err
	}
	return payout, nil
}

func (r *Repository) scanRule(row pgx.Row) (payouts.Rule, error) {
	rule, err := scanRuleRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return payouts.Rule{}, ruleNotFoundError()
		}
		return payouts.Rule{}, err
	}
	return rule, nil
}

type scanner interface{ Scan(dest ...any) error }

func scanPayoutRow(row scanner) (payouts.Payout, error) {
	var dto PayoutDTO
	if err := row.Scan(&dto.ID, &dto.ReportID, &dto.UserID, &dto.Amount, &dto.Status, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
		return payouts.Payout{}, fmt.Errorf("scan payout: %w", err)
	}
	return dto.toDomain()
}

func scanRuleRow(row scanner) (payouts.Rule, error) {
	var dto RuleDTO
	if err := row.Scan(&dto.ID, &dto.ViolationTypeID, &dto.Percent, &dto.IsActive, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
		return payouts.Rule{}, fmt.Errorf("scan payout rule: %w", err)
	}
	return dto.toDomain(), nil
}

func buildPayoutsWhere(filter application.ListPayoutsFilter) (string, []any) {
	conditions, args := make([]string, 0, 2), make([]any, 0, 2)
	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if filter.Status != nil {
		args = append(args, int16(*filter.Status))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	return joinWhere(conditions), args
}

func buildRulesWhere(filter application.ListRulesFilter) (string, []any) {
	conditions, args := make([]string, 0, 2), make([]any, 0, 2)
	if filter.ViolationTypeID != nil {
		args = append(args, *filter.ViolationTypeID)
		conditions = append(conditions, fmt.Sprintf("violation_type_id = $%d", len(args)))
	}
	if filter.OnlyActive != nil {
		args = append(args, *filter.OnlyActive)
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(args)))
	}
	return joinWhere(conditions), args
}

func joinWhere(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

func selectPayoutSQL() string {
	return `SELECT id, report_id, user_id, amount::text, status, created_at, updated_at FROM payouts`
}
func selectRuleSQL() string {
	return `SELECT id, violation_type_id, percent::text, is_active, created_at, updated_at FROM payout_rules`
}
func payoutNotFoundError() error {
	return fmt.Errorf("%w: %w", coreerrors.ErrPayoutNotFound, coreerrors.ErrNotFound)
}
func ruleNotFoundError() error {
	return fmt.Errorf("%w: %w", coreerrors.ErrPayoutRuleNotFound, coreerrors.ErrNotFound)
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
