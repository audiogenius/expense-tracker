#!/bin/bash

echo "=== Диагностика проблем с авторизацией ==="

# 1. Проверка статуса контейнеров
echo "1. Статус контейнеров:"
docker-compose ps
echo ""

# 2. Проверка переменных окружения
echo "2. Переменные окружения:"
echo "JWT_SECRET: $(grep JWT_SECRET .env | cut -d'=' -f2 | cut -c1-10)..."
echo "TELEGRAM_BOT_TOKEN: $(grep TELEGRAM_BOT_TOKEN .env | cut -d'=' -f2 | cut -c1-10)..."
echo "TELEGRAM_WHITELIST: $(grep TELEGRAM_WHITELIST .env | cut -d'=' -f2)"
echo ""

# 3. Проверка логов API сервиса
echo "3. Логи API сервиса (последние 20 строк):"
docker-compose logs --tail=20 api
echo ""

# 4. Тест health endpoint
echo "4. Тест health endpoint:"
curl -s http://localhost:8080/health || echo "Health endpoint недоступен"
echo ""

# 5. Тест categories endpoint
echo "5. Тест categories endpoint:"
curl -s http://localhost:8080/categories | head -c 100 || echo "Categories endpoint недоступен"
echo ""

# 6. Проверка базы данных
echo "6. Проверка подключения к базе данных:"
docker-compose exec -T db psql -U expense_user -d expense_tracker -c "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "Ошибка подключения к БД"
echo ""

# 7. Проверка пользователей в БД
echo "7. Пользователи в базе данных:"
docker-compose exec -T db psql -U expense_user -d expense_tracker -c "SELECT id, telegram_id, username FROM users;" 2>/dev/null || echo "Ошибка получения пользователей"
echo ""

# 8. Тест логина (замените на ваши данные)
echo "8. Тест логина (замените TELEGRAM_ID на ваш):"
TELEGRAM_ID="260144148"  # Замените на ваш Telegram ID
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$TELEGRAM_ID\",
    \"username\": \"testuser\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\"
  }" 2>/dev/null || echo "Ошибка тестирования логина"
echo ""

# 9. Проверка портов
echo "9. Проверка портов:"
netstat -tlnp | grep -E ":(80|8080|5432)" || echo "Порты не найдены"
echo ""

echo "=== Диагностика завершена ==="
