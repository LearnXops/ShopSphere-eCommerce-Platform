package search

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/utils"
)

// AnalyticsService implements search analytics functionality
type AnalyticsService struct {
	db *sql.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// RecordSearch records a search query for analytics
func (a *AnalyticsService) RecordSearch(ctx context.Context, query string, userID string, resultsCount int) error {
	searchID := uuid.New().String()
	now := time.Now()

	insertQuery := `
		INSERT INTO search_analytics (
			id, query, user_id, results_count, created_at
		) VALUES ($1, $2, $3, $4, $5)`

	_, err := a.db.ExecContext(ctx, insertQuery, searchID, query, userID, resultsCount, now)
	if err != nil {
		return utils.NewInternalError("failed to record search", err)
	}

	return nil
}

// GetPopularSearches returns popular search terms
func (a *AnalyticsService) GetPopularSearches(ctx context.Context, limit int) ([]SearchTerm, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT query, COUNT(*) as frequency
		FROM search_analytics
		WHERE created_at >= NOW() - INTERVAL '30 days'
		  AND query != ''
		GROUP BY query
		ORDER BY frequency DESC
		LIMIT $1`

	rows, err := a.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, utils.NewInternalError("failed to get popular searches", err)
	}
	defer rows.Close()

	var terms []SearchTerm
	for rows.Next() {
		var term SearchTerm
		if err := rows.Scan(&term.Term, &term.Frequency); err != nil {
			return nil, utils.NewInternalError("failed to scan search term", err)
		}
		terms = append(terms, term)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.NewInternalError("failed to iterate search terms", err)
	}

	return terms, nil
}

// GetSearchMetrics returns search performance metrics
func (a *AnalyticsService) GetSearchMetrics(ctx context.Context, from, to time.Time) (*SearchMetrics, error) {
	// Get basic metrics
	metricsQuery := `
		SELECT 
			COUNT(*) as total_searches,
			AVG(results_count) as average_results,
			COUNT(CASE WHEN results_count = 0 THEN 1 END)::float / COUNT(*)::float as zero_results_rate
		FROM search_analytics
		WHERE created_at >= $1 AND created_at <= $2`

	var metrics SearchMetrics
	err := a.db.QueryRowContext(ctx, metricsQuery, from, to).Scan(
		&metrics.TotalSearches,
		&metrics.AverageResults,
		&metrics.ZeroResultsRate,
	)
	if err != nil {
		return nil, utils.NewInternalError("failed to get search metrics", err)
	}

	// Get popular terms for the period
	popularTerms, err := a.getPopularTermsForPeriod(ctx, from, to, 10)
	if err != nil {
		return nil, err
	}
	metrics.PopularTerms = popularTerms

	// Note: Response time would need to be tracked separately in a real implementation
	metrics.AverageResponseTime = 50 * time.Millisecond

	return &metrics, nil
}

// getPopularTermsForPeriod gets popular terms for a specific time period
func (a *AnalyticsService) getPopularTermsForPeriod(ctx context.Context, from, to time.Time, limit int) ([]SearchTerm, error) {
	query := `
		SELECT query, COUNT(*) as frequency
		FROM search_analytics
		WHERE created_at >= $1 AND created_at <= $2
		  AND query != ''
		GROUP BY query
		ORDER BY frequency DESC
		LIMIT $3`

	rows, err := a.db.QueryContext(ctx, query, from, to, limit)
	if err != nil {
		return nil, utils.NewInternalError("failed to get popular terms for period", err)
	}
	defer rows.Close()

	var terms []SearchTerm
	for rows.Next() {
		var term SearchTerm
		if err := rows.Scan(&term.Term, &term.Frequency); err != nil {
			return nil, utils.NewInternalError("failed to scan search term", err)
		}
		terms = append(terms, term)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.NewInternalError("failed to iterate search terms", err)
	}

	return terms, nil
}