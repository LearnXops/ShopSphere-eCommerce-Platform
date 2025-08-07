package models

import (
	"errors"
	"regexp"
	"strings"
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
	ProductID     string    `json:"product_id" db:"product_id"`
	TotalReviews  int       `json:"total_reviews" db:"total_reviews"`
	AverageRating float64   `json:"average_rating" db:"average_rating"`
	Rating1Count  int       `json:"rating_1_count" db:"rating_1_count"`
	Rating2Count  int       `json:"rating_2_count" db:"rating_2_count"`
	Rating3Count  int       `json:"rating_3_count" db:"rating_3_count"`
	Rating4Count  int       `json:"rating_4_count" db:"rating_4_count"`
	Rating5Count  int       `json:"rating_5_count" db:"rating_5_count"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// VoteType represents the type of vote on a review
type VoteType string

const (
	VoteHelpful    VoteType = "helpful"
	VoteNotHelpful VoteType = "not_helpful"
)

// ReviewVote represents a user's vote on a review's helpfulness
type ReviewVote struct {
	ID       string    `json:"id" db:"id"`
	ReviewID string    `json:"review_id" db:"review_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	VoteType VoteType  `json:"vote_type" db:"vote_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewReviewVote creates a new review vote
func NewReviewVote(reviewID, userID string, voteType VoteType) *ReviewVote {
	return &ReviewVote{
		ID:       uuid.New().String(),
		ReviewID: reviewID,
		UserID:   userID,
		VoteType: voteType,
		CreatedAt: time.Now(),
	}
}

// ModerationAction represents the type of moderation action
type ModerationAction string

const (
	ModerationApprove ModerationAction = "approve"
	ModerationReject  ModerationAction = "reject"
	ModerationFlag    ModerationAction = "flag"
	ModerationUnflag  ModerationAction = "unflag"
)

// ReviewModerationLog represents a moderation action on a review
type ReviewModerationLog struct {
	ID             string            `json:"id" db:"id"`
	ReviewID       string            `json:"review_id" db:"review_id"`
	ModeratorID    *string           `json:"moderator_id" db:"moderator_id"`
	Action         ModerationAction  `json:"action" db:"action"`
	Reason         *string           `json:"reason" db:"reason"`
	PreviousStatus *ReviewStatus     `json:"previous_status" db:"previous_status"`
	NewStatus      *ReviewStatus     `json:"new_status" db:"new_status"`
	Automated      bool              `json:"automated" db:"automated"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
}

// NewModerationLog creates a new moderation log entry
func NewModerationLog(reviewID string, moderatorID *string, action ModerationAction, reason *string, prevStatus, newStatus *ReviewStatus, automated bool) *ReviewModerationLog {
	return &ReviewModerationLog{
		ID:             uuid.New().String(),
		ReviewID:       reviewID,
		ModeratorID:    moderatorID,
		Action:         action,
		Reason:         reason,
		PreviousStatus: prevStatus,
		NewStatus:      newStatus,
		Automated:      automated,
		CreatedAt:      time.Now(),
	}
}

// FilterType represents the type of content filter
type FilterType string

const (
	FilterProfanity    FilterType = "profanity"
	FilterSpam         FilterType = "spam"
	FilterPersonalInfo FilterType = "personal_info"
	FilterPromotional  FilterType = "promotional"
	FilterFakeReview   FilterType = "fake_review"
)

// FilterAction represents the action to take when a filter matches
type FilterAction string

const (
	FilterActionFlag   FilterAction = "flag"
	FilterActionReject FilterAction = "reject"
	FilterActionWarn   FilterAction = "warn"
)

// FilterSeverity represents the severity level of a content filter
type FilterSeverity string

const (
	FilterSeverityLow    FilterSeverity = "low"
	FilterSeverityMedium FilterSeverity = "medium"
	FilterSeverityHigh   FilterSeverity = "high"
)

// ReviewContentFilter represents a content moderation filter
type ReviewContentFilter struct {
	ID         string          `json:"id" db:"id"`
	FilterType FilterType      `json:"filter_type" db:"filter_type"`
	Pattern    string          `json:"pattern" db:"pattern"`
	Action     FilterAction    `json:"action" db:"action"`
	Severity   FilterSeverity  `json:"severity" db:"severity"`
	Active     bool            `json:"active" db:"active"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// NewContentFilter creates a new content filter
func NewContentFilter(filterType FilterType, pattern string, action FilterAction, severity FilterSeverity) *ReviewContentFilter {
	return &ReviewContentFilter{
		ID:         uuid.New().String(),
		FilterType: filterType,
		Pattern:    pattern,
		Action:     action,
		Severity:   severity,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// MatchesFilter checks if the content matches this filter
func (f *ReviewContentFilter) MatchesFilter(content string) bool {
	if !f.Active {
		return false
	}
	
	regex, err := regexp.Compile(f.Pattern)
	if err != nil {
		return false
	}
	
	return regex.MatchString(strings.ToLower(content))
}

// ReviewRequest represents a request to create a review
type ReviewRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	OrderID   string `json:"order_id" validate:"required,uuid"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Title     string `json:"title" validate:"required,min=5,max=255"`
	Content   string `json:"content" validate:"required,min=10,max=2000"`
}

// Validate validates the review request
func (r *ReviewRequest) Validate() error {
	if r.ProductID == "" {
		return errors.New("product_id is required")
	}
	if r.OrderID == "" {
		return errors.New("order_id is required")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if len(strings.TrimSpace(r.Title)) < 5 {
		return errors.New("title must be at least 5 characters")
	}
	if len(strings.TrimSpace(r.Content)) < 10 {
		return errors.New("content must be at least 10 characters")
	}
	if len(r.Title) > 255 {
		return errors.New("title must be less than 255 characters")
	}
	if len(r.Content) > 2000 {
		return errors.New("content must be less than 2000 characters")
	}
	return nil
}

// ReviewVoteRequest represents a request to vote on a review
type ReviewVoteRequest struct {
	VoteType VoteType `json:"vote_type" validate:"required,oneof=helpful not_helpful"`
}

// Validate validates the review vote request
func (r *ReviewVoteRequest) Validate() error {
	if r.VoteType != VoteHelpful && r.VoteType != VoteNotHelpful {
		return errors.New("vote_type must be either 'helpful' or 'not_helpful'")
	}
	return nil
}

// ModerationRequest represents a request to moderate a review
type ModerationRequest struct {
	Action ModerationAction `json:"action" validate:"required,oneof=approve reject flag unflag"`
	Reason *string          `json:"reason,omitempty"`
}

// Validate validates the moderation request
func (r *ModerationRequest) Validate() error {
	validActions := []ModerationAction{ModerationApprove, ModerationReject, ModerationFlag, ModerationUnflag}
	for _, action := range validActions {
		if r.Action == action {
			return nil
		}
	}
	return errors.New("invalid moderation action")
}