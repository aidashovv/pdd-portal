package violations

import "testing"

func TestViolationTypeActivation(t *testing.T) {
	violationType, err := NewViolationType("RED_LIGHT", "Red light", "Crossing red light", "5000")
	if err != nil {
		t.Fatalf("NewViolationType() error = %v", err)
	}
	if !violationType.IsAvailableForReports() {
		t.Fatal("new violation type should be active")
	}

	violationType.Deactivate()
	if violationType.IsAvailableForReports() {
		t.Fatal("deactivated violation type should not be available")
	}

	violationType.Activate()
	if !violationType.IsAvailableForReports() {
		t.Fatal("activated violation type should be available")
	}

	if err := violationType.Rename(""); err == nil {
		t.Fatal("Rename() expected error")
	}
	if err := violationType.Rename("Updated title"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
}
