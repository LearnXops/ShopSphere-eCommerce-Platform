package repository

import (
	"database/sql"
	"fmt"
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

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err := r.db.Exec(query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users WHERE id = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Phone,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users WHERE email = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Phone,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users WHERE username = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Phone,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, first_name = $4, last_name = $5, phone = $6, 
		    role = $7, status = $8, updated_at = $9
		WHERE id = $1
	`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.db.Exec(query,
		user.ID, user.Email, user.Username, user.FirstName, user.LastName,
		user.Phone, user.Role, user.Status, user.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("user")
	}
	
	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	
	result, err := r.db.Exec(query, userID, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("user")
	}
	
	return nil
}

// Delete soft deletes a user by setting status to deleted
func (r *UserRepository) Delete(id string) error {
	query := `UPDATE users SET status = $2, updated_at = $3 WHERE id = $1`
	
	result, err := r.db.Exec(query, id, models.StatusDeleted, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("user")
	}
	
	return nil
}

// List retrieves users with pagination and filtering
func (r *UserRepository) List(limit, offset int, status models.UserStatus, role models.UserRole) ([]*models.User, int, error) {
	// Build query with optional filters
	baseQuery := `FROM users WHERE 1=1`
	countQuery := `SELECT COUNT(*) ` + baseQuery
	selectQuery := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
	` + baseQuery
	
	args := []interface{}{}
	argIndex := 1
	
	if status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery = `SELECT COUNT(*) ` + baseQuery
		selectQuery = `
			SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		` + baseQuery
		args = append(args, status)
		argIndex++
	}
	
	if role != "" {
		baseQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		countQuery = `SELECT COUNT(*) ` + baseQuery
		selectQuery = `
			SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		` + baseQuery
		args = append(args, role)
		argIndex++
	}
	
	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}
	
	// Add pagination
	selectQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash,
			&user.FirstName, &user.LastName, &user.Phone,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}
	
	return users, total, nil
}

// Search searches users by email, username, first name, or last name
func (r *UserRepository) Search(query string, limit, offset int) ([]*models.User, int, error) {
	searchPattern := "%" + query + "%"
	
	countQuery := `
		SELECT COUNT(*) FROM users 
		WHERE (email ILIKE $1 OR username ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)
		AND status != $2
	`
	
	selectQuery := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, role, status, created_at, updated_at
		FROM users 
		WHERE (email ILIKE $1 OR username ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)
		AND status != $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	
	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, searchPattern, models.StatusDeleted).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get search count: %w", err)
	}
	
	rows, err := r.db.Query(selectQuery, searchPattern, models.StatusDeleted, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash,
			&user.FirstName, &user.LastName, &user.Phone,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating search results: %w", err)
	}
	
	return users, total, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return exists, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	
	return exists, nil
}

// CreatePasswordResetToken creates a password reset token
func (r *UserRepository) CreatePasswordResetToken(userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := r.db.Exec(query, uuid.New().String(), userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}
	
	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (r *UserRepository) GetPasswordResetToken(tokenHash string) (string, error) {
	query := `
		SELECT user_id FROM password_reset_tokens 
		WHERE token_hash = $1 AND expires_at > NOW() AND used = FALSE
	`
	
	var userID string
	err := r.db.QueryRow(query, tokenHash).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", utils.NewNotFoundError("password reset token")
		}
		return "", fmt.Errorf("failed to get password reset token: %w", err)
	}
	
	return userID, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (r *UserRepository) MarkPasswordResetTokenUsed(tokenHash string) error {
	query := `UPDATE password_reset_tokens SET used = TRUE WHERE token_hash = $1`
	
	_, err := r.db.Exec(query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}
	
	return nil
}

// CreateEmailVerificationToken creates an email verification token
func (r *UserRepository) CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := r.db.Exec(query, uuid.New().String(), userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create email verification token: %w", err)
	}
	
	return nil
}

// VerifyEmail verifies a user's email using the verification token
func (r *UserRepository) VerifyEmail(tokenHash string) error {
	return utils.ExecuteInTransaction(r.db, func(tx *sql.Tx) error {
		// Get user ID from token
		var userID string
		err := tx.QueryRow(`
			SELECT user_id FROM email_verification_tokens 
			WHERE token_hash = $1 AND expires_at > NOW() AND used = FALSE
		`, tokenHash).Scan(&userID)
		
		if err != nil {
			if err == sql.ErrNoRows {
				return utils.NewNotFoundError("email verification token")
			}
			return fmt.Errorf("failed to get email verification token: %w", err)
		}
		
		// Mark token as used
		_, err = tx.Exec(`UPDATE email_verification_tokens SET used = TRUE WHERE token_hash = $1`, tokenHash)
		if err != nil {
			return fmt.Errorf("failed to mark email verification token as used: %w", err)
		}
		
		// Update user status to active and mark email as verified
		_, err = tx.Exec(`
			UPDATE users SET status = $1, email_verified = TRUE, updated_at = $2 
			WHERE id = $3
		`, models.StatusActive, time.Now(), userID)
		if err != nil {
			return fmt.Errorf("failed to verify user email: %w", err)
		}
		
		return nil
	})
}