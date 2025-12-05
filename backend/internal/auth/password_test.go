package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "TestPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !CheckPasswordHash(password, hash) {
		t.Error("Password should match hash")
	}

	// Test incorrect password
	if CheckPasswordHash("WrongPassword", hash) {
		t.Error("Wrong password should not match hash")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"short", true},            // Too short
		{"12345678", true},         // No letters
		{"abcdefgh", true},         // No digits
		{"Password123", false},     // Valid
		{"Test1234", false},        // Valid
		{"MySecurePass999", false}, // Valid
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
