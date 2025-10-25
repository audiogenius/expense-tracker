# 🗄️ Database Setup Guide

## PostgreSQL
```bash
# Установка
sudo apt install postgresql postgresql-contrib

# Создание пользователя
sudo -u postgres createuser --interactive

# Создание базы данных
sudo -u postgres createdb expense_tracker

# Настройка пароля
sudo -u postgres psql
ALTER USER expense_user PASSWORD 'your_password';
```

## Миграции
```bash
# Выполнить миграции
docker-compose exec db psql -U expense_user -d expense_tracker -f /docker-entrypoint-initdb.d/init.sql

# Проверить таблицы
docker-compose exec db psql -U expense_user -d expense_tracker -c "\dt"
```

## Бэкап
```bash
# Создать бэкап
docker-compose exec db pg_dump -U expense_user expense_tracker > backup.sql

# Восстановить
docker-compose exec -T db psql -U expense_user expense_tracker < backup.sql
```
