package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/user-service/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(email, username, firstName, lastName, password string) (*models.User, string, error) {
	// Validate input
	if errs := utils.ValidateUserRegistration(email, username, firstName, lastName, password); errs.HasErrors() {
		return nil, "", utils.NewValidationError(errs.Error())
	}
	
	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, "", utils.NewInternalError("failed to check email existence", err)
	}
	if emailExists {
		return nil, "", utils.NewConflictError("email already exists")
	}
	
	// Check if username already exists
	usernameExists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return nil, "", utils.NewInternalError("failed to check username existence", err)
	}
	if usernameExists {
		return nil, "", utils.NewConflictError("username already exists")
	}
	
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", utils.NewInternalError("failed to hash password", err)
	}
	
	// Create user
	user := models.NewUser(email, username, firstName, lastName)
	user.PasswordHash = string(passwordHash)
	
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, "", utils.NewInternalError("failed to create user", err)
	}
	
	// Generate email verification token
	verificationToken, err := s.generateSecureToken()
	if err != nil {
		return nil, "", utils.NewInternalError("failed to generate verification token", err)
	}
	
	tokenHash := s.hashToken(verificationToken)
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	
	err = s.userRepo.CreateEmailVerificationToken(user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, "", utils.NewInternalError("failed to create email verification token", err)
	}
	
	// Clear password hash from response
	user.PasswordHash = ""
	
	return user, verificationToken, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Clear password hash from response
	user.PasswordHash = ""
	
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	
	// Clear password hash from response
	user.PasswordHash = ""
	
	return user, nil
}

// UpdateUser updates user profile information
func (s *UserService) UpdateUser(id string, updates map[string]interface{}) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Validate and apply updates
	validator := utils.NewValidator()
	
	if email, ok := updates["email"].(string); ok {
		validator.Required("email", email).Email("email", email)
		if !validator.HasErrors() {
			// Check if new email already exists (excluding current user)
			existingUser, err := s.userRepo.GetByEmail(email)
			if err == nil && existingUser.ID != id {
				return nil, utils.NewConflictError("email already exists")
			}
			user.Email = email
		}
	}
	
	if username, ok := updates["username"].(string); ok {
		validator.Username("username", username)
		if !validator.HasErrors() {
			// Check if new username already exists (excluding current user)
			existingUser, err := s.userRepo.GetByUsername(username)
			if err == nil && existingUser.ID != id {
				return nil, utils.NewConflictError("username already exists")
			}
			user.Username = username
		}
	}
	
	if firstName, ok := updates["first_name"].(string); ok {
		validator.Required("first_name", firstName).MaxLength("first_name", firstName, 50)
		if !validator.HasErrors() {
			user.FirstName = firstName
		}
	}
	
	if lastName, ok := updates["last_name"].(string); ok {
		validator.Required("last_name", lastName).MaxLength("last_name", lastName, 50)
		if !validator.HasErrors() {
			user.LastName = lastName
		}
	}
	
	if phone, ok := updates["phone"].(string); ok {
		if phone != "" {
			validator.Phone("phone", phone)
		}
		if !validator.HasErrors() {
			user.Phone = phone
		}
	}
	
	if validator.HasErrors() {
		return nil, utils.NewValidationError(validator.Errors().Error())
	}
	
	// Update user
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, utils.NewInternalError("failed to update user", err)
	}
	
	// Clear password hash from response
	user.PasswordHash = ""
	
	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(userID, currentPassword, newPassword string) error {
	// Get user with password hash
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	
	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return utils.NewValidationError("current password is incorrect")
	}
	
	// Validate new password
	validator := utils.NewValidator()
	validator.Password("password", newPassword)
	if validator.HasErrors() {
		return utils.NewValidationError(validator.Errors().Error())
	}
	
	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.NewInternalError("failed to hash password", err)
	}
	
	// Update password
	err = s.userRepo.UpdatePassword(userID, string(passwordHash))
	if err != nil {
		return utils.NewInternalError("failed to update password", err)
	}
	
	return nil
}

// RequestPasswordReset initiates a password reset process
func (s *UserService) RequestPasswordReset(email string) (string, error) {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return "", nil
	}
	
	// Generate reset token
	resetToken, err := s.generateSecureToken()
	if err != nil {
		return "", utils.NewInternalError("failed to generate reset token", err)
	}
	
	tokenHash := s.hashToken(resetToken)
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	
	err = s.userRepo.CreatePasswordResetToken(user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", utils.NewInternalError("failed to create password reset token", err)
	}
	
	return resetToken, nil
}

// ResetPassword resets a user's password using a reset token
func (s *UserService) ResetPassword(token, newPassword string) error {
	// Validate new password
	validator := utils.NewValidator()
	validator.Password("password", newPassword)
	if validator.HasErrors() {
		return utils.NewValidationError(validator.Errors().Error())
	}
	
	tokenHash := s.hashToken(token)
	
	// Get user ID from token
	userID, err := s.userRepo.GetPasswordResetToken(tokenHash)
	if err != nil {
		return err
	}
	
	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.NewInternalError("failed to hash password", err)
	}
	
	// Update password
	err = s.userRepo.UpdatePassword(userID, string(passwordHash))
	if err != nil {
		return utils.NewInternalError("failed to update password", err)
	}
	
	// Mark token as used
	err = s.userRepo.MarkPasswordResetTokenUsed(tokenHash)
	if err != nil {
		return utils.NewInternalError("failed to mark token as used", err)
	}
	
	return nil
}

// VerifyEmail verifies a user's email address
func (s *UserService) VerifyEmail(token string) error {
	tokenHash := s.hashToken(token)
	
	err := s.userRepo.VerifyEmail(tokenHash)
	if err != nil {
		return err
	}
	
	return nil
}

// UpdateUserStatus updates a user's status (admin operation)
func (s *UserService) UpdateUserStatus(userID string, status models.UserStatus) error {
	// Validate status
	validStatuses := []interface{}{
		models.StatusActive,
		models.StatusSuspended,
		models.StatusDeleted,
		models.StatusPending,
	}
	
	validator := utils.NewValidator()
	validator.OneOf("status", status, validStatuses)
	if validator.HasErrors() {
		return utils.NewValidationError(validator.Errors().Error())
	}
	
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	
	// Update status
	user.Status = status
	err = s.userRepo.Update(user)
	if err != nil {
		return utils.NewInternalError("failed to update user status", err)
	}
	
	return nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(userID string) error {
	err := s.userRepo.Delete(userID)
	if err != nil {
		return utils.NewInternalError("failed to delete user", err)
	}
	
	return nil
}

// ListUsers lists users with pagination and filtering (admin operation)
func (s *UserService) ListUsers(limit, offset int, status models.UserStatus, role models.UserRole) ([]*models.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}
	
	users, total, err := s.userRepo.List(limit, offset, status, role)
	if err != nil {
		return nil, 0, utils.NewInternalError("failed to list users", err)
	}
	
	// Clear password hashes from response
	for _, user := range users {
		user.PasswordHash = ""
	}
	
	return users, total, nil
}

// SearchUsers searches users by query (admin operation)
func (s *UserService) SearchUsers(query string, limit, offset int) ([]*models.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}
	
	if query == "" {
		return []*models.User{}, 0, nil
	}
	
	users, total, err := s.userRepo.Search(query, limit, offset)
	if err != nil {
		return nil, 0, utils.NewInternalError("failed to search users", err)
	}
	
	// Clear password hashes from response
	for _, user := range users {
		user.PasswordHash = ""
	}
	
	return users, total, nil
}

// generateSecureToken generates a cryptographically secure random token
func (s *UserService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken hashes a token using SHA-256
func (s *UserService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}