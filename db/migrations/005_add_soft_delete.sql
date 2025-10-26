-- Migration: Add soft delete support for expenses
-- Description: Add deleted_at column for soft delete functionality

-- Add deleted_at column to expenses table
ALTER TABLE expenses ADD COLUMN deleted_at TIMESTAMP NULL;

-- Add index for soft delete queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_deleted_at 
ON expenses (deleted_at) 
WHERE deleted_at IS NULL;

-- Add index for deleted expenses
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_deleted 
ON expenses (deleted_at) 
WHERE deleted_at IS NOT NULL;

-- Update existing expenses to have deleted_at = NULL
UPDATE expenses SET deleted_at = NULL WHERE deleted_at IS NULL;
