package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/auth-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// RBACMiddleware provides role-based access control
type RBACMiddleware struct {
	authService *service.AuthService
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(authService *service.AuthService) *RBACMiddleware {
	return &RBACMiddleware{
		authService: authService,
	}
}

// Authenticate validates JWT token and adds user info to context
func (m *RBACMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		tokenString, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.writeErrorResponse(w, r, err)
			return
		}

		// Validate token and get user
		user, err := m.authService.ValidateToken(r.Context(), tokenString)
		if err != nil {
			m.writeErrorResponse(w, r, err)
			return
		}

		// Add user info to context
		ctx := utils.WithUserID(r.Context(), user.ID)
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "user_role", user.Role)
		ctx = context.WithValue(ctx, "user_status", user.Status)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if user has the required role
func (m *RBACMiddleware) RequireRole(role models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil {
				m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "User not authenticated", nil))
				return
			}

			if userRole.(models.UserRole) != role {
				m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthorization, "Insufficient permissions", nil))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole checks if user has any of the required roles
func (m *RBACMiddleware) RequireAnyRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil {
				m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "User not authenticated", nil))
				return
			}

			currentRole := userRole.(models.UserRole)
			hasPermission := false
			for _, role := range roles {
				if currentRole == role {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthorization, "Insufficient permissions", nil))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin checks if user has admin role
func (m *RBACMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole(models.RoleAdmin)(next)
}

// RequireAdminOrModerator checks if user has admin or moderator role
func (m *RBACMiddleware) RequireAdminOrModerator(next http.Handler) http.Handler {
	return m.RequireAnyRole(models.RoleAdmin, models.RoleModerator)(next)
}

// RequireActiveUser checks if user account is active
func (m *RBACMiddleware) RequireActiveUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userStatus := r.Context().Value("user_status")
		if userStatus == nil {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthentication, "User not authenticated", nil))
			return
		}

		if userStatus.(models.UserStatus) != models.StatusActive {
			m.writeErrorResponse(w, r, utils.NewAppError(utils.ErrAuthorization, "Account is not active", nil))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCurrentUser retrieves the current user from context
func GetCurrentUser(ctx context.Context) (*models.User, bool) {
	user := ctx.Value("user")
	if user == nil {
		return nil, false
	}
	return user.(*models.User), true
}

// GetCurrentUserID retrieves the current user ID from context
func GetCurrentUserID(ctx context.Context) (string, bool) {
	userID := utils.GetUserID(ctx)
	if userID == "" {
		return "", false
	}
	return userID, true
}

// GetCurrentUserRole retrieves the current user role from context
func GetCurrentUserRole(ctx context.Context) (models.UserRole, bool) {
	role := ctx.Value("user_role")
	if role == nil {
		return "", false
	}
	return role.(models.UserRole), true
}

// writeErrorResponse writes an error response
func (m *RBACMiddleware) writeErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *utils.AppError
	var ok bool
	
	if appErr, ok = err.(*utils.AppError); !ok {
		appErr = utils.NewInternalError("Internal server error", err)
	}

	response := utils.ErrorResponse{
		Error: utils.ErrorDetail{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
		TraceID: utils.GetTraceID(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatusCode())
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.Logger.Error(r.Context(), "Failed to encode error response", err)
	}
}