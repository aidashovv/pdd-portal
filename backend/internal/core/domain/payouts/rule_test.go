package payouts

import (
	"testing"

	"github.com/google/uuid"
)

func TestRulePercentValidation(t *testing.T) {
	if _, err := NewRule(uuid.New(), "10"); err != nil {
		t.Fatalf("NewRule() error = %v", err)
	}
	if _, err := NewRule(uuid.New(), "0"); err == nil {
		t.Fatal("NewRule() expected error for zero percent")
	}
	if _, err := NewRule(uuid.New(), "101"); err == nil {
		t.Fatal("NewRule() expected error for percent above 100")
	}
	if _, err := NewRule(uuid.New(), "abc"); err == nil {
		t.Fatal("NewRule() expected error for non-numeric percent")
	}
}

func TestRuleActivation(t *testing.T) {
	rule, err := NewRule(uuid.New(), "25")
	if err != nil {
		t.Fatalf("NewRule() error = %v", err)
	}

	rule.Deactivate()
	if rule.IsActive {
		t.Fatal("rule should be inactive")
	}

	rule.Activate()
	if !rule.IsActive {
		t.Fatal("rule should be active")
	}
}
