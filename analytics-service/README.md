# 🤖 Analytics Service

## Описание
Сервис аналитики с интеграцией Ollama для умного анализа расходов.

## Функции
- AI анализ расходов через Ollama
- Автоматические уведомления в Telegram
- Fallback на правило-основанную аналитику
- Планировщик задач

## Архитектура
- Go + Ollama
- HTTP API
- Scheduler для периодических задач
- Telegram интеграция

## Запуск
```bash
docker-compose up analytics
```

## Конфигурация
- OLLAMA_URL
- OLLAMA_MODEL
- TELEGRAM_BOT_TOKEN
- DATABASE_URL

## API Endpoints
- GET /health - проверка здоровья
- GET /ollama/status - статус Ollama
- POST /analytics/process - обработка аналитики

## Логи
```bash
docker-compose logs -f analytics
```
