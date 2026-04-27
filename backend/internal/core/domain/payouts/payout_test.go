package payouts

import (
	"testing"

	"github.com/google/uuid"
)

func TestPayoutStatusTransitions(t *testing.T) {
	payout, err := NewPayout(uuid.New(), uuid.New(), "1500")
	if err != nil {
		t.Fatalf("NewPayout() error = %v", err)
	}
	if payout.Status != StatusPending {
		t.Fatalf("Status = %v, want %v", payout.Status, StatusPending)
	}
	if !payout.CanBePaid() {
		t.Fatal("pending payout should be payable")
	}

	if err := payout.MarkFailed(); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	if payout.Status != StatusFailed {
		t.Fatalf("Status = %v, want %v", payout.Status, StatusFailed)
	}
	if !payout.CanBePaid() {
		t.Fatal("failed payout should be payable")
	}

	if err := payout.MarkPending(); err != nil {
		t.Fatalf("MarkPending() error = %v", err)
	}
	if err := payout.MarkPaid(); err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}
	if payout.Status != StatusPaid {
		t.Fatalf("Status = %v, want %v", payout.Status, StatusPaid)
	}
	if payout.CanBePaid() {
		t.Fatal("paid payout should not be payable")
	}
	if err := payout.MarkFailed(); err == nil {
		t.Fatal("MarkFailed() expected error from paid")
	}
}
