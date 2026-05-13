ALTER TABLE orders
DROP COLUMN IF EXISTS pickup_address,
DROP COLUMN IF EXISTS delivery_cost;