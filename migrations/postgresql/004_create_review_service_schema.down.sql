-- Rollback review service database schema

-- Drop triggers
DROP TRIGGER IF EXISTS review_vote_counts_trigger ON review_votes;
DROP TRIGGER IF EXISTS product_review_summary_trigger ON reviews;
DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;
DROP TRIGGER IF EXISTS update_product_review_summaries_updated_at ON product_review_summaries;
DROP TRIGGER IF EXISTS update_review_responses_updated_at ON review_responses;
DROP TRIGGER IF EXISTS update_review_response_templates_updated_at ON review_response_templates;

-- Drop functions
DROP FUNCTION IF EXISTS update_review_vote_counts();
DROP FUNCTION IF EXISTS update_product_review_summary();

-- Drop indexes
DROP INDEX IF EXISTS idx_reviews_product_id;
DROP INDEX IF EXISTS idx_reviews_user_id;
DROP INDEX IF EXISTS idx_reviews_order_id;
DROP INDEX IF EXISTS idx_reviews_status;
DROP INDEX IF EXISTS idx_reviews_rating;
DROP INDEX IF EXISTS idx_reviews_verified_purchase;
DROP INDEX IF EXISTS idx_reviews_created_at;
DROP INDEX IF EXISTS idx_review_votes_review_id;
DROP INDEX IF EXISTS idx_review_votes_user_id;
DROP INDEX IF EXISTS idx_review_reports_review_id;
DROP INDEX IF EXISTS idx_review_reports_status;
DROP INDEX IF EXISTS idx_review_responses_review_id;
DROP INDEX IF EXISTS idx_review_response_templates_category;

-- Drop tables
DROP TABLE IF EXISTS review_response_templates;
DROP TABLE IF EXISTS review_responses;
DROP TABLE IF EXISTS product_review_summaries;
DROP TABLE IF EXISTS review_reports;
DROP TABLE IF EXISTS review_votes;
DROP TABLE IF EXISTS reviews;