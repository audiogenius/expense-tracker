-- Rollback for Migration 004: Remove Telegram groups support
-- Version: 004
-- Description: Removes Telegram groups tables and related columns

BEGIN;

-- 1. Drop indexes
DROP INDEX IF EXISTS idx_incomes_private;
DROP INDEX IF EXISTS idx_incomes_group;
DROP INDEX IF EXISTS idx_expenses_private;
DROP INDEX IF EXISTS idx_expenses_group;
DROP INDEX IF EXISTS idx_group_members_group;
DROP INDEX IF EXISTS idx_group_members_user;
DROP INDEX IF EXISTS idx_telegram_groups_type;

-- 2. Remove columns from incomes
ALTER TABLE incomes 
DROP COLUMN IF EXISTS is_private,
DROP COLUMN IF EXISTS group_id;

-- 3. Remove columns from expenses
ALTER TABLE expenses 
DROP COLUMN IF EXISTS is_private,
DROP COLUMN IF EXISTS group_id;

-- 4. Drop tables
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS telegram_groups;

COMMIT;

