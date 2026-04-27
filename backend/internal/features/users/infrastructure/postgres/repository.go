package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	pgpool "pdd-service/internal/core/repository/postgres/pool"
	"pdd-service/internal/features/users/application"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	pool *pgpool.ConnectionPool
}

func NewRepository(pool *pgpool.ConnectionPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (users.User, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	return r.scanUser(r.pool.Pool().QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id))
}

func (r *Repository) ListUsers(ctx context.Context, filter application.ListUsersFilter) ([]users.User, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, len(args)-1, len(args))

	rows, err := r.pool.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]users.User, 0, filter.Limit)
	for rows.Next() {
		user, err := scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users rows: %w", err)
	}

	return result, nil
}

func (r *Repository) CountUsers(ctx context.Context, filter application.ListUsersFilter) (int64, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	whereSQL, args := buildListWhere(filter)
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users
		%s
	`, whereSQL)

	var total int64
	if err := r.pool.Pool().QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return total, nil
}

func (r *Repository) UpdateUserRole(ctx context.Context, id uuid.UUID, role users.Role) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	tag, err := r.pool.Pool().Exec(ctx, `
		UPDATE users
		SET role = $2, updated_at = NOW()
		WHERE id = $1
	`, id, int16(role))
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return coreerrors.ErrUserNotFound
	}

	return nil
}

func (r *Repository) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.QueryTimeout())
}

func (r *Repository) scanUser(row pgx.Row) (users.User, error) {
	user, err := scanUserRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.User{}, coreerrors.ErrUserNotFound
		}
		return users.User{}, err
	}

	return user, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUserRow(row userScanner) (users.User, error) {
	var dto UserDTO
	err := row.Scan(&dto.ID, &dto.Email, &dto.PasswordHash, &dto.FullName, &dto.Role, &dto.CreatedAt, &dto.UpdatedAt)
	if err != nil {
		return users.User{}, fmt.Errorf("scan user: %w", err)
	}

	return dto.toDomain()
}

func buildListWhere(filter application.ListUsersFilter) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filter.Role != nil {
		args = append(args, int16(*filter.Role))
		conditions = append(conditions, fmt.Sprintf("role = $%d", len(args)))
	}

	search := strings.TrimSpace(filter.Search)
	if search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR full_name ILIKE $%d)", len(args), len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
