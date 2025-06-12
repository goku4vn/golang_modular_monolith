#!/bin/bash

# Wrapper script to run the application with environment variables

export CUSTOMER_DATABASE_HOST=localhost
export CUSTOMER_DATABASE_PORT=5433
export CUSTOMER_DATABASE_USER=postgres
export CUSTOMER_DATABASE_PASSWORD=postgres
export CUSTOMER_DATABASE_NAME=modular_monolith_customer
export CUSTOMER_DATABASE_SSLMODE=disable

export ORDER_DATABASE_HOST=localhost
export ORDER_DATABASE_PORT=5433
export ORDER_DATABASE_USER=postgres
export ORDER_DATABASE_PASSWORD=postgres
export ORDER_DATABASE_NAME=modular_monolith_order
export ORDER_DATABASE_SSLMODE=disable

export GIN_MODE=debug

# Run the binary
exec ./tmp/main 