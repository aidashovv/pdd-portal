package violations

import (
	"fmt"
	"strings"

	coreerrors "pdd-service/internal/core/errors"
)

type Source int16

const (
	sourceUnknown Source = iota
	SourceExternalURL
	SourceS3Upload
)

func (s Source) String() string {
	switch s {
	case SourceExternalURL:
		return "EXTERNAL_URL"
	case SourceS3Upload:
		return "S3_UPLOAD"
	default:
		return "UNKNOWN"
	}
}

func (s Source) IsValid() bool {
	return s == SourceExternalURL || s == SourceS3Upload
}

func ParseSource(raw string) (Source, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "EXTERNAL_URL":
		return SourceExternalURL, nil
	case "S3_UPLOAD":
		return SourceS3Upload, nil
	default:
		return SourceExternalURL, fmt.Errorf("parse video source %q: %w", raw, coreerrors.ErrInvalidDomainValue)
	}
}
