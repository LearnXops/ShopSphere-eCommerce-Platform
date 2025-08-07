package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/review-service/internal/repository"
)

// ReviewService defines the interface for review business logic
type ReviewService interface {
	// Review operations
	CreateReview(ctx context.Context, userID string, request *models.ReviewRequest) (*models.Review, error)
	GetReviewByID(ctx context.Context, id string) (*models.Review, error)
	GetReviewsByProductID(ctx context.Context, productID string, limit, offset int) ([]*models.Review, error)
	GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error)
	UpdateReview(ctx context.Context, userID, reviewID string, request *models.ReviewRequest) (*models.Review, error)
	DeleteReview(ctx context.Context, userID, reviewID string) error
	
	// Review voting operations
	VoteOnReview(ctx context.Context, userID, reviewID string, request *models.ReviewVoteRequest) error
	RemoveVote(ctx context.Context, userID, reviewID string) error
	
	// Rating summary operations
	GetProductRatingSummary(ctx context.Context, productID string) (*models.ReviewSummary, error)
	
	// Moderation operations
	ModerateReview(ctx context.Context, moderatorID, reviewID string, request *models.ModerationRequest) error
	GetPendingReviews(ctx context.Context, limit, offset int) ([]*models.Review, error)
	GetModerationLogs(ctx context.Context, reviewID string) ([]*models.ReviewModerationLog, error)
	
	// Content moderation
	CheckContentModeration(ctx context.Context, content string) (bool, string, error)
}

// ReviewServiceImpl implements ReviewService
type ReviewServiceImpl struct {
	repo   repository.ReviewRepository
	logger *utils.StructuredLogger
}

// NewReviewService creates a new review service
func NewReviewService(repo repository.ReviewRepository, logger *utils.StructuredLogger) ReviewService {
	return &ReviewServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// CreateReview creates a new product review
func (s *ReviewServiceImpl) CreateReview(ctx context.Context, userID string, request *models.ReviewRequest) (*models.Review, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		s.logger.Error(ctx, "Invalid review request", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Verify purchase
	verified, err := s.repo.VerifyPurchase(ctx, userID, request.ProductID, request.OrderID)
	if err != nil {
		s.logger.Error(ctx, "Failed to verify purchase", err, map[string]interface{}{
			"user_id":    userID,
			"product_id": request.ProductID,
			"order_id":   request.OrderID,
		})
		return nil, fmt.Errorf("failed to verify purchase: %w", err)
	}
	
	if !verified {
		s.logger.Info(ctx, "Purchase verification failed", map[string]interface{}{
			"user_id":    userID,
			"product_id": request.ProductID,
			"order_id":   request.OrderID,
		})
		return nil, fmt.Errorf("purchase not verified")
	}
	
	// Check if user already reviewed this product
	existingReviews, err := s.repo.GetReviewsByUserID(ctx, userID, 100, 0)
	if err != nil {
		s.logger.Error(ctx, "Failed to check existing reviews", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to check existing reviews: %w", err)
	}
	
	for _, review := range existingReviews {
		if review.ProductID == request.ProductID {
			return nil, fmt.Errorf("user has already reviewed this product")
		}
	}
	
	// Create review
	review := models.NewReview(request.ProductID, userID, request.OrderID, request.Rating, request.Title, request.Content)
	
	// Check content moderation
	needsModeration, reason, err := s.CheckContentModeration(ctx, request.Title+" "+request.Content)
	if err != nil {
		s.logger.Error(ctx, "Failed to check content moderation", err, map[string]interface{}{
			"user_id": userID,
		})
		// Continue with creation but flag for manual review
		needsModeration = true
		reason = "automatic moderation check failed"
	}
	
	if needsModeration {
		review.Status = models.ReviewFlagged
		s.logger.Info(ctx, "Review flagged for moderation", map[string]interface{}{
			"review_id": review.ID,
			"reason":    reason,
		})
		
		// Create moderation log
		prevStatus := models.ReviewPending
		newStatus := models.ReviewFlagged
		moderationLog := models.NewModerationLog(
			review.ID, nil, models.ModerationFlag, &reason, &prevStatus, &newStatus, true,
		)
		if err := s.repo.CreateModerationLog(ctx, moderationLog); err != nil {
			s.logger.Error(ctx, "Failed to create moderation log", err)
		}
	} else {
		review.Status = models.ReviewApproved
	}
	
	// Save review
	if err := s.repo.CreateReview(ctx, review); err != nil {
		s.logger.Error(ctx, "Failed to create review", err, map[string]interface{}{
			"user_id":    userID,
			"product_id": request.ProductID,
		})
		return nil, fmt.Errorf("failed to create review: %w", err)
	}
	
	s.logger.Info(ctx, "Review created successfully", map[string]interface{}{
		"review_id":  review.ID,
		"user_id":    userID,
		"product_id": request.ProductID,
		"status":     review.Status,
	})
	
	return review, nil
}

// GetReviewByID retrieves a review by ID
func (s *ReviewServiceImpl) GetReviewByID(ctx context.Context, id string) (*models.Review, error) {
	review, err := s.repo.GetReviewByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get review by ID", err, map[string]interface{}{
			"review_id": id,
		})
		return nil, fmt.Errorf("failed to get review: %w", err)
	}
	
	return review, nil
}

// GetReviewsByProductID retrieves reviews for a product
func (s *ReviewServiceImpl) GetReviewsByProductID(ctx context.Context, productID string, limit, offset int) ([]*models.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	reviews, err := s.repo.GetReviewsByProductID(ctx, productID, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to get reviews by product ID", err, map[string]interface{}{
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	
	return reviews, nil
}

// GetReviewsByUserID retrieves reviews by a user
func (s *ReviewServiceImpl) GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	reviews, err := s.repo.GetReviewsByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to get reviews by user ID", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	
	return reviews, nil
}

// UpdateReview updates an existing review
func (s *ReviewServiceImpl) UpdateReview(ctx context.Context, userID, reviewID string, request *models.ReviewRequest) (*models.Review, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		s.logger.Error(ctx, "Invalid review update request", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Get existing review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get review for update", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return nil, fmt.Errorf("failed to get review: %w", err)
	}
	
	// Check ownership
	if review.UserID != userID {
		s.logger.Info(ctx, "Unauthorized review update attempt", map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
			"owner_id":  review.UserID,
		})
		return nil, fmt.Errorf("unauthorized: user does not own this review")
	}
	
	// Update review fields
	review.Rating = request.Rating
	review.Title = request.Title
	review.Content = request.Content
	review.UpdatedAt = time.Now()
	
	// Check content moderation for updated content
	needsModeration, reason, err := s.CheckContentModeration(ctx, request.Title+" "+request.Content)
	if err != nil {
		s.logger.Error(ctx, "Failed to check content moderation for update", err, map[string]interface{}{
			"review_id": reviewID,
		})
		// Continue with update but flag for manual review
		needsModeration = true
		reason = "automatic moderation check failed"
	}
	
	prevStatus := review.Status
	if needsModeration {
		review.Status = models.ReviewFlagged
		
		// Create moderation log
		newStatus := models.ReviewFlagged
		moderationLog := models.NewModerationLog(
			review.ID, nil, models.ModerationFlag, &reason, &prevStatus, &newStatus, true,
		)
		if err := s.repo.CreateModerationLog(ctx, moderationLog); err != nil {
			s.logger.Error(ctx, "Failed to create moderation log for update", err)
		}
	} else if prevStatus == models.ReviewFlagged {
		review.Status = models.ReviewApproved
	}
	
	// Save updated review
	if err := s.repo.UpdateReview(ctx, review); err != nil {
		s.logger.Error(ctx, "Failed to update review", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return nil, fmt.Errorf("failed to update review: %w", err)
	}
	
	s.logger.Info(ctx, "Review updated successfully", map[string]interface{}{
		"review_id": reviewID,
		"user_id":   userID,
		"status":    review.Status,
	})
	
	return review, nil
}

// DeleteReview deletes a review
func (s *ReviewServiceImpl) DeleteReview(ctx context.Context, userID, reviewID string) error {
	// Get existing review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get review for deletion", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return fmt.Errorf("failed to get review: %w", err)
	}
	
	// Check ownership
	if review.UserID != userID {
		s.logger.Info(ctx, "Unauthorized review deletion attempt", map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
			"owner_id":  review.UserID,
		})
		return fmt.Errorf("unauthorized: user does not own this review")
	}
	
	// Delete review
	if err := s.repo.DeleteReview(ctx, reviewID); err != nil {
		s.logger.Error(ctx, "Failed to delete review", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return fmt.Errorf("failed to delete review: %w", err)
	}
	
	s.logger.Info(ctx, "Review deleted successfully", map[string]interface{}{
		"review_id": reviewID,
		"user_id":   userID,
	})
	
	return nil
}

// VoteOnReview allows a user to vote on a review's helpfulness
func (s *ReviewServiceImpl) VoteOnReview(ctx context.Context, userID, reviewID string, request *models.ReviewVoteRequest) error {
	// Validate request
	if err := request.Validate(); err != nil {
		s.logger.Error(ctx, "Invalid vote request", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		return fmt.Errorf("invalid request: %w", err)
	}
	
	// Check if review exists
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get review for voting", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return fmt.Errorf("failed to get review: %w", err)
	}
	
	// Users cannot vote on their own reviews
	if review.UserID == userID {
		return fmt.Errorf("users cannot vote on their own reviews")
	}
	
	// Check if user has already voted
	existingVote, err := s.repo.GetReviewVote(ctx, reviewID, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to check existing vote", err, map[string]interface{}{
			"review_id": reviewID,
			"user_id":   userID,
		})
		return fmt.Errorf("failed to check existing vote: %w", err)
	}
	
	if existingVote != nil {
		// Update existing vote
		existingVote.VoteType = request.VoteType
		if err := s.repo.UpdateReviewVote(ctx, existingVote); err != nil {
			s.logger.Error(ctx, "Failed to update vote", err, map[string]interface{}{
				"review_id": reviewID,
				"user_id":   userID,
			})
			return fmt.Errorf("failed to update vote: %w", err)
		}
	} else {
		// Create new vote
		vote := models.NewReviewVote(reviewID, userID, request.VoteType)
		if err := s.repo.CreateReviewVote(ctx, vote); err != nil {
			s.logger.Error(ctx, "Failed to create vote", err, map[string]interface{}{
				"review_id": reviewID,
				"user_id":   userID,
			})
			return fmt.Errorf("failed to create vote: %w", err)
		}
	}
	
	s.logger.Info(ctx, "Vote recorded successfully", map[string]interface{}{
		"review_id": reviewID,
		"user_id":   userID,
		"vote_type": request.VoteType,
	})
	
	return nil
}

// RemoveVote removes a user's vote on a review
func (s *ReviewServiceImpl) RemoveVote(ctx context.Context, userID, reviewID string) error {
	// Check if vote exists
	existingVote, err := s.repo.GetReviewVote(ctx, reviewID, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to check existing vote for removal", err, map[string]interface{}{
			"review_id": reviewID,
			"user_id":   userID,
		})
		return fmt.Errorf("failed to check existing vote: %w", err)
	}
	
	if existingVote == nil {
		return fmt.Errorf("no vote found to remove")
	}
	
	// Remove vote
	if err := s.repo.DeleteReviewVote(ctx, reviewID, userID); err != nil {
		s.logger.Error(ctx, "Failed to remove vote", err, map[string]interface{}{
			"review_id": reviewID,
			"user_id":   userID,
		})
		return fmt.Errorf("failed to remove vote: %w", err)
	}
	
	s.logger.Info(ctx, "Vote removed successfully", map[string]interface{}{
		"review_id": reviewID,
		"user_id":   userID,
	})
	
	return nil
}

// GetProductRatingSummary retrieves the rating summary for a product
func (s *ReviewServiceImpl) GetProductRatingSummary(ctx context.Context, productID string) (*models.ReviewSummary, error) {
	summary, err := s.repo.GetProductRatingSummary(ctx, productID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get product rating summary", err, map[string]interface{}{
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get rating summary: %w", err)
	}
	
	return summary, nil
}

// ModerateReview moderates a review (approve, reject, flag, unflag)
func (s *ReviewServiceImpl) ModerateReview(ctx context.Context, moderatorID, reviewID string, request *models.ModerationRequest) error {
	// Validate request
	if err := request.Validate(); err != nil {
		s.logger.Error(ctx, "Invalid moderation request", err, map[string]interface{}{
			"moderator_id": moderatorID,
			"review_id":    reviewID,
		})
		return fmt.Errorf("invalid request: %w", err)
	}
	
	// Get existing review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get review for moderation", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return fmt.Errorf("failed to get review: %w", err)
	}
	
	prevStatus := review.Status
	var newStatus models.ReviewStatus
	
	// Determine new status based on action
	switch request.Action {
	case models.ModerationApprove:
		newStatus = models.ReviewApproved
	case models.ModerationReject:
		newStatus = models.ReviewRejected
	case models.ModerationFlag:
		newStatus = models.ReviewFlagged
	case models.ModerationUnflag:
		newStatus = models.ReviewApproved
	default:
		return fmt.Errorf("invalid moderation action")
	}
	
	// Update review status
	if err := s.repo.UpdateReviewStatus(ctx, reviewID, newStatus, &moderatorID, request.Reason); err != nil {
		s.logger.Error(ctx, "Failed to update review status", err, map[string]interface{}{
			"review_id": reviewID,
			"action":    request.Action,
		})
		return fmt.Errorf("failed to update review status: %w", err)
	}
	
	// Create moderation log
	moderationLog := models.NewModerationLog(
		reviewID, &moderatorID, request.Action, request.Reason, &prevStatus, &newStatus, false,
	)
	if err := s.repo.CreateModerationLog(ctx, moderationLog); err != nil {
		s.logger.Error(ctx, "Failed to create moderation log", err)
		// Don't fail the operation if logging fails
	}
	
	s.logger.Info(ctx, "Review moderated successfully", map[string]interface{}{
		"review_id":     reviewID,
		"moderator_id":  moderatorID,
		"action":        request.Action,
		"prev_status":   prevStatus,
		"new_status":    newStatus,
	})
	
	return nil
}

// GetPendingReviews retrieves reviews pending moderation
func (s *ReviewServiceImpl) GetPendingReviews(ctx context.Context, limit, offset int) ([]*models.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	reviews, err := s.repo.GetReviewsByStatus(ctx, models.ReviewPending, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to get pending reviews", err)
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}
	
	// Also get flagged reviews
	flaggedReviews, err := s.repo.GetReviewsByStatus(ctx, models.ReviewFlagged, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to get flagged reviews", err)
		return nil, fmt.Errorf("failed to get flagged reviews: %w", err)
	}
	
	// Combine and return
	allReviews := append(reviews, flaggedReviews...)
	return allReviews, nil
}

// GetModerationLogs retrieves moderation logs for a review
func (s *ReviewServiceImpl) GetModerationLogs(ctx context.Context, reviewID string) ([]*models.ReviewModerationLog, error) {
	logs, err := s.repo.GetModerationLogs(ctx, reviewID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get moderation logs", err, map[string]interface{}{
			"review_id": reviewID,
		})
		return nil, fmt.Errorf("failed to get moderation logs: %w", err)
	}
	
	return logs, nil
}

// CheckContentModeration checks if content needs moderation
func (s *ReviewServiceImpl) CheckContentModeration(ctx context.Context, content string) (bool, string, error) {
	// Get active content filters
	filters, err := s.repo.GetContentFilters(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get content filters", err)
		return false, "", fmt.Errorf("failed to get content filters: %w", err)
	}
	
	// Check content against each filter
	for _, filter := range filters {
		if filter.MatchesFilter(content) {
			reason := fmt.Sprintf("Content matched %s filter: %s", filter.FilterType, filter.Pattern)
			
			s.logger.Info(ctx, "Content moderation triggered", map[string]interface{}{
				"filter_type": filter.FilterType,
				"action":      filter.Action,
				"severity":    filter.Severity,
			})
			
			// Return true for flag or reject actions
			if filter.Action == models.FilterActionFlag || filter.Action == models.FilterActionReject {
				return true, reason, nil
			}
		}
	}
	
	return false, "", nil
}
