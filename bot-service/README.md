# 🤖 Bot Service

## Описание
Telegram бот для управления расходами через чат.

## Функции
- Команды бота
- Запись расходов/доходов
- Уведомления
- Интеграция с API

## Архитектура
- Go + Telegram Bot API
- HTTP клиент для API
- Команды и обработчики

## Запуск
```bash
docker-compose up bot
```

## Конфигурация
- TELEGRAM_BOT_TOKEN
- API_URL
- BOT_API_KEY

## Команды бота
- /start - начать работу
- /help - помощь
- /expense - добавить расход
- /income - добавить доход
- /balance - баланс

## Логи
```bash
docker-compose logs -f bot
```
