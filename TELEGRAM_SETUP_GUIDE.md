# 🤖 Подробная инструкция по настройке Telegram Bot

## Шаг 1: Создание Telegram бота

### Для новичков - что такое BotFather?
BotFather - это официальный бот Telegram, который помогает создавать и управлять ботами.

### Инструкция:

1. **Откройте Telegram** на телефоне или в браузере https://web.telegram.org/

2. **Найдите BotFather:**
   - В поиске введите: `@BotFather`
   - Откройте чат с синей галочкой (официальный бот)

3. **Создайте нового бота:**
   - Отправьте команду: `/newbot`
   - BotFather спросит имя бота. Введите: `Expense Tracker` (можно любое)
   - BotFather попросит username (должен заканчиваться на `bot`). Введите, например: `your_expense_tracker_bot`
   
4. **Сохраните токен:**
   - BotFather пришлет сообщение с токеном вида: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`
   - **ВАЖНО:** Скопируйте этот токен! Он вам понадобится для `.env` файла
   - Этот токен - секретный ключ. Не публикуйте его нигде!

### Пример ответа BotFather:
```
Done! Congratulations on your new bot. You will find it at t.me/your_expense_tracker_bot. 
You can now add a description, about section and profile picture for your bot, see /help for a list of commands.

Use this token to access the HTTP API:
123456789:ABCdefGHIjklMNOpqrsTUVwxyz

For a description of the Bot API, see this page: https://core.telegram.org/bots/api
```

---

## Шаг 2: Получение вашего Telegram ID

### Зачем нужен Telegram ID?
Ваш Telegram ID - это уникальный числовой идентификатор вашего аккаунта. 
Он нужен для whitelist (чтобы только вы могли пользоваться ботом).

### Инструкция:

1. **Найдите бота для получения ID:**
   - В поиске Telegram введите: `@userinfobot` или `@get_id_bot`
   - Откройте чат

2. **Получите свой ID:**
   - Отправьте команду: `/start`
   - Бот пришлет ваш ID, например: `Your user ID: 123456789`
   - Скопируйте это число

3. **Для семьи - получите ID супруги/супруга:**
   - Попросите их также написать `/start` в @userinfobot
   - Скопируйте их ID

---

## Шаг 3: Настройка Telegram Login Widget (для веб-интерфейса)

### Что это?
Telegram Login Widget позволяет входить в веб-приложение через Telegram (OAuth).

### Инструкция:

1. **Вернитесь к @BotFather**

2. **Настройте домен (для production):**
   - Отправьте команду: `/setdomain`
   - Выберите вашего бота из списка
   - Введите домен, например: `expense-tracker.yourdomain.com`
   
   **Для локальной разработки:**
   - Можно пропустить или указать `localhost`

3. **Готово!** Widget уже работает в коде фронтенда

---

## Шаг 4: Настройка .env файла

### Откройте файл `.env` в корне проекта

Если его нет, скопируйте из примера:
```bash
cp env.example .env
```

### Заполните необходимые поля:

```env
# Database Configuration (можно оставить как есть для разработки)
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=your_strong_password_here   # Придумайте пароль
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz   # Токен из Шага 1
TELEGRAM_WHITELIST=123456789,987654321                    # Ваши ID из Шага 2 (через запятую)

# API Configuration
BOT_API_KEY=random_secure_key_here_32_characters   # Придумайте случайную строку
JWT_SECRET=another_random_key_here_32_chars       # Придумайте другую случайную строку

# API URLs (для Docker - оставить как есть)
API_URL=http://api:8080
```

### Как сгенерировать случайные ключи?

**Windows PowerShell:**
```powershell
# Для BOT_API_KEY
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})

# Для JWT_SECRET
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

**Linux/Mac:**
```bash
# Для BOT_API_KEY
openssl rand -hex 16

# Для JWT_SECRET
openssl rand -hex 16
```

**Онлайн генератор:**
https://www.random.org/strings/ (установите Length: 32, Number: 2)

---

## Шаг 5: Запуск проекта

### Убедитесь что Docker запущен:
```bash
docker --version
docker-compose --version
```

### Запустите проект:
```bash
# Сборка и запуск всех сервисов
docker-compose up --build -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### Проверка что бот запустился:
```bash
docker-compose logs -f bot
```

Вы должны увидеть:
```
Bot service starting...
API URL: http://api:8080
Bot Token: 123456789:...xyz
✅ Bot is running! Waiting for messages...
```

---

## Шаг 6: Тестирование бота

### 1. Найдите вашего бота в Telegram:
- Перейдите по ссылке: `t.me/your_expense_tracker_bot` (замените на username вашего бота)
- Или найдите в поиске по username

### 2. Начните диалог:
- Нажмите **START** или отправьте `/start`

### 3. Попробуйте команды:

```
/help           - Справка по командам
100             - Записать расход 100 руб.
100 продукты    - Расход 100 руб. с категорией
/total          - Показать общую сумму
/total week     - Расходы за неделю
/total month    - Расходы за месяц
```

### Пример диалога:
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
Бот: ✅ Записал расход: 50.50 руб. (категория: 3)

Вы: /total
Бот: 📊 Расходы всего: 150.50 руб.
```

---

## Шаг 7: Тестирование веб-интерфейса

### 1. Откройте браузер:
- Перейдите на: http://localhost (или http://localhost:80)

### 2. Войдите через Telegram:
- Нажмите на кнопку "Log in with Telegram"
- Откроется popup Telegram - подтвердите вход
- Вы будете перенаправлены обратно на сайт и увидите свои расходы

### Если не работает Telegram Login:
- Для локальной разработки используйте кнопку **"Simulate Login (dev)"**
- Это создаст тестовую сессию без реального входа через Telegram

---

## 🔧 Решение проблем

### Бот не отвечает:

1. **Проверьте логи:**
```bash
docker-compose logs bot
```

2. **Проверьте что токен правильный:**
- Откройте `.env`
- Убедитесь что `TELEGRAM_BOT_TOKEN` скопирован полностью
- Не должно быть лишних пробелов

3. **Проверьте whitelist:**
- Убедитесь что ваш Telegram ID добавлен в `TELEGRAM_WHITELIST`
- Формат: `123456789,987654321` (без пробелов)

### Бот отвечает "forbidden":

- Ваш Telegram ID не в whitelist
- Проверьте `.env` файл
- Перезапустите: `docker-compose restart bot`

### Веб-интерфейс не загружается:

1. **Проверьте что контейнеры запущены:**
```bash
docker-compose ps
```

2. **Проверьте логи frontend:**
```bash
docker-compose logs frontend
docker-compose logs proxy
```

3. **Проверьте API:**
```bash
curl http://localhost:8080/health
```

Должен вернуть: `{"status":"ok","service":"api"}`

### База данных не подключается:

```bash
# Проверьте логи БД
docker-compose logs db

# Проверьте что БД запустилась
docker-compose exec db psql -U expense_user -d expense_tracker -c "SELECT 1"
```

---

## 📱 Полезные команды

```bash
# Остановить все сервисы
docker-compose down

# Перезапустить конкретный сервис
docker-compose restart bot
docker-compose restart api

# Пересобрать и запустить
docker-compose up --build -d

# Посмотреть логи конкретного сервиса
docker-compose logs -f bot
docker-compose logs -f api
docker-compose logs -f db

# Подключиться к базе данных
docker-compose exec db psql -U expense_user -d expense_tracker

# Очистить все (ОСТОРОЖНО - удалит данные!)
docker-compose down -v
```

---

## 🎉 Готово!

Теперь у вас работает:
- ✅ Telegram бот для записи расходов
- ✅ Веб-интерфейс с графиками
- ✅ База данных для хранения данных
- ✅ API для обмена данными

**Следующие шаги:**
- Добавьте супругу/друзей в whitelist
- Начните записывать расходы
- Посмотрите статистику в веб-интерфейсе
- Настройте deployment на сервер (см. DEPLOY.md)

---

## 💡 Советы по использованию

1. **Быстрая запись:**
   - Просто отправьте число боту: `100`
   - Или с категорией: `100 продукты`

2. **Точность:**
   - Можно использовать копейки: `50.50` или `50,50`

3. **Категории:**
   - Бот автоматически определяет категорию по ключевым словам
   - Доступные категории: продукты, транспорт, кафе, развлечения, здоровье, одежда, коммуналка

4. **Просмотр статистики:**
   - В боте: `/total`, `/total week`, `/total month`
   - В веб-интерфейсе: графики и фильтры

---

Если возникли вопросы - создайте issue в GitHub! 🚀

