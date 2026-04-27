package violations

import "testing"

func TestVideoInvariants(t *testing.T) {
	if _, err := NewExternalVideo(""); err == nil {
		t.Fatal("NewExternalVideo() expected error for empty url")
	}

	if _, err := NewExternalVideo("https://example.com/video.mp4"); err != nil {
		t.Fatalf("NewExternalVideo() error = %v", err)
	}

	if _, err := NewS3Video("", "https://storage.example/video.mp4", "video/mp4", 1); err == nil {
		t.Fatal("NewS3Video() expected error for empty object key")
	}

	if _, err := NewS3Video("reports/user/video.mp4", "", "video/mp4", -1); err == nil {
		t.Fatal("NewS3Video() expected error for negative size")
	}

	if _, err := NewS3Video("reports/user/video.txt", "", "text/plain", 1); err == nil {
		t.Fatal("NewS3Video() expected error for invalid content type")
	}

	if _, err := NewS3Video("reports/user/video.webm", "", "video/webm", 1); err != nil {
		t.Fatalf("NewS3Video() error = %v", err)
	}

	if err := (Video{}).Validate(); err == nil {
		t.Fatal("Video{}.Validate() expected error")
	}
}
