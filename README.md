# 💰 Expense Tracker - Умный трекер расходов

> **Версия**: v0.1 (25.10.2025)  
> **Статус**: ✅ Готов к продакшену  
> **Оптимизация**: Для серверов с 2GB RAM

## 🚀 Быстрый старт

### Для разработки (локально)
```bash
# Клонирование
git clone https://github.com/audiogenius/expense-tracker.git
cd expense-tracker

# Настройка
cp env.example .env
# Отредактируйте .env файл

# Запуск
docker-compose up --build -d
```

### Для продакшена (сервер)
См. [HOSTING_SETUP.md](HOSTING_SETUP.md) - полная инструкция по развертыванию на хостинге.

---

## ✨ Основные возможности

### 🤖 Telegram бот
- Запись расходов: `100`, `100 продукты`
- Команды: `/total`, `/debts`, `/help`
- Автоопределение категорий
- Умные подсказки

### 🌐 Веб-интерфейс
- **Графики**: Line chart (расходы по дням), Pie chart (по категориям)
- **Фильтры**: По периодам и категориям
- **Автодополнение**: Умные подсказки категорий
- **Виртуализация**: Быстрая работа с большими списками
- **Telegram Login**: Вход через Telegram

### 🧠 Умная аналитика
- **Ollama интеграция**: Локальная AI для анализа расходов
- **Автоматические уведомления**: Telegram с аналитикой
- **Fallback**: Правило-основанная аналитика при недоступности AI

### ⚡ Оптимизация производительности
- **Keyset pagination**: Быстрая загрузка больших списков
- **In-memory кэш**: 5-минутный кэш для частых запросов
- **Виртуализация**: Рендеринг только видимых элементов
- **Мемоизация**: Оптимизированные компоненты React

---

## 🏗 Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Telegram Bot  │    │   Web Frontend  │    │   Analytics     │
│                 │    │                 │    │                 │
│  - Команды      │    │  - React SPA    │    │  - Ollama AI    │
│  - Запись       │    │  - Charts.js    │    │  - Уведомления  │
│  - Уведомления  │    │  - Виртуализация │    │  - Fallback     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                               │
                    ┌─────────────────┐
                    │   API Service   │
                    │                 │
                    │  - REST API     │
                    │  - JWT Auth     │
                    │  - Кэширование  │
                    │  - Подсказки    │
                    └─────────────────┘
                               │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │                 │
                    │  - Оптимизированные индексы
                    │  - pg_trgm поиск
                    │  - Constraints
                    └─────────────────┘
```

---

## 🛠 Технологии

### Backend
- **Go** - API сервис
- **PostgreSQL** - База данных с оптимизированными индексами
- **Docker** - Контейнеризация
- **JWT** - Авторизация

### Frontend
- **React + TypeScript** - SPA приложение
- **Chart.js** - Графики и диаграммы
- **react-window** - Виртуализация списков
- **Vite** - Сборка

### AI & Analytics
- **Ollama** - Локальная LLM (qwen2.5:0.5b)
- **Go** - Analytics сервис
- **Telegram Bot API** - Уведомления

### DevOps
- **Docker Compose** - Оркестрация
- **Nginx** - Reverse proxy
- **SSL/TLS** - Безопасность
- **Health checks** - Мониторинг

---

## 📁 Структура проекта

```
expense-tracker/
├── api-service/           # Go API сервис
│   ├── cmd/api/          # Точка входа
│   ├── internal/         # Внутренние пакеты
│   │   ├── auth/         # JWT авторизация
│   │   ├── handlers/     # HTTP обработчики
│   │   └── cache/        # In-memory кэш
│   └── Dockerfile
├── bot-service/          # Telegram бот
│   ├── cmd/              # Точка входа
│   └── Dockerfile
├── frontend-service/     # React приложение
│   ├── src/
│   │   ├── components/   # React компоненты
│   │   │   ├── Suggestions/    # Автодополнение
│   │   │   ├── Transactions/   # Виртуализированные списки
│   │   │   └── Charts/         # Мемоизированные графики
│   │   ├── utils/         # Утилиты (кэш, API)
│   │   └── styles/        # CSS модули
│   └── Dockerfile
├── analytics-service/    # AI аналитика
│   ├── internal/
│   │   ├── ollama/       # Ollama клиент
│   │   ├── handlers/     # HTTP обработчики
│   │   └── scheduler/    # Планировщик задач
│   └── Dockerfile
├── db/                   # База данных
│   ├── init.sql         # Схема БД
│   └── migrations/      # Миграции
├── deploy/              # Скрипты развертывания
│   ├── timeweb-setup.sh
│   ├── docker-compose.timeweb.yml
│   └── init-ollama-timeweb.sh
├── docs/                # Документация
│   └── updates/         # История обновлений
├── docker-compose.yml   # Основная конфигурация
├── docker-compose.prod.yml  # Продакшен конфигурация
└── env.example          # Пример переменных окружения
```

---

## 🔧 Настройка

### Переменные окружения

```env
# База данных
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=your_strong_password
POSTGRES_DB=expense_tracker

# Telegram бот
TELEGRAM_BOT_TOKEN=your_bot_token_from_botfather
TELEGRAM_WHITELIST=your_telegram_id

# API ключи
BOT_API_KEY=random_secure_key
JWT_SECRET=another_random_secure_key

# Ollama (для аналитики)
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=qwen2.5:0.5b
OLLAMA_NUM_PARALLEL=1
OLLAMA_MAX_LOADED_MODELS=1
```

### Получение Telegram токенов

1. **Создание бота**: [@BotFather](https://t.me/BotFather) → `/newbot`
2. **Получение ID**: [@userinfobot](https://t.me/userinfobot) → `/start`

---

## 🚀 Развертывание

### Локальная разработка
```bash
# Клонирование
git clone https://github.com/audiogenius/expense-tracker.git
cd expense-tracker

# Настройка
cp env.example .env
# Отредактируйте .env

# Запуск
docker-compose up --build -d
```

### Продакшен (Timeweb Cloud MSK 30)
```bash
# Подключение к серверу
ssh root@147.45.246.210

# Установка (автоматический скрипт)
curl -sSL https://raw.githubusercontent.com/audiogenius/expense-tracker/main/deploy/timeweb-setup.sh | bash

# Или ручная установка
git clone https://github.com/audiogenius/expense-tracker.git
cd expense-tracker
cp deploy/docker-compose.timeweb.yml docker-compose.yml
# Настройте .env и запустите
```

Подробная инструкция: [HOSTING_SETUP.md](HOSTING_SETUP.md)

---

## 📊 Мониторинг

### Health checks
```bash
# API
curl http://localhost:8080/api/health

# Analytics
curl http://localhost:8081/health

# Ollama
curl http://localhost:11434/api/tags
```

### Логи
```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f api
docker-compose logs -f bot
docker-compose logs -f analytics
```

### Ресурсы
```bash
# Использование памяти
free -h

# Docker статистика
docker stats

# Диск
df -h
```

---

## 🔄 Обновления

### Текущая версия: v0.1 (25.10.2025)
- ✅ Система умных подсказок
- ✅ Оптимизация для 2GB RAM
- ✅ Ollama интеграция
- ✅ Виртуализация и кэширование

### История обновлений
- [v0.1 - 25.10.2025](docs/updates/v0.1-2025-10-25.md) - Первая оптимизированная версия

---

## 🐛 Решение проблем

### Высокое потребление памяти
```bash
# Перезапуск тяжелых сервисов
docker-compose restart ollama analytics
```

### Бот не отвечает
```bash
# Проверка логов
docker-compose logs -f bot

# Перезапуск
docker-compose restart bot
```

### Медленная загрузка
```bash
# Очистка кэша
docker-compose restart api

# Проверка индексов БД
docker-compose exec db psql -U expense_user -d expense_tracker -c "\di"
```

---

## 📞 Поддержка

- **GitHub Issues**: https://github.com/audiogenius/expense-tracker/issues
- **Документация**: [docs/](docs/)
- **Обновления**: [docs/updates/](docs/updates/)

---

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE)

---

**Expense Tracker v0.1 - Умный трекер расходов с AI аналитикой! 🚀**