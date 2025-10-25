-- Migration: Add subcategories and operation_type functionality
-- Version: 001
-- Description: Adds subcategories table, operation_type to expenses, and subcategory_id
-- Compatibility: PostgreSQL 16+

BEGIN;

-- 1. Add operation_type field to expenses table
-- This will allow expenses to be marked as 'expense' or 'income'
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS operation_type VARCHAR(10) DEFAULT 'expense' 
CHECK (operation_type IN ('expense', 'income'));

-- 2. Create subcategories table with relationship to categories
CREATE TABLE IF NOT EXISTS subcategories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  aliases JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(name, category_id)
);

-- 3. Add subcategory_id to expenses table
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS subcategory_id INT REFERENCES subcategories(id);

-- 4. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subcategories_category ON subcategories(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_operation_type ON expenses(operation_type);
CREATE INDEX IF NOT EXISTS idx_expenses_subcategory ON expenses(subcategory_id);
CREATE INDEX IF NOT EXISTS idx_expenses_user_operation_timestamp ON expenses(user_id, operation_type, timestamp DESC);

-- 5. Add some default subcategories for existing categories
INSERT INTO subcategories (name, category_id, aliases) VALUES 
-- Продукты subcategories
('Молочные продукты', (SELECT id FROM categories WHERE name = 'Продукты'), '["молоко", "сыр", "йогурт", "творог"]'),
('Мясо и рыба', (SELECT id FROM categories WHERE name = 'Продукты'), '["мясо", "рыба", "курица", "говядина", "свинина"]'),
('Овощи и фрукты', (SELECT id FROM categories WHERE name = 'Продукты'), '["овощи", "фрукты", "яблоки", "бананы", "помидоры"]'),
('Хлеб и выпечка', (SELECT id FROM categories WHERE name = 'Продукты'), '["хлеб", "булочки", "печенье", "торт"]'),

-- Транспорт subcategories  
('Общественный транспорт', (SELECT id FROM categories WHERE name = 'Транспорт'), '["метро", "автобус", "троллейбус", "трамвай"]'),
('Такси', (SELECT id FROM categories WHERE name = 'Транспорт'), '["такси", "uber", "яндекс.такси"]'),
('Бензин', (SELECT id FROM categories WHERE name = 'Транспорт'), '["бензин", "заправка", "АЗС"]'),
('Парковка', (SELECT id FROM categories WHERE name = 'Транспорт'), '["парковка", "стоянка"]'),

-- Кафе и рестораны subcategories
('Рестораны', (SELECT id FROM categories WHERE name = 'Кафе и рестораны'), '["ресторан", "ужин", "обед в ресторане"]'),
('Кафе', (SELECT id FROM categories WHERE name = 'Кафе и рестораны'), '["кафе", "кофе", "завтрак"]'),
('Фастфуд', (SELECT id FROM categories WHERE name = 'Кафе и рестораны'), '["фастфуд", "бургер", "пицца", "шаурма"]'),
('Доставка', (SELECT id FROM categories WHERE name = 'Кафе и рестораны'), '["доставка", "еда на дом"]'),

-- Развлечения subcategories
('Кино', (SELECT id FROM categories WHERE name = 'Развлечения'), '["кино", "фильм", "кинотеатр"]'),
('Театр', (SELECT id FROM categories WHERE name = 'Развлечения'), '["театр", "спектакль", "балет", "опера"]'),
('Концерты', (SELECT id FROM categories WHERE name = 'Развлечения'), '["концерт", "музыка", "группа"]'),
('Игры', (SELECT id FROM categories WHERE name = 'Развлечения'), '["игры", "игровые автоматы", "бильярд"]'),

-- Здоровье subcategories
('Аптека', (SELECT id FROM categories WHERE name = 'Здоровье'), '["аптека", "лекарства", "таблетки", "витамины"]'),
('Врачи', (SELECT id FROM categories WHERE name = 'Здоровье'), '["врач", "доктор", "прием", "консультация"]'),
('Стоматология', (SELECT id FROM categories WHERE name = 'Здоровье'), '["стоматолог", "зубы", "лечение зубов"]'),
('Спорт', (SELECT id FROM categories WHERE name = 'Здоровье'), '["спорт", "тренажерный зал", "фитнес"]'),

-- Одежда subcategories
('Верхняя одежда', (SELECT id FROM categories WHERE name = 'Одежда'), '["куртка", "пальто", "плащ", "шуба"]'),
('Обувь', (SELECT id FROM categories WHERE name = 'Одежда'), '["обувь", "кроссовки", "ботинки", "туфли"]'),
('Нижнее белье', (SELECT id FROM categories WHERE name = 'Одежда'), '["белье", "трусы", "бюстгальтер", "носки"]'),
('Аксессуары', (SELECT id FROM categories WHERE name = 'Одежда'), '["сумка", "ремень", "шарф", "перчатки"]'),

-- Коммунальные услуги subcategories
('Электричество', (SELECT id FROM categories WHERE name = 'Коммунальные услуги'), '["свет", "электричество", "электроэнергия"]'),
('Вода', (SELECT id FROM categories WHERE name = 'Коммунальные услуги'), '["вода", "водоснабжение", "канализация"]'),
('Газ', (SELECT id FROM categories WHERE name = 'Коммунальные услуги'), '["газ", "газоснабжение"]'),
('Интернет и связь', (SELECT id FROM categories WHERE name = 'Коммунальные услуги'), '["интернет", "телефон", "мобильная связь"]')
ON CONFLICT (name, category_id) DO NOTHING;

-- 6. Update existing expenses to have operation_type = 'expense' (default value)
UPDATE expenses 
SET operation_type = 'expense' 
WHERE operation_type IS NULL;

-- 7. Create a view for backward compatibility with existing code
-- This view will show expenses and incomes as separate tables
CREATE OR REPLACE VIEW v_expenses AS
SELECT 
  id,
  user_id,
  amount_cents,
  category_id,
  subcategory_id,
  timestamp,
  is_shared
FROM expenses 
WHERE operation_type = 'expense';

CREATE OR REPLACE VIEW v_incomes AS
SELECT 
  id,
  user_id,
  amount_cents,
  category_id,
  subcategory_id,
  timestamp,
  is_shared,
  'other' as income_type,
  NULL as description,
  NULL as related_debt_id
FROM expenses 
WHERE operation_type = 'income';

-- 8. Create a function to migrate existing incomes table data to expenses
CREATE OR REPLACE FUNCTION migrate_incomes_to_expenses()
RETURNS INTEGER AS $$
DECLARE
  migrated_count INTEGER := 0;
  income_record RECORD;
BEGIN
  -- Migrate each income record to expenses table
  FOR income_record IN 
    SELECT * FROM incomes 
  LOOP
    INSERT INTO expenses (
      user_id, 
      amount_cents, 
      category_id, 
      timestamp, 
      is_shared,
      operation_type
    ) VALUES (
      income_record.user_id,
      income_record.amount_cents,
      NULL, -- No category for incomes in original schema
      income_record.timestamp,
      FALSE, -- Default for incomes
      'income'
    );
    
    migrated_count := migrated_count + 1;
  END LOOP;
  
  RETURN migrated_count;
END;
$$ LANGUAGE plpgsql;

-- 9. Add constraints for data integrity
DO $$ 
BEGIN
  -- Ensure operation_type is not null
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_operation_type_not_null') THEN
    ALTER TABLE expenses ADD CONSTRAINT check_operation_type_not_null CHECK (operation_type IS NOT NULL);
  END IF;
  
  -- Ensure subcategory belongs to the same category as expense
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_subcategory_category_match') THEN
    ALTER TABLE expenses ADD CONSTRAINT check_subcategory_category_match 
    CHECK (
      subcategory_id IS NULL OR 
      EXISTS (
        SELECT 1 FROM subcategories s 
        WHERE s.id = expenses.subcategory_id 
        AND s.category_id = expenses.category_id
      )
    );
  END IF;
END $$;

COMMIT;
