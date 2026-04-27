package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pdd-service/internal/core/domain/sessions"
	"pdd-service/internal/core/domain/users"
	coreerrors "pdd-service/internal/core/errors"
	pgpool "pdd-service/internal/core/repository/postgres/pool"

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

func (r *Repository) CreateUser(ctx context.Context, user users.User) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := userToDTO(user)
	_, err := r.pool.Pool().Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, dto.ID, dto.Email, dto.PasswordHash, dto.FullName, dto.Role, dto.CreatedAt, dto.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return coreerrors.ErrEmailAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
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

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (users.User, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	return r.scanUser(r.pool.Pool().QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email))
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	var exists bool
	if err := r.pool.Pool().QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}

	return exists, nil
}

func (r *Repository) CreateSession(ctx context.Context, session sessions.Session) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	dto := sessionToDTO(session)
	_, err := r.pool.Pool().Exec(ctx, `
		INSERT INTO user_sessions (id, user_id, refresh_token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, dto.ID, dto.UserID, dto.RefreshTokenHash, dto.ExpiresAt, dto.RevokedAt, dto.CreatedAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *Repository) GetSessionByRefreshTokenHash(ctx context.Context, hash string) (sessions.Session, error) {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	var dto SessionDTO
	err := r.pool.Pool().QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
	`, hash).Scan(&dto.ID, &dto.UserID, &dto.RefreshTokenHash, &dto.ExpiresAt, &dto.RevokedAt, &dto.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessions.Session{}, coreerrors.ErrSessionNotFound
		}
		return sessions.Session{}, fmt.Errorf("get session by refresh token hash: %w", err)
	}

	return dto.toDomain(), nil
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	tag, err := r.pool.Pool().Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = $2
		WHERE id = $1
	`, sessionID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return coreerrors.ErrSessionNotFound
	}

	return nil
}

func (r *Repository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := r.queryContext(ctx)
	defer cancel()

	_, err := r.pool.Pool().Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}

	return nil
}

func (r *Repository) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.QueryTimeout())
}

func (r *Repository) scanUser(row pgx.Row) (users.User, error) {
	var dto UserDTO
	err := row.Scan(&dto.ID, &dto.Email, &dto.PasswordHash, &dto.FullName, &dto.Role, &dto.CreatedAt, &dto.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.User{}, coreerrors.ErrUserNotFound
		}
		return users.User{}, fmt.Errorf("scan user: %w", err)
	}

	return dto.toDomain()
}
