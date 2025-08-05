package repository

import (
	"time"

	"github.com/shopsphere/shared/models"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	UpdatePassword(userID, passwordHash string) error
	Delete(id string) error
	List(limit, offset int, status models.UserStatus, role models.UserRole) ([]*models.User, int, error)
	Search(query string, limit, offset int) ([]*models.User, int, error)
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
	CreatePasswordResetToken(userID, tokenHash string, expiresAt time.Time) error
	GetPasswordResetToken(tokenHash string) (string, error)
	MarkPasswordResetTokenUsed(tokenHash string) error
	CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time) error
	VerifyEmail(tokenHash string) error
}