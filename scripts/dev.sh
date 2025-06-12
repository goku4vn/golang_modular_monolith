#!/bin/bash

# Development environment setup script

set -e

echo "🚀 Starting development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start PostgreSQL if not running
if ! docker ps | grep -q postgres-tmm; then
    echo "📦 Starting PostgreSQL container..."
    make docker-up
else
    echo "✅ PostgreSQL container is already running"
fi

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 3

# Run migrations
echo "🔄 Running database migrations..."
CUSTOMER_DATABASE_HOST=localhost \
CUSTOMER_DATABASE_PORT=5433 \
CUSTOMER_DATABASE_USER=postgres \
CUSTOMER_DATABASE_PASSWORD=postgres \
CUSTOMER_DATABASE_NAME=modular_monolith_customer \
CUSTOMER_DATABASE_SSLMODE=disable \
ORDER_DATABASE_HOST=localhost \
ORDER_DATABASE_PORT=5433 \
ORDER_DATABASE_USER=postgres \
ORDER_DATABASE_PASSWORD=postgres \
ORDER_DATABASE_NAME=modular_monolith_order \
ORDER_DATABASE_SSLMODE=disable \
make migrate-all-up

# Start development server with hot reload
echo "🔥 Starting development server with hot reload..."
echo "📝 Server will be available at: http://localhost:8080"
echo "🏥 Health check: http://localhost:8080/health"
echo "📚 API endpoints: http://localhost:8080/api/v1/"
echo ""
echo "Press Ctrl+C to stop the development server"
echo ""

make run-dev 