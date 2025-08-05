package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/user-service/internal/handlers"
	"github.com/shopsphere/user-service/internal/repository"
	"github.com/shopsphere/user-service/internal/service"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB     *sql.DB
	testRouter *mux.Router
	container  testcontainers.Container
)

func TestMain(m *testing.M) {
	// Setup test database container
	if err := setupTestDatabase(); err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := teardownTestDatabase(); err != nil {
		fmt.Printf("Failed to cleanup test database: %v\n", err)
	}

	os.Exit(code)
}

func setupTestDatabase() error {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "user_service_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	var err error
	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get container port: %w", err)
	}

	// Connect to test database
	dbConfig := &utils.DatabaseConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "test",
		Password: "test",
		DBName:   "user_service_test",
		SSLMode:  "disable",
	}

	testDB, err = dbConfig.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Run migrations
	if err := runTestMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup test router
	setupTestRouter()

	return nil
}

func teardownTestDatabase() error {
	if testDB != nil {
		testDB.Close()
	}

	if container != nil {
		ctx := context.Background()
		return container.Terminate(ctx)
	}

	return nil
}

func runTestMigrations() error {
	// Read and execute the migration file
	migrationSQL := `
		-- Users table
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone VARCHAR(20),
			role VARCHAR(20) DEFAULT 'customer' CHECK (role IN ('customer', 'admin', 'moderator')),
			status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('active', 'suspended', 'deleted', 'pending')),
			email_verified BOOLEAN DEFAULT FALSE,
			phone_verified BOOLEAN DEFAULT FALSE,
			last_login_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Password reset tokens table
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Email verification tokens table
		CREATE TABLE IF NOT EXISTS email_verification_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Create indexes
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_username ON users(username);
		CREATE INDEX idx_users_status ON users(status);
		CREATE INDEX idx_users_role ON users(role);

		-- Create trigger to update updated_at timestamp
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := testDB.Exec(migrationSQL)
	return err
}

func setupTestRouter() {
	userRepo := repository.NewUserRepository(testDB)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	testRouter = mux.NewRouter()
	api := testRouter.PathPrefix("/api/v1").Subrouter()

	// Setup routes
	api.HandleFunc("/users/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/users/verify-email", userHandler.VerifyEmail).Methods("POST")
	api.HandleFunc("/users/password-reset/request", userHandler.RequestPasswordReset).Methods("POST")
	api.HandleFunc("/users/password-reset/confirm", userHandler.ResetPassword).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}/password", userHandler.ChangePassword).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	api.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
	api.HandleFunc("/users/search", userHandler.SearchUsers).Methods("GET")
	api.HandleFunc("/users/{id}/status", userHandler.UpdateUserStatus).Methods("PUT")
}

func clearTestData() {
	testDB.Exec("DELETE FROM email_verification_tokens")
	testDB.Exec("DELETE FROM password_reset_tokens")
	testDB.Exec("DELETE FROM users")
}

func TestUserRegistration(t *testing.T) {
	clearTestData()

	tests := []struct {
		name           string
		payload        handlers.RegisterRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid registration",
			payload: handlers.RegisterRequest{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "Password123!",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Invalid email",
			payload: handlers.RegisterRequest{
				Email:     "invalid-email",
				Username:  "testuser2",
				FirstName: "Test",
				LastName:  "User",
				Password:  "Password123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Weak password",
			payload: handlers.RegisterRequest{
				Email:     "test2@example.com",
				Username:  "testuser3",
				FirstName: "Test",
				LastName:  "User",
				Password:  "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Duplicate email",
			payload: handlers.RegisterRequest{
				Email:     "test@example.com", // Same as first test
				Username:  "testuser4",
				FirstName: "Test",
				LastName:  "User",
				Password:  "Password123!",
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Environment", "development")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				var response handlers.RegisterResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response.User == nil {
					t.Error("Expected user in response")
				}

				if response.VerificationToken == "" {
					t.Error("Expected verification token in development mode")
				}
			}
		})
	}
}

func TestEmailVerification(t *testing.T) {
	clearTestData()

	// First register a user
	payload := handlers.RegisterRequest{
		Email:     "verify@example.com",
		Username:  "verifyuser",
		FirstName: "Verify",
		LastName:  "User",
		Password:  "Password123!",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Environment", "development")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Registration failed with status %d", w.Code)
	}

	var registerResponse handlers.RegisterResponse
	json.NewDecoder(w.Body).Decode(&registerResponse)

	// Now verify email
	verifyPayload := handlers.VerifyEmailRequest{
		Token: registerResponse.VerificationToken,
	}

	body, _ = json.Marshal(verifyPayload)
	req = httptest.NewRequest("POST", "/api/v1/users/verify-email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify user status is now active
	var user models.User
	err := testDB.QueryRow("SELECT status FROM users WHERE email = $1", payload.Email).Scan(&user.Status)
	if err != nil {
		t.Errorf("Failed to query user status: %v", err)
	}

	if user.Status != models.StatusActive {
		t.Errorf("Expected user status to be active, got %s", user.Status)
	}
}

func TestPasswordReset(t *testing.T) {
	clearTestData()

	// First register and verify a user
	userRepo := repository.NewUserRepository(testDB)
	user := models.NewUser("reset@example.com", "resetuser", "Reset", "User")
	user.Status = models.StatusActive
	user.PasswordHash = "$2a$10$example" // dummy hash
	userRepo.Create(user)

	// Request password reset
	resetRequest := handlers.PasswordResetRequest{
		Email: "reset@example.com",
	}

	body, _ := json.Marshal(resetRequest)
	req := httptest.NewRequest("POST", "/api/v1/users/password-reset/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Environment", "development")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resetResponse handlers.PasswordResetResponse
	json.NewDecoder(w.Body).Decode(&resetResponse)

	if resetResponse.Token == "" {
		t.Error("Expected reset token in development mode")
	}

	// Now reset password
	confirmRequest := handlers.ResetPasswordRequest{
		Token:       resetResponse.Token,
		NewPassword: "NewPassword123!",
	}

	body, _ = json.Marshal(confirmRequest)
	req = httptest.NewRequest("POST", "/api/v1/users/password-reset/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUserProfileOperations(t *testing.T) {
	clearTestData()

	// Create a test user
	userRepo := repository.NewUserRepository(testDB)
	user := models.NewUser("profile@example.com", "profileuser", "Profile", "User")
	user.Status = models.StatusActive
	user.PasswordHash = "$2a$10$example" // dummy hash
	userRepo.Create(user)

	// Test getting user
	req := httptest.NewRequest("GET", "/api/v1/users/"+user.ID, nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var retrievedUser models.User
	json.NewDecoder(w.Body).Decode(&retrievedUser)

	if retrievedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrievedUser.Email)
	}

	// Test updating user
	updateRequest := handlers.UpdateUserRequest{
		FirstName: stringPtr("Updated"),
		LastName:  stringPtr("Name"),
		Phone:     stringPtr("+1234567890"),
	}

	body, _ := json.Marshal(updateRequest)
	req = httptest.NewRequest("PUT", "/api/v1/users/"+user.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var updatedUser models.User
	json.NewDecoder(w.Body).Decode(&updatedUser)

	if updatedUser.FirstName != "Updated" {
		t.Errorf("Expected first name 'Updated', got %s", updatedUser.FirstName)
	}
}

func TestUserListing(t *testing.T) {
	clearTestData()

	// Create test users
	userRepo := repository.NewUserRepository(testDB)
	users := []*models.User{
		models.NewUser("user1@example.com", "user1", "User", "One"),
		models.NewUser("user2@example.com", "user2", "User", "Two"),
		models.NewUser("user3@example.com", "user3", "User", "Three"),
	}

	for _, user := range users {
		user.Status = models.StatusActive
		user.PasswordHash = "$2a$10$example"
		userRepo.Create(user)
	}

	// Test listing users
	req := httptest.NewRequest("GET", "/api/v1/users?limit=2&offset=0", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response handlers.ListUsersResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(response.Users))
	}

	if response.Total != 3 {
		t.Errorf("Expected total 3, got %d", response.Total)
	}
}

func TestUserSearch(t *testing.T) {
	clearTestData()

	// Create test users
	userRepo := repository.NewUserRepository(testDB)
	users := []*models.User{
		models.NewUser("john@example.com", "john_doe", "John", "Doe"),
		models.NewUser("jane@example.com", "jane_smith", "Jane", "Smith"),
		models.NewUser("bob@example.com", "bob_jones", "Bob", "Jones"),
	}

	for _, user := range users {
		user.Status = models.StatusActive
		user.PasswordHash = "$2a$10$example"
		userRepo.Create(user)
	}

	// Test searching users
	req := httptest.NewRequest("GET", "/api/v1/users/search?q=john", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response handlers.ListUsersResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(response.Users))
	}

	if response.Users[0].FirstName != "John" {
		t.Errorf("Expected first name 'John', got %s", response.Users[0].FirstName)
	}
}

func TestUserStatusUpdate(t *testing.T) {
	clearTestData()

	// Create a test user
	userRepo := repository.NewUserRepository(testDB)
	user := models.NewUser("status@example.com", "statususer", "Status", "User")
	user.Status = models.StatusActive
	user.PasswordHash = "$2a$10$example"
	userRepo.Create(user)

	// Test updating user status
	statusRequest := handlers.UpdateStatusRequest{
		Status: models.StatusSuspended,
	}

	body, _ := json.Marshal(statusRequest)
	req := httptest.NewRequest("PUT", "/api/v1/users/"+user.ID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify status was updated
	var updatedStatus models.UserStatus
	err := testDB.QueryRow("SELECT status FROM users WHERE id = $1", user.ID).Scan(&updatedStatus)
	if err != nil {
		t.Errorf("Failed to query user status: %v", err)
	}

	if updatedStatus != models.StatusSuspended {
		t.Errorf("Expected status suspended, got %s", updatedStatus)
	}
}

func TestUserDeletion(t *testing.T) {
	clearTestData()

	// Create a test user
	userRepo := repository.NewUserRepository(testDB)
	user := models.NewUser("delete@example.com", "deleteuser", "Delete", "User")
	user.Status = models.StatusActive
	user.PasswordHash = "$2a$10$example"
	userRepo.Create(user)

	// Test deleting user
	req := httptest.NewRequest("DELETE", "/api/v1/users/"+user.ID, nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify user was soft deleted
	var status models.UserStatus
	err := testDB.QueryRow("SELECT status FROM users WHERE id = $1", user.ID).Scan(&status)
	if err != nil {
		t.Errorf("Failed to query user status: %v", err)
	}

	if status != models.StatusDeleted {
		t.Errorf("Expected status deleted, got %s", status)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}