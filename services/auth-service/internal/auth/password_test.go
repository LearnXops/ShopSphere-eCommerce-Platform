package auth

import (
	"strings"
	"testing"
)

func TestPasswordService_HashPassword(t *testing.T) {
	ps := NewPasswordService()

	password := "TestPassword123!"
	hash, err := ps.HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hash == "" {
		t.Error("Expected hash to be generated")
	}

	if hash == password {
		t.Error("Hash should not be the same as password")
	}
}

func TestPasswordService_VerifyPassword(t *testing.T) {
	ps := NewPasswordService()

	password := "TestPassword123!"
	hash, err := ps.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	err = ps.VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("Expected no error for correct password, got %v", err)
	}

	// Test incorrect password
	err = ps.VerifyPassword("WrongPassword123!", hash)
	if err == nil {
		t.Error("Expected error for incorrect password")
	}
}

func TestPasswordService_ValidatePassword(t *testing.T) {
	ps := NewPasswordService()

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid password",
			password:    "TestPassword123!",
			expectError: false,
		},
		{
			name:        "Too short",
			password:    "Test1!",
			expectError: true,
			errorMsg:    "Password must be at least 8 characters long",
		},
		{
			name:        "Too long",
			password:    strings.Repeat("a", 129) + "A1!",
			expectError: true,
			errorMsg:    "Password must be no more than 128 characters long",
		},
		{
			name:        "No uppercase",
			password:    "testpassword123!",
			expectError: true,
			errorMsg:    "Password must contain at least one uppercase letter",
		},
		{
			name:        "No lowercase",
			password:    "TESTPASSWORD123!",
			expectError: true,
			errorMsg:    "Password must contain at least one lowercase letter",
		},
		{
			name:        "No digit",
			password:    "TestPassword!",
			expectError: true,
			errorMsg:    "Password must contain at least one digit",
		},
		{
			name:        "No special character",
			password:    "TestPassword123",
			expectError: true,
			errorMsg:    "Password must contain at least one special character",
		},
		{
			name:        "Minimum valid password",
			password:    "Test123!",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ps.ValidatePassword(tt.password)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPasswordService_GenerateResetToken(t *testing.T) {
	ps := NewPasswordService()

	token, err := ps.GenerateResetToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	if len(token) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("Expected token length to be 64, got %d", len(token))
	}

	// Generate another token to ensure they're different
	token2, err := ps.GenerateResetToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == token2 {
		t.Error("Expected different tokens to be generated")
	}
}

func TestPasswordService_HashToken(t *testing.T) {
	ps := NewPasswordService()

	token := "test-token-123"
	hash, err := ps.HashToken(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hash == "" {
		t.Error("Expected hash to be generated")
	}

	if hash == token {
		t.Error("Hash should not be the same as token")
	}
}

func TestPasswordService_VerifyToken(t *testing.T) {
	ps := NewPasswordService()

	token := "test-token-123"
	hash, err := ps.HashToken(token)
	if err != nil {
		t.Fatalf("Failed to hash token: %v", err)
	}

	// Test correct token
	err = ps.VerifyToken(token, hash)
	if err != nil {
		t.Errorf("Expected no error for correct token, got %v", err)
	}

	// Test incorrect token
	err = ps.VerifyToken("wrong-token", hash)
	if err == nil {
		t.Error("Expected error for incorrect token")
	}
}