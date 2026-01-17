-- Clean all data from database
-- 
-- To run this file:
--   docker exec -i jonosize-postgres psql -U jonosize -d jonosize < clean_database.sql
-- 
-- Or connect directly:
--   psql -h localhost -U jonosize -d jonosize -f clean_database.sql
--
-- Option 1: Using DELETE (respects foreign key constraints with CASCADE)
-- Delete in order to respect foreign key relationships

DELETE FROM clicks;
DELETE FROM links;
DELETE FROM campaign_products;
DELETE FROM offers;
DELETE FROM campaigns;
DELETE FROM products;

-- Option 2: Using TRUNCATE (faster, resets sequences, but requires CASCADE for foreign keys)
-- Uncomment below if you prefer TRUNCATE instead of DELETE

-- TRUNCATE TABLE clicks CASCADE;
-- TRUNCATE TABLE links CASCADE;
-- TRUNCATE TABLE campaign_products CASCADE;
-- TRUNCATE TABLE offers CASCADE;
-- TRUNCATE TABLE campaigns CASCADE;
-- TRUNCATE TABLE products CASCADE;

-- Option 3: Reset all sequences (if using auto-increment IDs, but this project uses UUID)
-- Not needed for this project as it uses UUID

-- Verify tables are empty
-- SELECT 
--     (SELECT COUNT(*) FROM clicks) as clicks_count,
--     (SELECT COUNT(*) FROM links) as links_count,
--     (SELECT COUNT(*) FROM campaign_products) as campaign_products_count,
--     (SELECT COUNT(*) FROM offers) as offers_count,
--     (SELECT COUNT(*) FROM campaigns) as campaigns_count,
--     (SELECT COUNT(*) FROM products) as products_count;
