-- ============================================================================
-- iKampus Database Schema - Social Features
-- ============================================================================
-- Posts, comments, reactions, communities, events
-- Services: PostService, FeedService, CommunityService, EventService
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE post_visibility AS ENUM ('public', 'university', 'community', 'followers', 'private');
CREATE TYPE post_type AS ENUM ('text', 'image', 'video', 'poll', 'confession');
CREATE TYPE reaction_type AS ENUM ('like', 'love', 'celebrate', 'insightful', 'support', 'funny');
CREATE TYPE community_type AS ENUM ('course', 'society', 'hall', 'interest', 'official');
CREATE TYPE member_role AS ENUM ('member', 'moderator', 'admin', 'owner');
CREATE TYPE event_status AS ENUM ('draft', 'published', 'cancelled', 'completed');

-- ============================================================================
-- COMMUNITIES
-- ============================================================================

CREATE TABLE communities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,
    description TEXT,
    type community_type NOT NULL,
    university_id UUID REFERENCES universities(id) ON DELETE CASCADE,
    avatar_url VARCHAR(500),
    cover_photo_url VARCHAR(500),
    rules TEXT[],
    member_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    is_private BOOLEAN DEFAULT FALSE,
    requires_approval BOOLEAN DEFAULT FALSE,
    is_official BOOLEAN DEFAULT FALSE, -- Official university communities
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT communities_slug_check CHECK (slug ~ '^[a-z0-9-]+$')
);

CREATE INDEX idx_communities_slug ON communities(slug);
CREATE INDEX idx_communities_university_id ON communities(university_id);
CREATE INDEX idx_communities_type ON communities(type);
CREATE INDEX idx_communities_is_private ON communities(is_private);
CREATE INDEX idx_communities_name_search ON communities USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

COMMENT ON TABLE communities IS 'Groups, societies, and communities within universities';
COMMENT ON COLUMN communities.is_official IS 'Verified official university communities';

-- ============================================================================
-- COMMUNITY MEMBERS
-- ============================================================================

CREATE TABLE community_members (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,

    PRIMARY KEY (community_id, user_id)
);

CREATE INDEX idx_community_members_user_id ON community_members(user_id);
CREATE INDEX idx_community_members_community_id ON community_members(community_id);
CREATE INDEX idx_community_members_role ON community_members(role);

COMMENT ON TABLE community_members IS 'Community membership and roles';

-- ============================================================================
-- POSTS
-- ============================================================================

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    community_id UUID REFERENCES communities(id) ON DELETE CASCADE,
    type post_type DEFAULT 'text',
    visibility post_visibility DEFAULT 'public',
    content TEXT NOT NULL,
    media_urls TEXT[], -- Array of S3/GCS URLs for images/videos
    hashtags TEXT[], -- Extracted hashtags for search
    mentions UUID[], -- Array of mentioned user IDs
    is_anonymous BOOLEAN DEFAULT FALSE,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    location GEOGRAPHY(POINT, 4326), -- Campus location (PostGIS)
    location_name VARCHAR(255), -- Human-readable location
    reaction_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    engagement_score FLOAT DEFAULT 0.0, -- Calculated score for ranking
    flagged BOOLEAN DEFAULT FALSE,
    flagged_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT posts_content_check CHECK (char_length(content) <= 5000),
    CONSTRAINT posts_anonymous_community_check CHECK (
        NOT (is_anonymous = TRUE AND community_id IS NOT NULL)
    )
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_community_id ON posts(community_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_engagement_score ON posts(engagement_score DESC);
CREATE INDEX idx_posts_hashtags ON posts USING gin(hashtags);
CREATE INDEX idx_posts_mentions ON posts USING gin(mentions);
CREATE INDEX idx_posts_visibility ON posts(visibility);
CREATE INDEX idx_posts_flagged ON posts(flagged) WHERE flagged = TRUE;
CREATE INDEX idx_posts_content_search ON posts USING gin(to_tsvector('english', content));
CREATE INDEX idx_posts_location ON posts USING gist(location) WHERE location IS NOT NULL;

COMMENT ON TABLE posts IS 'User-generated posts across the platform';
COMMENT ON COLUMN posts.engagement_score IS 'Calculated score for feed ranking algorithm';
COMMENT ON COLUMN posts.is_anonymous IS 'Anonymous posts (confessions)';
COMMENT ON COLUMN posts.location IS 'Geographic location using PostGIS';

-- ============================================================================
-- COMMENTS
-- ============================================================================

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    reaction_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    flagged BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT comments_content_check CHECK (char_length(content) <= 2000),
    CONSTRAINT comments_no_deep_nesting CHECK (
        parent_comment_id IS NULL OR
        (SELECT parent_comment_id FROM comments WHERE id = parent_comment_id) IS NULL
    )
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_comment_id ON comments(parent_comment_id);
CREATE INDEX idx_comments_created_at ON comments(created_at DESC);

COMMENT ON TABLE comments IS 'Comments on posts (max 2 levels deep)';
COMMENT ON CONSTRAINT comments_no_deep_nesting ON comments IS 'Limit comment nesting to 2 levels';

-- ============================================================================
-- REACTIONS (Posts and Comments)
-- ============================================================================

CREATE TABLE post_reactions (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type reaction_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_post_reactions_post_id ON post_reactions(post_id);
CREATE INDEX idx_post_reactions_user_id ON post_reactions(user_id);
CREATE INDEX idx_post_reactions_type ON post_reactions(reaction_type);

COMMENT ON TABLE post_reactions IS 'User reactions to posts (one per user per post)';

CREATE TABLE comment_reactions (
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type reaction_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX idx_comment_reactions_comment_id ON comment_reactions(comment_id);
CREATE INDEX idx_comment_reactions_user_id ON comment_reactions(user_id);

COMMENT ON TABLE comment_reactions IS 'User reactions to comments';

-- ============================================================================
-- SAVED POSTS
-- ============================================================================

CREATE TABLE saved_posts (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    collection_name VARCHAR(100), -- Optional collections/folders
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX idx_saved_posts_user_id ON saved_posts(user_id);
CREATE INDEX idx_saved_posts_post_id ON saved_posts(post_id);
CREATE INDEX idx_saved_posts_collection ON saved_posts(collection_name) WHERE collection_name IS NOT NULL;

COMMENT ON TABLE saved_posts IS 'Posts saved/bookmarked by users';

-- ============================================================================
-- POLLS
-- ============================================================================

CREATE TABLE polls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL UNIQUE REFERENCES posts(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    allow_multiple BOOLEAN DEFAULT FALSE,
    total_votes INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_polls_post_id ON polls(post_id);
CREATE INDEX idx_polls_expires_at ON polls(expires_at);

COMMENT ON TABLE polls IS 'Poll data for poll-type posts';

CREATE TABLE poll_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    option_text VARCHAR(255) NOT NULL,
    vote_count INTEGER DEFAULT 0,
    display_order INTEGER NOT NULL,

    CONSTRAINT poll_options_unique_order UNIQUE(poll_id, display_order)
);

CREATE INDEX idx_poll_options_poll_id ON poll_options(poll_id);

COMMENT ON TABLE poll_options IS 'Options for polls';

CREATE TABLE poll_votes (
    poll_option_id UUID NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (poll_option_id, user_id)
);

CREATE INDEX idx_poll_votes_poll_option_id ON poll_votes(poll_option_id);
CREATE INDEX idx_poll_votes_user_id ON poll_votes(user_id);

COMMENT ON TABLE poll_votes IS 'User votes on poll options';

-- ============================================================================
-- EVENTS
-- ============================================================================

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID REFERENCES communities(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location_name VARCHAR(255),
    location GEOGRAPHY(POINT, 4326),
    location_details TEXT, -- Room number, building, etc.
    cover_image_url VARCHAR(500),
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status event_status DEFAULT 'published',
    is_online BOOLEAN DEFAULT FALSE,
    online_link VARCHAR(500),
    capacity INTEGER,
    attendee_count INTEGER DEFAULT 0,
    waitlist_count INTEGER DEFAULT 0,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT events_dates_check CHECK (ends_at > starts_at),
    CONSTRAINT events_capacity_check CHECK (capacity IS NULL OR capacity > 0)
);

CREATE INDEX idx_events_community_id ON events(community_id);
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_starts_at ON events(starts_at);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_tags ON events USING gin(tags);
CREATE INDEX idx_events_location ON events USING gist(location) WHERE location IS NOT NULL;

COMMENT ON TABLE events IS 'University and community events';

CREATE TYPE rsvp_status AS ENUM ('going', 'interested', 'not_going', 'waitlist');

CREATE TABLE event_attendees (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status rsvp_status DEFAULT 'going',
    checked_in BOOLEAN DEFAULT FALSE,
    checked_in_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX idx_event_attendees_event_id ON event_attendees(event_id);
CREATE INDEX idx_event_attendees_user_id ON event_attendees(user_id);
CREATE INDEX idx_event_attendees_status ON event_attendees(status);

COMMENT ON TABLE event_attendees IS 'Event RSVPs and attendance tracking';

-- ============================================================================
-- HASHTAGS (Aggregate Table)
-- ============================================================================

CREATE TABLE hashtags (
    tag VARCHAR(100) PRIMARY KEY,
    post_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT hashtags_tag_check CHECK (tag ~ '^[a-zA-Z0-9_]+$')
);

CREATE INDEX idx_hashtags_post_count ON hashtags(post_count DESC);
CREATE INDEX idx_hashtags_last_used_at ON hashtags(last_used_at DESC);

COMMENT ON TABLE hashtags IS 'Aggregated hashtag statistics';

-- ============================================================================
-- REPORTS (Content Moderation)
-- ============================================================================

CREATE TYPE report_type AS ENUM ('spam', 'harassment', 'hate_speech', 'misinformation', 'inappropriate', 'other');
CREATE TYPE report_status AS ENUM ('pending', 'reviewed', 'actioned', 'dismissed');

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(50) NOT NULL, -- 'post', 'comment', 'user', 'community'
    entity_id UUID NOT NULL,
    report_type report_type NOT NULL,
    description TEXT,
    status report_status DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    action_taken TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT reports_entity_check CHECK (entity_type IN ('post', 'comment', 'user', 'community', 'event'))
);

CREATE INDEX idx_reports_reporter_id ON reports(reporter_id);
CREATE INDEX idx_reports_reported_user_id ON reports(reported_user_id);
CREATE INDEX idx_reports_entity ON reports(entity_type, entity_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_created_at ON reports(created_at);

COMMENT ON TABLE reports IS 'User-reported content for moderation';
