-- Drop trigger
DROP TRIGGER IF EXISTS update_customers_updated_at ON "public"."customers";

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS "public"."customers";

-- Drop enum type
DROP TYPE IF EXISTS "public"."customer_status";
