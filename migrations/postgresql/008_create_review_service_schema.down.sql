-- Drop triggers
DROP TRIGGER IF EXISTS update_product_ratings_summary_trigger ON reviews;
DROP TRIGGER IF EXISTS update_review_content_filters_updated_at ON review_content_filters;
DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;

-- Drop functions
DROP FUNCTION IF EXISTS update_product_ratings_summary();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_product_ratings_summary_average_rating;
DROP INDEX IF EXISTS idx_review_moderation_logs_moderator_id;
DROP INDEX IF EXISTS idx_review_moderation_logs_review_id;
DROP INDEX IF EXISTS idx_review_votes_user_id;
DROP INDEX IF EXISTS idx_review_votes_review_id;
DROP INDEX IF EXISTS idx_reviews_verified_purchase;
DROP INDEX IF EXISTS idx_reviews_created_at;
DROP INDEX IF EXISTS idx_reviews_rating;
DROP INDEX IF EXISTS idx_reviews_status;
DROP INDEX IF EXISTS idx_reviews_order_id;
DROP INDEX IF EXISTS idx_reviews_product_id;
DROP INDEX IF EXISTS idx_reviews_user_id;

-- Drop tables
DROP TABLE IF EXISTS review_content_filters;
DROP TABLE IF EXISTS product_ratings_summary;
DROP TABLE IF EXISTS review_moderation_logs;
DROP TABLE IF EXISTS review_votes;
DROP TABLE IF EXISTS reviews;
