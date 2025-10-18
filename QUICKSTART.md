# 🚀 Быстрый запуск Expense Tracker

## ⚡ За 5 минут

### 1. Клонирование и настройка

```bash
# Клонируйте репозиторий
git clone https://github.com/YOUR_USERNAME/expense-tracker.git
cd expense-tracker

# Создайте .env файл
cp env.example .env
nano .env
```

### 2. Настройка .env

```env
# Обязательные переменные
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=your_strong_password
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

TELEGRAM_BOT_TOKEN=YOUR_BOT_TOKEN_FROM_BOTFATHER
TELEGRAM_WHITELIST=YOUR_TELEGRAM_ID

BOT_API_KEY=random_secure_key_here
JWT_SECRET=another_random_secure_key_here

API_URL=http://api:8080
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

1. Найдите вашего бота в Telegram
2. Отправьте `/help`
3. Отправьте `100` - должен записать расход
4. Отправьте `/total` - должен показать сумму

## 🔧 Полезные команды

```bash
# Остановка
docker-compose down

# Перезапуск
docker-compose restart

# Обновление
git pull && docker-compose up --build -d

# Логи бота
docker-compose logs -f bot

# Логи API
docker-compose logs -f api
```

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

## 📱 Использование бота

- `100` - записать расход 100 руб.
- `100 продукты` - записать расход с категорией
- `/total` - показать общую сумму
- `/total week` - расходы за неделю
- `/total month` - расходы за месяц
- `/debts` - показать долги
- `/help` - справка

## 🌐 Для продакшена

См. [DEPLOY.md](DEPLOY.md) для развертывания на сервере с доменом и SSL.
