-- ============================================================================
-- iKampus Database Schema - Messaging & Notifications
-- ============================================================================
-- Direct messaging, group chats, push notifications
-- Services: MessagingService, NotificationService
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE conversation_type AS ENUM ('direct', 'group');
CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read');
CREATE TYPE notification_type AS ENUM (
    'follow',
    'post_like',
    'post_comment',
    'comment_reply',
    'mention',
    'message',
    'event_reminder',
    'startup_update',
    'cofounder_match',
    'mentorship_request',
    'investor_interest',
    'system'
);
CREATE TYPE notification_priority AS ENUM ('low', 'medium', 'high', 'urgent');

-- ============================================================================
-- CONVERSATIONS
-- ============================================================================

CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type conversation_type DEFAULT 'direct',
    title VARCHAR(255), -- For group chats
    avatar_url VARCHAR(500), -- For group chats
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    last_message_id UUID, -- Will be updated via trigger
    last_message_at TIMESTAMP WITH TIME ZONE,
    message_count INTEGER DEFAULT 0,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_conversations_type ON conversations(type);
CREATE INDEX idx_conversations_last_message_at ON conversations(last_message_at DESC);
CREATE INDEX idx_conversations_created_by ON conversations(created_by);

COMMENT ON TABLE conversations IS 'Direct and group chat conversations';
COMMENT ON COLUMN conversations.last_message_id IS 'Cached for performance';

-- ============================================================================
-- CONVERSATION PARTICIPANTS
-- ============================================================================

CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'member', 'admin' for group chats
    last_read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_read_message_id UUID,
    unread_count INTEGER DEFAULT 0,
    is_muted BOOLEAN DEFAULT FALSE,
    muted_until TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_conversation_participants_user_id ON conversation_participants(user_id);
CREATE INDEX idx_conversation_participants_conversation_id ON conversation_participants(conversation_id);
CREATE INDEX idx_conversation_participants_unread ON conversation_participants(user_id, unread_count) WHERE unread_count > 0;

COMMENT ON TABLE conversation_participants IS 'Users participating in conversations';
COMMENT ON COLUMN conversation_participants.unread_count IS 'Cached unread message count';

-- ============================================================================
-- MESSAGES
-- ============================================================================

CREATE TYPE message_content_type AS ENUM ('text', 'image', 'video', 'file', 'audio', 'location');

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    content_type message_content_type DEFAULT 'text',
    media_url VARCHAR(500),
    media_metadata JSONB, -- File size, dimensions, duration, etc.
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT messages_content_check CHECK (
        (content_type = 'text' AND content IS NOT NULL) OR
        (content_type != 'text' AND media_url IS NOT NULL)
    )
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_reply_to ON messages(reply_to_message_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);

COMMENT ON TABLE messages IS 'Messages in conversations';
COMMENT ON COLUMN messages.media_metadata IS 'JSON metadata for media messages';

-- ============================================================================
-- MESSAGE READ RECEIPTS
-- ============================================================================

CREATE TABLE message_reads (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);
CREATE INDEX idx_message_reads_user_id ON message_reads(user_id);

COMMENT ON TABLE message_reads IS 'Message read receipts for group chats';

-- ============================================================================
-- MESSAGE REQUESTS (For non-connections)
-- ============================================================================

CREATE TABLE message_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    status connection_status DEFAULT 'pending',
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT message_requests_unique UNIQUE(from_user_id, to_user_id),
    CONSTRAINT message_requests_no_self_request CHECK (from_user_id != to_user_id)
);

CREATE INDEX idx_message_requests_to_user_id ON message_requests(to_user_id);
CREATE INDEX idx_message_requests_from_user_id ON message_requests(from_user_id);
CREATE INDEX idx_message_requests_status ON message_requests(status);

COMMENT ON TABLE message_requests IS 'Message requests between non-connected users';

-- ============================================================================
-- NOTIFICATIONS
-- ============================================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    priority notification_priority DEFAULT 'medium',
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    action_url VARCHAR(500), -- Deep link to related content
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL, -- User who triggered the notification
    entity_type VARCHAR(50), -- 'post', 'comment', 'event', 'startup', etc.
    entity_id UUID,
    metadata JSONB, -- Additional context data
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    is_sent BOOLEAN DEFAULT FALSE, -- For push notifications
    sent_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- Auto-delete old notifications
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT notifications_entity_check CHECK (
        (entity_type IS NULL AND entity_id IS NULL) OR
        (entity_type IS NOT NULL AND entity_id IS NOT NULL)
    )
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_is_read ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_actor_id ON notifications(actor_id);
CREATE INDEX idx_notifications_entity ON notifications(entity_type, entity_id);
CREATE INDEX idx_notifications_expires_at ON notifications(expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE notifications IS 'User notifications for all activities';
COMMENT ON COLUMN notifications.metadata IS 'JSON data for notification rendering';

-- ============================================================================
-- NOTIFICATION PREFERENCES
-- ============================================================================

CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    -- Push notification settings
    push_enabled BOOLEAN DEFAULT TRUE,
    push_follows BOOLEAN DEFAULT TRUE,
    push_likes BOOLEAN DEFAULT TRUE,
    push_comments BOOLEAN DEFAULT TRUE,
    push_mentions BOOLEAN DEFAULT TRUE,
    push_messages BOOLEAN DEFAULT TRUE,
    push_events BOOLEAN DEFAULT TRUE,
    push_startup_updates BOOLEAN DEFAULT TRUE,

    -- Email notification settings
    email_enabled BOOLEAN DEFAULT TRUE,
    email_digest_frequency VARCHAR(20) DEFAULT 'daily', -- 'instant', 'daily', 'weekly', 'never'
    email_follows BOOLEAN DEFAULT TRUE,
    email_likes BOOLEAN DEFAULT FALSE,
    email_comments BOOLEAN DEFAULT TRUE,
    email_mentions BOOLEAN DEFAULT TRUE,
    email_messages BOOLEAN DEFAULT TRUE,
    email_events BOOLEAN DEFAULT TRUE,
    email_startup_updates BOOLEAN DEFAULT FALSE,

    -- Do Not Disturb
    dnd_enabled BOOLEAN DEFAULT FALSE,
    dnd_start_time TIME,
    dnd_end_time TIME,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notification_preferences_email_digest ON notification_preferences(email_digest_frequency)
    WHERE email_enabled = TRUE;

COMMENT ON TABLE notification_preferences IS 'User notification preferences and settings';

-- ============================================================================
-- PUSH NOTIFICATION TOKENS (FCM/APNS)
-- ============================================================================

CREATE TABLE push_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL UNIQUE,
    platform VARCHAR(20) NOT NULL, -- 'ios', 'android', 'web'
    device_id VARCHAR(255),
    device_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT push_tokens_platform_check CHECK (platform IN ('ios', 'android', 'web'))
);

CREATE INDEX idx_push_tokens_user_id ON push_tokens(user_id);
CREATE INDEX idx_push_tokens_token ON push_tokens(token);
CREATE INDEX idx_push_tokens_active ON push_tokens(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE push_tokens IS 'Device tokens for push notifications';

-- ============================================================================
-- EMAIL QUEUE
-- ============================================================================

CREATE TYPE email_status AS ENUM ('pending', 'sent', 'failed', 'bounced');

CREATE TABLE email_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    to_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    template_name VARCHAR(100),
    template_data JSONB,
    status email_status DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error_message TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    scheduled_for TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_email_queue_status ON email_queue(status, scheduled_for);
CREATE INDEX idx_email_queue_user_id ON email_queue(user_id);

COMMENT ON TABLE email_queue IS 'Email queue for asynchronous sending';

-- ============================================================================
-- ACTIVITY FEED (Aggregated Feed Data)
-- ============================================================================

CREATE TABLE activity_feed (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_activity_feed_user_id ON activity_feed(user_id, created_at DESC);
CREATE INDEX idx_activity_feed_actor_id ON activity_feed(actor_id);
CREATE INDEX idx_activity_feed_entity ON activity_feed(entity_type, entity_id);
CREATE INDEX idx_activity_feed_created_at ON activity_feed(created_at DESC);

COMMENT ON TABLE activity_feed IS 'Aggregated activity feed for users';
COMMENT ON COLUMN activity_feed.metadata IS 'Activity-specific data for rendering';

-- ============================================================================
-- TYPING INDICATORS (Ephemeral - can be Redis, included for completeness)
-- ============================================================================

CREATE TABLE typing_indicators (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '5 seconds'),

    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_typing_indicators_conversation_id ON typing_indicators(conversation_id);
CREATE INDEX idx_typing_indicators_expires_at ON typing_indicators(expires_at);

COMMENT ON TABLE typing_indicators IS 'Temporary typing indicators (use Redis in production)';
