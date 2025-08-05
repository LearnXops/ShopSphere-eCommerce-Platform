-- Create review service database schema
-- This migration creates the product review and rating system tables

-- Reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36), -- for verified purchase reviews
    order_item_id VARCHAR(36), -- specific order item being reviewed
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    content TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'flagged', 'hidden')),
    
    -- Review metadata
    verified_purchase BOOLEAN DEFAULT FALSE,
    helpful_count INTEGER DEFAULT 0 CHECK (helpful_count >= 0),
    not_helpful_count INTEGER DEFAULT 0 CHECK (not_helpful_count >= 0),
    
    -- Moderation information
    moderated_by VARCHAR(36), -- admin user who moderated
    moderated_at TIMESTAMP,
    moderation_reason TEXT,
    
    -- Review attributes
    pros TEXT,
    cons TEXT,
    images TEXT[], -- array of review image URLs
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one review per user per product (unless multiple orders)
    UNIQUE(product_id, user_id, order_item_id)
);

-- Review helpfulness votes table
CREATE TABLE IF NOT EXISTS review_votes (
    id VARCHAR(36) PRIMARY KEY,
    review_id VARCHAR(36) NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('helpful', 'not_helpful')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id) -- one vote per user per review
);

-- Review reports table for user-reported inappropriate reviews
CREATE TABLE IF NOT EXISTS review_reports (
    id VARCHAR(36) PRIMARY KEY,
    review_id VARCHAR(36) NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    reported_by VARCHAR(36) NOT NULL,
    report_reason VARCHAR(50) NOT NULL CHECK (report_reason IN ('spam', 'inappropriate', 'fake', 'offensive', 'other')),
    report_details TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
    reviewed_by VARCHAR(36), -- admin who reviewed the report
    reviewed_at TIMESTAMP,
    resolution_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product review summaries table for aggregated review data
CREATE TABLE IF NOT EXISTS product_review_summaries (
    product_id VARCHAR(36) PRIMARY KEY,
    total_reviews INTEGER DEFAULT 0 CHECK (total_reviews >= 0),
    average_rating DECIMAL(3,2) DEFAULT 0 CHECK (average_rating >= 0 AND average_rating <= 5),
    rating_1_count INTEGER DEFAULT 0 CHECK (rating_1_count >= 0),
    rating_2_count INTEGER DEFAULT 0 CHECK (rating_2_count >= 0),
    rating_3_count INTEGER DEFAULT 0 CHECK (rating_3_count >= 0),
    rating_4_count INTEGER DEFAULT 0 CHECK (rating_4_count >= 0),
    rating_5_count INTEGER DEFAULT 0 CHECK (rating_5_count >= 0),
    verified_reviews_count INTEGER DEFAULT 0 CHECK (verified_reviews_count >= 0),
    last_review_date TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Review responses table for merchant/admin responses to reviews
CREATE TABLE IF NOT EXISTS review_responses (
    id VARCHAR(36) PRIMARY KEY,
    review_id VARCHAR(36) NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    responder_id VARCHAR(36) NOT NULL, -- admin or merchant user
    responder_type VARCHAR(20) NOT NULL CHECK (responder_type IN ('admin', 'merchant', 'support')),
    response_text TEXT NOT NULL,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Review templates table for common review response templates
CREATE TABLE IF NOT EXISTS review_response_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    template_text TEXT NOT NULL,
    category VARCHAR(50), -- complaint, praise, question, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_reviews_product_id ON reviews(product_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_order_id ON reviews(order_id);
CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_verified_purchase ON reviews(verified_purchase);
CREATE INDEX idx_reviews_created_at ON reviews(created_at);
CREATE INDEX idx_review_votes_review_id ON review_votes(review_id);
CREATE INDEX idx_review_votes_user_id ON review_votes(user_id);
CREATE INDEX idx_review_reports_review_id ON review_reports(review_id);
CREATE INDEX idx_review_reports_status ON review_reports(status);
CREATE INDEX idx_review_responses_review_id ON review_responses(review_id);
CREATE INDEX idx_review_response_templates_category ON review_response_templates(category);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_review_summaries_updated_at BEFORE UPDATE ON product_review_summaries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_responses_updated_at BEFORE UPDATE ON review_responses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_response_templates_updated_at BEFORE UPDATE ON review_response_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update review vote counts
CREATE OR REPLACE FUNCTION update_review_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.vote_type = 'helpful' THEN
            UPDATE reviews SET helpful_count = helpful_count + 1 WHERE id = NEW.review_id;
        ELSIF NEW.vote_type = 'not_helpful' THEN
            UPDATE reviews SET not_helpful_count = not_helpful_count + 1 WHERE id = NEW.review_id;
        END IF;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle vote type changes
        IF OLD.vote_type = 'helpful' AND NEW.vote_type = 'not_helpful' THEN
            UPDATE reviews SET 
                helpful_count = helpful_count - 1,
                not_helpful_count = not_helpful_count + 1 
            WHERE id = NEW.review_id;
        ELSIF OLD.vote_type = 'not_helpful' AND NEW.vote_type = 'helpful' THEN
            UPDATE reviews SET 
                helpful_count = helpful_count + 1,
                not_helpful_count = not_helpful_count - 1 
            WHERE id = NEW.review_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.vote_type = 'helpful' THEN
            UPDATE reviews SET helpful_count = helpful_count - 1 WHERE id = OLD.review_id;
        ELSIF OLD.vote_type = 'not_helpful' THEN
            UPDATE reviews SET not_helpful_count = not_helpful_count - 1 WHERE id = OLD.review_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER review_vote_counts_trigger 
    AFTER INSERT OR UPDATE OR DELETE ON review_votes
    FOR EACH ROW EXECUTE FUNCTION update_review_vote_counts();

-- Function to update product review summaries
CREATE OR REPLACE FUNCTION update_product_review_summary()
RETURNS TRIGGER AS $$
DECLARE
    product_id_to_update VARCHAR(36);
BEGIN
    -- Determine which product to update
    IF TG_OP = 'DELETE' THEN
        product_id_to_update := OLD.product_id;
    ELSE
        product_id_to_update := NEW.product_id;
    END IF;
    
    -- Only update for approved reviews
    IF (TG_OP = 'INSERT' AND NEW.status = 'approved') OR 
       (TG_OP = 'UPDATE' AND NEW.status = 'approved' AND OLD.status != 'approved') OR
       (TG_OP = 'DELETE' AND OLD.status = 'approved') OR
       (TG_OP = 'UPDATE' AND OLD.status = 'approved' AND NEW.status != 'approved') THEN
        
        -- Upsert product review summary
        INSERT INTO product_review_summaries (
            product_id,
            total_reviews,
            average_rating,
            rating_1_count,
            rating_2_count,
            rating_3_count,
            rating_4_count,
            rating_5_count,
            verified_reviews_count,
            last_review_date,
            updated_at
        )
        SELECT 
            product_id_to_update,
            COUNT(*),
            ROUND(AVG(rating), 2),
            COUNT(*) FILTER (WHERE rating = 1),
            COUNT(*) FILTER (WHERE rating = 2),
            COUNT(*) FILTER (WHERE rating = 3),
            COUNT(*) FILTER (WHERE rating = 4),
            COUNT(*) FILTER (WHERE rating = 5),
            COUNT(*) FILTER (WHERE verified_purchase = TRUE),
            MAX(created_at),
            CURRENT_TIMESTAMP
        FROM reviews 
        WHERE product_id = product_id_to_update AND status = 'approved'
        ON CONFLICT (product_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            rating_1_count = EXCLUDED.rating_1_count,
            rating_2_count = EXCLUDED.rating_2_count,
            rating_3_count = EXCLUDED.rating_3_count,
            rating_4_count = EXCLUDED.rating_4_count,
            rating_5_count = EXCLUDED.rating_5_count,
            verified_reviews_count = EXCLUDED.verified_reviews_count,
            last_review_date = EXCLUDED.last_review_date,
            updated_at = EXCLUDED.updated_at;
    END IF;
    
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ language 'plpgsql';

CREATE TRIGGER product_review_summary_trigger 
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_product_review_summary();