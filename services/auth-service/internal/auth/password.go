package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"github.com/shopsphere/shared/utils"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 12
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum password length
	MaxPasswordLength = 128
)

// PasswordService handles password operations
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: DefaultCost,
	}
}

// HashPassword hashes a password using bcrypt
func (p *PasswordService) HashPassword(password string) (string, error) {
	if err := p.ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", utils.NewInternalError("Failed to hash password", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordService) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return utils.NewAppError(utils.ErrAuthentication, "Invalid password", nil)
		}
		return utils.NewInternalError("Failed to verify password", err)
	}
	return nil
}

// ValidatePassword validates password strength
func (p *PasswordService) ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return utils.NewValidationError(fmt.Sprintf("Password must be at least %d characters long", MinPasswordLength))
	}

	if len(password) > MaxPasswordLength {
		return utils.NewValidationError(fmt.Sprintf("Password must be no more than %d characters long", MaxPasswordLength))
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 32 && char <= 126: // Printable ASCII characters
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	if !hasUpper {
		return utils.NewValidationError("Password must contain at least one uppercase letter")
	}

	if !hasLower {
		return utils.NewValidationError("Password must contain at least one lowercase letter")
	}

	if !hasDigit {
		return utils.NewValidationError("Password must contain at least one digit")
	}

	if !hasSpecial {
		return utils.NewValidationError("Password must contain at least one special character")
	}

	return nil
}

// GenerateResetToken generates a secure password reset token
func (p *PasswordService) GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", utils.NewInternalError("Failed to generate reset token", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken hashes a token for secure storage
func (p *PasswordService) HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", utils.NewInternalError("Failed to hash token", err)
	}
	return string(hash), nil
}

// VerifyToken verifies a token against its hash
func (p *PasswordService) VerifyToken(token, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return utils.NewAppError(utils.ErrAuthentication, "Invalid token", nil)
		}
		return utils.NewInternalError("Failed to verify token", err)
	}
	return nil
}