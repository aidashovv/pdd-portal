package postgres

import (
	"time"

	"github.com/google/uuid"
)

type ViolationTypeDTO struct {
	ID             uuid.UUID `db:"id"`
	Code           string    `db:"code"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	BaseFineAmount string    `db:"base_fine_amount"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
