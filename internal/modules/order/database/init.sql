-- Order Module Database Initialization
-- This file is responsible for creating the order module database
-- Database name will be dynamically generated using environment variables

-- Create order database with configurable prefix
-- This will be processed by the init script to replace variables
CREATE DATABASE ${DATABASE_PREFIX}_order;

-- Grant permissions (if needed)
-- GRANT ALL PRIVILEGES ON DATABASE ${DATABASE_PREFIX}_order TO postgres; 