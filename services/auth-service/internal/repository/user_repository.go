package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		return utils.NewInternalError("Failed to create user", err)
	}
	
	return nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND status != 'deleted'
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("User")
		}
		return nil, utils.NewInternalError("Failed to get user by email", err)
	}
	
	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users
		WHERE id = $1 AND status != 'deleted'
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("User")
		}
		return nil, utils.NewInternalError("Failed to get user by ID", err)
	}
	
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users
		WHERE username = $1 AND status != 'deleted'
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("User")
		}
		return nil, utils.NewInternalError("Failed to get user by username", err)
	}
	
	return user, nil
}

// UpdateUserLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateUserLastLogin(userID string) error {
	query := `
		UPDATE users
		SET last_login_at = $1, updated_at = $2
		WHERE id = $3
	`
	
	now := time.Now()
	_, err := r.db.Exec(query, now, now, userID)
	if err != nil {
		return utils.NewInternalError("Failed to update user last login", err)
	}
	
	return nil
}

// UpdateUserStatus updates the user's status
func (r *UserRepository) UpdateUserStatus(userID string, status models.UserStatus) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := r.db.Exec(query, status, time.Now(), userID)
	if err != nil {
		return utils.NewInternalError("Failed to update user status", err)
	}
	
	return nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND status != 'deleted')`
	
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, utils.NewInternalError("Failed to check email existence", err)
	}
	
	return exists, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND status != 'deleted')`
	
	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, utils.NewInternalError("Failed to check username existence", err)
	}
	
	return exists, nil
}

// CreateUserSession creates a new user session
func (r *UserRepository) CreateUserSession(userID, tokenHash, deviceInfo, ipAddress string, expiresAt time.Time) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, device_info, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	sessionID := uuid.New().String()
	_, err := r.db.Exec(query, sessionID, userID, tokenHash, deviceInfo, ipAddress, expiresAt, time.Now())
	if err != nil {
		return utils.NewInternalError("Failed to create user session", err)
	}
	
	return nil
}

// ValidateUserSession validates if a session exists and is not expired
func (r *UserRepository) ValidateUserSession(userID, tokenHash string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_sessions 
			WHERE user_id = $1 AND token_hash = $2 AND expires_at > NOW()
		)
	`
	
	var exists bool
	err := r.db.QueryRow(query, userID, tokenHash).Scan(&exists)
	if err != nil {
		return false, utils.NewInternalError("Failed to validate user session", err)
	}
	
	return exists, nil
}

// DeleteUserSession deletes a user session
func (r *UserRepository) DeleteUserSession(userID, tokenHash string) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1 AND token_hash = $2`
	
	_, err := r.db.Exec(query, userID, tokenHash)
	if err != nil {
		return utils.NewInternalError("Failed to delete user session", err)
	}
	
	return nil
}

// DeleteExpiredSessions deletes expired sessions
func (r *UserRepository) DeleteExpiredSessions() error {
	query := `DELETE FROM user_sessions WHERE expires_at <= NOW()`
	
	_, err := r.db.Exec(query)
	if err != nil {
		return utils.NewInternalError("Failed to delete expired sessions", err)
	}
	
	return nil
}