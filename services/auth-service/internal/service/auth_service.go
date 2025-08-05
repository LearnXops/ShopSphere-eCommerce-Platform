package service

import (
	"context"
	"time"

	"github.com/shopsphere/auth-service/internal/auth"
	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// UserRepository interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUserLastLogin(userID string) error
	UpdateUserStatus(userID string, status models.UserStatus) error
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
	CreateUserSession(userID, tokenHash, deviceInfo, ipAddress string, expiresAt time.Time) error
	ValidateUserSession(userID, tokenHash string) (bool, error)
	DeleteUserSession(userID, tokenHash string) error
	DeleteExpiredSessions() error
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User   *models.User `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthService handles authentication operations
type AuthService struct {
	userRepo        UserRepository
	jwtService      *jwt.JWTService
	passwordService *auth.PasswordService
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo UserRepository, jwtService *jwt.JWTService, passwordService *auth.PasswordService) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, utils.NewValidationError("Email is required")
	}
	if req.Password == "" {
		return nil, utils.NewValidationError("Password is required")
	}

	// Get user by email
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == utils.ErrNotFound {
			return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid email or password", nil)
		}
		return nil, err
	}

	// Check user status
	if user.Status != models.StatusActive {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Account is not active", nil)
	}

	// Verify password
	if err := s.passwordService.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid email or password", nil)
	}

	// Generate token pair
	tokens, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token session
	refreshClaims, err := s.jwtService.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Create session record
	expiresAt := time.Now().Add(24 * time.Hour * 7) // 7 days
	if err := s.userRepo.CreateUserSession(user.ID, refreshClaims.TokenHash, "", "", expiresAt); err != nil {
		utils.Logger.Error(ctx, "Failed to create user session", err)
		// Don't fail login if session creation fails
	}

	// Update last login
	if err := s.userRepo.UpdateUserLastLogin(user.ID); err != nil {
		utils.Logger.Error(ctx, "Failed to update last login", err)
		// Don't fail login if last login update fails
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &LoginResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *RefreshRequest) (*jwt.TokenPair, error) {
	if req.RefreshToken == "" {
		return nil, utils.NewValidationError("Refresh token is required")
	}

	// Validate refresh token
	refreshClaims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Validate session exists
	exists, err := s.userRepo.ValidateUserSession(refreshClaims.UserID, refreshClaims.TokenHash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid refresh token session", nil)
	}

	// Get user
	user, err := s.userRepo.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, err
	}

	// Check user status
	if user.Status != models.StatusActive {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Account is not active", nil)
	}

	// Generate new token pair
	tokens, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Delete old session
	if err := s.userRepo.DeleteUserSession(refreshClaims.UserID, refreshClaims.TokenHash); err != nil {
		utils.Logger.Error(ctx, "Failed to delete old session", err)
	}

	// Create new session
	newRefreshClaims, err := s.jwtService.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(24 * time.Hour * 7) // 7 days
	if err := s.userRepo.CreateUserSession(user.ID, newRefreshClaims.TokenHash, "", "", expiresAt); err != nil {
		utils.Logger.Error(ctx, "Failed to create new session", err)
	}

	return tokens, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, req *LogoutRequest) error {
	if req.RefreshToken == "" {
		return utils.NewValidationError("Refresh token is required")
	}

	// Validate refresh token
	refreshClaims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		// If token is invalid, consider logout successful
		return nil
	}

	// Delete session
	if err := s.userRepo.DeleteUserSession(refreshClaims.UserID, refreshClaims.TokenHash); err != nil {
		utils.Logger.Error(ctx, "Failed to delete session during logout", err)
		// Don't fail logout if session deletion fails
	}

	return nil
}

// ValidateToken validates an access token and returns user info
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// Validate access token
	claims, err := s.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check user status
	if user.Status != models.StatusActive {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Account is not active", nil)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	return s.userRepo.DeleteExpiredSessions()
}