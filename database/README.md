# iKampus Database

Production-ready PostgreSQL database schema for the iKampus platform.

---

## Quick Start

### Prerequisites

- PostgreSQL 14+ installed
- PostGIS extension available
- Superuser access to create database

### Initialize Database

```bash
cd database
./init_database.sh ikampus ikampus_user your_password
```

This will:
1. Create the database and user
2. Apply all schema files in order
3. Create tables, indexes, functions, triggers
4. Initialize materialized views

### Docker Setup

```bash
docker run -d \
  --name ikampus-postgres \
  -e POSTGRES_DB=ikampus \
  -e POSTGRES_USER=ikampus_user \
  -e POSTGRES_PASSWORD=changeme \
  -p 5432:5432 \
  -v ikampus-data:/var/lib/postgresql/data \
  postgis/postgis:14-3.3
```

Then run the init script:
```bash
./init_database.sh ikampus ikampus_user changeme
```

---

## Directory Structure

```
database/
├── schemas/              # SQL schema files
│   ├── 00_extensions.sql
│   ├── 01_auth_users.sql
│   ├── 02_social.sql
│   ├── 03_startup_hub.sql
│   ├── 04_messaging_notifications.sql
│   ├── 05_functions_triggers.sql
│   └── 06_views_analytics.sql
├── migrations/           # Database migrations (Flyway/Alembic)
├── seeds/                # Seed data for development
├── init_database.sh      # Initialization script
└── README.md             # This file
```

---

## Schema Overview

### 1. Authentication & Users (`01_auth_users.sql`)

**10 tables:**
- `universities` - Verified university list
- `users` - Core authentication
- `profiles` - Extended user profiles
- `student_verifications` - Student ID verification
- `refresh_tokens` - JWT session management
- `follows` - Social graph
- `blocks` - User blocking
- `password_reset_tokens` - Password recovery
- `user_audit_log` - Security audit trail

### 2. Social Features (`02_social.sql`)

**18 tables:**
- `communities` - Groups, societies, halls
- `community_members` - Membership
- `posts` - User-generated content
- `comments` - Nested comments (2 levels)
- `post_reactions` - Likes, loves, etc.
- `comment_reactions` - Comment engagement
- `saved_posts` - Bookmarks
- `polls`, `poll_options`, `poll_votes` - Polling
- `events`, `event_attendees` - Campus events
- `hashtags` - Trending hashtags
- `reports` - Content moderation

### 3. Startup Hub (`03_startup_hub.sql`)

**14 tables:**
- `startups` - Startup profiles
- `startup_members` - Team composition
- `pitches` - Pitch decks
- `cofounder_profiles`, `cofounder_matches` - Co-founder matching
- `mentor_profiles`, `mentorship_connections`, `mentorship_sessions` - Mentorship
- `investor_profiles`, `startup_investor_connections` - Investment pipeline
- `startup_updates`, `startup_followers`, `startup_likes` - Engagement

### 4. Messaging & Notifications (`04_messaging_notifications.sql`)

**11 tables:**
- `conversations`, `conversation_participants` - Chats
- `messages`, `message_reads` - Messaging
- `message_requests` - Connection requests
- `notifications`, `notification_preferences` - Multi-channel notifications
- `push_tokens` - FCM/APNS tokens
- `email_queue` - Async email sending
- `activity_feed` - Aggregated feed
- `typing_indicators` - Real-time typing

### 5. Functions & Triggers (`05_functions_triggers.sql`)

**15+ functions and 20+ triggers:**
- Auto-update timestamps
- Counter maintenance (followers, reactions, etc.)
- Hashtag extraction
- Mention extraction
- Engagement score calculation
- Unread count management
- Data cleanup

### 6. Analytics & Views (`06_views_analytics.sql`)

**6 materialized views + analytics:**
- `user_engagement_metrics` - User engagement scores
- `trending_posts` - Hot posts by score
- `popular_hashtags` - Trending hashtags
- `active_communities` - Most active groups
- `startup_directory` - Searchable startups
- `university_statistics` - University metrics
- `analytics_events` - Event tracking
- `daily_statistics` - Daily aggregates

---

## Database Statistics

**Total Objects:**
- 53 tables
- 6 materialized views
- 15+ functions
- 20+ triggers
- 200+ indexes
- 15+ enum types

**Estimated Size:**
- Fresh install: ~50 MB
- 1,000 users: ~500 MB
- 10,000 users: ~5 GB
- 100,000 users: ~50 GB

---

## Connection

### Connection String

```
postgresql://ikampus_user:changeme@localhost:5432/ikampus
```

### psql

```bash
psql -U ikampus_user -d ikampus
```

### Environment Variables

```bash
export DATABASE_URL="postgresql://ikampus_user:changeme@localhost:5432/ikampus"
export DATABASE_HOST="localhost"
export DATABASE_PORT="5432"
export DATABASE_NAME="ikampus"
export DATABASE_USER="ikampus_user"
export DATABASE_PASSWORD="changeme"
export DATABASE_SSL_MODE="prefer"
```

---

## Maintenance

### Daily Tasks

```bash
# Clean up expired tokens
psql -U ikampus_user -d ikampus -c "SELECT cleanup_expired_tokens();"
```

### Hourly Tasks

```bash
# Refresh materialized views
psql -U ikampus_user -d ikampus -c "SELECT refresh_materialized_views();"
```

### Weekly Tasks

```bash
# Vacuum and analyze
psql -U ikampus_user -d ikampus -c "VACUUM ANALYZE;"

# Reindex
psql -U ikampus_user -d ikampus -c "REINDEX DATABASE ikampus;"
```

### Cron Jobs

```cron
# Cleanup expired tokens (daily at 2am)
0 2 * * * psql -U ikampus_user -d ikampus -c "SELECT cleanup_expired_tokens();"

# Refresh materialized views (hourly)
0 * * * * psql -U ikampus_user -d ikampus -c "SELECT refresh_materialized_views();"

# Vacuum and analyze (weekly on Sunday at 3am)
0 3 * * 0 psql -U ikampus_user -d ikampus -c "VACUUM ANALYZE;"
```

---

## Backup & Restore

### Full Backup

```bash
# Custom format (recommended)
pg_dump -U ikampus_user -d ikampus -F c -f ikampus_$(date +%Y%m%d).dump

# SQL format
pg_dump -U ikampus_user -d ikampus -f ikampus_$(date +%Y%m%d).sql

# Compressed
pg_dump -U ikampus_user -d ikampus | gzip > ikampus_$(date +%Y%m%d).sql.gz
```

### Restore

```bash
# From custom format
pg_restore -U ikampus_user -d ikampus -c ikampus_20251118.dump

# From SQL
psql -U ikampus_user -d ikampus < ikampus_20251118.sql

# From compressed
gunzip -c ikampus_20251118.sql.gz | psql -U ikampus_user -d ikampus
```

### Automated Backups (cron)

```bash
# Daily backup at 1am
0 1 * * * pg_dump -U ikampus_user -d ikampus -F c -f /backups/ikampus_$(date +\%Y\%m\%d).dump

# Weekly backup (keep for 30 days)
0 0 * * 0 pg_dump -U ikampus_user -d ikampus -F c -f /backups/weekly/ikampus_$(date +\%Y\%m\%d).dump

# Cleanup old backups (older than 30 days)
0 3 * * * find /backups -name "ikampus_*.dump" -mtime +30 -delete
```

---

## Performance Tuning

### PostgreSQL Configuration

**postgresql.conf:**
```ini
# Memory
shared_buffers = 4GB            # 25% of RAM
effective_cache_size = 12GB     # 75% of RAM
work_mem = 64MB
maintenance_work_mem = 1GB

# Checkpoints
checkpoint_completion_target = 0.9
wal_buffers = 16MB

# Query Planner
random_page_cost = 1.1          # For SSD
effective_io_concurrency = 200  # For SSD

# Connections
max_connections = 200

# Logging
log_min_duration_statement = 1000  # Log queries > 1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

### Connection Pooling (PgBouncer)

**pgbouncer.ini:**
```ini
[databases]
ikampus = host=localhost port=5432 dbname=ikampus

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
reserve_pool_size = 5
reserve_pool_timeout = 3
```

---

## Monitoring

### Useful Queries

**Active Queries:**
```sql
SELECT
    pid,
    now() - query_start AS duration,
    query,
    state
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;
```

**Database Size:**
```sql
SELECT
    pg_size_pretty(pg_database_size('ikampus')) AS db_size;
```

**Table Sizes:**
```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;
```

**Index Usage:**
```sql
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC, pg_relation_size(indexrelid) DESC
LIMIT 20;
```

**Cache Hit Ratio:**
```sql
SELECT
    sum(heap_blks_read) AS heap_read,
    sum(heap_blks_hit) AS heap_hit,
    sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) AS ratio
FROM pg_statio_user_tables;
```

---

## Troubleshooting

### Reset Database

```bash
# Drop and recreate
dropdb -U postgres ikampus
./init_database.sh ikampus ikampus_user
```

### Fix Sequences

```sql
-- Reset all sequences
SELECT 'SELECT SETVAL(' ||
       quote_literal(quote_ident(PGT.schemaname) || '.' || quote_ident(S.relname)) ||
       ', MAX(' || quote_ident(C.attname) || ') ) FROM ' ||
       quote_ident(PGT.schemaname) || '.' || quote_ident(T.relname) || ';'
FROM pg_class AS S,
     pg_depend AS D,
     pg_class AS T,
     pg_attribute AS C,
     pg_tables AS PGT
WHERE S.relkind = 'S'
  AND S.oid = D.objid
  AND D.refobjid = T.oid
  AND D.refobjid = C.attrelid
  AND D.refobjsubid = C.attnum
  AND T.relname = PGT.tablename
  AND PGT.schemaname = 'public';
```

### Rebuild Indexes

```sql
REINDEX DATABASE ikampus;
```

### Analyze Tables

```sql
ANALYZE;
```

---

## Development

### Seed Data

```bash
# Load seed data
psql -U ikampus_user -d ikampus < seeds/01_universities.sql
psql -U ikampus_user -d ikampus < seeds/02_users.sql
psql -U ikampus_user -d ikampus < seeds/03_posts.sql
```

### Test Data Generation

```sql
-- Generate 1000 test users
INSERT INTO users (email, password_hash, email_verified, university_id)
SELECT
    'user' || generate_series || '@test.ac.uk',
    '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5aqCBxJjwDj4u', -- "password"
    TRUE,
    (SELECT id FROM universities ORDER BY RANDOM() LIMIT 1)
FROM generate_series(1, 1000);
```

---

## Migration Strategy

### Using Flyway

**flyway.conf:**
```properties
flyway.url=jdbc:postgresql://localhost:5432/ikampus
flyway.user=ikampus_user
flyway.password=changeme
flyway.locations=filesystem:./migrations
flyway.baselineOnMigrate=true
```

**Migration File:**
```sql
-- V001__add_user_bio_field.sql
ALTER TABLE profiles ADD COLUMN bio TEXT;
```

### Using Alembic (Python)

```bash
pip install alembic psycopg2-binary

# Initialize
alembic init alembic

# Create migration
alembic revision -m "add_user_bio_field"

# Apply migrations
alembic upgrade head
```

---

## Security

### SSL/TLS

```bash
# Enable SSL
ssl = on
ssl_cert_file = '/path/to/server.crt'
ssl_key_file = '/path/to/server.key'
ssl_ca_file = '/path/to/root.crt'
```

### Row-Level Security (RLS)

```sql
-- Enable RLS on posts
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see public posts or their own
CREATE POLICY posts_select_policy ON posts
FOR SELECT
USING (
    visibility = 'public'
    OR user_id = current_setting('app.current_user_id')::UUID
);
```

### Audit Logging

```sql
-- Enable pgaudit extension
CREATE EXTENSION pgaudit;

-- Configure audit logging
ALTER SYSTEM SET pgaudit.log = 'write, ddl';
ALTER SYSTEM SET pgaudit.log_catalog = off;
```

---

## Additional Resources

- **Full Schema Documentation:** `/docs/database/schema.md`
- **ER Diagram:** `/docs/database/erd.png` (to be generated)
- **API Integration Guide:** `/docs/api/database-integration.md`

---

**Version:** 1.0
**Last Updated:** 2025-11-18
**Maintained by:** iKampus Engineering Team
