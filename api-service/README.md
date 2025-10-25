# 🔌 API Service

## Описание
Основной API сервис для управления расходами и доходами.

## Функции
- REST API для CRUD операций
- JWT авторизация
- Умные подсказки категорий
- In-memory кэширование
- Keyset pagination

## Архитектура
- Go + Gin
- PostgreSQL
- JWT токены
- Кэширование

## Запуск
```bash
docker-compose up api
```

## Конфигурация
- DATABASE_URL
- JWT_SECRET
- API_PORT

## API Endpoints
- GET /health - проверка здоровья
- GET /api/categories - список категорий
- GET /api/expenses - список расходов
- POST /api/expenses - добавить расход
- GET /api/suggestions/categories - подсказки категорий

## Логи
```bash
docker-compose logs -f api
```
