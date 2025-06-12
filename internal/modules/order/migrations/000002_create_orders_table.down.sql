-- Drop trigger
DROP TRIGGER IF EXISTS trigger_orders_updated_at ON orders;

-- Drop function
DROP FUNCTION IF EXISTS update_orders_updated_at();

-- Drop table
DROP TABLE IF EXISTS orders;

-- Drop custom types
DROP TYPE IF EXISTS order_status;
