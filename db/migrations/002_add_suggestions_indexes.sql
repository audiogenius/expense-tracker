-- Migration: Add indexes for smart suggestions
-- Description: Add indexes for category and subcategory search optimization

-- Enable pg_trgm extension for similarity search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add indexes for category search optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_name_trgm 
ON categories USING gin (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_name_lower 
ON categories (lower(name));

-- Add indexes for subcategory search optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subcategories_name_trgm 
ON subcategories USING gin (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subcategories_name_lower 
ON subcategories (lower(name));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subcategories_category_name 
ON subcategories (category_id, name);

-- Add indexes for usage statistics
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_user_timestamp_category 
ON expenses (user_id, timestamp DESC, category_id) 
WHERE category_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_user_timestamp_subcategory 
ON expenses (user_id, timestamp DESC, subcategory_id) 
WHERE subcategory_id IS NOT NULL;

-- Add composite index for recent operations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_user_recent_operations 
ON expenses (user_id, timestamp DESC) 
WHERE timestamp >= NOW() - INTERVAL '30 days';

-- Add index for category usage frequency
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_category_usage_frequency 
ON expenses (category_id, user_id, timestamp DESC) 
WHERE category_id IS NOT NULL AND timestamp >= NOW() - INTERVAL '30 days';

-- Add index for subcategory usage frequency  
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subcategory_usage_frequency 
ON expenses (subcategory_id, user_id, timestamp DESC) 
WHERE subcategory_id IS NOT NULL AND timestamp >= NOW() - INTERVAL '30 days';
