package sessions

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSessionActiveExpiredRevokedBehavior(t *testing.T) {
	now := time.Now().UTC()
	session, err := NewSession(uuid.New(), "refresh-hash", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if session.IsExpired(now) {
		t.Fatal("session should not be expired")
	}
	if !session.IsActive(now) {
		t.Fatal("session should be active")
	}
	if !session.IsExpired(now.Add(2 * time.Hour)) {
		t.Fatal("session should be expired")
	}
	if session.IsActive(now.Add(2 * time.Hour)) {
		t.Fatal("expired session should not be active")
	}

	if err := session.Revoke(now); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if !session.IsRevoked() {
		t.Fatal("session should be revoked")
	}
	if session.IsActive(now) {
		t.Fatal("revoked session should not be active")
	}
	if err := session.Revoke(now); err == nil {
		t.Fatal("Revoke() expected error for already revoked session")
	}
}
