# ⚡ Быстрый запуск Expense Tracker

## 🚀 За 5 минут (локальная разработка)

### 1. Клонирование и настройка

```bash
# Клонируйте репозиторий
git clone https://github.com/audiogenius/expense-tracker.git
cd expense-tracker

# Создайте .env файл
cp env.example .env
nano .env
```

### 2. Настройка .env (минимально)

```env
# Telegram (обязательно)
TELEGRAM_BOT_TOKEN=ваш_токен_от_BotFather
TELEGRAM_WHITELIST=ваш_telegram_id

# База данных
POSTGRES_PASSWORD=любой_пароль

# Ключи безопасности (любые случайные строки)
BOT_API_KEY=случайная_строка_32_символа
JWT_SECRET=другая_случайная_строка_32
```

### 3. Создание Telegram бота

1. Перейдите к [@BotFather](https://t.me/BotFather)
2. Отправьте `/newbot`
3. Введите имя: `Expense Tracker`
4. Введите username: `your_expense_tracker_bot`
5. Скопируйте токен в `.env`

### 4. Получение Telegram ID

1. Найдите [@userinfobot](https://t.me/userinfobot)
2. Отправьте `/start`
3. Скопируйте ID в `TELEGRAM_WHITELIST`

### 5. Запуск

```bash
# Запуск проекта
docker-compose up --build -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 6. Тестирование

1. **Telegram бот:**
   - Найдите бота по username
   - Нажмите START
   - Попробуйте: `/help`, `100`, `100 продукты`

2. **Веб-интерфейс:**
   - Откройте: http://localhost
   - Нажмите "Simulate Login (dev)" для теста
   - Или войдите через Telegram Login Widget

---

## 🌐 Для продакшена

**Полная инструкция**: [HOSTING_SETUP.md](HOSTING_SETUP.md)

### Timeweb Cloud MSK 30 (2GB RAM)

```bash
# Подключение к серверу
ssh root@147.45.246.210

# Автоматическая установка
curl -sSL https://raw.githubusercontent.com/audiogenius/expense-tracker/main/deploy/timeweb-setup.sh | bash
```

---

## 🔧 Полезные команды

```bash
# Остановка
docker-compose down

# Перезапуск
docker-compose restart

# Обновление
git pull && docker-compose up --build -d

# Логи
docker-compose logs -f bot
docker-compose logs -f api

# Подключение к БД
docker-compose exec db psql -U expense_user -d expense_tracker
```

---

## 🐛 Решение проблем

**Бот не отвечает:**
- Проверьте `TELEGRAM_BOT_TOKEN` в `.env`
- Проверьте `TELEGRAM_WHITELIST` содержит ваш ID
- Проверьте логи: `docker-compose logs -f bot`

**Ошибка 500:**
- Проверьте `BOT_API_KEY` в `.env`
- Проверьте логи API: `docker-compose logs -f api`

**База данных не работает:**
- Проверьте `POSTGRES_PASSWORD` в `.env`
- Проверьте логи БД: `docker-compose logs -f db`

---

## 📱 Использование бота

- `100` - записать расход 100 руб.
- `100 продукты` - записать расход с категорией
- `/total` - показать общую сумму
- `/total week` - расходы за неделю
- `/total month` - расходы за месяц
- `/debts` - показать долги
- `/help` - справка

---

## 🎯 Что работает

### ✅ Полный функционал
- **Telegram бот** с командами и автоопределением категорий
- **Веб-интерфейс** с графиками и фильтрами
- **Умные подсказки** категорий с автодополнением
- **AI аналитика** с Ollama интеграцией
- **Оптимизация** для серверов с 2GB RAM
- **Виртуализация** списков для быстрой работы
- **Кэширование** для улучшения производительности

### 📊 Графики и аналитика
- Line chart - расходы за последние 7 дней
- Pie chart - распределение по категориям
- Фильтры по периодам и категориям
- Автоматические уведомления с аналитикой

---

## 📞 Нужна помощь?

1. **Документация**: [README.md](README.md)
2. **Развертывание**: [HOSTING_SETUP.md](HOSTING_SETUP.md)
3. **Обновления**: [docs/updates/](docs/updates/)
4. **GitHub Issues**: https://github.com/audiogenius/expense-tracker/issues

---

**Готово! Ваш Expense Tracker запущен! 🚀**