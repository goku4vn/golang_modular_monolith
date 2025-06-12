-- Customer Module Database Initialization
-- This file is responsible for creating the customer module database
-- Database name will be dynamically generated using environment variables

-- Create customer database with configurable prefix
-- This will be processed by the init script to replace variables
CREATE DATABASE ${DATABASE_PREFIX}_customer;

-- Grant permissions (if needed)
-- GRANT ALL PRIVILEGES ON DATABASE ${DATABASE_PREFIX}_customer TO postgres; 