# 🔧 Исправление ошибки 401 при аутентификации

## 🚨 Проблема
При развертывании на хостинге возникает ошибка:
```
Login failed: AxiosError: Request failed with status code 401
```

## ✅ Решение

### 1. **Проверьте переменные окружения**

Убедитесь, что в `.env` файле правильно настроены:

```env
# КРИТИЧЕСКИ ВАЖНО: JWT Secret должен быть сильным
JWT_SECRET=your_very_secure_jwt_secret_key_here_minimum_32_characters

# Telegram Bot Token от @BotFather
TELEGRAM_BOT_TOKEN=your_bot_token_from_botfather

# Whitelist: ваш Telegram ID или '*' для всех пользователей
TELEGRAM_WHITELIST=your_telegram_id_here
# или для всех пользователей:
TELEGRAM_WHITELIST=*
```

### 2. **Получите свой Telegram ID**

```bash
# Отправьте /start боту @userinfobot в Telegram
# Скопируйте ваш ID и добавьте в TELEGRAM_WHITELIST
```

### 3. **Перезапустите сервисы**

```bash
# Остановить все
docker-compose down

# Запустить заново
docker-compose up -d

# Проверить логи
docker-compose logs -f api
```

### 4. **Проверьте логи**

```bash
# Следить за логами API
docker-compose logs -f api

# Проверить все сервисы
docker-compose ps
```

## 🛠️ Архитектурные улучшения

### Добавлены новые модули:

1. **`internal/auth/errors.go`** - Централизованная обработка ошибок аутентификации
2. **`internal/auth/validator.go`** - Валидация конфигурации и whitelist
3. **`internal/auth/response.go`** - Стандартизированные ответы API
4. **`internal/handlers/auth_handler.go`** - Отдельный обработчик аутентификации
5. **`internal/middleware/cors.go`** - CORS middleware для cross-origin запросов

### Принципы чистой архитектуры:

- ✅ **Разделение ответственности**: Каждый модуль отвечает за свою область
- ✅ **Инверсия зависимостей**: Модули не зависят от конкретных реализаций
- ✅ **Единая ответственность**: Каждый файл решает одну задачу
- ✅ **Открытость/закрытость**: Легко расширять без изменения существующего кода

## 🧪 Тестирование

### Автоматический тест:
```bash
# Запустить тест аутентификации
./scripts/test-auth.sh

# Диагностика проблем
./scripts/debug-auth.sh
```

### Ручная проверка:
```bash
# 1. Проверить здоровье API
curl http://your-domain.com/health

# 2. Тест логина
curl -X POST http://your-domain.com/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "your_telegram_id",
    "username": "your_username",
    "first_name": "Your",
    "last_name": "Name"
  }'
```

## 🔍 Диагностика

### Частые проблемы:

1. **403 Forbidden** → Пользователь не в whitelist
2. **401 Unauthorized** → Проблема с JWT или Telegram токеном
3. **CORS ошибки** → Проблемы с cross-origin запросами
4. **500 Internal Server Error** → Проблемы с базой данных

### Решения:

```bash
# 1. Проверить переменные
echo $JWT_SECRET
echo $TELEGRAM_BOT_TOKEN
echo $TELEGRAM_WHITELIST

# 2. Проверить логи
docker-compose logs api | grep -i error

# 3. Перезапустить с очисткой
docker-compose down
docker system prune -f
docker-compose up --build -d
```

## 📱 Frontend исправления

### TelegramLogin.tsx обновлен:
- Улучшена обработка ошибок
- Добавлена диагностика проблем с доменом
- Лучшая обработка CORS

### API клиент обновлен:
- Добавлены правильные заголовки
- Улучшена обработка ошибок
- Кэширование для производительности

## 🚀 Развертывание

### Для продакшена:
```bash
# 1. Настройте .env файл
cp env.example .env
nano .env

# 2. Установите сильные пароли
JWT_SECRET=$(openssl rand -base64 32)
TELEGRAM_WHITELIST=your_telegram_id

# 3. Запустите
docker-compose up -d

# 4. Проверьте
./scripts/test-auth.sh
```

## 📞 Поддержка

Если проблема не решается:

1. Проверьте все логи: `docker-compose logs`
2. Убедитесь, что база данных доступна
3. Проверьте сетевые настройки
4. Создайте issue в GitHub с логами

---

**✅ После применения этих исправлений ошибка 401 должна быть устранена!**
