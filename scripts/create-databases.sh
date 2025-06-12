#!/bin/bash

# Create Databases Script
# This script creates databases for enabled modules based on configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5433}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
DATABASE_PREFIX="${DATABASE_PREFIX:-modular_monolith}"

echo -e "${BLUE}🗄️ Database Creation Script${NC}"
echo -e "${BLUE}================================${NC}"

# Function to create database if not exists
create_database() {
    local db_name=$1
    echo -e "${YELLOW}📦 Checking database: ${db_name}${NC}"
    
    # Check if database exists
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt | cut -d \| -f 1 | grep -qw $db_name; then
        echo -e "${GREEN}✅ Database ${db_name} already exists${NC}"
    else
        echo -e "${YELLOW}🔨 Creating database: ${db_name}${NC}"
        PGPASSWORD=$POSTGRES_PASSWORD createdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $db_name
        echo -e "${GREEN}✅ Database ${db_name} created successfully${NC}"
    fi
}

# Function to get enabled modules from config
get_enabled_modules() {
    # Try to load from Go binary if available
    if command -v go &> /dev/null && [ -f "cmd/tools/list-modules.go" ]; then
        go run cmd/tools/list-modules.go 2>/dev/null || echo ""
    else
        # Fallback: parse YAML manually (basic parsing)
        if [ -f "config/modules.yaml" ]; then
            # Extract only modules section and parse enabled modules
            awk '
            /^modules:/ { in_modules=1; next }
            /^[a-zA-Z]/ && in_modules { in_modules=0 }
            in_modules && /^  [a-zA-Z_]+:/ {
                if ($0 ~ /: *true/ || $0 ~ /: *{/) {
                    gsub(/^  /, ""); gsub(/:.*/, ""); print
                }
            }
            ' config/modules.yaml || echo ""
        fi
    fi
}

# Check if PostgreSQL is accessible
echo -e "${YELLOW}🔍 Checking PostgreSQL connection...${NC}"
if ! PGPASSWORD=$POSTGRES_PASSWORD pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER &>/dev/null; then
    echo -e "${RED}❌ Cannot connect to PostgreSQL at ${POSTGRES_HOST}:${POSTGRES_PORT}${NC}"
    echo -e "${RED}   Make sure PostgreSQL is running: make docker-dev${NC}"
    exit 1
fi
echo -e "${GREEN}✅ PostgreSQL connection successful${NC}"

# Get enabled modules
echo -e "${YELLOW}🔍 Discovering enabled modules...${NC}"
enabled_modules=$(get_enabled_modules)

if [ -z "$enabled_modules" ]; then
    echo -e "${YELLOW}⚠️ No enabled modules found. Creating default databases...${NC}"
    enabled_modules="customer order"
fi

echo -e "${BLUE}📋 Enabled modules: ${enabled_modules}${NC}"

# Create databases for enabled modules
for module in $enabled_modules; do
    db_name="${DATABASE_PREFIX}_${module}"
    create_database $db_name
done

echo -e "${GREEN}🎉 Database creation completed!${NC}"
echo -e "${BLUE}📋 Summary:${NC}"
for module in $enabled_modules; do
    db_name="${DATABASE_PREFIX}_${module}"
    echo -e "   • ${db_name}"
done 