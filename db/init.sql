-- Initial schema for Expense Tracker v1.1

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

CREATE TABLE IF NOT EXISTS expenses (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  amount_cents INT NOT NULL,
  category_id INT REFERENCES categories(id),
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
  amount_cents INT
);
