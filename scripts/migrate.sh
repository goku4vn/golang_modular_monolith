#!/bin/bash

# Migration script to run migrations inside Docker container
# This ensures proper network connectivity to database services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

# Check if app container is running
if ! docker ps | grep -q "tmm-dev"; then
    print_error "Application container is not running. Please run 'make docker-dev' first."
    exit 1
fi

# Default values
MODULE=""
ACTION="up"
VERSION=""
NAME=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--module)
            MODULE="$2"
            shift 2
            ;;
        -a|--action)
            ACTION="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -n|--name)
            NAME="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -m, --module MODULE    Module name or 'all' for all enabled modules"
            echo "  -a, --action ACTION    Migration action (up, down, version, reset, create)"
            echo "  -v, --version VERSION  Target version for migrate"
            echo "  -n, --name NAME        Migration name for create action"
            echo "  -h, --help            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 -m customer -a up                    # Migrate customer module up"
            echo "  $0 -m all -a version                    # Show version for all modules"
            echo "  $0 -m customer -a create -n add_email   # Create new migration"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use -h or --help for usage information."
            exit 1
            ;;
    esac
done

# Show available modules if no module specified
if [[ -z "$MODULE" ]]; then
    print_info "Getting available modules..."
    docker exec tmm-dev go run cmd/migrate/main.go 2>/dev/null || true
    exit 1
fi

# Build migration command
MIGRATE_CMD="go run cmd/migrate/main.go -module=$MODULE -action=$ACTION"

if [[ -n "$VERSION" ]]; then
    MIGRATE_CMD="$MIGRATE_CMD -version=$VERSION"
fi

if [[ -n "$NAME" ]]; then
    MIGRATE_CMD="$MIGRATE_CMD -name=$NAME"
fi

print_info "Running migration command: $MIGRATE_CMD"
print_info "Module: $MODULE, Action: $ACTION"

# Execute migration inside Docker container
if docker exec tmm-dev $MIGRATE_CMD; then
    print_success "Migration completed successfully!"
else
    print_error "Migration failed!"
    exit 1
fi 