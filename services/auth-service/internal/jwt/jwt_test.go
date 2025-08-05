package jwt

import (
	"testing"
	"time"

	"github.com/shopsphere/shared/models"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	jwtService := NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	user := &models.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Role:   models.RoleCustomer,
		Status: models.StatusActive,
	}

	tokens, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if tokens.RefreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	if tokens.TokenType != "Bearer" {
		t.Errorf("Expected token type to be 'Bearer', got %s", tokens.TokenType)
	}

	if tokens.ExpiresIn != int64((15 * time.Minute).Seconds()) {
		t.Errorf("Expected expires_in to be %d, got %d", int64((15*time.Minute).Seconds()), tokens.ExpiresIn)
	}
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService := NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	user := &models.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Role:   models.RoleCustomer,
		Status: models.StatusActive,
	}

	tokens, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := jwtService.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	if claims.Status != user.Status {
		t.Errorf("Expected status %s, got %s", user.Status, claims.Status)
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	jwtService := NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	user := &models.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Role:   models.RoleCustomer,
		Status: models.StatusActive,
	}

	tokens, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := jwtService.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.TokenHash == "" {
		t.Error("Expected token hash to be present")
	}
}

func TestJWTService_ValidateAccessToken_InvalidToken(t *testing.T) {
	jwtService := NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	_, err := jwtService.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestJWTService_ValidateAccessToken_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService(
		"test-access-secret-1",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	jwtService2 := NewJWTService(
		"test-access-secret-2",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)

	user := &models.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Role:   models.RoleCustomer,
		Status: models.StatusActive,
	}

	tokens, err := jwtService1.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	_, err = jwtService2.ValidateAccessToken(tokens.AccessToken)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret")
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedErr bool
		expected    string
	}{
		{
			name:        "Valid Bearer token",
			authHeader:  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedErr: false,
			expected:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:        "Empty header",
			authHeader:  "",
			expectedErr: true,
		},
		{
			name:        "Invalid format - no Bearer",
			authHeader:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedErr: true,
		},
		{
			name:        "Invalid format - wrong prefix",
			authHeader:  "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedErr: true,
		},
		{
			name:        "Invalid format - only Bearer",
			authHeader:  "Bearer",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.authHeader)
			
			if tt.expectedErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token != tt.expected {
					t.Errorf("Expected token %s, got %s", tt.expected, token)
				}
			}
		})
	}
}