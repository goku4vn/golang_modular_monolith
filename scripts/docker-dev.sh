#!/bin/bash

# Docker development environment setup script

set -e

echo "🐳 Starting Docker development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker compose is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose (modern) is not available. Please update Docker to latest version."
    exit 1
fi

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker compose -f docker/docker-compose.dev.yml down || true

# Build development image
echo "🔨 Building development Docker image..."
docker compose -f docker/docker-compose.dev.yml build

# Start PostgreSQL first
echo "📦 Starting PostgreSQL container..."
docker compose -f docker/docker-compose.dev.yml up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

echo "📋 PostgreSQL is ready!"
echo "💡 To create databases manually, run: make create-databases"
echo "💡 To run migrations, run: make migrate-up"

# Start application with hot reload
echo "🔥 Starting application with hot reload..."
echo "📝 Server will be available at: http://localhost:8080"
echo "🏥 Health check: http://localhost:8080/health"
echo "📚 API endpoints: http://localhost:8080/api/v1/"
echo ""
echo "🐳 Docker containers:"
echo "  - Application: tmm-dev"
echo "  - PostgreSQL: tmm-postgres-dev"
echo "  - Redis: tmm-redis-dev"
echo "  - Vault: tmm-vault-dev"
echo ""
echo "📋 Useful commands:"
echo "  - View logs: make docker-dev-logs"
echo "  - Access shell: make docker-dev-shell"
echo "  - Stop environment: make docker-dev-down"
echo ""
echo "Press Ctrl+C to stop the development server"
echo ""

# Start application (this will show logs)
docker compose -f docker/docker-compose.dev.yml up app 