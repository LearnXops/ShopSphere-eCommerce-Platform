-- Create search analytics table
CREATE TABLE IF NOT EXISTS search_analytics (
    id UUID PRIMARY KEY,
    query TEXT NOT NULL,
    user_id UUID,
    results_count INTEGER NOT NULL DEFAULT 0,
    response_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indices for search analytics
CREATE INDEX IF NOT EXISTS idx_search_analytics_query ON search_analytics(query);
CREATE INDEX IF NOT EXISTS idx_search_analytics_user_id ON search_analytics(user_id);
CREATE INDEX IF NOT EXISTS idx_search_analytics_created_at ON search_analytics(created_at);
CREATE INDEX IF NOT EXISTS idx_search_analytics_query_created_at ON search_analytics(query, created_at);

-- Create search suggestions table for caching popular searches
CREATE TABLE IF NOT EXISTS search_suggestions (
    id UUID PRIMARY KEY,
    suggestion TEXT NOT NULL UNIQUE,
    frequency INTEGER NOT NULL DEFAULT 1,
    last_used TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indices for search suggestions
CREATE INDEX IF NOT EXISTS idx_search_suggestions_suggestion ON search_suggestions(suggestion);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_frequency ON search_suggestions(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_last_used ON search_suggestions(last_used);

-- Create function to update search suggestions
CREATE OR REPLACE FUNCTION update_search_suggestions()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO search_suggestions (id, suggestion, frequency, last_used)
    VALUES (gen_random_uuid(), NEW.query, 1, NEW.created_at)
    ON CONFLICT (suggestion)
    DO UPDATE SET
        frequency = search_suggestions.frequency + 1,
        last_used = NEW.created_at,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search suggestions
CREATE TRIGGER trigger_update_search_suggestions
    AFTER INSERT ON search_analytics
    FOR EACH ROW
    WHEN (NEW.query IS NOT NULL AND NEW.query != '')
    EXECUTE FUNCTION update_search_suggestions();

-- Create materialized view for search analytics dashboard
CREATE MATERIALIZED VIEW IF NOT EXISTS search_analytics_daily AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_searches,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(results_count) as avg_results,
    COUNT(CASE WHEN results_count = 0 THEN 1 END)::float / COUNT(*)::float as zero_results_rate,
    AVG(response_time_ms) as avg_response_time_ms
FROM search_analytics
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Create index on materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_search_analytics_daily_date ON search_analytics_daily(date);

-- Create function to refresh search analytics materialized view
CREATE OR REPLACE FUNCTION refresh_search_analytics_daily()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY search_analytics_daily;
END;
$$ LANGUAGE plpgsql;

-- Add comments
COMMENT ON TABLE search_analytics IS 'Stores search query analytics for performance monitoring and insights';
COMMENT ON TABLE search_suggestions IS 'Caches popular search terms for autocomplete functionality';
COMMENT ON MATERIALIZED VIEW search_analytics_daily IS 'Daily aggregated search analytics for dashboard reporting';