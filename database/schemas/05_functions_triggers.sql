-- ============================================================================
-- iKampus Database Schema - Functions & Triggers
-- ============================================================================
-- Database functions, triggers, and automation
-- Version: 1.0
-- Created: 2025-11-18

-- ============================================================================
-- UTILITY FUNCTIONS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically update updated_at timestamp';

-- ============================================================================
-- USER TRIGGERS
-- ============================================================================

-- Auto-update updated_at on users table
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Auto-update updated_at on profiles table
CREATE TRIGGER profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COUNTER UPDATE FUNCTIONS
-- ============================================================================

-- Update follower/following counts
CREATE OR REPLACE FUNCTION update_follow_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment follower count for following user
        UPDATE profiles SET follower_count = follower_count + 1
        WHERE user_id = NEW.following_id;

        -- Increment following count for follower user
        UPDATE profiles SET following_count = following_count + 1
        WHERE user_id = NEW.follower_id;

    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement follower count
        UPDATE profiles SET follower_count = GREATEST(follower_count - 1, 0)
        WHERE user_id = OLD.following_id;

        -- Decrement following count
        UPDATE profiles SET following_count = GREATEST(following_count - 1, 0)
        WHERE user_id = OLD.follower_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER follow_counts_trigger
    AFTER INSERT OR DELETE ON follows
    FOR EACH ROW
    EXECUTE FUNCTION update_follow_counts();

COMMENT ON FUNCTION update_follow_counts() IS 'Maintain follower/following counts';

-- ============================================================================
-- POST COUNTER FUNCTIONS
-- ============================================================================

-- Update post reaction count
CREATE OR REPLACE FUNCTION update_post_reaction_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE posts SET reaction_count = reaction_count + 1
        WHERE id = NEW.post_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE posts SET reaction_count = GREATEST(reaction_count - 1, 0)
        WHERE id = OLD.post_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER post_reaction_count_trigger
    AFTER INSERT OR DELETE ON post_reactions
    FOR EACH ROW
    EXECUTE FUNCTION update_post_reaction_count();

-- Update post comment count
CREATE OR REPLACE FUNCTION update_post_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE posts SET comment_count = comment_count + 1
        WHERE id = NEW.post_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0)
        WHERE id = OLD.post_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER post_comment_count_trigger
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_post_comment_count();

-- Update comment reaction count
CREATE OR REPLACE FUNCTION update_comment_reaction_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE comments SET reaction_count = reaction_count + 1
        WHERE id = NEW.comment_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE comments SET reaction_count = GREATEST(reaction_count - 1, 0)
        WHERE id = OLD.comment_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_reaction_count_trigger
    AFTER INSERT OR DELETE ON comment_reactions
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_reaction_count();

-- Update comment reply count
CREATE OR REPLACE FUNCTION update_comment_reply_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.parent_comment_id IS NOT NULL THEN
        UPDATE comments SET reply_count = reply_count + 1
        WHERE id = NEW.parent_comment_id;
    ELSIF TG_OP = 'DELETE' AND OLD.parent_comment_id IS NOT NULL THEN
        UPDATE comments SET reply_count = GREATEST(reply_count - 1, 0)
        WHERE id = OLD.parent_comment_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_reply_count_trigger
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_reply_count();

-- Update user post count
CREATE OR REPLACE FUNCTION update_user_post_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE profiles SET post_count = post_count + 1
        WHERE user_id = NEW.user_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE profiles SET post_count = GREATEST(post_count - 1, 0)
        WHERE user_id = OLD.user_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_post_count_trigger
    AFTER INSERT OR DELETE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_user_post_count();

-- ============================================================================
-- COMMUNITY COUNTER FUNCTIONS
-- ============================================================================

-- Update community member count
CREATE OR REPLACE FUNCTION update_community_member_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE communities SET member_count = member_count + 1
        WHERE id = NEW.community_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE communities SET member_count = GREATEST(member_count - 1, 0)
        WHERE id = OLD.community_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER community_member_count_trigger
    AFTER INSERT OR DELETE ON community_members
    FOR EACH ROW
    EXECUTE FUNCTION update_community_member_count();

-- Update community post count
CREATE OR REPLACE FUNCTION update_community_post_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.community_id IS NOT NULL THEN
        UPDATE communities SET post_count = post_count + 1
        WHERE id = NEW.community_id;
    ELSIF TG_OP = 'DELETE' AND OLD.community_id IS NOT NULL THEN
        UPDATE communities SET post_count = GREATEST(post_count - 1, 0)
        WHERE id = OLD.community_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER community_post_count_trigger
    AFTER INSERT OR DELETE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_community_post_count();

-- ============================================================================
-- STARTUP COUNTER FUNCTIONS
-- ============================================================================

-- Update startup follower count
CREATE OR REPLACE FUNCTION update_startup_follower_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE startups SET follower_count = follower_count + 1
        WHERE id = NEW.startup_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE startups SET follower_count = GREATEST(follower_count - 1, 0)
        WHERE id = OLD.startup_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER startup_follower_count_trigger
    AFTER INSERT OR DELETE ON startup_followers
    FOR EACH ROW
    EXECUTE FUNCTION update_startup_follower_count();

-- Update startup like count
CREATE OR REPLACE FUNCTION update_startup_like_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE startups SET like_count = like_count + 1
        WHERE id = NEW.startup_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE startups SET like_count = GREATEST(like_count - 1, 0)
        WHERE id = OLD.startup_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER startup_like_count_trigger
    AFTER INSERT OR DELETE ON startup_likes
    FOR EACH ROW
    EXECUTE FUNCTION update_startup_like_count();

-- ============================================================================
-- HASHTAG EXTRACTION AND MANAGEMENT
-- ============================================================================

-- Extract and update hashtags from post content
CREATE OR REPLACE FUNCTION extract_and_update_hashtags()
RETURNS TRIGGER AS $$
DECLARE
    tag TEXT;
BEGIN
    -- Extract hashtags from content using regex
    FOR tag IN
        SELECT DISTINCT lower(regexp_replace(match[1], '^#', ''))
        FROM regexp_matches(NEW.content, '#([a-zA-Z0-9_]+)', 'g') AS match
    LOOP
        -- Insert or update hashtag
        INSERT INTO hashtags (tag, post_count, last_used_at)
        VALUES (tag, 1, NOW())
        ON CONFLICT (tag) DO UPDATE
        SET post_count = hashtags.post_count + 1,
            last_used_at = NOW();
    END LOOP;

    -- Store hashtags in post
    NEW.hashtags := ARRAY(
        SELECT lower(regexp_replace(match[1], '^#', ''))
        FROM regexp_matches(NEW.content, '#([a-zA-Z0-9_]+)', 'g') AS match
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER extract_hashtags_trigger
    BEFORE INSERT OR UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION extract_and_update_hashtags();

-- ============================================================================
-- MENTION EXTRACTION
-- ============================================================================

-- Extract mentions from post content
CREATE OR REPLACE FUNCTION extract_mentions()
RETURNS TRIGGER AS $$
BEGIN
    -- Extract @mentions and lookup user IDs
    NEW.mentions := ARRAY(
        SELECT DISTINCT p.user_id
        FROM regexp_matches(NEW.content, '@([a-zA-Z0-9_-]+)', 'g') AS match
        JOIN profiles p ON p.username = match[1]
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER extract_mentions_trigger
    BEFORE INSERT OR UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION extract_mentions();

-- ============================================================================
-- ENGAGEMENT SCORE CALCULATION
-- ============================================================================

-- Calculate engagement score for posts (for feed ranking)
CREATE OR REPLACE FUNCTION calculate_engagement_score()
RETURNS TRIGGER AS $$
DECLARE
    time_decay FLOAT;
    hours_old FLOAT;
BEGIN
    -- Calculate hours since post creation
    hours_old := EXTRACT(EPOCH FROM (NOW() - NEW.created_at)) / 3600.0;

    -- Time decay factor (exponential decay over 48 hours)
    time_decay := EXP(-hours_old / 48.0);

    -- Engagement score formula:
    -- (reactions * 1 + comments * 2 + shares * 3) * time_decay
    NEW.engagement_score := (
        (NEW.reaction_count * 1.0) +
        (NEW.comment_count * 2.0) +
        (NEW.share_count * 3.0)
    ) * time_decay;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_engagement_score_trigger
    BEFORE INSERT OR UPDATE OF reaction_count, comment_count, share_count, created_at ON posts
    FOR EACH ROW
    EXECUTE FUNCTION calculate_engagement_score();

-- ============================================================================
-- MESSAGING TRIGGERS
-- ============================================================================

-- Update conversation last_message metadata
CREATE OR REPLACE FUNCTION update_conversation_last_message()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE conversations
        SET last_message_id = NEW.id,
            last_message_at = NEW.created_at,
            message_count = message_count + 1
        WHERE id = NEW.conversation_id;

        -- Update unread count for all participants except sender
        UPDATE conversation_participants
        SET unread_count = unread_count + 1
        WHERE conversation_id = NEW.conversation_id
          AND user_id != NEW.sender_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_conversation_last_message_trigger
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_conversation_last_message();

-- Reset unread count when user reads messages
CREATE OR REPLACE FUNCTION reset_unread_count()
RETURNS TRIGGER AS $$
BEGIN
    NEW.unread_count := 0;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reset_unread_count_trigger
    BEFORE UPDATE OF last_read_at ON conversation_participants
    FOR EACH ROW
    WHEN (NEW.last_read_at > OLD.last_read_at)
    EXECUTE FUNCTION reset_unread_count();

-- ============================================================================
-- UNIVERSITY STUDENT COUNT
-- ============================================================================

-- Update university student count
CREATE OR REPLACE FUNCTION update_university_student_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.university_id IS NOT NULL THEN
        UPDATE universities SET student_count = student_count + 1
        WHERE id = NEW.university_id;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.university_id IS NOT NULL AND NEW.university_id != OLD.university_id THEN
            UPDATE universities SET student_count = GREATEST(student_count - 1, 0)
            WHERE id = OLD.university_id;
        END IF;
        IF NEW.university_id IS NOT NULL AND NEW.university_id != OLD.university_id THEN
            UPDATE universities SET student_count = student_count + 1
            WHERE id = NEW.university_id;
        END IF;
    ELSIF TG_OP = 'DELETE' AND OLD.university_id IS NOT NULL THEN
        UPDATE universities SET student_count = GREATEST(student_count - 1, 0)
        WHERE id = OLD.university_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_university_student_count_trigger
    AFTER INSERT OR UPDATE OF university_id OR DELETE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_university_student_count();

-- ============================================================================
-- POLL VOTE COUNT
-- ============================================================================

-- Update poll option vote count
CREATE OR REPLACE FUNCTION update_poll_vote_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment option vote count
        UPDATE poll_options SET vote_count = vote_count + 1
        WHERE id = NEW.poll_option_id;

        -- Increment total poll votes
        UPDATE polls SET total_votes = total_votes + 1
        WHERE id = (SELECT poll_id FROM poll_options WHERE id = NEW.poll_option_id);

    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement option vote count
        UPDATE poll_options SET vote_count = GREATEST(vote_count - 1, 0)
        WHERE id = OLD.poll_option_id;

        -- Decrement total poll votes
        UPDATE polls SET total_votes = GREATEST(total_votes - 1, 0)
        WHERE id = (SELECT poll_id FROM poll_options WHERE id = OLD.poll_option_id);
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER poll_vote_count_trigger
    AFTER INSERT OR DELETE ON poll_votes
    FOR EACH ROW
    EXECUTE FUNCTION update_poll_vote_count();

-- ============================================================================
-- EVENT ATTENDEE COUNT
-- ============================================================================

-- Update event attendee count
CREATE OR REPLACE FUNCTION update_event_attendee_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.status = 'going' THEN
            UPDATE events SET attendee_count = attendee_count + 1
            WHERE id = NEW.event_id;
        ELSIF NEW.status = 'waitlist' THEN
            UPDATE events SET waitlist_count = waitlist_count + 1
            WHERE id = NEW.event_id;
        END IF;

    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle status changes
        IF OLD.status = 'going' AND NEW.status != 'going' THEN
            UPDATE events SET attendee_count = GREATEST(attendee_count - 1, 0)
            WHERE id = NEW.event_id;
        ELSIF OLD.status != 'going' AND NEW.status = 'going' THEN
            UPDATE events SET attendee_count = attendee_count + 1
            WHERE id = NEW.event_id;
        END IF;

        IF OLD.status = 'waitlist' AND NEW.status != 'waitlist' THEN
            UPDATE events SET waitlist_count = GREATEST(waitlist_count - 1, 0)
            WHERE id = NEW.event_id;
        ELSIF OLD.status != 'waitlist' AND NEW.status = 'waitlist' THEN
            UPDATE events SET waitlist_count = waitlist_count + 1
            WHERE id = NEW.event_id;
        END IF;

    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.status = 'going' THEN
            UPDATE events SET attendee_count = GREATEST(attendee_count - 1, 0)
            WHERE id = OLD.event_id;
        ELSIF OLD.status = 'waitlist' THEN
            UPDATE events SET waitlist_count = GREATEST(waitlist_count - 1, 0)
            WHERE id = OLD.event_id;
        END IF;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER event_attendee_count_trigger
    AFTER INSERT OR UPDATE OF status OR DELETE ON event_attendees
    FOR EACH ROW
    EXECUTE FUNCTION update_event_attendee_count();

-- ============================================================================
-- AUTO-DELETE EXPIRED TOKENS
-- ============================================================================

-- Clean up expired tokens (run via cron job)
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $$
BEGIN
    -- Delete expired email verification tokens
    UPDATE users
    SET email_verification_token = NULL,
        email_verification_expires_at = NULL
    WHERE email_verification_expires_at < NOW();

    -- Delete expired password reset tokens
    DELETE FROM password_reset_tokens
    WHERE expires_at < NOW() AND used = FALSE;

    -- Delete expired refresh tokens
    DELETE FROM refresh_tokens
    WHERE expires_at < NOW();

    -- Delete old notifications (older than 90 days)
    DELETE FROM notifications
    WHERE created_at < NOW() - INTERVAL '90 days';

    -- Delete old audit logs (older than 1 year)
    DELETE FROM user_audit_log
    WHERE created_at < NOW() - INTERVAL '1 year';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_tokens() IS 'Clean up expired tokens and old data (run daily via cron)';
