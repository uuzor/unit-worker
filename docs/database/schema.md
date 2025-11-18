## iKampus Database Schema Documentation

**Version:** 1.0
**Last Updated:** 2025-11-18
**Database:** PostgreSQL 14+

---

## Overview

The iKampus database is designed as a robust, scalable foundation for a university-exclusive social platform. The schema supports:

- **Authentication & User Management** - Secure student verification and profiles
- **Social Features** - Posts, comments, reactions, communities
- **Startup Hub** - Startup profiles, team management, investor matching
- **Messaging** - Direct and group conversations
- **Notifications** - Multi-channel notification system
- **Analytics** - Comprehensive event tracking and metrics

---

## Database Architecture

### Design Principles

1. **Normalization** - 3NF normalized to reduce redundancy
2. **Performance** - Strategic denormalization for read-heavy operations (cached counts)
3. **Integrity** - Foreign keys, constraints, and triggers maintain data consistency
4. **Scalability** - Indexed for common query patterns
5. **Security** - Row-level security ready, encrypted sensitive data
6. **Observability** - Audit logging and analytics built-in

### Technology Stack

- **Database:** PostgreSQL 14+
- **Extensions:**
  - `uuid-ossp` - UUID generation
  - `pg_trgm` - Full-text search
  - `citext` - Case-insensitive text
  - `pgcrypto` - Cryptographic functions
  - `postgis` - Geospatial data (campus locations)

---

## Schema Organization

### Schema Files

| File | Purpose | Tables |
|------|---------|--------|
| `00_extensions.sql` | PostgreSQL extensions | - |
| `01_auth_users.sql` | Authentication & users | 10 tables |
| `02_social.sql` | Social features | 18 tables |
| `03_startup_hub.sql` | Startup ecosystem | 14 tables |
| `04_messaging_notifications.sql` | Messaging & notifications | 11 tables |
| `05_functions_triggers.sql` | Functions & triggers | 15+ functions |
| `06_views_analytics.sql` | Materialized views & analytics | 6 views |

### Total Database Objects

- **Tables:** 53
- **Materialized Views:** 6
- **Functions:** 15+
- **Triggers:** 20+
- **Indexes:** 200+
- **Enum Types:** 15+

---

## Core Domains

### 1. Authentication & Users

**Purpose:** Secure authentication, student verification, and user profiles

**Key Tables:**

```sql
universities          -- Verified universities
users                 -- Core authentication
profiles              -- Extended user profiles
student_verifications -- Student ID verification
refresh_tokens        -- JWT session management
follows               -- Social graph
blocks                -- User blocking
```

**Key Features:**
- University email verification (`.ac.uk` domains)
- Student ID document upload
- Multi-device session management
- Password reset with tokens
- Audit logging for security events

**Relationships:**
```
universities (1) ←→ (N) users
users (1) ←→ (1) profiles
users (1) ←→ (1) student_verifications
users (1) ←→ (N) refresh_tokens
users (N) ←→ (N) follows [self-referencing]
```

---

### 2. Social Features

**Purpose:** Posts, comments, communities, events, and engagement

**Key Tables:**

```sql
communities       -- Groups, societies, halls
posts             -- User-generated posts
comments          -- Post comments (2-level nesting)
post_reactions    -- Likes, loves, etc.
comment_reactions -- Comment reactions
polls             -- Polls within posts
poll_options      -- Poll choices
poll_votes        -- User votes
events            -- Campus events
event_attendees   -- RSVPs and attendance
hashtags          -- Trending hashtags
saved_posts       -- Bookmarked posts
reports           -- Content moderation
```

**Key Features:**
- Multiple post types (text, image, video, poll, confession)
- Visibility controls (public, university, community, followers)
- Nested comments (max 2 levels)
- 6 reaction types (like, love, celebrate, insightful, support, funny)
- Geospatial event locations (PostGIS)
- Anonymous posting (confessions)
- Content moderation and reporting

**Engagement Score Algorithm:**
```sql
engagement_score = (
    (reaction_count * 1) +
    (comment_count * 2) +
    (share_count * 3)
) * time_decay
```

**Relationships:**
```
universities (1) ←→ (N) communities
communities (1) ←→ (N) posts
communities (N) ←→ (N) community_members
posts (1) ←→ (N) comments
posts (1) ←→ (N) post_reactions
posts (1) ←→ (1) polls
comments (1) ←→ (N) comment_reactions
events (1) ←→ (N) event_attendees
```

---

### 3. Startup Hub

**Purpose:** Student entrepreneurship, team building, investor connections

**Key Tables:**

```sql
startups                       -- Startup profiles
startup_members                -- Team composition
pitches                        -- Pitch decks
cofounder_profiles             -- Co-founder seeking profiles
cofounder_matches              -- AI-powered matching
mentor_profiles                -- Alumni mentors
mentorship_connections         -- Mentorship relationships
mentorship_sessions            -- Scheduled sessions
investor_profiles              -- Alumni investors
startup_investor_connections   -- Investment pipeline
startup_updates                -- Progress updates
startup_followers              -- Followers
startup_likes                  -- Likes/endorsements
```

**Key Features:**
- Startup lifecycle stages (idea → MVP → launched → scaling → funded)
- Team equity management
- AI-powered co-founder matching
- Mentor availability and booking
- Investor interest tracking
- Pitch deck storage and versioning
- Traction metrics (JSONB)

**Matching Algorithm:**
```sql
-- Co-founder match score (0.0 - 1.0)
- Skills overlap
- Industry alignment
- Commitment level compatibility
- Geographic proximity
```

**Relationships:**
```
startups (1) ←→ (N) startup_members
startups (1) ←→ (N) pitches
startups (1) ←→ (N) startup_updates
users (1) ←→ (1) cofounder_profiles
users (N) ←→ (N) cofounder_matches
users (1) ←→ (1) mentor_profiles
users (N) ←→ (N) mentorship_connections
startups (N) ←→ (N) startup_investor_connections
```

---

### 4. Messaging & Notifications

**Purpose:** Real-time messaging and multi-channel notifications

**Key Tables:**

```sql
conversations              -- DMs and group chats
conversation_participants  -- Participants with unread counts
messages                   -- Messages with media
message_reads              -- Read receipts
message_requests           -- Requests from non-connections
notifications              -- All notification types
notification_preferences   -- User preferences
push_tokens                -- FCM/APNS tokens
email_queue                -- Async email sending
activity_feed              -- Aggregated feed
typing_indicators          -- Real-time typing (ephemeral)
```

**Key Features:**
- Direct and group conversations
- Read receipts and typing indicators
- Message requests for non-connections
- Rich media messages (images, videos, files, audio, voice notes)
- Reply threading
- Multi-channel notifications (push, email, in-app)
- Do Not Disturb schedules
- Email digest options (instant, daily, weekly)
- Notification preferences per type

**Notification Types:**
```sql
- follow
- post_like, post_comment, comment_reply
- mention
- message
- event_reminder
- startup_update
- cofounder_match
- mentorship_request
- investor_interest
- system
```

**Relationships:**
```
conversations (1) ←→ (N) messages
conversations (N) ←→ (N) conversation_participants
messages (N) ←→ (N) message_reads
users (1) ←→ (N) notifications
users (1) ←→ (1) notification_preferences
users (1) ←→ (N) push_tokens
```

---

### 5. Analytics & Metrics

**Purpose:** Event tracking, user engagement, platform insights

**Key Tables:**

```sql
analytics_events         -- Event tracking
daily_statistics         -- Daily aggregates
user_audit_log          -- Security audit trail
```

**Materialized Views:**

```sql
user_engagement_metrics  -- Refresh hourly
trending_posts           -- Refresh every 15 min
popular_hashtags         -- Refresh hourly
active_communities       -- Refresh hourly
startup_directory        -- Refresh every 6 hours
university_statistics    -- Refresh daily
```

**Event Tracking:**
```javascript
{
  event_name: "post_created",
  event_category: "engagement",
  entity_type: "post",
  entity_id: "uuid",
  properties: {
    post_type: "image",
    has_hashtags: true,
    word_count: 42
  }
}
```

---

## Performance Optimization

### Indexing Strategy

**Primary Indexes:**
- Primary keys on all tables (UUID)
- Foreign keys for relationship queries
- Unique constraints on business keys (username, email, slug)

**Query Optimization:**
- GIN indexes for arrays (hashtags, mentions, interests)
- GiST indexes for geospatial (PostGIS)
- Full-text search indexes (to_tsvector)
- Partial indexes for filtered queries (WHERE active = TRUE)

**Example Indexes:**
```sql
-- Feed queries
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_engagement_score ON posts(engagement_score DESC);

-- Search queries
CREATE INDEX idx_profiles_username ON profiles(username);
CREATE INDEX idx_posts_hashtags ON posts USING gin(hashtags);
CREATE INDEX idx_posts_content_search ON posts USING gin(to_tsvector('english', content));

-- Geospatial queries
CREATE INDEX idx_posts_location ON posts USING gist(location);
CREATE INDEX idx_events_location ON events USING gist(location);
```

### Denormalized Counts

**Cached Counters (for performance):**
- `profiles.follower_count`
- `profiles.following_count`
- `profiles.post_count`
- `posts.reaction_count`
- `posts.comment_count`
- `communities.member_count`
- `communities.post_count`
- `startups.follower_count`
- `conversations.message_count`
- `conversation_participants.unread_count`

**Maintained by triggers** - See `05_functions_triggers.sql`

### Materialized Views

**Refresh Schedule:**
- `trending_posts` - Every 15 minutes (cron)
- `popular_hashtags` - Hourly
- `user_engagement_metrics` - Hourly
- `active_communities` - Hourly
- `startup_directory` - Every 6 hours
- `university_statistics` - Daily

**Refresh Command:**
```sql
SELECT refresh_materialized_views();
```

---

## Data Integrity

### Constraints

**Foreign Keys:**
- All relationships enforced with `ON DELETE CASCADE` or `ON DELETE SET NULL`
- Self-referencing constraints (follows, blocks)

**Check Constraints:**
```sql
-- Email format validation
CONSTRAINT users_email_check CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')

-- Username format (alphanumeric, underscore, hyphen)
CONSTRAINT profiles_username_check CHECK (username ~ '^[a-zA-Z0-9_-]{3,50}$')

-- No self-follows
CONSTRAINT follows_no_self_follow CHECK (follower_id != following_id)

-- Event date logic
CONSTRAINT events_dates_check CHECK (ends_at > starts_at)

-- Positive values
CONSTRAINT startups_team_size_check CHECK (team_size > 0)
```

### Triggers

**Automatic Updates:**
```sql
-- Updated_at timestamps
users_updated_at
profiles_updated_at
... (all mutable tables)

-- Counter maintenance
follow_counts_trigger
post_reaction_count_trigger
community_member_count_trigger
startup_follower_count_trigger
... (20+ counter triggers)

-- Content extraction
extract_hashtags_trigger
extract_mentions_trigger

-- Engagement scoring
calculate_engagement_score_trigger

-- Messaging
update_conversation_last_message_trigger
reset_unread_count_trigger
```

---

## Security Considerations

### Authentication

**Password Hashing:**
- bcrypt with cost factor 12
- Stored in `users.password_hash`

**JWT Tokens:**
- Access tokens: 15-minute expiry
- Refresh tokens: 30-day expiry (stored hashed)
- Device tracking (IP, user agent, device ID)

**Email Verification:**
- Cryptographically secure tokens
- Expiry timestamps
- Single-use tokens

### Data Protection

**Sensitive Data:**
- Passwords: bcrypt hashed
- Refresh tokens: SHA-256 hashed
- Email: case-insensitive (citext)
- Student verification docs: Encrypted URLs

**Audit Logging:**
```sql
user_audit_log
- Login attempts
- Password changes
- Email changes
- Permission changes
- Content moderation actions
```

### Account Security

**Brute Force Protection:**
```sql
users.failed_login_attempts
users.locked_until
```

**Rate Limiting:**
- Implemented at application layer
- Redis-based counters

---

## Data Lifecycle

### Soft Deletes

Some tables use soft deletes:
```sql
users.status = 'deleted'
messages.is_deleted = TRUE
```

### Hard Deletes

`ON DELETE CASCADE` for:
- User deletes all their content
- Community deletes all posts
- Conversation deletes all messages

### Data Retention

**Auto-cleanup (via cron):**
```sql
-- Expired tokens (daily)
DELETE FROM refresh_tokens WHERE expires_at < NOW();
DELETE FROM password_reset_tokens WHERE expires_at < NOW();

-- Old notifications (90 days)
DELETE FROM notifications WHERE created_at < NOW() - INTERVAL '90 days';

-- Old audit logs (1 year)
DELETE FROM user_audit_log WHERE created_at < NOW() - INTERVAL '1 year';
```

---

## Querying Patterns

### Common Queries

**User Feed (personalized):**
```sql
SELECT p.*
FROM posts p
JOIN follows f ON p.user_id = f.following_id
WHERE f.follower_id = $user_id
  AND p.created_at > NOW() - INTERVAL '7 days'
  AND p.visibility IN ('public', 'followers')
ORDER BY p.engagement_score DESC
LIMIT 50;
```

**Trending Posts (university-specific):**
```sql
SELECT * FROM trending_posts
WHERE university_id = $university_id
ORDER BY trending_score DESC
LIMIT 20;
```

**Community Feed:**
```sql
SELECT p.*
FROM posts p
WHERE p.community_id = $community_id
ORDER BY p.created_at DESC
LIMIT 50;
```

**Search Users:**
```sql
SELECT p.*
FROM profiles p
WHERE p.username ILIKE '%search%'
   OR p.bio ILIKE '%search%'
   OR p.course ILIKE '%search%'
ORDER BY p.follower_count DESC
LIMIT 20;
```

**Search Startups:**
```sql
SELECT * FROM startup_directory
WHERE name ILIKE '%search%'
   OR tagline ILIKE '%search%'
   OR $tag = ANY(categories)
ORDER BY visibility_score DESC
LIMIT 20;
```

**Unread Messages:**
```sql
SELECT c.*, cp.unread_count
FROM conversations c
JOIN conversation_participants cp ON c.id = cp.conversation_id
WHERE cp.user_id = $user_id
  AND cp.unread_count > 0
ORDER BY c.last_message_at DESC;
```

---

## Maintenance

### Regular Tasks

**Daily:**
```sql
SELECT cleanup_expired_tokens();
```

**Hourly:**
```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY user_engagement_metrics;
REFRESH MATERIALIZED VIEW CONCURRENTLY popular_hashtags;
REFRESH MATERIALIZED VIEW CONCURRENTLY active_communities;
```

**Every 15 minutes:**
```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY trending_posts;
```

**Weekly:**
```sql
VACUUM ANALYZE;
REINDEX DATABASE ikampus;
```

### Backup Strategy

**Full Backup (daily):**
```bash
pg_dump -U ikampus_user -d ikampus -F c -f ikampus_$(date +%Y%m%d).dump
```

**Point-in-Time Recovery:**
```bash
# Enable WAL archiving
archive_mode = on
archive_command = 'cp %p /var/lib/postgresql/wal_archive/%f'
```

**Restore:**
```bash
pg_restore -U ikampus_user -d ikampus ikampus_20251118.dump
```

---

## Scaling Considerations

### Read Replicas

For read-heavy workloads:
- Primary (master): Writes
- Replicas (slaves): Reads

```sql
-- Route reads to replicas
SELECT * FROM posts;  -- Read replica

-- Route writes to primary
INSERT INTO posts VALUES (...);  -- Primary
```

### Partitioning

**Time-based partitioning for large tables:**

```sql
-- Partition analytics_events by month
CREATE TABLE analytics_events_2025_11 PARTITION OF analytics_events
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

### Connection Pooling

**PgBouncer configuration:**
```ini
[databases]
ikampus = host=localhost port=5432 dbname=ikampus

[pgbouncer]
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
```

---

## Migration Strategy

### Initial Setup

```bash
./database/init_database.sh ikampus ikampus_user
```

### Schema Versioning

Use migration tools:
- **Flyway** (Java-based)
- **Liquibase** (XML/YAML)
- **migrate** (Go-based)
- **Alembic** (Python-based)

**Example Migration:**
```sql
-- V001__add_user_bio_length.sql
ALTER TABLE profiles
ADD CONSTRAINT profiles_bio_length_check
CHECK (char_length(bio) <= 500);
```

### Zero-Downtime Deployments

1. Add new column (nullable)
2. Backfill data
3. Make column NOT NULL
4. Add indexes concurrently
5. Drop old column

```sql
-- Step 1: Add column
ALTER TABLE profiles ADD COLUMN new_bio TEXT;

-- Step 2: Backfill
UPDATE profiles SET new_bio = old_bio;

-- Step 3: Make NOT NULL
ALTER TABLE profiles ALTER COLUMN new_bio SET NOT NULL;

-- Step 4: Drop old
ALTER TABLE profiles DROP COLUMN old_bio;
```

---

## Troubleshooting

### Slow Queries

**Identify slow queries:**
```sql
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

**Enable query logging:**
```sql
ALTER DATABASE ikampus SET log_min_duration_statement = 1000; -- Log queries > 1s
```

### Index Usage

**Check unused indexes:**
```sql
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Bloat Analysis

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## Appendix

### Complete Table List

**Authentication & Users (10 tables):**
1. universities
2. users
3. profiles
4. student_verifications
5. refresh_tokens
6. follows
7. blocks
8. password_reset_tokens
9. user_audit_log
10. (user_engagement_metrics - view)

**Social Features (18 tables):**
11. communities
12. community_members
13. posts
14. comments
15. post_reactions
16. comment_reactions
17. saved_posts
18. polls
19. poll_options
20. poll_votes
21. events
22. event_attendees
23. hashtags
24. reports

**Startup Hub (14 tables):**
25. startups
26. startup_members
27. pitches
28. cofounder_profiles
29. cofounder_matches
30. mentor_profiles
31. mentorship_connections
32. mentorship_sessions
33. investor_profiles
34. startup_investor_connections
35. startup_updates
36. startup_followers
37. startup_likes

**Messaging & Notifications (11 tables):**
38. conversations
39. conversation_participants
40. messages
41. message_reads
42. message_requests
43. notifications
44. notification_preferences
45. push_tokens
46. email_queue
47. activity_feed
48. typing_indicators

**Analytics (2 tables + 6 views):**
49. analytics_events
50. daily_statistics

**Materialized Views:**
51. user_engagement_metrics
52. trending_posts
53. popular_hashtags
54. active_communities
55. startup_directory
56. university_statistics

### Connection Examples

**Python (psycopg2):**
```python
import psycopg2

conn = psycopg2.connect(
    dbname="ikampus",
    user="ikampus_user",
    password="changeme",
    host="localhost",
    port="5432"
)
```

**Node.js (pg):**
```javascript
const { Pool } = require('pg');

const pool = new Pool({
    database: 'ikampus',
    user: 'ikampus_user',
    password: 'changeme',
    host: 'localhost',
    port: 5432,
    max: 20
});
```

**Go (lib/pq):**
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres",
    "postgres://ikampus_user:changeme@localhost:5432/ikampus?sslmode=disable")
```

---

**Document Version:** 1.0
**Last Updated:** 2025-11-18
**Maintained by:** iKampus Engineering Team
