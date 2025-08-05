package service

import (
	"testing"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/user-service/internal/repository"
)

// MockUserRepository is a mock implementation of the user repository
type MockUserRepository struct {
	users                map[string]*models.User
	emailExists          map[string]bool
	usernameExists       map[string]bool
	passwordResetTokens  map[string]string
	emailVerifyTokens    map[string]string
}

// Ensure MockUserRepository implements the interface
var _ repository.UserRepositoryInterface = (*MockUserRepository)(nil)

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:               make(map[string]*models.User),
		emailExists:         make(map[string]bool),
		usernameExists:      make(map[string]bool),
		passwordResetTokens: make(map[string]string),
		emailVerifyTokens:   make(map[string]string),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	m.users[user.ID] = user
	m.emailExists[user.Email] = true
	m.usernameExists[user.Username] = true
	return nil
}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, utils.NewNotFoundError("user")
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, utils.NewNotFoundError("user")
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, utils.NewNotFoundError("user")
}

func (m *MockUserRepository) Update(user *models.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return utils.NewNotFoundError("user")
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdatePassword(userID, passwordHash string) error {
	if user, exists := m.users[userID]; exists {
		user.PasswordHash = passwordHash
		user.UpdatedAt = time.Now()
		return nil
	}
	return utils.NewNotFoundError("user")
}

func (m *MockUserRepository) Delete(id string) error {
	if user, exists := m.users[id]; exists {
		user.Status = models.StatusDeleted
		return nil
	}
	return utils.NewNotFoundError("user")
}

func (m *MockUserRepository) List(limit, offset int, status models.UserStatus, role models.UserRole) ([]*models.User, int, error) {
	var users []*models.User
	for _, user := range m.users {
		if (status == "" || user.Status == status) && (role == "" || user.Role == role) {
			users = append(users, user)
		}
	}
	return users, len(users), nil
}

func (m *MockUserRepository) Search(query string, limit, offset int) ([]*models.User, int, error) {
	var users []*models.User
	for _, user := range m.users {
		if user.Status != models.StatusDeleted {
			users = append(users, user)
		}
	}
	return users, len(users), nil
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	return m.emailExists[email], nil
}

func (m *MockUserRepository) UsernameExists(username string) (bool, error) {
	return m.usernameExists[username], nil
}

func (m *MockUserRepository) CreatePasswordResetToken(userID, tokenHash string, expiresAt time.Time) error {
	m.passwordResetTokens[tokenHash] = userID
	return nil
}

func (m *MockUserRepository) GetPasswordResetToken(tokenHash string) (string, error) {
	if userID, exists := m.passwordResetTokens[tokenHash]; exists {
		return userID, nil
	}
	return "", utils.NewNotFoundError("password reset token")
}

func (m *MockUserRepository) MarkPasswordResetTokenUsed(tokenHash string) error {
	delete(m.passwordResetTokens, tokenHash)
	return nil
}

func (m *MockUserRepository) CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time) error {
	m.emailVerifyTokens[tokenHash] = userID
	return nil
}

func (m *MockUserRepository) VerifyEmail(tokenHash string) error {
	if userID, exists := m.emailVerifyTokens[tokenHash]; exists {
		if user, exists := m.users[userID]; exists {
			user.Status = models.StatusActive
			delete(m.emailVerifyTokens, tokenHash)
			return nil
		}
	}
	return utils.NewNotFoundError("email verification token")
}

func TestUserService_RegisterUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	tests := []struct {
		name        string
		email       string
		username    string
		firstName   string
		lastName    string
		password    string
		expectError bool
	}{
		{
			name:        "Valid registration",
			email:       "test@example.com",
			username:    "testuser",
			firstName:   "Test",
			lastName:    "User",
			password:    "Password123!",
			expectError: false,
		},
		{
			name:        "Invalid email",
			email:       "invalid-email",
			username:    "testuser2",
			firstName:   "Test",
			lastName:    "User",
			password:    "Password123!",
			expectError: true,
		},
		{
			name:        "Weak password",
			email:       "test2@example.com",
			username:    "testuser3",
			firstName:   "Test",
			lastName:    "User",
			password:    "weak",
			expectError: true,
		},
		{
			name:        "Duplicate email",
			email:       "test@example.com", // Same as first test
			username:    "testuser4",
			firstName:   "Test",
			lastName:    "User",
			password:    "Password123!",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.RegisterUser(tt.email, tt.username, tt.firstName, tt.lastName, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if user == nil {
					t.Errorf("Expected user but got nil")
				}
				if token == "" {
					t.Errorf("Expected verification token but got empty string")
				}
				if user != nil && user.Status != models.StatusPending {
					t.Errorf("Expected user status to be pending, got %s", user.Status)
				}
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create a test user
	user := models.NewUser("test@example.com", "testuser", "Test", "User")
	user.Status = models.StatusActive
	mockRepo.Create(user)

	tests := []struct {
		name        string
		userID      string
		updates     map[string]interface{}
		expectError bool
	}{
		{
			name:   "Valid update",
			userID: user.ID,
			updates: map[string]interface{}{
				"first_name": "Updated",
				"last_name":  "Name",
				"phone":      "+1234567890",
			},
			expectError: false,
		},
		{
			name:   "Invalid email",
			userID: user.ID,
			updates: map[string]interface{}{
				"email": "invalid-email",
			},
			expectError: true,
		},
		{
			name:        "User not found",
			userID:      "nonexistent",
			updates:     map[string]interface{}{"first_name": "Test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser, err := service.UpdateUser(tt.userID, tt.updates)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if updatedUser == nil {
					t.Errorf("Expected updated user but got nil")
				}
				if updatedUser != nil {
					if firstName, ok := tt.updates["first_name"].(string); ok {
						if updatedUser.FirstName != firstName {
							t.Errorf("Expected first name %s, got %s", firstName, updatedUser.FirstName)
						}
					}
				}
			}
		})
	}
}

func TestUserService_PasswordReset(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create a test user
	user := models.NewUser("reset@example.com", "resetuser", "Reset", "User")
	user.Status = models.StatusActive
	user.PasswordHash = "$2a$10$example" // dummy hash
	mockRepo.Create(user)

	// Test password reset request
	token, err := service.RequestPasswordReset(user.Email)
	if err != nil {
		t.Errorf("Unexpected error requesting password reset: %v", err)
	}
	if token == "" {
		t.Errorf("Expected reset token but got empty string")
	}

	// Test password reset with token
	newPassword := "NewPassword123!"
	err = service.ResetPassword(token, newPassword)
	if err != nil {
		t.Errorf("Unexpected error resetting password: %v", err)
	}

	// Verify password was updated (in a real scenario, we'd check the hash)
	updatedUser, _ := mockRepo.GetByID(user.ID)
	if updatedUser.PasswordHash == "$2a$10$example" {
		t.Errorf("Password hash should have been updated")
	}
}

func TestUserService_EmailVerification(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Register a user (which creates verification token)
	user, token, err := service.RegisterUser("verify@example.com", "verifyuser", "Verify", "User", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Verify the user's status is pending
	if user.Status != models.StatusPending {
		t.Errorf("Expected user status to be pending, got %s", user.Status)
	}

	// Verify email
	err = service.VerifyEmail(token)
	if err != nil {
		t.Errorf("Unexpected error verifying email: %v", err)
	}

	// Check that user status is now active
	verifiedUser, _ := mockRepo.GetByID(user.ID)
	if verifiedUser.Status != models.StatusActive {
		t.Errorf("Expected user status to be active after verification, got %s", verifiedUser.Status)
	}
}

func TestUserService_UserStatusUpdate(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create a test user
	user := models.NewUser("status@example.com", "statususer", "Status", "User")
	user.Status = models.StatusActive
	mockRepo.Create(user)

	// Test status update
	err := service.UpdateUserStatus(user.ID, models.StatusSuspended)
	if err != nil {
		t.Errorf("Unexpected error updating user status: %v", err)
	}

	// Verify status was updated
	updatedUser, _ := mockRepo.GetByID(user.ID)
	if updatedUser.Status != models.StatusSuspended {
		t.Errorf("Expected user status to be suspended, got %s", updatedUser.Status)
	}

	// Test invalid status
	err = service.UpdateUserStatus(user.ID, "invalid_status")
	if err == nil {
		t.Errorf("Expected error for invalid status but got none")
	}
}

func TestUserService_ListUsers(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create test users
	users := []*models.User{
		models.NewUser("user1@example.com", "user1", "User", "One"),
		models.NewUser("user2@example.com", "user2", "User", "Two"),
		models.NewUser("user3@example.com", "user3", "User", "Three"),
	}

	for _, user := range users {
		user.Status = models.StatusActive
		mockRepo.Create(user)
	}

	// Test listing users
	listedUsers, total, err := service.ListUsers(10, 0, "", "")
	if err != nil {
		t.Errorf("Unexpected error listing users: %v", err)
	}

	if len(listedUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(listedUsers))
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	// Test filtering by status
	listedUsers, total, err = service.ListUsers(10, 0, models.StatusActive, "")
	if err != nil {
		t.Errorf("Unexpected error listing users with status filter: %v", err)
	}

	if len(listedUsers) != 3 {
		t.Errorf("Expected 3 active users, got %d", len(listedUsers))
	}
}