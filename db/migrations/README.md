# Database Migrations for Expense Tracker

This directory contains database migration scripts for the Expense Tracker application.

## Migration 001: Add Subcategories and Operation Type

### Description
Adds support for subcategories and operation type (expense/income) to the expenses table.

### Changes Made
1. **Added `operation_type` field** to `expenses` table with values 'expense' or 'income'
2. **Created `subcategories` table** with relationship to `categories`
3. **Added `subcategory_id` field** to `expenses` table
4. **Created default subcategories** for existing categories
5. **Added performance indexes** for new fields
6. **Created backward compatibility views** (`v_expenses`, `v_incomes`)
7. **Added data integrity constraints**

### Files
- `001_add_subcategories_and_operation_type.sql` - Main migration script
- `001_rollback.sql` - Rollback script to revert changes
- `test_migration.sql` - Test script to verify migration
- `README.md` - This documentation

### Usage

#### Apply Migration
```sql
-- Run the migration
\i db/migrations/001_add_subcategories_and_operation_type.sql
```

#### Test Migration
```sql
-- Test the migration
\i db/migrations/test_migration.sql
```

#### Rollback Migration
```sql
-- Rollback the migration (WARNING: This will remove all subcategories and operation_type data!)
\i db/migrations/001_rollback.sql
```

### Backward Compatibility

The migration maintains backward compatibility through:

1. **Views**: `v_expenses` and `v_incomes` provide the same interface as before
2. **Default Values**: Existing expenses get `operation_type = 'expense'` by default
3. **Original Tables**: The original `incomes` table is preserved
4. **Migration Function**: `migrate_incomes_to_expenses()` can migrate existing income data

### New Features

#### Subcategories
- Each category can have multiple subcategories
- Subcategories have aliases for flexible matching
- Subcategories are linked to expenses via `subcategory_id`

#### Operation Types
- Expenses can be marked as 'expense' or 'income'
- Allows unified handling of all financial transactions
- Maintains separation through views for backward compatibility

### Data Migration

The migration includes a function to migrate existing income data:

```sql
-- Migrate existing incomes to expenses table
SELECT migrate_incomes_to_expenses();
```

### Performance Considerations

New indexes added:
- `idx_expenses_operation_type` - For filtering by operation type
- `idx_expenses_subcategory` - For filtering by subcategory
- `idx_expenses_user_operation_timestamp` - For user-specific operation queries
- `idx_subcategories_category` - For subcategory lookups

### Constraints Added

1. **Operation Type Constraint**: Ensures only 'expense' or 'income' values
2. **Subcategory Category Match**: Ensures subcategory belongs to the same category as expense
3. **Not Null Constraints**: Ensures operation_type is always specified

### Testing

The test script verifies:
- All fields were added correctly
- Indexes were created
- Constraints are working
- Backward compatibility views function
- Data integrity is maintained
- Performance indexes are being used

### Rollback Considerations

⚠️ **WARNING**: Rolling back this migration will:
- Remove all subcategories and their data
- Remove operation_type field and data
- Remove all related indexes and constraints
- Drop backward compatibility views

Make sure to backup your data before applying or rolling back migrations.
