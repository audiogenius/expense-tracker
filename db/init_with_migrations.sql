-- Initial schema for Expense Tracker v1.2 with subcategories and operation_type
-- This file includes all the migration changes in the initial schema

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  telegram_id BIGINT UNIQUE NOT NULL,
  username VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS categories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  aliases JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default categories
INSERT INTO categories (name, aliases) VALUES 
('Продукты', '["еда", "продукты", "магазин", "супермаркет", "продуктовый"]'),
('Транспорт', '["транспорт", "бензин", "такси", "метро", "автобус", "поезд"]'),
('Кафе и рестораны', '["кафе", "ресторан", "еда", "обед", "ужин", "кофе"]'),
('Развлечения', '["кино", "театр", "концерт", "игры", "развлечения"]'),
('Здоровье', '["аптека", "врач", "медицина", "здоровье", "лекарства"]'),
('Одежда', '["одежда", "обувь", "магазин", "шопинг"]'),
('Коммунальные услуги', '["коммуналка", "свет", "вода", "газ", "интернет", "телефон"]'),
('Прочее', '["прочее", "другое", "разное"]')
ON CONFLICT DO NOTHING;

-- Create subcategories table
CREATE TABLE IF NOT EXISTS subcategories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  aliases JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(name, category_id)
);

-- Create expenses table with operation_type and subcategory_id
CREATE TABLE IF NOT EXISTS expenses (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  amount_cents INT NOT NULL,
  category_id INT REFERENCES categories(id),
  subcategory_id INT REFERENCES subcategories(id),
  operation_type VARCHAR(10) DEFAULT 'expense' CHECK (operation_type IN ('expense', 'income')),
  timestamp TIMESTAMPTZ DEFAULT NOW(),
  is_shared BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS receipts (
  id SERIAL PRIMARY KEY,
  owner_id INT REFERENCES users(id),
  image_path TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receipt_items (
  id SERIAL PRIMARY KEY,
  receipt_id INT REFERENCES receipts(id),
  name TEXT,
  price_cents INT,
  selected_by JSONB DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS debts (
  id SERIAL PRIMARY KEY,
  from_user INT REFERENCES users(id),
  to_user INT REFERENCES users(id),
  amount_cents INT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  is_paid BOOLEAN DEFAULT FALSE,
  paid_at TIMESTAMPTZ
);

-- Keep the original incomes table for backward compatibility
-- Income types: salary, debt_return, prize, gift, refund, other
CREATE TABLE IF NOT EXISTS incomes (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  amount_cents INT NOT NULL,
  income_type VARCHAR(50) NOT NULL DEFAULT 'other',
  description TEXT,
  related_debt_id INT REFERENCES debts(id),
  timestamp TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default subcategories
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

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_expenses_user_timestamp ON expenses(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_timestamp ON expenses(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_operation_type ON expenses(operation_type);
CREATE INDEX IF NOT EXISTS idx_expenses_subcategory ON expenses(subcategory_id);
CREATE INDEX IF NOT EXISTS idx_expenses_user_operation_timestamp ON expenses(user_id, operation_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_subcategories_category ON subcategories(category_id);
CREATE INDEX IF NOT EXISTS idx_debts_from_user ON debts(from_user);
CREATE INDEX IF NOT EXISTS idx_debts_to_user ON debts(to_user);
CREATE INDEX IF NOT EXISTS idx_receipt_items_receipt ON receipt_items(receipt_id);
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_incomes_user_timestamp ON incomes(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_incomes_timestamp ON incomes(timestamp DESC);

-- Add constraints for data integrity
DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_amount_positive') THEN
    ALTER TABLE expenses ADD CONSTRAINT check_amount_positive CHECK (amount_cents > 0);
  END IF;
  
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_debt_amount_positive') THEN
    ALTER TABLE debts ADD CONSTRAINT check_debt_amount_positive CHECK (amount_cents > 0);
  END IF;
  
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_different_users') THEN
    ALTER TABLE debts ADD CONSTRAINT check_different_users CHECK (from_user != to_user);
  END IF;
  
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_income_amount_positive') THEN
    ALTER TABLE incomes ADD CONSTRAINT check_income_amount_positive CHECK (amount_cents > 0);
  END IF;
  
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_operation_type_not_null') THEN
    ALTER TABLE expenses ADD CONSTRAINT check_operation_type_not_null CHECK (operation_type IS NOT NULL);
  END IF;
  
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

-- Create backward compatibility views
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
