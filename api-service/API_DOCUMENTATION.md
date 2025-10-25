# API Documentation - Expense Tracker v1.2

## Новые эндпоинты и функциональность

### 1. Обновленная модель Expense

#### POST /expenses
Создание транзакции (расход/доход) с поддержкой новых полей.

**Request Body:**
```json
{
  "amount_cents": 150000,
  "category_id": 1,
  "subcategory_id": 5,
  "operation_type": "expense",
  "timestamp": "2024-01-15T10:30:00Z",
  "is_shared": false
}
```

**Поля:**
- `amount_cents` (int, обязательное) - сумма в копейках
- `category_id` (int, опциональное) - ID категории
- `subcategory_id` (int, опциональное) - ID подкатегории
- `operation_type` (string, опциональное) - "expense" или "income" (по умолчанию "expense")
- `timestamp` (string, опциональное) - время в RFC3339 формате
- `is_shared` (bool, опциональное) - общая транзакция

**Response:**
```json
{
  "id": 123
}
```

### 2. CRUD для подкатегорий

#### POST /subcategories
Создание новой подкатегории.

**Request Body:**
```json
{
  "name": "Молочные продукты",
  "category_id": 1,
  "aliases": ["молоко", "сыр", "йогурт"]
}
```

**Response:**
```json
{
  "id": 5,
  "name": "Молочные продукты",
  "category_id": 1,
  "aliases": ["молоко", "сыр", "йогурт"],
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### GET /subcategories
Получение списка подкатегорий.

**Query Parameters:**
- `category_id` (опциональное) - фильтр по категории

**Response:**
```json
[
  {
    "id": 5,
    "name": "Молочные продукты",
    "category_id": 1,
    "category_name": "Продукты",
    "aliases": ["молоко", "сыр", "йогурт"],
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

#### PUT /subcategories/{id}
Обновление подкатегории.

**Request Body:** (аналогично POST)

#### DELETE /subcategories/{id}
Удаление подкатегории (только если не используется в транзакциях).

### 3. Унифицированный эндпоинт транзакций

#### GET /transactions
Получение транзакций с пагинацией и фильтрами.

**Query Parameters:**
- `operation_type` (опциональное) - "expense", "income", "both"
- `category_id` (опциональное) - фильтр по категории
- `subcategory_id` (опциональное) - фильтр по подкатегории
- `start_date` (опциональное) - начальная дата (RFC3339)
- `end_date` (опциональное) - конечная дата (RFC3339)
- `page` (опциональное) - номер страницы (по умолчанию 1)
- `limit` (опциональное) - количество записей на странице (по умолчанию 50, максимум 200)

**Пример запроса:**
```
GET /transactions?operation_type=expense&category_id=1&page=1&limit=20
```

**Response:**
```json
{
  "transactions": [
    {
      "id": 123,
      "user_id": 1,
      "amount_cents": 150000,
      "category_id": 1,
      "subcategory_id": 5,
      "operation_type": "expense",
      "timestamp": "2024-01-15T10:30:00Z",
      "is_shared": false,
      "username": "user1",
      "category_name": "Продукты",
      "subcategory_name": "Молочные продукты"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_count": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  },
  "filters": {
    "operation_type": "expense",
    "category_id": "1",
    "subcategory_id": "",
    "start_date": "",
    "end_date": ""
  }
}
```

### 4. Умные подсказки категорий

#### GET /suggestions/categories
Получение подсказок категорий на основе текстового запроса.

**Query Parameters:**
- `query` (обязательное) - поисковый запрос

**Пример запроса:**
```
GET /suggestions/categories?query=молоко
```

**Response:**
```json
[
  {
    "id": 5,
    "name": "Молочные продукты",
    "type": "subcategory",
    "score": 3
  },
  {
    "id": 1,
    "name": "Продукты",
    "type": "category",
    "score": 2
  }
]
```

**Алгоритм поиска:**
- Поиск по названию категории/подкатегории (вес 3/2)
- Поиск по алиасам (вес 2/1)
- Сортировка по релевантности (score)
- Ограничение до 10 результатов

## Валидация и обработка ошибок

### Коды ошибок:
- `400 Bad Request` - неверные параметры запроса
- `401 Unauthorized` - отсутствует или неверный токен
- `404 Not Found` - ресурс не найден
- `500 Internal Server Error` - внутренняя ошибка сервера

### Валидация:
1. **operation_type** - только "expense" или "income"
2. **subcategory_id** - должен принадлежать указанной категории
3. **amount_cents** - положительное число
4. **timestamp** - валидный RFC3339 формат
5. **category_id** - существующая категория
6. **subcategory_id** - существующая подкатегория

### Обработка ошибок:
- Все ошибки логируются с контекстом
- Пользователю возвращается понятное сообщение об ошибке
- Валидация на уровне базы данных (constraints)
- Проверка существования связанных сущностей

## Обратная совместимость

Все существующие эндпоинты сохранены:
- `/expenses` - работает как раньше, но с новыми полями
- `/incomes` - сохранен для совместимости
- `/categories` - без изменений
- `/debts` - без изменений
- `/balance` - без изменений

## Производительность

### Индексы:
- `idx_expenses_operation_type` - для фильтрации по типу операции
- `idx_expenses_subcategory` - для фильтрации по подкатегории
- `idx_expenses_user_operation_timestamp` - для пользовательских запросов
- `idx_subcategories_category` - для поиска подкатегорий

### Оптимизации:
- Пагинация для больших наборов данных
- Эффективные JOIN'ы с индексами
- Кэширование частых запросов
- Ограничение лимита записей (максимум 200)

## Примеры использования

### Создание расхода с подкатегорией:
```bash
curl -X POST /expenses \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 150000,
    "category_id": 1,
    "subcategory_id": 5,
    "operation_type": "expense"
  }'
```

### Поиск транзакций за месяц:
```bash
curl "/transactions?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&page=1&limit=50"
```

### Получение подсказок:
```bash
curl "/suggestions/categories?query=транспорт"
```
