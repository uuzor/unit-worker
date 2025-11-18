#!/bin/bash
# ============================================================================
# iKampus Database Initialization Script
# ============================================================================
# This script initializes the iKampus PostgreSQL database
# Usage: ./init_database.sh [database_name] [database_user]
# Version: 1.0
# Created: 2025-11-18

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_NAME="${1:-ikampus}"
DB_USER="${2:-ikampus_user}"
DB_PASSWORD="${3:-changeme}"
SCHEMA_DIR="$(dirname "$0")/schemas"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}iKampus Database Initialization${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Database Name: ${YELLOW}${DB_NAME}${NC}"
echo -e "Database User: ${YELLOW}${DB_USER}${NC}"
echo -e "Schema Directory: ${YELLOW}${SCHEMA_DIR}${NC}"
echo ""

# Check if PostgreSQL is running
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: PostgreSQL client (psql) not found${NC}"
    exit 1
fi

# Create database and user (requires superuser privileges)
echo -e "${YELLOW}Creating database and user...${NC}"
psql -U postgres <<EOF
-- Create user if not exists
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '${DB_USER}') THEN
        CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
    END IF;
END
\$\$;

-- Create database if not exists
SELECT 'CREATE DATABASE ${DB_NAME} OWNER ${DB_USER}'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_NAME}')\gexec

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
EOF

echo -e "${GREEN}✓ Database and user created${NC}"

# Apply schema files in order
echo -e "${YELLOW}Applying schema files...${NC}"

SCHEMA_FILES=(
    "00_extensions.sql"
    "01_auth_users.sql"
    "02_social.sql"
    "03_startup_hub.sql"
    "04_messaging_notifications.sql"
    "05_functions_triggers.sql"
    "06_views_analytics.sql"
)

for file in "${SCHEMA_FILES[@]}"; do
    filepath="${SCHEMA_DIR}/${file}"
    if [ -f "$filepath" ]; then
        echo -e "  ${YELLOW}→${NC} Applying ${file}..."
        psql -U "$DB_USER" -d "$DB_NAME" -f "$filepath" > /dev/null
        echo -e "  ${GREEN}✓${NC} ${file} applied"
    else
        echo -e "  ${RED}✗${NC} ${file} not found"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}✓ All schema files applied successfully${NC}"

# Display database statistics
echo ""
echo -e "${YELLOW}Database Statistics:${NC}"
psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT
    schemaname,
    COUNT(*) as table_count
FROM pg_tables
WHERE schemaname = 'public'
GROUP BY schemaname;
"

echo -e "${YELLOW}Table List:${NC}"
psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT tablename
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;
"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Database initialization complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Connection string:"
echo -e "${YELLOW}postgresql://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}${NC}"
echo ""
echo -e "To connect:"
echo -e "${YELLOW}psql -U ${DB_USER} -d ${DB_NAME}${NC}"
echo ""
