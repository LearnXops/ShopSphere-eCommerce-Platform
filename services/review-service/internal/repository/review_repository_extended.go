package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopsphere/shared/models"
)

// UpdateReviewStatus updates the status of a review
func (r *PostgresReviewRepository) UpdateReviewStatus(ctx context.Context, reviewID string, status models.ReviewStatus, moderatorID *string, reason *string) error {
	query := `
		UPDATE reviews 
		SET status = $2, moderated_by = $3, moderation_reason = $4, moderated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, reviewID, status, moderatorID, reason)
	if err != nil {
		r.logger.Error(ctx, "Failed to update review status", err, map[string]interface{}{
			"review_id": reviewID,
			"status":    status,
		})
		return fmt.Errorf("failed to update review status: %w", err)
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

// GetReviewsByStatus retrieves reviews by status
func (r *PostgresReviewRepository) GetReviewsByStatus(ctx context.Context, status models.ReviewStatus, limit, offset int) ([]*models.Review, error) {
	query := `
		SELECT id, user_id, product_id, order_id, rating, title, content,
			   status, moderation_reason, moderated_by, moderated_at,
			   helpful_count, not_helpful_count, verified_purchase,
			   created_at, updated_at
		FROM reviews 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "Failed to get reviews by status", err, map[string]interface{}{
			"status": status,
		})
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()
	
	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		var moderationReason, moderatedBy sql.NullString
		var moderatedAt sql.NullTime
		
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

// VerifyPurchase verifies if a user has purchased a product
func (r *PostgresReviewRepository) VerifyPurchase(ctx context.Context, userID, productID, orderID string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		WHERE o.user_id = $1 AND o.id = $2 AND oi.product_id = $3 AND o.status = 'completed'`
	
	var verified bool
	err := r.db.QueryRowContext(ctx, query, userID, orderID, productID).Scan(&verified)
	if err != nil {
		r.logger.Error(ctx, "Failed to verify purchase", err, map[string]interface{}{
			"user_id":    userID,
			"product_id": productID,
			"order_id":   orderID,
		})
		return false, fmt.Errorf("failed to verify purchase: %w", err)
	}
	
	return verified, nil
}

// CreateReviewVote creates a new review vote
func (r *PostgresReviewRepository) CreateReviewVote(ctx context.Context, vote *models.ReviewVote) error {
	query := `
		INSERT INTO review_votes (id, review_id, user_id, vote_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (review_id, user_id) 
		DO UPDATE SET vote_type = EXCLUDED.vote_type`
	
	_, err := r.db.ExecContext(ctx, query,
		vote.ID, vote.ReviewID, vote.UserID, vote.VoteType, vote.CreatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to create review vote", err, map[string]interface{}{
			"review_id": vote.ReviewID,
			"user_id":   vote.UserID,
		})
		return fmt.Errorf("failed to create review vote: %w", err)
	}
	
	// Update helpful counts
	return r.UpdateReviewHelpfulCounts(ctx, vote.ReviewID)
}

// GetReviewVote retrieves a user's vote on a review
func (r *PostgresReviewRepository) GetReviewVote(ctx context.Context, reviewID, userID string) (*models.ReviewVote, error) {
	query := `
		SELECT id, review_id, user_id, vote_type, created_at
		FROM review_votes
		WHERE review_id = $1 AND user_id = $2`
	
	vote := &models.ReviewVote{}
	err := r.db.QueryRowContext(ctx, query, reviewID, userID).Scan(
		&vote.ID, &vote.ReviewID, &vote.UserID, &vote.VoteType, &vote.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vote found
		}
		r.logger.Error(ctx, "Failed to get review vote", err, map[string]interface{}{
			"review_id": reviewID,
			"user_id":   userID,
		})
		return nil, fmt.Errorf("failed to get review vote: %w", err)
	}
	
	return vote, nil
}

// UpdateReviewVote updates an existing review vote
func (r *PostgresReviewRepository) UpdateReviewVote(ctx context.Context, vote *models.ReviewVote) error {
	query := `
		UPDATE review_votes 
		SET vote_type = $3
		WHERE review_id = $1 AND user_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, vote.ReviewID, vote.UserID, vote.VoteType)
	if err != nil {
		r.logger.Error(ctx, "Failed to update review vote", err, map[string]interface{}{
			"review_id": vote.ReviewID,
			"user_id":   vote.UserID,
		})
		return fmt.Errorf("failed to update review vote: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("review vote not found")
	}
	
	// Update helpful counts
	return r.UpdateReviewHelpfulCounts(ctx, vote.ReviewID)
}

// DeleteReviewVote deletes a review vote
func (r *PostgresReviewRepository) DeleteReviewVote(ctx context.Context, reviewID, userID string) error {
	query := `DELETE FROM review_votes WHERE review_id = $1 AND user_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, reviewID, userID)
	if err != nil {
		r.logger.Error(ctx, "Failed to delete review vote", err, map[string]interface{}{
			"review_id": reviewID,
			"user_id":   userID,
		})
		return fmt.Errorf("failed to delete review vote: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("review vote not found")
	}
	
	// Update helpful counts
	return r.UpdateReviewHelpfulCounts(ctx, reviewID)
}

// UpdateReviewHelpfulCounts updates the helpful and not helpful counts for a review
func (r *PostgresReviewRepository) UpdateReviewHelpfulCounts(ctx context.Context, reviewID string) error {
	query := `
		UPDATE reviews 
		SET helpful_count = (
			SELECT COUNT(*) FROM review_votes 
			WHERE review_id = $1 AND vote_type = 'helpful'
		),
		not_helpful_count = (
			SELECT COUNT(*) FROM review_votes 
			WHERE review_id = $1 AND vote_type = 'not_helpful'
		)
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, reviewID)
	if err != nil {
		r.logger.Error(ctx, "Failed to update review helpful counts", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return fmt.Errorf("failed to update review helpful counts: %w", err)
	}
	
	return nil
}

// GetProductRatingSummary retrieves the rating summary for a product
func (r *PostgresReviewRepository) GetProductRatingSummary(ctx context.Context, productID string) (*models.ReviewSummary, error) {
	query := `
		SELECT product_id, total_reviews, average_rating,
			   rating_1_count, rating_2_count, rating_3_count,
			   rating_4_count, rating_5_count, last_updated
		FROM product_ratings_summary
		WHERE product_id = $1`
	
	summary := &models.ReviewSummary{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&summary.ProductID, &summary.TotalReviews, &summary.AverageRating,
		&summary.Rating1Count, &summary.Rating2Count, &summary.Rating3Count,
		&summary.Rating4Count, &summary.Rating5Count, &summary.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty summary if not found
			return &models.ReviewSummary{
				ProductID:     productID,
				TotalReviews:  0,
				AverageRating: 0.0,
			}, nil
		}
		r.logger.Error(ctx, "Failed to get product rating summary", err, map[string]interface{}{
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get product rating summary: %w", err)
	}
	
	return summary, nil
}

// UpdateProductRatingSummary manually updates the rating summary for a product
func (r *PostgresReviewRepository) UpdateProductRatingSummary(ctx context.Context, productID string) error {
	query := `
		INSERT INTO product_ratings_summary (
			product_id, total_reviews, average_rating,
			rating_1_count, rating_2_count, rating_3_count,
			rating_4_count, rating_5_count, last_updated
		)
		SELECT 
			$1,
			COUNT(*),
			COALESCE(ROUND(AVG(rating::DECIMAL), 2), 0),
			COUNT(*) FILTER (WHERE rating = 1),
			COUNT(*) FILTER (WHERE rating = 2),
			COUNT(*) FILTER (WHERE rating = 3),
			COUNT(*) FILTER (WHERE rating = 4),
			COUNT(*) FILTER (WHERE rating = 5),
			CURRENT_TIMESTAMP
		FROM reviews 
		WHERE product_id = $1 AND status = 'approved'
		ON CONFLICT (product_id) DO UPDATE SET
			total_reviews = EXCLUDED.total_reviews,
			average_rating = EXCLUDED.average_rating,
			rating_1_count = EXCLUDED.rating_1_count,
			rating_2_count = EXCLUDED.rating_2_count,
			rating_3_count = EXCLUDED.rating_3_count,
			rating_4_count = EXCLUDED.rating_4_count,
			rating_5_count = EXCLUDED.rating_5_count,
			last_updated = EXCLUDED.last_updated`
	
	_, err := r.db.ExecContext(ctx, query, productID)
	if err != nil {
		r.logger.Error(ctx, "Failed to update product rating summary", err, map[string]interface{}{
			"product_id": productID,
		})
		return fmt.Errorf("failed to update product rating summary: %w", err)
	}
	
	return nil
}

// CreateModerationLog creates a new moderation log entry
func (r *PostgresReviewRepository) CreateModerationLog(ctx context.Context, log *models.ReviewModerationLog) error {
	query := `
		INSERT INTO review_moderation_logs (
			id, review_id, moderator_id, action, reason,
			previous_status, new_status, automated, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.ReviewID, log.ModeratorID, log.Action, log.Reason,
		log.PreviousStatus, log.NewStatus, log.Automated, log.CreatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to create moderation log", err, map[string]interface{}{
			"review_id": log.ReviewID,
			"action":    log.Action,
		})
		return fmt.Errorf("failed to create moderation log: %w", err)
	}
	
	return nil
}

// GetModerationLogs retrieves moderation logs for a review
func (r *PostgresReviewRepository) GetModerationLogs(ctx context.Context, reviewID string) ([]*models.ReviewModerationLog, error) {
	query := `
		SELECT id, review_id, moderator_id, action, reason,
			   previous_status, new_status, automated, created_at
		FROM review_moderation_logs
		WHERE review_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, reviewID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get moderation logs", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return nil, fmt.Errorf("failed to get moderation logs: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.ReviewModerationLog
	for rows.Next() {
		log := &models.ReviewModerationLog{}
		var moderatorID, reason sql.NullString
		var previousStatus, newStatus sql.NullString
		
		err := rows.Scan(
			&log.ID, &log.ReviewID, &moderatorID, &log.Action, &reason,
			&previousStatus, &newStatus, &log.Automated, &log.CreatedAt,
		)
		if err != nil {
			r.logger.Error(ctx, "Failed to scan moderation log", err)
			continue
		}
		
		if moderatorID.Valid {
			log.ModeratorID = &moderatorID.String
		}
		if reason.Valid {
			log.Reason = &reason.String
		}
		if previousStatus.Valid {
			status := models.ReviewStatus(previousStatus.String)
			log.PreviousStatus = &status
		}
		if newStatus.Valid {
			status := models.ReviewStatus(newStatus.String)
			log.NewStatus = &status
		}
		
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetContentFilters retrieves all active content filters
func (r *PostgresReviewRepository) GetContentFilters(ctx context.Context) ([]*models.ReviewContentFilter, error) {
	query := `
		SELECT id, filter_type, pattern, action, severity, active, created_at, updated_at
		FROM review_content_filters
		WHERE active = true
		ORDER BY severity DESC, created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error(ctx, "Failed to get content filters", err)
		return nil, fmt.Errorf("failed to get content filters: %w", err)
	}
	defer rows.Close()
	
	var filters []*models.ReviewContentFilter
	for rows.Next() {
		filter := &models.ReviewContentFilter{}
		
		err := rows.Scan(
			&filter.ID, &filter.FilterType, &filter.Pattern,
			&filter.Action, &filter.Severity, &filter.Active,
			&filter.CreatedAt, &filter.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(ctx, "Failed to scan content filter", err)
			continue
		}
		
		filters = append(filters, filter)
	}
	
	return filters, nil
}

// CreateContentFilter creates a new content filter
func (r *PostgresReviewRepository) CreateContentFilter(ctx context.Context, filter *models.ReviewContentFilter) error {
	query := `
		INSERT INTO review_content_filters (
			id, filter_type, pattern, action, severity, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.ExecContext(ctx, query,
		filter.ID, filter.FilterType, filter.Pattern, filter.Action,
		filter.Severity, filter.Active, filter.CreatedAt, filter.UpdatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to create content filter", err, map[string]interface{}{
			"filter_type": filter.FilterType,
			"pattern":     filter.Pattern,
		})
		return fmt.Errorf("failed to create content filter: %w", err)
	}
	
	return nil
}

// UpdateContentFilter updates an existing content filter
func (r *PostgresReviewRepository) UpdateContentFilter(ctx context.Context, filter *models.ReviewContentFilter) error {
	query := `
		UPDATE review_content_filters 
		SET filter_type = $2, pattern = $3, action = $4, severity = $5, active = $6, updated_at = $7
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		filter.ID, filter.FilterType, filter.Pattern, filter.Action,
		filter.Severity, filter.Active, filter.UpdatedAt,
	)
	
	if err != nil {
		r.logger.Error(ctx, "Failed to update content filter", err, map[string]interface{}{
			"filter_id": filter.ID,
		})
		return fmt.Errorf("failed to update content filter: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("content filter not found")
	}
	
	return nil
}

// DeleteContentFilter deletes a content filter
func (r *PostgresReviewRepository) DeleteContentFilter(ctx context.Context, id string) error {
	query := `DELETE FROM review_content_filters WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to delete content filter", err, map[string]interface{}{
			"filter_id": id,
		})
		return fmt.Errorf("failed to delete content filter: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("content filter not found")
	}
	
	return nil
}
