# Примеры тестирования API

## Тестирование новых эндпоинтов

### 1. Тестирование создания транзакций

#### Создание расхода с подкатегорией:
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 150000,
    "category_id": 1,
    "subcategory_id": 5,
    "operation_type": "expense",
    "timestamp": "2024-01-15T10:30:00Z",
    "is_shared": false
  }'
```

#### Создание дохода:
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 500000,
    "category_id": 1,
    "operation_type": "income",
    "timestamp": "2024-01-15T10:30:00Z"
  }'
```

### 2. Тестирование CRUD подкатегорий

#### Создание подкатегории:
```bash
curl -X POST http://localhost:8080/subcategories \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Молочные продукты",
    "category_id": 1,
    "aliases": ["молоко", "сыр", "йогурт", "творог"]
  }'
```

#### Получение всех подкатегорий:
```bash
curl -X GET http://localhost:8080/subcategories \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Получение подкатегорий конкретной категории:
```bash
curl -X GET "http://localhost:8080/subcategories?category_id=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Обновление подкатегории:
```bash
curl -X PUT http://localhost:8080/subcategories/5 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Молочные продукты (обновлено)",
    "category_id": 1,
    "aliases": ["молоко", "сыр", "йогурт", "творог", "кефир"]
  }'
```

#### Удаление подкатегории:
```bash
curl -X DELETE http://localhost:8080/subcategories/5 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. Тестирование эндпоинта транзакций

#### Получение всех транзакций:
```bash
curl -X GET http://localhost:8080/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Фильтрация по типу операции:
```bash
curl -X GET "http://localhost:8080/transactions?operation_type=expense" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Фильтрация по категории:
```bash
curl -X GET "http://localhost:8080/transactions?category_id=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Фильтрация по подкатегории:
```bash
curl -X GET "http://localhost:8080/transactions?subcategory_id=5" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Фильтрация по дате:
```bash
curl -X GET "http://localhost:8080/transactions?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Пагинация:
```bash
curl -X GET "http://localhost:8080/transactions?page=2&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Комбинированные фильтры:
```bash
curl -X GET "http://localhost:8080/transactions?operation_type=expense&category_id=1&page=1&limit=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 4. Тестирование подсказок категорий

#### Поиск по ключевому слову:
```bash
curl -X GET "http://localhost:8080/suggestions/categories?query=молоко" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Поиск по категории:
```bash
curl -X GET "http://localhost:8080/suggestions/categories?query=продукты" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Поиск по транспорту:
```bash
curl -X GET "http://localhost:8080/suggestions/categories?query=транспорт" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Тестирование валидации

### Неверные данные:

#### Неверный operation_type:
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 150000,
    "operation_type": "invalid_type"
  }'
```

#### Неверная подкатегория:
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 150000,
    "category_id": 1,
    "subcategory_id": 999
  }'
```

#### Отсутствие обязательного поля:
```bash
curl -X POST http://localhost:8080/subcategories \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1
  }'
```

## Тестирование производительности

### Нагрузочное тестирование:
```bash
# Тест пагинации с большим количеством данных
for i in {1..10}; do
  curl -X GET "http://localhost:8080/transactions?page=$i&limit=50" \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" &
done
wait
```

### Тест подсказок:
```bash
# Тест различных поисковых запросов
queries=("молоко" "транспорт" "кафе" "здоровье" "одежда")
for query in "${queries[@]}"; do
  curl -X GET "http://localhost:8080/suggestions/categories?query=$query" \
    -H "Authorization: Bearer YOUR_JWT_TOKEN"
done
```

## Ожидаемые результаты

### Успешные ответы:
- **201 Created** - для создания ресурсов
- **200 OK** - для получения данных
- **204 No Content** - для удаления

### Ошибки:
- **400 Bad Request** - неверные данные
- **401 Unauthorized** - проблемы с аутентификацией
- **404 Not Found** - ресурс не найден
- **500 Internal Server Error** - серверные ошибки

### Формат ответов:
- Все ответы в JSON
- Временные метки в RFC3339
- Суммы в копейках
- Пагинация с метаданными
- Детальная информация об ошибках
