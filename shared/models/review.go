package models

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the status of a review
type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
	ReviewFlagged  ReviewStatus = "flagged"
)

// Review represents a product review
type Review struct {
	ID        string       `json:"id" db:"id"`
	ProductID string       `json:"product_id" db:"product_id"`
	UserID    string       `json:"user_id" db:"user_id"`
	OrderID   string       `json:"order_id" db:"order_id"`
	Rating    int          `json:"rating" db:"rating"` // 1-5 stars
	Title     string       `json:"title" db:"title"`
	Content   string       `json:"content" db:"content"`
	Status    ReviewStatus `json:"status" db:"status"`
	Helpful   int          `json:"helpful" db:"helpful"`
	NotHelpful int         `json:"not_helpful" db:"not_helpful"`
	Verified  bool         `json:"verified" db:"verified"` // verified purchase
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// NewReview creates a new review with default values
func NewReview(productID, userID, orderID string, rating int, title, content string) *Review {
	return &Review{
		ID:        uuid.New().String(),
		ProductID: productID,
		UserID:    userID,
		OrderID:   orderID,
		Rating:    rating,
		Title:     title,
		Content:   content,
		Status:    ReviewPending,
		Verified:  true, // assuming verified if orderID is provided
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ReviewSummary represents aggregated review data for a product
type ReviewSummary struct {
	ProductID     string  `json:"product_id" db:"product_id"`
	TotalReviews  int     `json:"total_reviews" db:"total_reviews"`
	AverageRating float64 `json:"average_rating" db:"average_rating"`
	Rating1Count  int     `json:"rating_1_count" db:"rating_1_count"`
	Rating2Count  int     `json:"rating_2_count" db:"rating_2_count"`
	Rating3Count  int     `json:"rating_3_count" db:"rating_3_count"`
	Rating4Count  int     `json:"rating_4_count" db:"rating_4_count"`
	Rating5Count  int     `json:"rating_5_count" db:"rating_5_count"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}