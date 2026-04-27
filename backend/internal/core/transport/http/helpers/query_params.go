package helpers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	value, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse path param %q as uuid: %w", key, err)
	}

	return value, nil
}

func GetIntQueryParam(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse query param %q as int: %w", key, err)
	}

	return value, nil
}

func GetDateQueryParam(r *http.Request, key string) (time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return time.Time{}, nil
	}

	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse query param %q as date: %w", key, err)
	}

	return value, nil
}
