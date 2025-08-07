-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    order_id UUID NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    moderation_reason TEXT,
    moderated_by UUID,
    moderated_at TIMESTAMP,
    helpful_count INTEGER DEFAULT 0,
    not_helpful_count INTEGER DEFAULT 0,
    verified_purchase BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create review_votes table for helpfulness voting
CREATE TABLE IF NOT EXISTS review_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    vote_type VARCHAR(20) NOT NULL CHECK (vote_type IN ('helpful', 'not_helpful')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- Create review_moderation_logs table
CREATE TABLE IF NOT EXISTS review_moderation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    moderator_id UUID,
    action VARCHAR(50) NOT NULL,
    reason TEXT,
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    automated BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create product_ratings_summary table for caching
CREATE TABLE IF NOT EXISTS product_ratings_summary (
    product_id UUID PRIMARY KEY,
    total_reviews INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.00,
    rating_1_count INTEGER DEFAULT 0,
    rating_2_count INTEGER DEFAULT 0,
    rating_3_count INTEGER DEFAULT 0,
    rating_4_count INTEGER DEFAULT 0,
    rating_5_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create review_content_filters table for moderation rules
CREATE TABLE IF NOT EXISTS review_content_filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filter_type VARCHAR(50) NOT NULL,
    pattern TEXT NOT NULL,
    action VARCHAR(50) NOT NULL DEFAULT 'flag',
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_order_id ON reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at);
CREATE INDEX IF NOT EXISTS idx_reviews_verified_purchase ON reviews(verified_purchase);

CREATE INDEX IF NOT EXISTS idx_review_votes_review_id ON review_votes(review_id);
CREATE INDEX IF NOT EXISTS idx_review_votes_user_id ON review_votes(user_id);

CREATE INDEX IF NOT EXISTS idx_review_moderation_logs_review_id ON review_moderation_logs(review_id);
CREATE INDEX IF NOT EXISTS idx_review_moderation_logs_moderator_id ON review_moderation_logs(moderator_id);

CREATE INDEX IF NOT EXISTS idx_product_ratings_summary_average_rating ON product_ratings_summary(average_rating);

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_content_filters_updated_at BEFORE UPDATE ON review_content_filters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update product ratings summary
CREATE OR REPLACE FUNCTION update_product_ratings_summary()
RETURNS TRIGGER AS $$
BEGIN
    -- Handle INSERT and UPDATE
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        INSERT INTO product_ratings_summary (
            product_id,
            total_reviews,
            average_rating,
            rating_1_count,
            rating_2_count,
            rating_3_count,
            rating_4_count,
            rating_5_count,
            last_updated
        )
        SELECT 
            NEW.product_id,
            COUNT(*),
            ROUND(AVG(rating::DECIMAL), 2),
            COUNT(*) FILTER (WHERE rating = 1),
            COUNT(*) FILTER (WHERE rating = 2),
            COUNT(*) FILTER (WHERE rating = 3),
            COUNT(*) FILTER (WHERE rating = 4),
            COUNT(*) FILTER (WHERE rating = 5),
            CURRENT_TIMESTAMP
        FROM reviews 
        WHERE product_id = NEW.product_id AND status = 'approved'
        ON CONFLICT (product_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            rating_1_count = EXCLUDED.rating_1_count,
            rating_2_count = EXCLUDED.rating_2_count,
            rating_3_count = EXCLUDED.rating_3_count,
            rating_4_count = EXCLUDED.rating_4_count,
            rating_5_count = EXCLUDED.rating_5_count,
            last_updated = EXCLUDED.last_updated;
        
        RETURN NEW;
    END IF;
    
    -- Handle DELETE
    IF TG_OP = 'DELETE' THEN
        INSERT INTO product_ratings_summary (
            product_id,
            total_reviews,
            average_rating,
            rating_1_count,
            rating_2_count,
            rating_3_count,
            rating_4_count,
            rating_5_count,
            last_updated
        )
        SELECT 
            OLD.product_id,
            COUNT(*),
            COALESCE(ROUND(AVG(rating::DECIMAL), 2), 0),
            COUNT(*) FILTER (WHERE rating = 1),
            COUNT(*) FILTER (WHERE rating = 2),
            COUNT(*) FILTER (WHERE rating = 3),
            COUNT(*) FILTER (WHERE rating = 4),
            COUNT(*) FILTER (WHERE rating = 5),
            CURRENT_TIMESTAMP
        FROM reviews 
        WHERE product_id = OLD.product_id AND status = 'approved'
        ON CONFLICT (product_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            rating_1_count = EXCLUDED.rating_1_count,
            rating_2_count = EXCLUDED.rating_2_count,
            rating_3_count = EXCLUDED.rating_3_count,
            rating_4_count = EXCLUDED.rating_4_count,
            rating_5_count = EXCLUDED.rating_5_count,
            last_updated = EXCLUDED.last_updated;
        
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic ratings summary updates
CREATE TRIGGER update_product_ratings_summary_trigger
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW
    WHEN (
        (TG_OP = 'INSERT' AND NEW.status = 'approved') OR
        (TG_OP = 'UPDATE' AND (OLD.status != NEW.status OR OLD.rating != NEW.rating)) OR
        (TG_OP = 'DELETE' AND OLD.status = 'approved')
    )
    EXECUTE FUNCTION update_product_ratings_summary();

-- Insert default content filters
INSERT INTO review_content_filters (filter_type, pattern, action, severity) VALUES
('profanity', '(?i)\b(damn|hell|crap|stupid|idiot)\b', 'flag', 'low'),
('profanity', '(?i)\b(shit|fuck|bitch|asshole)\b', 'reject', 'high'),
('spam', '(?i)\b(buy now|click here|visit our website|www\.|http)\b', 'flag', 'medium'),
('personal_info', '(?i)\b(\d{3}-\d{3}-\d{4}|\d{10}|\w+@\w+\.\w+)\b', 'flag', 'medium'),
('promotional', '(?i)\b(discount|coupon|promo|sale|offer|deal)\b', 'flag', 'low'),
('fake_review', '(?i)\b(fake|paid|sponsored|advertisement)\b', 'flag', 'high')
ON CONFLICT DO NOTHING;
