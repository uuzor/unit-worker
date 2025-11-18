-- ============================================================================
-- iKampus Database Schema - Views & Analytics
-- ============================================================================
-- Materialized views, analytics tables, and performance optimizations
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- ANALYTICS TABLES
-- ============================================================================

CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    event_name VARCHAR(100) NOT NULL,
    event_category VARCHAR(50),
    entity_type VARCHAR(50),
    entity_id UUID,
    properties JSONB,
    device_type VARCHAR(50),
    platform VARCHAR(50),
    app_version VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    referrer VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_events_event_name ON analytics_events(event_name);
CREATE INDEX idx_analytics_events_created_at ON analytics_events(created_at DESC);
CREATE INDEX idx_analytics_events_entity ON analytics_events(entity_type, entity_id);
CREATE INDEX idx_analytics_events_properties ON analytics_events USING gin(properties);

-- Partition by month for better performance
CREATE INDEX idx_analytics_events_created_at_month ON analytics_events(
    date_trunc('month', created_at)
);

COMMENT ON TABLE analytics_events IS 'User behavior analytics and event tracking';

-- ============================================================================
-- USER ENGAGEMENT METRICS (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW user_engagement_metrics AS
SELECT
    u.id AS user_id,
    u.created_at AS user_since,
    p.post_count,
    p.follower_count,
    p.following_count,
    COALESCE(post_stats.total_reactions, 0) AS total_post_reactions,
    COALESCE(post_stats.total_comments_received, 0) AS total_comments_received,
    COALESCE(comment_stats.total_comments_made, 0) AS total_comments_made,
    COALESCE(community_stats.communities_joined, 0) AS communities_joined,
    COALESCE(event_stats.events_attending, 0) AS events_attending,
    COALESCE(startup_stats.startups_count, 0) AS startups_count,
    COALESCE(activity_stats.last_active_at, u.created_at) AS last_active_at,
    -- Engagement score calculation
    (
        (p.post_count * 5) +
        (COALESCE(comment_stats.total_comments_made, 0) * 2) +
        (p.follower_count * 3) +
        (p.following_count * 1) +
        (COALESCE(community_stats.communities_joined, 0) * 2) +
        (COALESCE(event_stats.events_attending, 0) * 2)
    )::FLOAT AS engagement_score
FROM users u
JOIN profiles p ON u.id = p.user_id
LEFT JOIN (
    SELECT
        user_id,
        SUM(reaction_count) AS total_reactions,
        SUM(comment_count) AS total_comments_received
    FROM posts
    GROUP BY user_id
) post_stats ON u.id = post_stats.user_id
LEFT JOIN (
    SELECT
        user_id,
        COUNT(*) AS total_comments_made
    FROM comments
    GROUP BY user_id
) comment_stats ON u.id = comment_stats.user_id
LEFT JOIN (
    SELECT
        user_id,
        COUNT(*) AS communities_joined
    FROM community_members
    GROUP BY user_id
) community_stats ON u.id = community_stats.user_id
LEFT JOIN (
    SELECT
        user_id,
        COUNT(*) AS events_attending
    FROM event_attendees
    WHERE status = 'going'
    GROUP BY user_id
) event_stats ON u.id = event_stats.user_id
LEFT JOIN (
    SELECT
        created_by,
        COUNT(*) AS startups_count
    FROM startups
    GROUP BY created_by
) startup_stats ON u.id = startup_stats.created_by
LEFT JOIN (
    SELECT
        user_id,
        MAX(created_at) AS last_active_at
    FROM analytics_events
    GROUP BY user_id
) activity_stats ON u.id = activity_stats.user_id;

CREATE UNIQUE INDEX idx_user_engagement_metrics_user_id ON user_engagement_metrics(user_id);
CREATE INDEX idx_user_engagement_metrics_score ON user_engagement_metrics(engagement_score DESC);

COMMENT ON MATERIALIZED VIEW user_engagement_metrics IS 'Aggregated user engagement metrics (refresh hourly)';

-- ============================================================================
-- TRENDING POSTS (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW trending_posts AS
SELECT
    p.id,
    p.user_id,
    p.community_id,
    p.content,
    p.type,
    p.created_at,
    p.reaction_count,
    p.comment_count,
    p.share_count,
    p.engagement_score,
    u.university_id,
    -- Trending score (weighted by recency and engagement)
    (
        (p.reaction_count * 1.0) +
        (p.comment_count * 2.0) +
        (p.share_count * 3.0)
    ) * EXP(-EXTRACT(EPOCH FROM (NOW() - p.created_at)) / (6 * 3600.0)) AS trending_score
FROM posts p
JOIN users u ON p.user_id = u.id
WHERE p.created_at > NOW() - INTERVAL '7 days'
  AND p.flagged = FALSE
ORDER BY trending_score DESC
LIMIT 1000;

CREATE UNIQUE INDEX idx_trending_posts_id ON trending_posts(id);
CREATE INDEX idx_trending_posts_score ON trending_posts(trending_score DESC);
CREATE INDEX idx_trending_posts_university ON trending_posts(university_id);
CREATE INDEX idx_trending_posts_community ON trending_posts(community_id);

COMMENT ON MATERIALIZED VIEW trending_posts IS 'Trending posts (refresh every 15 minutes)';

-- ============================================================================
-- POPULAR HASHTAGS (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW popular_hashtags AS
SELECT
    h.tag,
    h.post_count,
    h.last_used_at,
    COUNT(DISTINCT p.user_id) AS unique_users,
    COUNT(DISTINCT p.community_id) AS communities_used,
    -- Popularity score (weighted by recency)
    h.post_count * EXP(-EXTRACT(EPOCH FROM (NOW() - h.last_used_at)) / (24 * 3600.0)) AS popularity_score
FROM hashtags h
LEFT JOIN posts p ON h.tag = ANY(p.hashtags)
WHERE h.last_used_at > NOW() - INTERVAL '30 days'
GROUP BY h.tag, h.post_count, h.last_used_at
ORDER BY popularity_score DESC
LIMIT 500;

CREATE UNIQUE INDEX idx_popular_hashtags_tag ON popular_hashtags(tag);
CREATE INDEX idx_popular_hashtags_score ON popular_hashtags(popularity_score DESC);

COMMENT ON MATERIALIZED VIEW popular_hashtags IS 'Trending hashtags (refresh hourly)';

-- ============================================================================
-- ACTIVE COMMUNITIES (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW active_communities AS
SELECT
    c.id,
    c.name,
    c.slug,
    c.type,
    c.university_id,
    c.member_count,
    c.post_count,
    COALESCE(recent_activity.posts_last_7_days, 0) AS posts_last_7_days,
    COALESCE(recent_activity.active_members_last_7_days, 0) AS active_members_last_7_days,
    COALESCE(recent_activity.last_post_at, c.created_at) AS last_post_at,
    -- Activity score
    (
        (c.member_count * 0.5) +
        (COALESCE(recent_activity.posts_last_7_days, 0) * 10.0) +
        (COALESCE(recent_activity.active_members_last_7_days, 0) * 5.0)
    ) AS activity_score
FROM communities c
LEFT JOIN (
    SELECT
        community_id,
        COUNT(*) AS posts_last_7_days,
        COUNT(DISTINCT user_id) AS active_members_last_7_days,
        MAX(created_at) AS last_post_at
    FROM posts
    WHERE created_at > NOW() - INTERVAL '7 days'
    GROUP BY community_id
) recent_activity ON c.id = recent_activity.community_id
WHERE c.is_private = FALSE
ORDER BY activity_score DESC;

CREATE UNIQUE INDEX idx_active_communities_id ON active_communities(id);
CREATE INDEX idx_active_communities_score ON active_communities(activity_score DESC);
CREATE INDEX idx_active_communities_university ON active_communities(university_id);

COMMENT ON MATERIALIZED VIEW active_communities IS 'Most active communities (refresh hourly)';

-- ============================================================================
-- STARTUP DIRECTORY (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW startup_directory AS
SELECT
    s.id,
    s.name,
    s.slug,
    s.tagline,
    s.stage,
    s.industry,
    s.categories,
    s.logo_url,
    s.university_id,
    s.follower_count,
    s.like_count,
    s.team_size,
    s.looking_for_cofounders,
    s.created_at,
    u.name AS university_name,
    COALESCE(team_data.founder_names, ARRAY[]::TEXT[]) AS founder_names,
    COALESCE(update_data.last_update_at, s.created_at) AS last_update_at,
    -- Visibility score (for featured/recommended)
    (
        (s.follower_count * 2.0) +
        (s.like_count * 1.0) +
        (s.team_size * 5.0) +
        CASE WHEN s.is_verified THEN 50.0 ELSE 0.0 END
    ) AS visibility_score
FROM startups s
LEFT JOIN universities u ON s.university_id = u.id
LEFT JOIN (
    SELECT
        sm.startup_id,
        ARRAY_AGG(p.username) AS founder_names
    FROM startup_members sm
    JOIN profiles p ON sm.user_id = p.user_id
    WHERE sm.role IN ('founder', 'co_founder')
      AND sm.is_active = TRUE
    GROUP BY sm.startup_id
) team_data ON s.id = team_data.startup_id
LEFT JOIN (
    SELECT
        startup_id,
        MAX(created_at) AS last_update_at
    FROM startup_updates
    GROUP BY startup_id
) update_data ON s.id = update_data.startup_id
WHERE s.is_public = TRUE
ORDER BY visibility_score DESC;

CREATE UNIQUE INDEX idx_startup_directory_id ON startup_directory(id);
CREATE INDEX idx_startup_directory_score ON startup_directory(visibility_score DESC);
CREATE INDEX idx_startup_directory_university ON startup_directory(university_id);
CREATE INDEX idx_startup_directory_stage ON startup_directory(stage);
CREATE INDEX idx_startup_directory_industry ON startup_directory(industry);

COMMENT ON MATERIALIZED VIEW startup_directory IS 'Searchable startup directory (refresh every 6 hours)';

-- ============================================================================
-- DAILY STATISTICS (Aggregate Table)
-- ============================================================================

CREATE TABLE daily_statistics (
    stat_date DATE PRIMARY KEY,
    total_users INTEGER DEFAULT 0,
    new_users INTEGER DEFAULT 0,
    active_users INTEGER DEFAULT 0, -- Users who logged in or posted
    total_posts INTEGER DEFAULT 0,
    new_posts INTEGER DEFAULT 0,
    total_comments INTEGER DEFAULT 0,
    new_comments INTEGER DEFAULT 0,
    total_communities INTEGER DEFAULT 0,
    new_communities INTEGER DEFAULT 0,
    total_startups INTEGER DEFAULT 0,
    new_startups INTEGER DEFAULT 0,
    total_events INTEGER DEFAULT 0,
    new_events INTEGER DEFAULT 0,
    engagement_metrics JSONB, -- Additional detailed metrics
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_daily_statistics_date ON daily_statistics(stat_date DESC);

COMMENT ON TABLE daily_statistics IS 'Daily aggregated platform statistics';

-- ============================================================================
-- UNIVERSITY STATISTICS (Materialized View)
-- ============================================================================

CREATE MATERIALIZED VIEW university_statistics AS
SELECT
    u.id,
    u.name,
    u.domain,
    u.country,
    u.student_count,
    COALESCE(community_data.community_count, 0) AS community_count,
    COALESCE(post_data.post_count, 0) AS post_count,
    COALESCE(startup_data.startup_count, 0) AS startup_count,
    COALESCE(event_data.event_count, 0) AS event_count,
    COALESCE(active_data.active_users_last_7_days, 0) AS active_users_last_7_days,
    -- University engagement score
    (
        (u.student_count * 1.0) +
        (COALESCE(community_data.community_count, 0) * 5.0) +
        (COALESCE(post_data.post_count, 0) * 0.1) +
        (COALESCE(startup_data.startup_count, 0) * 10.0)
    ) AS engagement_score
FROM universities u
LEFT JOIN (
    SELECT university_id, COUNT(*) AS community_count
    FROM communities
    GROUP BY university_id
) community_data ON u.id = community_data.university_id
LEFT JOIN (
    SELECT users.university_id, COUNT(posts.*) AS post_count
    FROM posts
    JOIN users ON posts.user_id = users.id
    GROUP BY users.university_id
) post_data ON u.id = post_data.university_id
LEFT JOIN (
    SELECT university_id, COUNT(*) AS startup_count
    FROM startups
    GROUP BY university_id
) startup_data ON u.id = startup_data.university_id
LEFT JOIN (
    SELECT university_id, COUNT(*) AS event_count
    FROM events
    JOIN communities ON events.community_id = communities.id
    GROUP BY communities.university_id
) event_data ON u.id = event_data.university_id
LEFT JOIN (
    SELECT
        u2.university_id,
        COUNT(DISTINCT ae.user_id) AS active_users_last_7_days
    FROM analytics_events ae
    JOIN users u2 ON ae.user_id = u2.id
    WHERE ae.created_at > NOW() - INTERVAL '7 days'
    GROUP BY u2.university_id
) active_data ON u.id = active_data.university_id
WHERE u.active = TRUE
ORDER BY engagement_score DESC;

CREATE UNIQUE INDEX idx_university_statistics_id ON university_statistics(id);
CREATE INDEX idx_university_statistics_score ON university_statistics(engagement_score DESC);

COMMENT ON MATERIALIZED VIEW university_statistics IS 'University engagement statistics (refresh daily)';

-- ============================================================================
-- REFRESH FUNCTIONS
-- ============================================================================

-- Function to refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_materialized_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY user_engagement_metrics;
    REFRESH MATERIALIZED VIEW CONCURRENTLY trending_posts;
    REFRESH MATERIALIZED VIEW CONCURRENTLY popular_hashtags;
    REFRESH MATERIALIZED VIEW CONCURRENTLY active_communities;
    REFRESH MATERIALIZED VIEW CONCURRENTLY startup_directory;
    REFRESH MATERIALIZED VIEW CONCURRENTLY university_statistics;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_materialized_views() IS 'Refresh all materialized views (run via cron)';

-- ============================================================================
-- ANALYTICS HELPER FUNCTIONS
-- ============================================================================

-- Get user activity summary
CREATE OR REPLACE FUNCTION get_user_activity_summary(p_user_id UUID, p_days INTEGER DEFAULT 30)
RETURNS TABLE (
    posts_created BIGINT,
    comments_made BIGINT,
    reactions_given BIGINT,
    events_attended BIGINT,
    communities_joined BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        (SELECT COUNT(*) FROM posts WHERE user_id = p_user_id AND created_at > NOW() - (p_days || ' days')::INTERVAL),
        (SELECT COUNT(*) FROM comments WHERE user_id = p_user_id AND created_at > NOW() - (p_days || ' days')::INTERVAL),
        (SELECT COUNT(*) FROM post_reactions WHERE user_id = p_user_id AND created_at > NOW() - (p_days || ' days')::INTERVAL),
        (SELECT COUNT(*) FROM event_attendees WHERE user_id = p_user_id AND status = 'going' AND created_at > NOW() - (p_days || ' days')::INTERVAL),
        (SELECT COUNT(*) FROM community_members WHERE user_id = p_user_id AND joined_at > NOW() - (p_days || ' days')::INTERVAL);
END;
$$ LANGUAGE plpgsql;

-- Get startup performance metrics
CREATE OR REPLACE FUNCTION get_startup_metrics(p_startup_id UUID)
RETURNS TABLE (
    total_views BIGINT,
    total_followers BIGINT,
    total_likes BIGINT,
    team_size INTEGER,
    updates_count BIGINT,
    last_update_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.view_count::BIGINT,
        s.follower_count::BIGINT,
        s.like_count::BIGINT,
        s.team_size,
        (SELECT COUNT(*) FROM startup_updates WHERE startup_id = p_startup_id),
        (SELECT MAX(created_at) FROM startup_updates WHERE startup_id = p_startup_id)
    FROM startups s
    WHERE s.id = p_startup_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_user_activity_summary IS 'Get user activity metrics for a time period';
COMMENT ON FUNCTION get_startup_metrics IS 'Get comprehensive startup metrics';
