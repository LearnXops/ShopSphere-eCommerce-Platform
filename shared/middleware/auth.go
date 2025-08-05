package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shopsphere/shared/utils"
)

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	secretKey []byte
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: []byte(secretKey),
	}
}

// Authenticate validates JWT token and adds user info to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "Authorization header required", nil))
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "Invalid authorization header format", nil))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.secretKey, nil
		})

		if err != nil {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "Invalid token", err))
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "Invalid token claims", nil))
			return
		}

		// Add user info to context
		ctx := utils.WithUserID(r.Context(), claims.UserID)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if user has required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil || userRole.(string) != role {
				m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthorization, "Insufficient permissions", nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, r *http.Request, err *utils.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	
	response := utils.ErrorResponse{
		Error: utils.ErrorDetail{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
		TraceID: utils.GetTraceID(r.Context()),
	}
	
	// In a real implementation, you'd use a JSON encoder here
	// For now, we'll keep it simple
}