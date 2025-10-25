-- Rollback migration 003: Remove performance indexes
-- Description: Remove indexes added for performance optimization

-- Drop analytics indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_suggestions;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_analytics;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_incomes_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_expenses_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_shared;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_recent;

-- Drop usage statistics indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_subcategory_usage;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_category_usage;

-- Drop balance calculation indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_balance_calc;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_pagination;

-- Drop main indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_operation_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_subcategory_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_category_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_expenses_user_timestamp;

-- Drop table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_subcategories_category_name;
DROP INDEX CONCURRENTLY IF EXISTS idx_categories_name_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_telegram_id;
