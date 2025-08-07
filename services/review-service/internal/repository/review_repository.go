package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// ReviewRepository defines the interface for review data operations
type ReviewRepository interface {
	// Review operations
	CreateReview(ctx context.Context, review *models.Review) error
	GetReviewByID(ctx context.Context, id string) (*models.Review, error)
	GetReviewsByProductID(ctx context.Context, productID string, limit, offset int) ([]*models.Review, error)
	GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error)
	UpdateReview(ctx context.Context, review *models.Review) error
	DeleteReview(ctx context.Context, id string) error
	
	// Review status operations
	UpdateReviewStatus(ctx context.Context, reviewID string, status models.ReviewStatus, moderatorID *string, reason *string) error
	GetReviewsByStatus(ctx context.Context, status models.ReviewStatus, limit, offset int) ([]*models.Review, error)
	
	// Purchase verification
	VerifyPurchase(ctx context.Context, userID, productID, orderID string) (bool, error)
	
	// Review voting operations
	CreateReviewVote(ctx context.Context, vote *models.ReviewVote) error
	GetReviewVote(ctx context.Context, reviewID, userID string) (*models.ReviewVote, error)
	UpdateReviewVote(ctx context.Context, vote *models.ReviewVote) error
	DeleteReviewVote(ctx context.Context, reviewID, userID string) error
	UpdateReviewHelpfulCounts(ctx context.Context, reviewID string) error
	
	// Rating summary operations
	GetProductRatingSummary(ctx context.Context, productID string) (*models.ReviewSummary, error)
	UpdateProductRatingSummary(ctx context.Context, productID string) error
	
	// Moderation operations
	CreateModerationLog(ctx context.Context, log *models.ReviewModerationLog) error
	GetModerationLogs(ctx context.Context, reviewID string) ([]*models.ReviewModerationLog, error)
	
	// Content filter operations
	GetContentFilters(ctx context.Context) ([]*models.ReviewContentFilter, error)
	CreateContentFilter(ctx context.Context, filter *models.ReviewContentFilter) error
	UpdateContentFilter(ctx context.Context, filter *models.ReviewContentFilter) error
	DeleteContentFilter(ctx context.Context, id string) error
}

// PostgresReviewRepository implements ReviewRepository using PostgreSQL
type PostgresReviewRepository struct {
	db     *sql.DB
	logger *utils.StructuredLogger
}

// NewPostgresReviewRepository creates a new PostgreSQL review repository
func NewPostgresReviewRepository(db *sql.DB, logger *utils.StructuredLogger) ReviewRepository {
	return &PostgresReviewRepository{
		db:     db,
		logger: logger,
	}
}

// CreateReview creates a new review
func (r *PostgresReviewRepository) CreateReview(ctx context.Context, review *models.Review) error {
	query := `
		INSERT INTO reviews (
			id, user_id, product_id, order_id, rating, title, content, 
			status, verified_purchase, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.UserID, review.ProductID, review.OrderID,
		review.Rating, review.Title, review.Content, review.Status,
		review.Verified, review.CreatedAt, review.UpdatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to create review", err, map[string]interface{}{
			"review_id": review.ID,
			"user_id":   review.UserID,
			"product_id": review.ProductID,
		})
		return fmt.Errorf("failed to create review: %w", err)
	}
	
	return nil
}

// GetReviewByID retrieves a review by ID
func (r *PostgresReviewRepository) GetReviewByID(ctx context.Context, id string) (*models.Review, error) {
	query := `
		SELECT id, user_id, product_id, order_id, rating, title, content,
			   status, moderation_reason, moderated_by, moderated_at,
			   helpful_count, not_helpful_count, verified_purchase,
			   created_at, updated_at
		FROM reviews WHERE id = $1`
	
	review := &models.Review{}
	var moderationReason, moderatedBy sql.NullString
	var moderatedAt pq.NullTime
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
		&review.Rating, &review.Title, &review.Content, &review.Status,
		&moderationReason, &moderatedBy, &moderatedAt,
		&review.Helpful, &review.NotHelpful, &review.Verified,
		&review.CreatedAt, &review.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review not found")
		}
		r.logger.Error(ctx, "Failed to get review by ID", err, map[string]interface{}{
			"review_id": id,
		})
		return nil, fmt.Errorf("failed to get review: %w", err)
	}
	
	return review, nil
}

// GetReviewsByProductID retrieves reviews for a product
func (r *PostgresReviewRepository) GetReviewsByProductID(ctx context.Context, productID string, limit, offset int) ([]*models.Review, error) {
	query := `
		SELECT id, user_id, product_id, order_id, rating, title, content,
			   status, moderation_reason, moderated_by, moderated_at,
			   helpful_count, not_helpful_count, verified_purchase,
			   created_at, updated_at
		FROM reviews 
		WHERE product_id = $1 AND status = 'approved'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.QueryContext(ctx, query, productID, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "Failed to get reviews by product ID", err, map[string]interface{}{
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()
	
	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		var moderationReason, moderatedBy sql.NullString
		var moderatedAt pq.NullTime
		
		err := rows.Scan(
			&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
			&review.Rating, &review.Title, &review.Content, &review.Status,
			&moderationReason, &moderatedBy, &moderatedAt,
			&review.Helpful, &review.NotHelpful, &review.Verified,
			&review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(ctx, "Failed to scan review", err)
			continue
		}
		
		reviews = append(reviews, review)
	}
	
	return reviews, nil
}

// GetReviewsByUserID retrieves reviews by a user
func (r *PostgresReviewRepository) GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error) {
	query := `
		SELECT id, user_id, product_id, order_id, rating, title, content,
			   status, moderation_reason, moderated_by, moderated_at,
			   helpful_count, not_helpful_count, verified_purchase,
			   created_at, updated_at
		FROM reviews 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "Failed to get reviews by user ID", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()
	
	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		var moderationReason, moderatedBy sql.NullString
		var moderatedAt pq.NullTime
		
		err := rows.Scan(
			&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
			&review.Rating, &review.Title, &review.Content, &review.Status,
			&moderationReason, &moderatedBy, &moderatedAt,
			&review.Helpful, &review.NotHelpful, &review.Verified,
			&review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(ctx, "Failed to scan review", err)
			continue
		}
		
		reviews = append(reviews, review)
	}
	
	return reviews, nil
}

// UpdateReview updates an existing review
func (r *PostgresReviewRepository) UpdateReview(ctx context.Context, review *models.Review) error {
	query := `
		UPDATE reviews 
		SET rating = $2, title = $3, content = $4, status = $5, updated_at = $6
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		review.ID, review.Rating, review.Title, review.Content,
		review.Status, review.UpdatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to update review", err, map[string]interface{}{
			"review_id": review.ID,
		})
		return fmt.Errorf("failed to update review: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}
	
	return nil
}

// DeleteReview deletes a review
func (r *PostgresReviewRepository) DeleteReview(ctx context.Context, id string) error {
	query := `DELETE FROM reviews WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to delete review", err, map[string]interface{}{
			"review_id": id,
		})
		return fmt.Errorf("failed to delete review: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}
	
	return nil
}

// Continue with remaining methods in separate files due to length constraints...
