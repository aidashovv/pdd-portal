package violations

import (
	"fmt"
	"net/url"
	"strings"

	coreerrors "pdd-service/internal/core/errors"
)

type Video struct {
	Source      Source
	URL         string
	ObjectKey   string
	ContentType string
	Size        int64
}

func NewExternalVideo(rawURL string) (Video, error) {
	video := Video{
		Source: SourceExternalURL,
		URL:    strings.TrimSpace(rawURL),
	}

	if err := video.Validate(); err != nil {
		return Video{}, err
	}

	return video, nil
}

func NewS3Video(objectKey string, rawURL string, contentType string, size int64) (Video, error) {
	video := Video{
		Source:      SourceS3Upload,
		URL:         strings.TrimSpace(rawURL),
		ObjectKey:   strings.TrimSpace(objectKey),
		ContentType: strings.TrimSpace(contentType),
		Size:        size,
	}

	if err := video.Validate(); err != nil {
		return Video{}, err
	}

	return video, nil
}

func (v Video) Validate() error {
	if !v.Source.IsValid() {
		return fmt.Errorf("invalid video source: %w", coreerrors.ErrInvalidDomainValue)
	}
	if v.Size < 0 {
		return fmt.Errorf("video size cannot be negative: %w", coreerrors.ErrInvalidDomainValue)
	}

	switch v.Source {
	case SourceExternalURL:
		if strings.TrimSpace(v.URL) == "" {
			return fmt.Errorf("external video url is required: %w", coreerrors.ErrInvalidDomainValue)
		}
		if !isValidURL(v.URL) {
			return fmt.Errorf("external video url is invalid: %w", coreerrors.ErrInvalidDomainValue)
		}
	case SourceS3Upload:
		if strings.TrimSpace(v.ObjectKey) == "" {
			return fmt.Errorf("s3 video object key is required: %w", coreerrors.ErrInvalidDomainValue)
		}
		if v.ContentType != "" && !isAllowedVideoContentType(v.ContentType) {
			return fmt.Errorf("video content type is invalid: %w", coreerrors.ErrInvalidDomainValue)
		}
	}

	return nil
}

func isValidURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func isAllowedVideoContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "video/mp4", "video/webm", "video/quicktime":
		return true
	default:
		return false
	}
}
