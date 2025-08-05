package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopsphere/auth-service/internal/auth"
	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users    map[string]*models.User
	sessions map[string]map[string]time.Time // userID -> tokenHash -> expiresAt
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:    make(map[string]*models.User),
		sessions: make(map[string]map[string]time.Time),
	}
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.Status != models.StatusDeleted {
			return user, nil
		}
	}
	return nil, utils.NewNotFoundError("User")
}

func (m *MockUserRepository) GetUserByID(id string) (*models.User, error) {
	user, exists := m.users[id]
	if !exists || user.Status == models.StatusDeleted {
		return nil, utils.NewNotFoundError("User")
	}
	return user, nil
}

func (m *MockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username && user.Status != models.StatusDeleted {
			return user, nil
		}
	}
	return nil, utils.NewNotFoundError("User")
}

func (m *MockUserRepository) UpdateUserLastLogin(userID string) error {
	user, exists := m.users[userID]
	if !exists {
		return utils.NewNotFoundError("User")
	}
	now := time.Now()
	user.UpdatedAt = now
	return nil
}

func (m *MockUserRepository) UpdateUserStatus(userID string, status models.UserStatus) error {
	user, exists := m.users[userID]
	if !exists {
		return utils.NewNotFoundError("User")
	}
	user.Status = status
	user.UpdatedAt = time.Now()
	return nil
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	for _, user := range m.users {
		if user.Email == email && user.Status != models.StatusDeleted {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) UsernameExists(username string) (bool, error) {
	for _, user := range m.users {
		if user.Username == username && user.Status != models.StatusDeleted {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) CreateUserSession(userID, tokenHash, deviceInfo, ipAddress string, expiresAt time.Time) error {
	if m.sessions[userID] == nil {
		m.sessions[userID] = make(map[string]time.Time)
	}
	m.sessions[userID][tokenHash] = expiresAt
	return nil
}

func (m *MockUserRepository) ValidateUserSession(userID, tokenHash string) (bool, error) {
	userSessions, exists := m.sessions[userID]
	if !exists {
		return false, nil
	}
	
	expiresAt, exists := userSessions[tokenHash]
	if !exists {
		return false, nil
	}
	
	return time.Now().Before(expiresAt), nil
}

func (m *MockUserRepository) DeleteUserSession(userID, tokenHash string) error {
	if userSessions, exists := m.sessions[userID]; exists {
		delete(userSessions, tokenHash)
	}
	return nil
}

func (m *MockUserRepository) DeleteExpiredSessions() error {
	now := time.Now()
	for _, userSessions := range m.sessions {
		for tokenHash, expiresAt := range userSessions {
			if now.After(expiresAt) {
				delete(userSessions, tokenHash)
			}
		}
	}
	return nil
}

func setupTestAuthService() (*AuthService, *MockUserRepository) {
	userRepo := NewMockUserRepository()
	passwordService := auth.NewPasswordService()
	jwtService := jwt.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)
	
	authService := NewAuthService(userRepo, jwtService, passwordService)
	return authService, userRepo
}

func TestAuthService_Login_Success(t *testing.T) {
	authService, userRepo := setupTestAuthService()
	ctx := context.Background()

	// Create test user
	password := "TestPassword123!"
	passwordService := auth.NewPasswordService()
	hashedPassword, err := passwordService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     "User",
		Role:         models.RoleCustomer,
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	userRepo.CreateUser(user)

	// Test login
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	response, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.User.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, response.User.ID)
	}

	if response.User.PasswordHash != "" {
		t.Error("Password hash should be removed from response")
	}

	if response.Tokens.AccessToken == "" {
		t.Error("Expected access token to be present")
	}

	if response.Tokens.RefreshToken == "" {
		t.Error("Expected refresh token to be present")
	}
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	authService, _ := setupTestAuthService()
	ctx := context.Background()

	loginReq := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "TestPassword123!",
	}

	_, err := authService.Login(ctx, loginReq)
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Errorf("Expected AppError, got %T", err)
	}

	if appErr.Code != utils.ErrAuthentication {
		t.Errorf("Expected authentication error, got %s", appErr.Code)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	authService, userRepo := setupTestAuthService()
	ctx := context.Background()

	// Create test user
	password := "TestPassword123!"
	passwordService := auth.NewPasswordService()
	hashedPassword, err := passwordService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     "User",
		Role:         models.RoleCustomer,
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	userRepo.CreateUser(user)

	// Test login with wrong password
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}

	_, err = authService.Login(ctx, loginReq)
	if err == nil {
		t.Error("Expected error for invalid password")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Errorf("Expected AppError, got %T", err)
	}

	if appErr.Code != utils.ErrAuthentication {
		t.Errorf("Expected authentication error, got %s", appErr.Code)
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	authService, userRepo := setupTestAuthService()
	ctx := context.Background()

	// Create test user with suspended status
	password := "TestPassword123!"
	passwordService := auth.NewPasswordService()
	hashedPassword, err := passwordService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     "User",
		Role:         models.RoleCustomer,
		Status:       models.StatusSuspended,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	userRepo.CreateUser(user)

	// Test login
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	_, err = authService.Login(ctx, loginReq)
	if err == nil {
		t.Error("Expected error for inactive user")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Errorf("Expected AppError, got %T", err)
	}

	if appErr.Code != utils.ErrAuthentication {
		t.Errorf("Expected authentication error, got %s", appErr.Code)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	authService, userRepo := setupTestAuthService()
	ctx := context.Background()

	// Create test user
	user := &models.User{
		ID:        "test-user-id",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleCustomer,
		Status:    models.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	userRepo.CreateUser(user)

	// Generate initial tokens
	jwtService := jwt.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)
	
	tokens, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Create session
	refreshClaims, err := jwtService.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour * 7)
	userRepo.CreateUserSession(user.ID, refreshClaims.TokenHash, "", "", expiresAt)

	// Test refresh
	refreshReq := &RefreshRequest{
		RefreshToken: tokens.RefreshToken,
	}

	newTokens, err := authService.RefreshToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if newTokens.AccessToken == "" {
		t.Error("Expected new access token to be present")
	}

	if newTokens.RefreshToken == "" {
		t.Error("Expected new refresh token to be present")
	}

	// Tokens should be different (though they might be the same if generated at the exact same time)
	// The important thing is that we got new valid tokens
	if newTokens.AccessToken == tokens.AccessToken {
		t.Log("Warning: New access token is the same as old one (this can happen if generated at the same time)")
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	authService, userRepo := setupTestAuthService()
	ctx := context.Background()

	// Create test user
	user := &models.User{
		ID:        "test-user-id",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleCustomer,
		Status:    models.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	userRepo.CreateUser(user)

	// Generate tokens
	jwtService := jwt.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)
	
	tokens, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Create session
	refreshClaims, err := jwtService.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour * 7)
	userRepo.CreateUserSession(user.ID, refreshClaims.TokenHash, "", "", expiresAt)

	// Test logout
	logoutReq := &LogoutRequest{
		RefreshToken: tokens.RefreshToken,
	}

	err = authService.Logout(ctx, logoutReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify session is deleted
	exists, err := userRepo.ValidateUserSession(user.ID, refreshClaims.TokenHash)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if exists {
		t.Error("Expected session to be deleted after logout")
	}
}