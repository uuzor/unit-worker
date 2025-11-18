-- ============================================================================
-- iKampus Database Schema - Authentication & Users
-- ============================================================================
-- Core authentication and user management tables
-- Services: AuthService, UserService
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE user_role AS ENUM ('student', 'alumni', 'moderator', 'admin');
CREATE TYPE verification_status AS ENUM ('pending', 'approved', 'rejected', 'expired');
CREATE TYPE account_status AS ENUM ('active', 'suspended', 'deleted', 'banned');

-- ============================================================================
-- UNIVERSITIES
-- ============================================================================

CREATE TABLE universities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(100),
    domain VARCHAR(100) NOT NULL UNIQUE, -- e.g., 'ox.ac.uk'
    country VARCHAR(2) NOT NULL, -- ISO 3166-1 alpha-2
    city VARCHAR(100),
    logo_url VARCHAR(500),
    website_url VARCHAR(500),
    verified BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    student_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT universities_domain_check CHECK (domain ~ '^[a-z0-9.-]+$')
);

CREATE INDEX idx_universities_domain ON universities(domain);
CREATE INDEX idx_universities_country ON universities(country);
CREATE INDEX idx_universities_active ON universities(active) WHERE active = TRUE;

COMMENT ON TABLE universities IS 'Verified universities that can use the platform';
COMMENT ON COLUMN universities.domain IS 'Email domain for student verification (e.g., ox.ac.uk)';
COMMENT ON COLUMN universities.verified IS 'Whether university is officially verified';

-- ============================================================================
-- USERS (Authentication)
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email CITEXT NOT NULL UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    email_verification_expires_at TIMESTAMP WITH TIME ZONE,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'student',
    status account_status DEFAULT 'active',
    university_id UUID REFERENCES universities(id) ON DELETE SET NULL,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT users_email_check CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_university_id ON users(university_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_email_verification_token ON users(email_verification_token) WHERE email_verification_token IS NOT NULL;

COMMENT ON TABLE users IS 'Core authentication table for all platform users';
COMMENT ON COLUMN users.email IS 'Case-insensitive email address';
COMMENT ON COLUMN users.locked_until IS 'Account locked due to failed login attempts';

-- ============================================================================
-- USER PROFILES
-- ============================================================================

CREATE TABLE profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100),
    bio TEXT,
    course VARCHAR(150),
    year_of_study INTEGER,
    graduation_year INTEGER,
    avatar_url VARCHAR(500),
    cover_photo_url VARCHAR(500),
    interests TEXT[], -- Array of interest tags
    website_url VARCHAR(500),
    linkedin_url VARCHAR(500),
    github_url VARCHAR(500),
    is_public BOOLEAN DEFAULT TRUE,
    is_searchable BOOLEAN DEFAULT TRUE,
    show_email BOOLEAN DEFAULT FALSE,
    follower_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT profiles_username_check CHECK (username ~ '^[a-zA-Z0-9_-]{3,50}$'),
    CONSTRAINT profiles_year_of_study_check CHECK (year_of_study >= 1 AND year_of_study <= 10),
    CONSTRAINT profiles_graduation_year_check CHECK (graduation_year >= 2000 AND graduation_year <= 2100)
);

CREATE INDEX idx_profiles_username ON profiles(username);
CREATE INDEX idx_profiles_university_course ON profiles USING gin(to_tsvector('english', course));
CREATE INDEX idx_profiles_interests ON profiles USING gin(interests);
CREATE INDEX idx_profiles_searchable ON profiles(is_searchable) WHERE is_searchable = TRUE;

COMMENT ON TABLE profiles IS 'Extended user profile information';
COMMENT ON COLUMN profiles.username IS 'Unique username for the platform';
COMMENT ON COLUMN profiles.interests IS 'Array of interest tags for matching and recommendations';

-- ============================================================================
-- STUDENT VERIFICATION
-- ============================================================================

CREATE TABLE student_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id VARCHAR(50),
    document_url VARCHAR(500), -- S3/GCS URL to uploaded verification document
    verification_method VARCHAR(50) NOT NULL, -- 'email', 'student_id', 'manual'
    status verification_status DEFAULT 'pending',
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE, -- For time-limited verifications
    notes TEXT,

    CONSTRAINT student_verifications_user_id_unique UNIQUE(user_id)
);

CREATE INDEX idx_student_verifications_user_id ON student_verifications(user_id);
CREATE INDEX idx_student_verifications_status ON student_verifications(status);
CREATE INDEX idx_student_verifications_submitted_at ON student_verifications(submitted_at);

COMMENT ON TABLE student_verifications IS 'Student verification records and status';
COMMENT ON COLUMN student_verifications.document_url IS 'Encrypted URL to uploaded verification documents';

-- ============================================================================
-- REFRESH TOKENS (Session Management)
-- ============================================================================

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_id VARCHAR(255),
    device_name VARCHAR(255),
    device_type VARCHAR(50), -- 'ios', 'android', 'web', 'desktop'
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_active ON refresh_tokens(user_id, revoked, expires_at) WHERE revoked = FALSE;

COMMENT ON TABLE refresh_tokens IS 'JWT refresh tokens for session management';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 hash of the refresh token';

-- ============================================================================
-- FOLLOWS (Social Graph)
-- ============================================================================

CREATE TABLE follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT follows_no_self_follow CHECK (follower_id != following_id)
);

CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);
CREATE INDEX idx_follows_created_at ON follows(created_at);

COMMENT ON TABLE follows IS 'User follow relationships';
COMMENT ON CONSTRAINT follows_no_self_follow ON follows IS 'Users cannot follow themselves';

-- ============================================================================
-- BLOCKS (User Blocking)
-- ============================================================================

CREATE TABLE blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (blocker_id, blocked_id),
    CONSTRAINT blocks_no_self_block CHECK (blocker_id != blocked_id)
);

CREATE INDEX idx_blocks_blocker_id ON blocks(blocker_id);
CREATE INDEX idx_blocks_blocked_id ON blocks(blocked_id);

COMMENT ON TABLE blocks IS 'User blocking relationships';

-- ============================================================================
-- PASSWORD RESET TOKENS
-- ============================================================================

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

COMMENT ON TABLE password_reset_tokens IS 'Password reset tokens with expiration';

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE user_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_audit_log_user_id ON user_audit_log(user_id);
CREATE INDEX idx_user_audit_log_action ON user_audit_log(action);
CREATE INDEX idx_user_audit_log_created_at ON user_audit_log(created_at);
CREATE INDEX idx_user_audit_log_metadata ON user_audit_log USING gin(metadata);

COMMENT ON TABLE user_audit_log IS 'Audit trail for security-sensitive actions';
