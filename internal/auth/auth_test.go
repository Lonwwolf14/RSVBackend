package auth

import (
	"testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
	password := "testpassword123"

	// Test hashing
	hashed, err := HashPassword(password)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if hashed == "" {
		t.Error("expected non-empty hash")
	}

	// Test correct password
	err = CheckPassword(hashed, password)
	if err != nil {
		t.Errorf("expected no error for correct password, got %v", err)
	}

	// Test wrong password
	err = CheckPassword(hashed, "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}
