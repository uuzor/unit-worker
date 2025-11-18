-- ============================================================================
-- iKampus Database Schema - Extensions
-- ============================================================================
-- This file sets up PostgreSQL extensions required for the platform
-- Version: 1.0
-- Created: 2025-11-18

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Full-text search with trigrams
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Case-insensitive text type
CREATE EXTENSION IF NOT EXISTS "citext";

-- Cryptographic functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- PostGIS for geospatial queries (campus location features)
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Comment
COMMENT ON EXTENSION "uuid-ossp" IS 'UUID generation functions';
COMMENT ON EXTENSION "pg_trgm" IS 'Trigram similarity for text search';
COMMENT ON EXTENSION "citext" IS 'Case-insensitive text type';
COMMENT ON EXTENSION "pgcrypto" IS 'Cryptographic functions';
COMMENT ON EXTENSION "postgis" IS 'Geographic objects support';
