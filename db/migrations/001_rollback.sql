-- Rollback script for migration 001
-- This script reverts all changes made in 001_add_subcategories_and_operation_type.sql
-- WARNING: This will remove all subcategories and operation_type data!

BEGIN;

-- 1. Drop views first (they depend on the table structure)
DROP VIEW IF EXISTS v_expenses;
DROP VIEW IF EXISTS v_incomes;

-- 2. Drop the migration function
DROP FUNCTION IF EXISTS migrate_incomes_to_expenses();

-- 3. Drop indexes that were added
DROP INDEX IF EXISTS idx_subcategories_category;
DROP INDEX IF EXISTS idx_expenses_operation_type;
DROP INDEX IF EXISTS idx_expenses_subcategory;
DROP INDEX IF EXISTS idx_expenses_user_operation_timestamp;

-- 4. Drop constraints that were added
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS check_operation_type_not_null;
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS check_subcategory_category_match;

-- 5. Remove columns that were added
ALTER TABLE expenses DROP COLUMN IF EXISTS operation_type;
ALTER TABLE expenses DROP COLUMN IF EXISTS subcategory_id;

-- 6. Drop the subcategories table
DROP TABLE IF EXISTS subcategories;

COMMIT;
