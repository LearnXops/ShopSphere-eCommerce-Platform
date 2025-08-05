package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopsphere/auth-service/internal/auth"
	"github.com/shopsphere/auth-service/internal/handlers"
	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/auth-service/internal/middleware"
	"github.com/shopsphere/auth-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// MockUserRepository for integration tests
type MockUserRepository struct {
	users    map[string]*models.User
	sessions map[string]map[string]time.Time
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
	user.UpdatedAt = time.Now()
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

func setupTestServer() (*httptest.Server, *MockUserRepository) {
	// Initialize services
	userRepo := NewMockUserRepository()
	passwordService := auth.NewPasswordService()
	jwtService := jwt.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		"test-issuer",
		15*time.Minute,
		7*24*time.Hour,
	)
	authService := service.NewAuthService(userRepo, jwtService, passwordService)

	// Initialize handlers and middleware
	authHandlers := handlers.NewAuthHandlers(authService)
	rbacMiddleware := middleware.NewRBACMiddleware(authService)

	// Create router
	router := mux.NewRouter()

	// Authentication endpoints
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", authHandlers.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", authHandlers.RefreshToken).Methods("POST")
	authRouter.HandleFunc("/logout", authHandlers.Logout).Methods("POST")
	authRouter.HandleFunc("/validate", authHandlers.ValidateToken).Methods("POST")
	
	// Protected endpoints
	protectedRouter := router.PathPrefix("/auth").Subrouter()
	protectedRouter.Use(rbacMiddleware.Authenticate)
	protectedRouter.Use(rbacMiddleware.RequireActiveUser)
	protectedRouter.HandleFunc("/me", authHandlers.Me).Methods("GET")

	server := httptest.NewServer(router)
	return server, userRepo
}

func TestIntegration_LoginFlow(t *testing.T) {
	server, userRepo := setupTestServer()
	defer server.Close()

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
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": password,
	}
	
	loginJSON, _ := json.Marshal(loginData)
	resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("Failed to make login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var loginResponse service.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResponse.User.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, loginResponse.User.ID)
	}

	if loginResponse.Tokens.AccessToken == "" {
		t.Error("Expected access token to be present")
	}

	// Test /me endpoint with access token
	req, _ := http.NewRequest("GET", server.URL+"/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.Tokens.AccessToken)
	
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make /me request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for /me endpoint, got %d", resp.StatusCode)
	}

	var meResponse models.User
	if err := json.NewDecoder(resp.Body).Decode(&meResponse); err != nil {
		t.Fatalf("Failed to decode /me response: %v", err)
	}

	if meResponse.ID != user.ID {
		t.Errorf("Expected user ID %s from /me endpoint, got %s", user.ID, meResponse.ID)
	}
}

func TestIntegration_InvalidLogin(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Test login with invalid credentials
	loginData := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "WrongPassword123!",
	}
	
	loginJSON, _ := json.Marshal(loginData)
	resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("Failed to make login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_UnauthorizedAccess(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Test /me endpoint without token
	resp, err := http.Get(server.URL + "/auth/me")
	if err != nil {
		t.Fatalf("Failed to make /me request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthorized access, got %d", resp.StatusCode)
	}
}