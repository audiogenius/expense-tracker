# 💰 Expense Tracker v1.1 - Family Expense Management System

<div align="center">

![Version](https://img.shields.io/badge/version-1.1-blue)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
![Status](https://img.shields.io/badge/status-Production%20Ready-success)

**Минимальный семейный трекер расходов с Telegram ботом и веб-интерфейсом**

[Быстрый старт](#-быстрый-старт) • [Функции](#-основные-функции) • [Документация](#-документация) • [Скриншоты](#-что-внутри)

</div>

---

## 🎯 О проекте

Удобная система для отслеживания семейных расходов с:
- 🤖 **Telegram ботом** для быстрого ввода расходов
- 💻 **Веб-интерфейсом** с графиками и аналитикой
- 🏷️ **Автоопределением категорий** из текста
- 📊 **Красивыми графиками** Chart.js
- 🔐 **Безопасной авторизацией** через Telegram

---

## ⚡ Быстрый старт

### 1. Создайте Telegram бота

```
1. Откройте @BotFather в Telegram
2. Отправьте /newbot
3. Следуйте инструкциям
4. Скопируйте токен
```

### 2. Получите ваш Telegram ID

```
1. Откройте @userinfobot
2. Отправьте /start
3. Скопируйте ваш ID
```

### 3. Настройте .env

```bash
cp env.example .env
# Отредактируйте .env файл
```

### 4. Запустите проект

```bash
docker-compose up --build -d
```

### 5. Готово! 🎉

- **Telegram бот:** `t.me/your_bot_username`
- **Веб-интерфейс:** http://localhost

📖 **Подробная инструкция:** [START_HERE.md](START_HERE.md)

---

## ✨ Основные функции

### 🤖 Telegram Бот

```
/help          - Справка
/total         - Общая сумма расходов
/total week    - Расходы за неделю
/total month   - Расходы за месяц
/debts         - Показать долги

100            - Записать расход 100 руб.
100 продукты   - Расход с категорией
50.50 кафе     - Расход с копейками
```

### 💻 Веб-интерфейс

- 📊 **Графики:**
  - Line chart расходов за 7 дней
  - Pie chart по категориям
  
- 🔍 **Фильтры:**
  - По периодам (все/неделя/месяц)
  - По категориям
  
- ➕ **Добавление расходов:**
  - С выбором категории
  - С автоопределением категории
  
- 🔐 **Авторизация:**
  - Telegram Login Widget
  - JWT токены

### 🏷️ Категории (автоопределение)

- 🛒 Продукты
- 🚗 Транспорт
- ☕ Кафе и рестораны
- 🎭 Развлечения
- 💊 Здоровье
- 👔 Одежда
- 🏠 Коммунальные услуги
- 📦 Прочее

---

## 🏗️ Архитектура

```
┌─────────────┐
│   Telegram  │──────┐
│     Bot     │      │
└─────────────┘      │
                     │
┌─────────────┐      │    ┌──────────┐
│     Web     │──────┼───▶│   API    │
│  Frontend   │      │    │ Service  │
└─────────────┘      │    └──────────┘
                     │         │
┌─────────────┐      │         │
│     OCR     │──────┘         │
│   Service   │                │
└─────────────┘                ▼
                        ┌──────────┐
                        │PostgreSQL│
                        └──────────┘
```

### Технологии:

| Компонент | Стек |
|-----------|------|
| Backend | Go 1.23, Chi router, pgx |
| Frontend | React 18, TypeScript, Vite, Chart.js |
| Database | PostgreSQL 16 |
| Auth | JWT, Telegram OAuth |
| Infrastructure | Docker Compose, Nginx |
| OCR | Google Cloud Vision / Tesseract *(planned)* |

---

## 📁 Структура проекта

```
expense-tracker/
├── api-service/          # REST API (Go)
│   ├── cmd/api/         # Main entry point
│   └── internal/
│       ├── auth/        # JWT & Telegram auth
│       └── handlers/    # HTTP handlers
├── bot-service/         # Telegram bot (Go)
│   └── cmd/            
├── frontend-service/    # Web UI (React + TS)
│   └── src/
│       ├── App.tsx     # Main component
│       └── styles.css  # Styles
├── ocr-service/         # OCR processing (Go)
├── db/                  # Database
│   └── init.sql        # Schema + indexes
├── docker-compose.yml   # Docker config
└── docs/               # Documentation
```

---

## 🚀 Что внутри

### ✅ Полностью реализовано:

- [x] База данных с индексами и constraints
- [x] REST API с JWT авторизацией
- [x] Telegram Login Widget
- [x] Telegram бот с командами
- [x] Автоопределение категорий
- [x] Веб-интерфейс с графиками (Chart.js)
- [x] Фильтры по периодам и категориям
- [x] Responsive дизайн
- [x] Health checks для всех сервисов
- [x] Docker Compose с правильными зависимостями
- [x] Подробная документация

### ⏳ Запланировано:

- [ ] OCR обработка чеков
- [ ] Shared expenses (разделение счета)
- [ ] Unit тесты (70% coverage)
- [ ] Бюджеты и лимиты
- [ ] Экспорт данных (CSV/XLSX)
- [ ] ML категоризация

---

## 📚 Документация

| Файл | Описание |
|------|----------|
| [START_HERE.md](START_HERE.md) | 🚀 **Начните отсюда!** Быстрый запуск за 5 минут |
| [TELEGRAM_SETUP_GUIDE.md](TELEGRAM_SETUP_GUIDE.md) | 🤖 Подробная настройка Telegram (для новичков) |
| [CHANGES_SUMMARY.md](CHANGES_SUMMARY.md) | 📝 Список всех изменений в проекте |
| [QUICKSTART.md](QUICKSTART.md) | ⚡ Краткая инструкция по запуску |
| [DEPLOY.md](DEPLOY.md) | 🌐 Развертывание на сервере |
| [SETUP_GITHUB.md](SETUP_GITHUB.md) | 📦 Развертывание с GitHub |

---

## 🔧 Команды разработки

### Docker

```bash
# Запуск
docker-compose up -d

# Остановка
docker-compose down

# Пересборка
docker-compose up --build -d

# Логи
docker-compose logs -f bot
docker-compose logs -f api

# Статус (с health checks)
docker-compose ps
```

### База данных

```bash
# Подключение
docker-compose exec db psql -U expense_user -d expense_tracker

# Бэкап
docker-compose exec db pg_dump -U expense_user expense_tracker > backup.sql

# Восстановление
cat backup.sql | docker-compose exec -T db psql -U expense_user -d expense_tracker
```

### Frontend разработка

```bash
cd frontend-service
npm install
npm run dev      # Development server
npm run build    # Production build
```

---

## 🎨 Скриншоты

### Веб-интерфейс

```
┌──────────────────────────────────────┐
│  Expense Tracker                      │
│  Вошли как: @username           [X]  │
└──────────────────────────────────────┘

┌──────────────────────────────────────┐
│  📊 Общие расходы                    │
│  ────────────────────────────────    │
│  За выбранный период: 15,340.50 ₽   │
│  [Все] [Неделя] [Месяц]             │
└──────────────────────────────────────┘

┌──────────────────────────────────────┐
│  ➕ Добавить расход                  │
│  [100.00] [Категория ▼] [Добавить]  │
└──────────────────────────────────────┘

┌─────────────┬────────────────────────┐
│ 📈 Расходы  │  🥧 По категориям     │
│ за 7 дней   │                        │
│             │                        │
│  [График]   │      [Pie Chart]       │
│             │                        │
└─────────────┴────────────────────────┘

┌──────────────────────────────────────┐
│  📋 Последние расходы  [Фильтр ▼]   │
│  ────────────────────────────────    │
│  21.10.2025 18:30                    │
│  Продукты                   150.00 ₽│
│  ────────────────────────────────    │
│  21.10.2025 12:15                    │
│  Кафе                        85.50 ₽│
└──────────────────────────────────────┘
```

---

## 💡 Примеры использования

### Telegram бот

```
Вы: /help
Бот: 🤖 Expense Tracker Bot
     
     📋 Команды:
     /help - показать эту справку
     /total - показать общую сумму расходов
     ...

Вы: 100
Бот: ✅ Записал расход: 100 руб.

Вы: 50.50 кафе
Бот: ✅ Записал расход: 50.50 руб. (категория: Кафе и рестораны)

Вы: /total week
Бот: 📊 Расходы за неделю: 1,250.50 руб.
```

### REST API

```bash
# Получить категории
curl http://localhost:8080/categories

# Определить категорию
curl -X POST http://localhost:8080/categories/detect \
  -H "Content-Type: application/json" \
  -d '{"description":"купил хлеб и молоко"}'
# → {"id":1,"name":"Продукты","score":2}

# Health check
curl http://localhost:8080/health
# → {"status":"ok","service":"api"}
```

---

## 🐛 Решение проблем

### Бот не отвечает

```bash
# Проверьте логи
docker-compose logs bot

# Проверьте переменные окружения
cat .env | grep TELEGRAM

# Перезапустите
docker-compose restart bot
```

### База данных не работает

```bash
# Проверьте health
docker-compose ps db

# Посмотрите логи
docker-compose logs db

# Подключитесь напрямую
docker-compose exec db psql -U expense_user -d expense_tracker
```

### Frontend не загружается

```bash
# Проверьте статус
docker-compose ps

# Пересоберите
docker-compose up --build frontend -d

# Проверьте логи
docker-compose logs frontend proxy
```

📖 **Подробнее:** [TELEGRAM_SETUP_GUIDE.md](TELEGRAM_SETUP_GUIDE.md#-решение-проблем)

---

## 🤝 Вклад в проект

1. Fork проекта
2. Создайте ветку (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

---

## 📄 Лицензия

MIT License - используйте свободно!

---

## 🙏 Благодарности

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Chart.js](https://www.chartjs.org/)
- [Chi router](https://github.com/go-chi/chi)
- [pgx](https://github.com/jackc/pgx)

---

## 📞 Контакты

- **Issues:** [GitHub Issues](https://github.com/yourusername/expense-tracker/issues)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/expense-tracker/discussions)

---

<div align="center">

**Сделано с ❤️ для семейного бюджета**

⭐ Поставьте звезду если проект понравился!

[Telegram Setup](TELEGRAM_SETUP_GUIDE.md) • [Quick Start](START_HERE.md) • [Changes](CHANGES_SUMMARY.md)

</div>
