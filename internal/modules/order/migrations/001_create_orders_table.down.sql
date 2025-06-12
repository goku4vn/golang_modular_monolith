-- Drop indexes first
DROP INDEX IF EXISTS idx_orders_order_date;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer_id;

-- Drop orders table
DROP TABLE IF EXISTS orders; 