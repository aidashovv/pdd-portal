package users

import "testing"

func TestRoleValidationAndPermissions(t *testing.T) {
	tests := []struct {
		role                    Role
		wantString              string
		wantValid               bool
		wantModerateReports     bool
		wantManageViolationType bool
		wantCreateReport        bool
	}{
		{RoleUser, "USER", true, false, false, true},
		{RoleModerator, "MODERATOR", true, true, false, false},
		{RoleAdmin, "ADMIN", true, true, true, false},
		{Role(99), "UNKNOWN", false, false, false, false},
	}

	for _, tt := range tests {
		if got := tt.role.String(); got != tt.wantString {
			t.Fatalf("String() = %s, want %s", got, tt.wantString)
		}
		if got := tt.role.IsValid(); got != tt.wantValid {
			t.Fatalf("IsValid() = %v, want %v", got, tt.wantValid)
		}
		if got := tt.role.CanModerateReports(); got != tt.wantModerateReports {
			t.Fatalf("CanModerateReports() = %v, want %v", got, tt.wantModerateReports)
		}
		if got := tt.role.CanManageViolationTypes(); got != tt.wantManageViolationType {
			t.Fatalf("CanManageViolationTypes() = %v, want %v", got, tt.wantManageViolationType)
		}
		if got := tt.role.CanCreateReport(); got != tt.wantCreateReport {
			t.Fatalf("CanCreateReport() = %v, want %v", got, tt.wantCreateReport)
		}
	}
}

func TestParseRole(t *testing.T) {
	role, err := ParseRole("moderator")
	if err != nil {
		t.Fatalf("ParseRole() error = %v", err)
	}
	if role != RoleModerator {
		t.Fatalf("ParseRole() = %v, want %v", role, RoleModerator)
	}

	if _, err := ParseRole("invalid"); err == nil {
		t.Fatal("ParseRole() expected error")
	}
}

func TestNewUser(t *testing.T) {
	user, err := NewUser("user@example.com", "hash", "User Name", RoleUser)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	if user.ID.ID() == 0 {
		t.Fatal("NewUser() generated nil uuid")
	}
	if !user.CanCreateReport() {
		t.Fatal("user should be able to create report")
	}

	if _, err := NewUser("", "hash", "User Name", RoleUser); err == nil {
		t.Fatal("NewUser() expected error for empty email")
	}
}
