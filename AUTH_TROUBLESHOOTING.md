# 🔧 Исправление проблем с аутентификацией

## 🚨 Ошибка 401: Request failed with status code 401

### Возможные причины и решения:

#### 1. **Проблема с Whitelist (403 Forbidden)**
```bash
# Проверьте переменную TELEGRAM_WHITELIST в .env
echo $TELEGRAM_WHITELIST

# Добавьте свой Telegram ID или '*' для всех пользователей
TELEGRAM_WHITELIST=your_telegram_id_here
# или
TELEGRAM_WHITELIST=*
```

#### 2. **Проблема с JWT Secret**
```bash
# Убедитесь, что JWT_SECRET установлен
JWT_SECRET=your_very_secure_jwt_secret_key_here
```

#### 3. **Проблема с Telegram Bot Token**
```bash
# Проверьте токен бота
TELEGRAM_BOT_TOKEN=your_bot_token_from_botfather
```

#### 4. **Проблема с CORS**
- Добавлен CORS middleware в код
- Проверьте, что запросы идут с правильного домена

### 🔍 Диагностика

#### Проверка логов:
```bash
# Логи API сервиса
docker logs expense_api -f

# Логи всех сервисов
docker-compose logs -f
```

#### Проверка переменных окружения:
```bash
# Проверить .env файл
cat .env | grep -E "(JWT_SECRET|TELEGRAM_BOT_TOKEN|TELEGRAM_WHITELIST)"
```

#### Тест API:
```bash
# Проверка здоровья API
curl http://localhost:8080/health

# Тест логина (замените на свои данные)
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "your_telegram_id",
    "username": "your_username",
    "first_name": "Your",
    "last_name": "Name"
  }'
```

### 🛠️ Пошаговое исправление

#### Шаг 1: Проверьте .env файл
```bash
# Убедитесь, что все переменные установлены
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=your_strong_password
POSTGRES_DB=expense_tracker
JWT_SECRET=your_jwt_secret_key
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_WHITELIST=your_telegram_id
```

#### Шаг 2: Перезапустите сервисы
```bash
# Остановить все сервисы
docker-compose down

# Запустить заново
docker-compose up -d

# Проверить статус
docker-compose ps
```

#### Шаг 3: Проверьте логи
```bash
# Следить за логами в реальном времени
docker-compose logs -f api
```

#### Шаг 4: Тестирование
1. Откройте браузер на `http://your-domain.com`
2. Попробуйте войти через Telegram
3. Проверьте Network tab в DevTools на ошибки

### 🚀 Быстрое исправление

Если проблема критическая, выполните:

```bash
# 1. Остановить все
docker-compose down

# 2. Очистить кэш
docker system prune -f

# 3. Пересобрать и запустить
docker-compose up --build -d

# 4. Проверить логи
docker-compose logs -f api
```

### 📱 Проверка Telegram Widget

Убедитесь, что в TelegramLogin.tsx правильно указан bot username:
```typescript
script.setAttribute('data-telegram-login', 'your_bot_username')
```

### 🔐 Безопасность

Для продакшена:
1. Используйте сильные пароли для JWT_SECRET
2. Ограничьте TELEGRAM_WHITELIST конкретными ID
3. Настройте HTTPS
4. Регулярно обновляйте токены

### 📞 Поддержка

Если проблема не решается:
1. Проверьте все логи: `docker-compose logs`
2. Убедитесь, что база данных доступна
3. Проверьте сетевые настройки
4. Создайте issue в GitHub с логами
