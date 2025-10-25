-- Rollback migration 002: Remove suggestions indexes
-- Description: Remove indexes added for smart suggestions

-- Drop usage frequency indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_subcategory_usage_frequency;
DROP INDEX CONCURRENTLY IF EXISTS idx_category_usage_frequency;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_user_recent_operations;

-- Drop timestamp indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_user_timestamp_subcategory;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_user_timestamp_category;

-- Drop subcategory indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_subcategories_category_name;
DROP INDEX CONCURRENTLY IF EXISTS idx_subcategories_name_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_subcategories_name_trgm;

-- Drop category indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_categories_name_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_categories_name_trgm;

-- Note: pg_trgm extension is not dropped as it might be used by other features
