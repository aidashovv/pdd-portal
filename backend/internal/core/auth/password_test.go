package auth

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "secret-password" {
		t.Fatal("HashPassword() returned plaintext password")
	}
	if !VerifyPassword(hash, "secret-password") {
		t.Fatal("VerifyPassword() = false, want true")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("VerifyPassword() = true for wrong password")
	}
}
