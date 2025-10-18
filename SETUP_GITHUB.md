# 🚀 Развертывание Expense Tracker с GitHub

## ✅ Проект успешно загружен в GitHub!

**Репозиторий**: https://github.com/AbleevDinis/rd_expense_tracker

## 🛠 Быстрое развертывание на сервере

### 1. Подключение к серверу

```bash
# Подключение по SSH
ssh root@YOUR_SERVER_IP
```

### 2. Установка Docker (если не установлен)

```bash
# Обновление системы
apt update && apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Установка Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 3. Клонирование и настройка

```bash
# Клонирование репозитория
git clone https://github.com/AbleevDinis/rd_expense_tracker.git
cd rd_expense_tracker

# Создание .env файла
cp env.example .env
nano .env
```

### 4. Настройка .env

```env
# Обязательные переменные
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=your_strong_password_here
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

TELEGRAM_BOT_TOKEN=YOUR_BOT_TOKEN_FROM_BOTFATHER
TELEGRAM_WHITELIST=YOUR_TELEGRAM_ID,SPOUSE_TELEGRAM_ID

BOT_API_KEY=random_secure_key_here_32_chars
JWT_SECRET=another_random_secure_key_here_32_chars

API_URL=http://api:8080
```

### 5. Создание Telegram бота

1. Перейдите к [@BotFather](https://t.me/BotFather)
2. Отправьте `/newbot`
3. Введите имя: `Expense Tracker`
4. Введите username: `your_expense_tracker_bot`
5. Скопируйте токен в `TELEGRAM_BOT_TOKEN`

### 6. Получение Telegram ID

1. Найдите [@userinfobot](https://t.me/userinfobot)
2. Отправьте `/start`
3. Скопируйте ID в `TELEGRAM_WHITELIST`

### 7. Запуск проекта

```bash
# Запуск всех сервисов
docker-compose up --build -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 8. Тестирование

1. Найдите вашего бота в Telegram
2. Отправьте `/help` - должна появиться справка
3. Отправьте `100` - должен записать расход
4. Отправьте `/total` - должен показать сумму

## 🔧 Полезные команды

```bash
# Остановка
docker-compose down

# Перезапуск
docker-compose restart

# Обновление проекта
git pull && docker-compose up --build -d

# Логи бота
docker-compose logs -f bot

# Логи API
docker-compose logs -f api

# Подключение к базе данных
docker-compose exec db psql -U expense_user -d expense_tracker
```

## 🌐 Настройка домена (опционально)

### 1. Покупка домена

Рекомендуемые регистраторы:
- **REG.RU** - от 200₽/год
- **Timeweb** - от 150₽/год
- **Beget** - от 200₽/год

### 2. Настройка DNS

```
A-запись: your-domain.com → YOUR_SERVER_IP
CNAME: www.your-domain.com → your-domain.com
```

### 3. Установка SSL

```bash
# Установка Certbot
apt install certbot nginx -y

# Получение сертификата
certbot certonly --standalone -d your-domain.com

# Настройка Nginx
nano /etc/nginx/sites-available/expense-tracker
```

**Конфигурация Nginx:**
```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Активация конфигурации
ln -s /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## 📊 Мониторинг

```bash
# Просмотр ресурсов
htop

# Автозапуск при перезагрузке
systemctl enable docker
```

## 💰 Стоимость

**Минимальная конфигурация:**
- VPS: 200₽/месяц
- Домен: 200₽/год
- SSL: бесплатно
- **Итого**: ~220₽/месяц

## 🐛 Решение проблем

**Бот не отвечает:**
- Проверьте `TELEGRAM_BOT_TOKEN`
- Проверьте `TELEGRAM_WHITELIST`
- Проверьте логи: `docker-compose logs -f bot`

**Ошибка 500:**
- Проверьте `BOT_API_KEY`
- Проверьте логи API: `docker-compose logs -f api`

**База данных:**
- Проверьте `POSTGRES_PASSWORD`
- Проверьте логи БД: `docker-compose logs -f db`

## 📱 Использование бота

- `100` - записать расход 100 руб.
- `100 продукты` - записать расход с категорией
- `/total` - показать общую сумму
- `/total week` - расходы за неделю
- `/total month` - расходы за месяц
- `/debts` - показать долги
- `/help` - справка

**Проект готов к использованию!** 🎉
