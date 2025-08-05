package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string           `json:"user_id"`
	Role   models.UserRole  `json:"role"`
	Status models.UserStatus `json:"status"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims
type RefreshClaims struct {
	UserID    string `json:"user_id"`
	TokenHash string `json:"token_hash"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	accessSecret     []byte
	refreshSecret    []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	issuer          string
}

// NewJWTService creates a new JWT service
func NewJWTService(accessSecret, refreshSecret, issuer string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		accessSecret:    []byte(accessSecret),
		refreshSecret:   []byte(refreshSecret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		issuer:         issuer,
	}
}

// GenerateTokenPair generates access and refresh token pair
func (j *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()
	
	// Generate access token
	accessClaims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		Status: user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   user.ID,
			Audience:  []string{"shopsphere-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.accessSecret)
	if err != nil {
		return nil, utils.NewInternalError("Failed to generate access token", err)
	}

	// Generate refresh token hash
	refreshTokenHash, err := j.generateSecureToken()
	if err != nil {
		return nil, utils.NewInternalError("Failed to generate refresh token hash", err)
	}

	// Generate refresh token
	refreshClaims := &RefreshClaims{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   user.ID,
			Audience:  []string{"shopsphere-refresh"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.refreshSecret)
	if err != nil {
		return nil, utils.NewInternalError("Failed to generate refresh token", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates and parses access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.accessSecret, nil
	})

	if err != nil {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid access token", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid token claims", nil)
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.refreshSecret, nil
	})

	if err != nil {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid refresh token", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, utils.NewAppError(utils.ErrAuthentication, "Invalid refresh token claims", nil)
	}

	return claims, nil
}

// generateSecureToken generates a cryptographically secure random token
func (j *JWTService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", utils.NewAppError(utils.ErrAuthentication, "Authorization header required", nil)
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", utils.NewAppError(utils.ErrAuthentication, "Invalid authorization header format", nil)
	}

	return authHeader[len(bearerPrefix):], nil
}