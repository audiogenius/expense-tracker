# Настройка Analytics Service

## 🚀 Быстрый старт

### 1. Запуск с Docker Compose

```bash
# Клонируйте проект и перейдите в директорию
cd expense-tracker

# Скопируйте и настройте переменные окружения
cp env.example .env
# Отредактируйте .env файл с вашими настройками

# Запустите все сервисы
docker-compose up -d

# Инициализируйте Ollama модель (займет несколько минут)
docker-compose exec ollama ollama pull qwen2.5:0.5b

# Проверьте статус
curl http://localhost:8081/health
```

### 2. Проверка работы

```bash
# Проверка health check
curl http://localhost:8081/health

# Проверка статуса Ollama
curl http://localhost:8081/api/v1/ollama/status

# Ручной запуск анализа
curl -X POST http://localhost:8081/api/v1/analyze/trigger

# Список запланированных задач
curl http://localhost:8081/api/v1/scheduler/jobs
```

## ⚙️ Конфигурация

### Переменные окружения

```bash
# Обязательные
DATABASE_URL=postgres://user:pass@db:5432/expense_tracker
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_IDS=123456789,987654321

# Опциональные
ANALYTICS_PORT=8081
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=qwen2.5:0.5b
```

### Настройка Telegram

1. **Создайте бота** через @BotFather
2. **Получите токен** и добавьте в `.env`
3. **Получите Chat ID** пользователей:
   ```bash
   # Отправьте сообщение боту, затем:
   curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates"
   ```
4. **Добавьте Chat ID** в `TELEGRAM_CHAT_IDS`

## 📊 API Endpoints

### Health Check
```bash
GET /health
```

**Ответ:**
```json
{
  "service": "analytics-service",
  "status": "healthy",
  "ollama": true,
  "database": true,
  "last_check": "2024-01-15T20:00:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0"
}
```

### Анализ периода
```bash
GET /api/v1/analyze?period=day&days=7
```

**Параметры:**
- `period` - тип периода (day/week/month)
- `days` - количество дней для анализа

### Ручной запуск анализа
```bash
POST /api/v1/analyze/trigger?period=day
```

### Отправка сообщений
```bash
POST /api/v1/messages/send
Content-Type: application/json

{
  "chat_ids": [123456789],
  "type": "daily",
  "period": "day"
}
```

**Типы сообщений:**
- `daily` - ежедневный отчет
- `anomaly` - уведомление об аномалии
- `trend` - анализ трендов

### Статус Ollama
```bash
GET /api/v1/ollama/status
```

### Список задач
```bash
GET /api/v1/scheduler/jobs
```

## 🤖 Ollama настройка

### Автоматическая инициализация

```bash
# Модель автоматически загружается при первом запуске
docker-compose up -d ollama

# Проверка статуса
curl http://localhost:11434/api/tags
```

### Ручная инициализация

```bash
# Подключение к контейнеру Ollama
docker-compose exec ollama bash

# Загрузка модели
ollama pull qwen2.5:0.5b

# Проверка доступных моделей
ollama list
```

### Альтернативные модели

```bash
# Более быстрая модель (меньше функций)
ollama pull qwen2.5:0.5b

# Более мощная модель (больше функций)
ollama pull qwen2.5:1.5b

# Обновление конфигурации
# Измените OLLAMA_MODEL в .env файле
```

## 📅 Расписание задач

### Автоматические задачи

| Время | Задача | Описание |
|-------|--------|----------|
| 20:00 MSK | Ежедневный отчет | Анализ дня и отправка отчета |
| Каждые 6 часов | Проверка аномалий | Поиск необычных трат |
| Воскресенье 21:00 | Анализ трендов | Еженедельный анализ |
| Каждый час | Health check | Проверка состояния сервисов |

### Настройка расписания

```bash
# Просмотр текущих задач
curl http://localhost:8081/api/v1/scheduler/jobs

# Добавление кастомной задачи (через код)
# В scheduler.go добавьте:
_, err := s.cron.AddFunc("0 9 * * *", func() {
    // Ваша задача
})
```

## 🔍 Мониторинг

### Логи

```bash
# Просмотр логов analytics-service
docker-compose logs -f analytics

# Просмотр логов Ollama
docker-compose logs -f ollama

# Все логи
docker-compose logs -f
```

### Метрики

```bash
# Статус сервисов
curl http://localhost:8081/health

# Статус базы данных
docker-compose exec db pg_isready

# Статус Ollama
curl http://localhost:11434/api/tags
```

### Алерты

Система автоматически отправляет уведомления при:
- Недоступности Ollama (fallback на rule-based логику)
- Ошибках базы данных
- Высоких аномалиях в тратах (>2x среднего)
- Неудачной отправке сообщений

## 🛠️ Разработка

### Локальная разработка

```bash
# Установка зависимостей
cd analytics-service
go mod tidy

# Запуск тестов
go test ./...

# Запуск с hot reload
air

# Сборка
go build -o analytics-service ./cmd/analytics
```

### Структура проекта

```
analytics-service/
├── cmd/analytics/          # Точка входа
├── internal/
│   ├── analytics/         # Движок анализа
│   ├── messaging/         # Генератор сообщений
│   ├── ollama/           # Ollama клиент
│   ├── scheduler/        # Планировщик
│   ├── handlers/         # HTTP обработчики
│   └── types/            # Типы данных
├── scripts/              # Скрипты инициализации
└── Dockerfile
```

### Добавление новых анализов

1. **Расширьте типы** в `internal/types/types.go`
2. **Добавьте логику** в `internal/analytics/engine.go`
3. **Создайте шаблоны** в `internal/messaging/generator.go`
4. **Добавьте в планировщик** в `internal/scheduler/scheduler.go`

## 🚨 Устранение неполадок

### Ollama не запускается

```bash
# Проверка ресурсов
docker stats expense_ollama

# Увеличение лимитов памяти
# В docker-compose.yml:
deploy:
  resources:
    limits:
      memory: 4G
```

### База данных недоступна

```bash
# Проверка подключения
docker-compose exec analytics ping db

# Проверка переменных окружения
docker-compose exec analytics env | grep DATABASE
```

### Telegram сообщения не отправляются

```bash
# Проверка токена
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Проверка Chat ID
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates"
```

### Высокое потребление ресурсов

```bash
# Мониторинг ресурсов
docker stats

# Ограничение ресурсов Ollama
# В docker-compose.yml:
deploy:
  resources:
    limits:
      memory: 2G
      cpus: '1.0'
```

## 📈 Производительность

### Оптимизация Ollama

```bash
# Использование GPU (если доступно)
docker-compose exec ollama ollama serve --gpu

# Настройка количества потоков
export OLLAMA_NUM_PARALLEL=2
export OLLAMA_MAX_LOADED_MODELS=1
```

### Оптимизация базы данных

```sql
-- Добавление индексов для аналитики
CREATE INDEX CONCURRENTLY idx_expenses_timestamp_operation 
ON expenses(timestamp, operation_type);

CREATE INDEX CONCURRENTLY idx_expenses_category_timestamp 
ON expenses(category_id, timestamp);
```

## 🔒 Безопасность

### Переменные окружения

```bash
# Никогда не коммитьте .env файл
echo ".env" >> .gitignore

# Используйте сильные пароли
POSTGRES_PASSWORD=your_strong_password_here
JWT_SECRET=your_very_long_jwt_secret_here
```

### Сетевая безопасность

```yaml
# В docker-compose.yml ограничьте доступ:
services:
  ollama:
    networks:
      - internal
    # Не экспонируйте порт наружу
```

## 📞 Поддержка

При возникновении проблем:

1. **Проверьте логи**: `docker-compose logs analytics`
2. **Проверьте статус**: `curl http://localhost:8081/health`
3. **Проверьте ресурсы**: `docker stats`
4. **Перезапустите сервисы**: `docker-compose restart analytics ollama`

### Частые проблемы

- **Ollama не отвечает**: Увеличьте лимиты памяти
- **База недоступна**: Проверьте переменные окружения
- **Telegram не работает**: Проверьте токен и Chat ID
- **Высокая нагрузка**: Ограничьте ресурсы Ollama
