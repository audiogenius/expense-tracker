-- Migration: Add Telegram groups support
-- Version: 004
-- Description: Adds support for Telegram groups, group members, and private/shared expenses
-- Compatibility: PostgreSQL 16+

BEGIN;

-- 1. Create telegram_groups table
CREATE TABLE IF NOT EXISTS telegram_groups (
    id BIGINT PRIMARY KEY,  -- Telegram chat_id
    name VARCHAR(255),
    type VARCHAR(20) DEFAULT 'group' CHECK (type IN ('group', 'supergroup', 'channel')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create group_members table
CREATE TABLE IF NOT EXISTS group_members (
    id SERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES telegram_groups(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(group_id, user_id)
);

-- 3. Add group_id and is_private to expenses table
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS group_id BIGINT REFERENCES telegram_groups(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_private BOOLEAN DEFAULT false;

-- 4. Add group_id and is_private to incomes table
ALTER TABLE incomes 
ADD COLUMN IF NOT EXISTS group_id BIGINT REFERENCES telegram_groups(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_private BOOLEAN DEFAULT false;

-- 5. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_telegram_groups_type ON telegram_groups(type);
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_group ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_expenses_group ON expenses(group_id);
CREATE INDEX IF NOT EXISTS idx_expenses_private ON expenses(is_private);
CREATE INDEX IF NOT EXISTS idx_incomes_group ON incomes(group_id);
CREATE INDEX IF NOT EXISTS idx_incomes_private ON incomes(is_private);

-- 6. Add comment explaining the logic
COMMENT ON COLUMN expenses.is_private IS 'true = personal expense (visible only to creator), false = shared with group';
COMMENT ON COLUMN expenses.group_id IS 'Telegram group where expense was created. NULL = created via website or bot DM';
COMMENT ON TABLE telegram_groups IS 'Telegram groups/chats where bot is added';
COMMENT ON TABLE group_members IS 'Members of Telegram groups (family members)';

COMMIT;

