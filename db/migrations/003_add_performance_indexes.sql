-- Migration: Add performance indexes for 2GB RAM server optimization
-- Description: Add critical indexes for fast queries and memory efficiency

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Critical indexes for expenses table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_user_timestamp 
ON expenses (user_id, timestamp DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_category_id 
ON expenses (category_id) 
WHERE category_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_subcategory_id 
ON expenses (subcategory_id) 
WHERE subcategory_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_operation_type 
ON expenses (operation_type);

-- Composite index for keyset pagination
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_pagination 
ON expenses (user_id, timestamp DESC, id);

-- Index for balance calculations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_balance_calc 
ON expenses (user_id, operation_type, timestamp DESC);

-- Index for category usage statistics
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_category_usage 
ON expenses (user_id, category_id, timestamp DESC) 
WHERE category_id IS NOT NULL AND timestamp >= NOW() - INTERVAL '30 days';

-- Index for subcategory usage statistics  
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_subcategory_usage 
ON expenses (user_id, subcategory_id, timestamp DESC) 
WHERE subcategory_id IS NOT NULL AND timestamp >= NOW() - INTERVAL '30 days';

-- Index for recent transactions (last 100 per user)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_recent 
ON expenses (user_id, timestamp DESC) 
WHERE timestamp >= NOW() - INTERVAL '7 days';

-- Partial indexes for common filters
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_shared 
ON expenses (user_id, timestamp DESC) 
WHERE is_shared = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_expenses_only 
ON expenses (user_id, timestamp DESC) 
WHERE operation_type = 'expense';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_incomes_only 
ON expenses (user_id, timestamp DESC) 
WHERE operation_type = 'income';

-- Index for analytics queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_analytics 
ON expenses (user_id, timestamp, operation_type, category_id) 
WHERE timestamp >= NOW() - INTERVAL '90 days';

-- Index for suggestions (already exists but ensure it's optimized)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_suggestions 
ON expenses (user_id, category_id, subcategory_id, timestamp DESC) 
WHERE timestamp >= NOW() - INTERVAL '30 days';

-- Index for users table (if not exists)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_telegram_id 
ON users (telegram_id);

-- Index for categories table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_name_lower 
ON categories (lower(name));

-- Index for subcategories table
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subcategories_category_name 
ON subcategories (category_id, name);

-- Update table statistics for better query planning
ANALYZE expenses;
ANALYZE categories;
ANALYZE subcategories;
ANALYZE users;
