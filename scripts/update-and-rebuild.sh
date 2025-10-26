#!/bin/bash

# Скрипт для обновления проекта на сервере и пересборки

set -e

echo "=== ОБНОВЛЕНИЕ И ПЕРЕСБОРКА ПРОЕКТА ==="

# 1. Pull latest changes
echo "1. Получение последних изменений из GitHub..."
git pull origin main

# 2. Stop services
echo "2. Остановка сервисов..."
docker-compose down

# 3. Rebuild only changed services (API and Frontend)
echo "3. Пересборка API и Frontend..."
docker-compose build --no-cache api frontend

# 4. Start all services
echo "4. Запуск всех сервисов..."
docker-compose up -d

# 5. Wait for services to start
echo "5. Ожидание запуска сервисов (10 секунд)..."
sleep 10

# 6. Check status
echo "6. Проверка статуса:"
docker-compose ps

# 7. Test API
echo "7. Тест API:"
curl -s http://localhost:8080/health || echo "API недоступен!"

echo ""
echo "=== ОБНОВЛЕНИЕ ЗАВЕРШЕНО ==="
echo "Проверьте сайт: https://rd-expense-tracker-bot.ru"

