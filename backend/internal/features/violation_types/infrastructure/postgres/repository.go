package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"
	pgpool "pdd-service/internal/core/repository/postgres/pool"
	"pdd-service/internal/features/violation_types/application"

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

func (r *Repository) CreateViolationType(ctx context.Context, violationType violations.ViolationType) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := violationTypeToDTO(violationType)
	_, err := r.pool.Pool().Exec(ctx, `
		INSERT INTO violation_types (
			id, code, title, description, base_fine_amount, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, dto.ID, dto.Code, dto.Title, dto.Description, nullableString(dto.BaseFineAmount), dto.IsActive, dto.CreatedAt, dto.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %w", coreerrors.ErrViolationTypeAlreadyExists, coreerrors.ErrInvalidRequest)
		}
		return fmt.Errorf("create violation type: %w", err)
	}

	return nil
}

func (r *Repository) GetViolationTypeByID(ctx context.Context, id uuid.UUID) (violations.ViolationType, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	return r.scanViolationType(r.pool.Pool().QueryRow(ctx, `
		SELECT id, code, title, COALESCE(description, ''), COALESCE(base_fine_amount::text, ''), is_active, created_at, updated_at
		FROM violation_types
		WHERE id = $1
	`, id))
}

func (r *Repository) GetViolationTypeByCode(ctx context.Context, code string) (violations.ViolationType, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	return r.scanViolationType(r.pool.Pool().QueryRow(ctx, `
		SELECT id, code, title, COALESCE(description, ''), COALESCE(base_fine_amount::text, ''), is_active, created_at, updated_at
		FROM violation_types
		WHERE code = $1
	`, code))
}

func (r *Repository) ListViolationTypes(ctx context.Context, filter application.ListViolationTypesFilter) ([]violations.ViolationType, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`
		SELECT id, code, title, COALESCE(description, ''), COALESCE(base_fine_amount::text, ''), is_active, created_at, updated_at
		FROM violation_types
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, len(args)-1, len(args))

	rows, err := r.pool.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list violation types: %w", err)
	}
	defer rows.Close()

	result := make([]violations.ViolationType, 0, filter.Limit)
	for rows.Next() {
		violationType, err := scanViolationTypeRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, violationType)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate violation type rows: %w", err)
	}

	return result, nil
}

func (r *Repository) CountViolationTypes(ctx context.Context, filter application.ListViolationTypesFilter) (int64, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM violation_types
		%s
	`, whereSQL)

	var total int64
	if err := r.pool.Pool().QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count violation types: %w", err)
	}

	return total, nil
}

func (r *Repository) UpdateViolationType(ctx context.Context, violationType violations.ViolationType) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := violationTypeToDTO(violationType)
	tag, err := r.pool.Pool().Exec(ctx, `
		UPDATE violation_types
		SET title = $2,
			description = $3,
			base_fine_amount = $4,
			is_active = $5,
			updated_at = $6
		WHERE id = $1
	`, dto.ID, dto.Title, dto.Description, nullableString(dto.BaseFineAmount), dto.IsActive, dto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update violation type: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return notFoundError()
	}

	return nil
}

func (r *Repository) DeleteViolationType(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	tag, err := r.pool.Pool().Exec(ctx, `
		DELETE FROM violation_types
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete violation type: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return notFoundError()
	}

	return nil
}

func (r *Repository) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.QueryTimeout())
}

func (r *Repository) scanViolationType(row pgx.Row) (violations.ViolationType, error) {
	violationType, err := scanViolationTypeRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return violations.ViolationType{}, notFoundError()
		}
		return violations.ViolationType{}, err
	}

	return violationType, nil
}

type violationTypeScanner interface {
	Scan(dest ...any) error
}

func scanViolationTypeRow(row violationTypeScanner) (violations.ViolationType, error) {
	var dto ViolationTypeDTO
	err := row.Scan(
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
		return violations.ViolationType{}, fmt.Errorf("scan violation type: %w", err)
	}

	return dto.toDomain(), nil
}

func buildListWhere(filter application.ListViolationTypesFilter) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filter.OnlyActive != nil {
		args = append(args, *filter.OnlyActive)
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(args)))
	}

	search := strings.TrimSpace(filter.Search)
	if search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(code ILIKE $%d OR title ILIKE $%d)", len(args), len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func notFoundError() error {
	return fmt.Errorf("%w: %w", coreerrors.ErrViolationTypeNotFound, coreerrors.ErrNotFound)
}
