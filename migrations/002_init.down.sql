DROP table IF EXISTS reviews;

ALTER TABLE couriers
DROP COLUMN IF EXISTS rating,
DROP COLUMN IF EXISTS total_reviews;