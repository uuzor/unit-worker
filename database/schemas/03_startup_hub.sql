-- ============================================================================
-- iKampus Database Schema - Startup Hub
-- ============================================================================
-- Startups, teams, pitches, investor connections, co-founder matching
-- Services: StartupService, MatchmakingService, MentorshipService
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE startup_stage AS ENUM ('idea', 'mvp', 'beta', 'launched', 'scaling', 'funded');
CREATE TYPE funding_stage AS ENUM ('pre_seed', 'seed', 'series_a', 'series_b', 'series_c', 'bootstrapped');
CREATE TYPE team_member_role AS ENUM ('founder', 'co_founder', 'technical_lead', 'designer', 'marketing', 'advisor', 'intern');
CREATE TYPE connection_status AS ENUM ('pending', 'accepted', 'rejected', 'withdrawn');
CREATE TYPE match_status AS ENUM ('suggested', 'interested', 'matched', 'declined');

-- ============================================================================
-- STARTUPS
-- ============================================================================

CREATE TABLE startups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    tagline VARCHAR(255),
    description TEXT,
    problem_statement TEXT,
    solution TEXT,
    target_market TEXT,
    stage startup_stage DEFAULT 'idea',
    funding_stage funding_stage DEFAULT 'bootstrapped',
    industry VARCHAR(100),
    categories TEXT[], -- e.g., ['edtech', 'saas', 'mobile']
    website_url VARCHAR(500),
    demo_url VARCHAR(500),
    pitch_deck_url VARCHAR(500),
    video_url VARCHAR(500), -- Pitch video
    logo_url VARCHAR(500),
    cover_image_url VARCHAR(500),
    university_id UUID REFERENCES universities(id) ON DELETE SET NULL,
    founded_date DATE,
    team_size INTEGER DEFAULT 1,
    looking_for_cofounders BOOLEAN DEFAULT FALSE,
    looking_for_roles TEXT[], -- Roles seeking
    is_public BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE, -- Verified by platform
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    follower_count INTEGER DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT startups_slug_check CHECK (slug ~ '^[a-z0-9-]+$'),
    CONSTRAINT startups_team_size_check CHECK (team_size > 0)
);

CREATE INDEX idx_startups_slug ON startups(slug);
CREATE INDEX idx_startups_stage ON startups(stage);
CREATE INDEX idx_startups_industry ON startups(industry);
CREATE INDEX idx_startups_categories ON startups USING gin(categories);
CREATE INDEX idx_startups_university_id ON startups(university_id);
CREATE INDEX idx_startups_looking_for_cofounders ON startups(looking_for_cofounders) WHERE looking_for_cofounders = TRUE;
CREATE INDEX idx_startups_is_public ON startups(is_public) WHERE is_public = TRUE;
CREATE INDEX idx_startups_search ON startups USING gin(
    to_tsvector('english', name || ' ' || COALESCE(tagline, '') || ' ' || COALESCE(description, ''))
);

COMMENT ON TABLE startups IS 'Student startup profiles';
COMMENT ON COLUMN startups.looking_for_cofounders IS 'Whether startup is actively seeking co-founders';

-- ============================================================================
-- STARTUP TEAM MEMBERS
-- ============================================================================

CREATE TABLE startup_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role team_member_role NOT NULL,
    title VARCHAR(100), -- e.g., "CTO", "Lead Engineer"
    equity_percentage DECIMAL(5, 2), -- e.g., 25.50 for 25.5%
    is_active BOOLEAN DEFAULT TRUE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT startup_members_unique UNIQUE(startup_id, user_id),
    CONSTRAINT startup_members_equity_check CHECK (equity_percentage >= 0 AND equity_percentage <= 100)
);

CREATE INDEX idx_startup_members_startup_id ON startup_members(startup_id);
CREATE INDEX idx_startup_members_user_id ON startup_members(user_id);
CREATE INDEX idx_startup_members_role ON startup_members(role);
CREATE INDEX idx_startup_members_active ON startup_members(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE startup_members IS 'Startup team members and their roles';

-- ============================================================================
-- PITCHES
-- ============================================================================

CREATE TABLE pitches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    pitch_deck_url VARCHAR(500),
    video_url VARCHAR(500),
    slides_url VARCHAR(500),
    ask_amount DECIMAL(15, 2), -- Funding amount seeking
    currency VARCHAR(3) DEFAULT 'USD',
    valuation DECIMAL(15, 2), -- Current valuation
    use_of_funds TEXT,
    milestones TEXT[],
    traction_metrics JSONB, -- {revenue: 10000, users: 5000, mrr: 2000}
    is_public BOOLEAN DEFAULT TRUE,
    view_count INTEGER DEFAULT 0,
    submitted_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT pitches_ask_amount_check CHECK (ask_amount > 0),
    CONSTRAINT pitches_valuation_check CHECK (valuation >= 0)
);

CREATE INDEX idx_pitches_startup_id ON pitches(startup_id);
CREATE INDEX idx_pitches_submitted_by ON pitches(submitted_by);
CREATE INDEX idx_pitches_created_at ON pitches(created_at DESC);
CREATE INDEX idx_pitches_traction ON pitches USING gin(traction_metrics);

COMMENT ON TABLE pitches IS 'Startup pitch decks and presentations';
COMMENT ON COLUMN pitches.traction_metrics IS 'JSON object with key metrics (revenue, users, etc.)';

-- ============================================================================
-- CO-FOUNDER MATCHING
-- ============================================================================

CREATE TABLE cofounder_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_seeking BOOLEAN DEFAULT TRUE,
    roles_seeking TEXT[], -- e.g., ['technical_cofounder', 'business_cofounder']
    skills TEXT[], -- e.g., ['python', 'react', 'machine_learning']
    industries_interested TEXT[], -- e.g., ['fintech', 'healthtech', 'edtech']
    commitment_level VARCHAR(50), -- 'full_time', 'part_time', 'weekends'
    equity_expectations VARCHAR(100),
    location_preference VARCHAR(100),
    bio TEXT,
    portfolio_url VARCHAR(500),
    resume_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_cofounder_profiles_is_seeking ON cofounder_profiles(is_seeking) WHERE is_seeking = TRUE;
CREATE INDEX idx_cofounder_profiles_roles ON cofounder_profiles USING gin(roles_seeking);
CREATE INDEX idx_cofounder_profiles_skills ON cofounder_profiles USING gin(skills);
CREATE INDEX idx_cofounder_profiles_industries ON cofounder_profiles USING gin(industries_interested);

COMMENT ON TABLE cofounder_profiles IS 'Profiles for users seeking co-founders';

CREATE TABLE cofounder_matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_score FLOAT, -- 0.0 to 1.0 from AI matching algorithm
    status match_status DEFAULT 'suggested',
    initiated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT cofounder_matches_unique UNIQUE(user1_id, user2_id),
    CONSTRAINT cofounder_matches_no_self_match CHECK (user1_id != user2_id),
    CONSTRAINT cofounder_matches_score_check CHECK (match_score >= 0 AND match_score <= 1)
);

CREATE INDEX idx_cofounder_matches_user1_id ON cofounder_matches(user1_id);
CREATE INDEX idx_cofounder_matches_user2_id ON cofounder_matches(user2_id);
CREATE INDEX idx_cofounder_matches_status ON cofounder_matches(status);
CREATE INDEX idx_cofounder_matches_score ON cofounder_matches(match_score DESC);

COMMENT ON TABLE cofounder_matches IS 'AI-powered co-founder matching';

-- ============================================================================
-- ALUMNI & MENTORSHIP
-- ============================================================================

CREATE TABLE mentor_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_available BOOLEAN DEFAULT TRUE,
    expertise_areas TEXT[], -- e.g., ['product_management', 'fundraising', 'marketing']
    industries TEXT[], -- Industries they can help with
    company VARCHAR(255),
    job_title VARCHAR(255),
    years_experience INTEGER,
    linkedin_url VARCHAR(500),
    bio TEXT,
    mentorship_capacity INTEGER DEFAULT 5, -- Max mentees
    current_mentees INTEGER DEFAULT 0,
    session_price DECIMAL(10, 2), -- Price per session (0 = free)
    currency VARCHAR(3) DEFAULT 'USD',
    availability JSONB, -- Schedule availability
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT mentor_profiles_experience_check CHECK (years_experience >= 0),
    CONSTRAINT mentor_profiles_capacity_check CHECK (mentorship_capacity > 0),
    CONSTRAINT mentor_profiles_price_check CHECK (session_price >= 0)
);

CREATE INDEX idx_mentor_profiles_is_available ON mentor_profiles(is_available) WHERE is_available = TRUE;
CREATE INDEX idx_mentor_profiles_expertise ON mentor_profiles USING gin(expertise_areas);
CREATE INDEX idx_mentor_profiles_industries ON mentor_profiles USING gin(industries);

COMMENT ON TABLE mentor_profiles IS 'Alumni and professional mentors';

CREATE TYPE mentorship_status AS ENUM ('requested', 'accepted', 'active', 'completed', 'cancelled');

CREATE TABLE mentorship_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mentor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mentee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    startup_id UUID REFERENCES startups(id) ON DELETE SET NULL, -- Optional startup context
    status mentorship_status DEFAULT 'requested',
    message TEXT,
    focus_areas TEXT[],
    session_count INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT mentorship_connections_no_self_mentor CHECK (mentor_id != mentee_id)
);

CREATE INDEX idx_mentorship_connections_mentor_id ON mentorship_connections(mentor_id);
CREATE INDEX idx_mentorship_connections_mentee_id ON mentorship_connections(mentee_id);
CREATE INDEX idx_mentorship_connections_startup_id ON mentorship_connections(startup_id);
CREATE INDEX idx_mentorship_connections_status ON mentorship_connections(status);

COMMENT ON TABLE mentorship_connections IS 'Mentorship relationships';

CREATE TABLE mentorship_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connection_id UUID NOT NULL REFERENCES mentorship_connections(id) ON DELETE CASCADE,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_minutes INTEGER DEFAULT 60,
    location VARCHAR(255), -- 'zoom', 'in_person', etc.
    meeting_link VARCHAR(500),
    agenda TEXT,
    notes TEXT,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT mentorship_sessions_duration_check CHECK (duration_minutes > 0)
);

CREATE INDEX idx_mentorship_sessions_connection_id ON mentorship_sessions(connection_id);
CREATE INDEX idx_mentorship_sessions_scheduled_at ON mentorship_sessions(scheduled_at);
CREATE INDEX idx_mentorship_sessions_completed ON mentorship_sessions(completed);

COMMENT ON TABLE mentorship_sessions IS 'Scheduled mentorship sessions';

-- ============================================================================
-- INVESTOR CONNECTIONS
-- ============================================================================

CREATE TABLE investor_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    investor_type VARCHAR(50), -- 'angel', 'vc', 'fund', 'corporate'
    fund_name VARCHAR(255),
    investment_range_min DECIMAL(15, 2),
    investment_range_max DECIMAL(15, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    industries_focus TEXT[],
    stages_focus startup_stage[],
    geography_focus TEXT[], -- Countries/regions
    portfolio_companies TEXT[],
    notable_investments TEXT[],
    website_url VARCHAR(500),
    linkedin_url VARCHAR(500),
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_investor_profiles_is_active ON investor_profiles(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_investor_profiles_industries ON investor_profiles USING gin(industries_focus);
CREATE INDEX idx_investor_profiles_stages ON investor_profiles USING gin(stages_focus);

COMMENT ON TABLE investor_profiles IS 'Alumni investors and VCs';

CREATE TABLE startup_investor_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    investor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status connection_status DEFAULT 'pending',
    initiated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    pitch_id UUID REFERENCES pitches(id) ON DELETE SET NULL,
    message TEXT,
    meeting_scheduled BOOLEAN DEFAULT FALSE,
    meeting_date TIMESTAMP WITH TIME ZONE,
    investment_amount DECIMAL(15, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT startup_investor_connections_unique UNIQUE(startup_id, investor_id)
);

CREATE INDEX idx_startup_investor_connections_startup_id ON startup_investor_connections(startup_id);
CREATE INDEX idx_startup_investor_connections_investor_id ON startup_investor_connections(investor_id);
CREATE INDEX idx_startup_investor_connections_status ON startup_investor_connections(status);

COMMENT ON TABLE startup_investor_connections IS 'Connections between startups and investors';

-- ============================================================================
-- STARTUP UPDATES
-- ============================================================================

CREATE TABLE startup_updates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    update_type VARCHAR(50), -- 'milestone', 'funding', 'launch', 'team', 'general'
    is_public BOOLEAN DEFAULT TRUE,
    media_urls TEXT[],
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_startup_updates_startup_id ON startup_updates(startup_id);
CREATE INDEX idx_startup_updates_created_at ON startup_updates(created_at DESC);
CREATE INDEX idx_startup_updates_type ON startup_updates(update_type);

COMMENT ON TABLE startup_updates IS 'Progress updates from startups';

-- ============================================================================
-- STARTUP FOLLOWERS
-- ============================================================================

CREATE TABLE startup_followers (
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (startup_id, user_id)
);

CREATE INDEX idx_startup_followers_startup_id ON startup_followers(startup_id);
CREATE INDEX idx_startup_followers_user_id ON startup_followers(user_id);

COMMENT ON TABLE startup_followers IS 'Users following startups for updates';

-- ============================================================================
-- STARTUP LIKES
-- ============================================================================

CREATE TABLE startup_likes (
    startup_id UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (startup_id, user_id)
);

CREATE INDEX idx_startup_likes_startup_id ON startup_likes(startup_id);
CREATE INDEX idx_startup_likes_user_id ON startup_likes(user_id);

COMMENT ON TABLE startup_likes IS 'User likes/endorsements for startups';
