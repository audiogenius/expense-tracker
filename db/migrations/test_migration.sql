-- Test script for migration 001
-- This script tests the migration and ensures backward compatibility

-- Test 1: Check that operation_type field was added correctly
SELECT 
  column_name, 
  data_type, 
  is_nullable, 
  column_default
FROM information_schema.columns 
WHERE table_name = 'expenses' 
  AND column_name = 'operation_type';

-- Test 2: Check that subcategories table was created
SELECT 
  table_name, 
  column_name, 
  data_type, 
  is_nullable
FROM information_schema.columns 
WHERE table_name = 'subcategories'
ORDER BY ordinal_position;

-- Test 3: Check that subcategory_id was added to expenses
SELECT 
  column_name, 
  data_type, 
  is_nullable
FROM information_schema.columns 
WHERE table_name = 'expenses' 
  AND column_name = 'subcategory_id';

-- Test 4: Check that indexes were created
SELECT 
  indexname, 
  tablename, 
  indexdef
FROM pg_indexes 
WHERE tablename IN ('expenses', 'subcategories')
  AND indexname LIKE 'idx_%'
ORDER BY tablename, indexname;

-- Test 5: Check that constraints were added
SELECT 
  conname, 
  contype, 
  pg_get_constraintdef(oid) as definition
FROM pg_constraint 
WHERE conrelid = 'expenses'::regclass
  AND conname LIKE 'check_%'
ORDER BY conname;

-- Test 6: Check that default subcategories were inserted
SELECT 
  s.name as subcategory_name,
  c.name as category_name,
  s.aliases
FROM subcategories s
JOIN categories c ON s.category_id = c.id
ORDER BY c.name, s.name;

-- Test 7: Test backward compatibility views
-- Check that v_expenses view works
SELECT COUNT(*) as expense_count FROM v_expenses;

-- Check that v_incomes view works  
SELECT COUNT(*) as income_count FROM v_incomes;

-- Test 8: Test data integrity constraints
-- This should fail if constraints are working
-- INSERT INTO expenses (user_id, amount_cents, operation_type) VALUES (1, 100, 'invalid_type');

-- Test 9: Test subcategory constraint
-- This should fail if constraint is working
-- INSERT INTO expenses (user_id, amount_cents, category_id, subcategory_id) 
-- VALUES (1, 100, 1, (SELECT id FROM subcategories WHERE category_id != 1 LIMIT 1));

-- Test 10: Check that existing data is preserved
SELECT 
  operation_type,
  COUNT(*) as count
FROM expenses 
GROUP BY operation_type;

-- Test 11: Test migration function (if incomes table has data)
-- This is commented out to avoid actually running the migration
-- SELECT migrate_incomes_to_expenses();

-- Test 12: Verify that all existing expenses have operation_type = 'expense'
SELECT COUNT(*) as expenses_without_operation_type
FROM expenses 
WHERE operation_type IS NULL;

-- Test 13: Check that subcategory relationships are valid
SELECT 
  COUNT(*) as invalid_subcategory_relationships
FROM expenses e
JOIN subcategories s ON e.subcategory_id = s.id
WHERE e.category_id != s.category_id;

-- Test 14: Performance test - check that new indexes are being used
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM expenses 
WHERE operation_type = 'expense' 
  AND user_id = 1 
ORDER BY timestamp DESC;

-- Test 15: Test subcategory filtering
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM expenses 
WHERE subcategory_id IS NOT NULL 
  AND operation_type = 'expense';
