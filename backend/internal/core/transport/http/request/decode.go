package request

import (
	"encoding/json"
	"fmt"
	"net/http"

	coreerrors "pdd-service/internal/core/errors"
)

func DecodeAndValidate(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, coreerrors.ErrInvalidRequest)
	}

	if err := Validate(dest); err != nil {
		return fmt.Errorf("validate request: %v: %w", err, coreerrors.ErrInvalidRequest)
	}

	return nil
}
