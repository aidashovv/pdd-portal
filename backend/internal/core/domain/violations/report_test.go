package violations

import (
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"

	"github.com/google/uuid"
)

func newTestReport(t *testing.T) *Report {
	t.Helper()

	video, err := NewExternalVideo("https://example.com/video.mp4")
	if err != nil {
		t.Fatalf("NewExternalVideo() error = %v", err)
	}

	report, err := NewReport(
		uuid.New(),
		uuid.New(),
		"Red light",
		"Driver crossed red light",
		"Nevsky prospect",
		time.Now().UTC(),
		video,
	)
	if err != nil {
		t.Fatalf("NewReport() error = %v", err)
	}

	return report
}

func TestReportStatusTransitions(t *testing.T) {
	report := newTestReport(t)
	moderatorID := uuid.New()

	if !report.IsDraft() {
		t.Fatal("new report should be draft")
	}
	if err := report.Submit(); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if !report.IsSubmitted() {
		t.Fatal("report should be submitted")
	}
	if err := report.StartReview(moderatorID); err != nil {
		t.Fatalf("StartReview() error = %v", err)
	}
	if !report.IsInReview() {
		t.Fatal("report should be in review")
	}
	if err := report.Approve(moderatorID, "valid"); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if !report.IsApproved() {
		t.Fatal("report should be approved")
	}
	if err := report.MarkPaid(); err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}
	if report.Status != StatusPaid {
		t.Fatalf("Status = %v, want %v", report.Status, StatusPaid)
	}
}

func TestReportRejectTransition(t *testing.T) {
	report := newTestReport(t)
	moderatorID := uuid.New()

	if err := report.Submit(); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if err := report.StartReview(moderatorID); err != nil {
		t.Fatalf("StartReview() error = %v", err)
	}
	if err := report.Reject(moderatorID, "not enough evidence"); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if report.Status != StatusRejected {
		t.Fatalf("Status = %v, want %v", report.Status, StatusRejected)
	}
}

func TestReportInvalidTransitions(t *testing.T) {
	report := newTestReport(t)
	if err := report.StartReview(uuid.New()); err == nil {
		t.Fatal("StartReview() expected error from draft")
	}
	if err := report.Approve(uuid.New(), "valid"); err == nil {
		t.Fatal("Approve() expected error from draft")
	}
	if err := report.MarkPaid(); err == nil {
		t.Fatal("MarkPaid() expected error from draft")
	}
	if err := report.Submit(); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if err := report.Submit(); err == nil {
		t.Fatal("Submit() expected error from submitted")
	}
}

func TestReportPermissions(t *testing.T) {
	report := newTestReport(t)
	ownerID := report.UserID
	otherUserID := uuid.New()

	if !report.CanBeEditedBy(ownerID, users.RoleUser) {
		t.Fatal("owner should edit draft report")
	}
	if !report.CanBeDeletedBy(ownerID, users.RoleUser) {
		t.Fatal("owner should delete draft report")
	}
	if report.CanBeEditedBy(otherUserID, users.RoleUser) {
		t.Fatal("other user should not edit report")
	}
	if !report.CanBeViewedBy(otherUserID, users.RoleModerator) {
		t.Fatal("moderator should view all reports")
	}
	if !report.CanBeViewedBy(otherUserID, users.RoleAdmin) {
		t.Fatal("admin should view all reports")
	}

	if err := report.Submit(); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if !report.CanBeEditedBy(ownerID, users.RoleUser) {
		t.Fatal("owner should edit submitted report")
	}
	if report.CanBeDeletedBy(ownerID, users.RoleUser) {
		t.Fatal("owner should not delete submitted report")
	}

	if err := report.StartReview(uuid.New()); err != nil {
		t.Fatalf("StartReview() error = %v", err)
	}
	if err := report.Approve(uuid.New(), "valid"); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if report.CanBeEditedBy(ownerID, users.RoleUser) {
		t.Fatal("owner should not edit approved report")
	}
}
