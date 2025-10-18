# 🚀 Руководство по развертыванию Expense Tracker

## 📋 Содержание
1. [Подготовка к деплою](#подготовка-к-деплою)
2. [Российские хостинги](#российские-хостинги)
3. [Развертывание на VPS](#развертывание-на-vps)
4. [Настройка домена и SSL](#настройка-домена-и-ssl)
5. [Мониторинг и бэкапы](#мониторинг-и-бэкапы)

## 🛠 Подготовка к деплою

### 1. Создание GitHub репозитория

```bash
# Инициализация Git
git init
git add .
git commit -m "Initial commit: Expense Tracker v1.1"

# Создание репозитория на GitHub
# Перейдите на https://github.com/new
# Создайте репозиторий с именем "expense-tracker"

# Подключение к GitHub
git remote add origin https://github.com/YOUR_USERNAME/expense-tracker.git
git branch -M main
git push -u origin main
```

### 2. Подготовка переменных окружения

Создайте файл `.env` на сервере:

```bash
# Скопируйте env.example в .env
cp env.example .env

# Отредактируйте .env
nano .env
```

**Обязательные переменные:**
```env
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=STRONG_PASSWORD_HERE
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

TELEGRAM_BOT_TOKEN=YOUR_BOT_TOKEN_FROM_BOTFATHER
TELEGRAM_WHITELIST=YOUR_TELEGRAM_ID,SPOUSE_TELEGRAM_ID

BOT_API_KEY=RANDOM_SECURE_KEY_HERE
JWT_SECRET=ANOTHER_RANDOM_SECURE_KEY_HERE

API_URL=http://api:8080
```

## 🇷🇺 Российские хостинги

### Рекомендуемые варианты:

#### 1. **Timeweb** (Рекомендуется)
- **Цена**: от 200₽/месяц
- **VPS**: Ubuntu 20.04/22.04
- **RAM**: 1GB (достаточно для MVP)
- **SSD**: 20GB
- **Сеть**: 100 Мбит/с
- **Сайт**: https://timeweb.com

#### 2. **Beget**
- **Цена**: от 150₽/месяц
- **VPS**: Ubuntu 20.04
- **RAM**: 512MB-1GB
- **SSD**: 10GB
- **Сайт**: https://beget.com

#### 3. **REG.RU**
- **Цена**: от 300₽/месяц
- **VPS**: Ubuntu 20.04
- **RAM**: 1GB
- **SSD**: 20GB
- **Сайт**: https://reg.ru

#### 4. **FirstVDS**
- **Цена**: от 200₽/месяц
- **VPS**: Ubuntu 20.04
- **RAM**: 1GB
- **SSD**: 20GB
- **Сайт**: https://firstvds.ru

## 🖥 Развертывание на VPS

### 1. Подключение к серверу

```bash
# Подключение по SSH
ssh root@YOUR_SERVER_IP

# Обновление системы
apt update && apt upgrade -y
```

### 2. Установка Docker

```bash
# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Установка Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Проверка установки
docker --version
docker-compose --version
```

### 3. Клонирование проекта

```bash
# Установка Git
apt install git -y

# Клонирование репозитория
git clone https://github.com/YOUR_USERNAME/expense-tracker.git
cd expense-tracker

# Создание .env файла
cp env.example .env
nano .env
```

### 4. Настройка Telegram бота

1. Перейдите к [@BotFather](https://t.me/BotFather)
2. Отправьте `/newbot`
3. Введите имя бота: `Expense Tracker`
4. Введите username: `your_expense_tracker_bot`
5. Скопируйте токен в `.env`

### 5. Получение Telegram ID

1. Найдите [@userinfobot](https://t.me/userinfobot)
2. Отправьте `/start`
3. Скопируйте ваш ID в `TELEGRAM_WHITELIST`

### 6. Запуск проекта

```bash
# Сборка и запуск
docker-compose up --build -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

## 🌐 Настройка домена и SSL

### 1. Покупка домена

**Рекомендуемые регистраторы:**
- **REG.RU** - от 200₽/год
- **Timeweb** - от 150₽/год
- **Beget** - от 200₽/год

### 2. Настройка DNS

```bash
# A-запись
your-domain.com → YOUR_SERVER_IP

# CNAME-запись (опционально)
www.your-domain.com → your-domain.com
```

### 3. Установка SSL сертификата

```bash
# Установка Certbot
apt install certbot -y

# Получение сертификата
certbot certonly --standalone -d your-domain.com

# Автообновление
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### 4. Настройка Nginx

```bash
# Установка Nginx
apt install nginx -y

# Создание конфигурации
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

## 📊 Мониторинг и бэкапы

### 1. Мониторинг

```bash
# Установка htop для мониторинга
apt install htop -y

# Просмотр ресурсов
htop

# Просмотр логов
docker-compose logs -f
```

### 2. Автоматические бэкапы

```bash
# Создание скрипта бэкапа
nano /root/backup.sh
```

**Скрипт бэкапа:**
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/root/backups"
PROJECT_DIR="/root/expense-tracker"

mkdir -p $BACKUP_DIR

# Бэкап базы данных
docker-compose exec -T db pg_dump -U expense_user expense_tracker > $BACKUP_DIR/db_$DATE.sql

# Бэкап файлов проекта
tar -czf $BACKUP_DIR/project_$DATE.tar.gz -C $PROJECT_DIR .

# Удаление старых бэкапов (старше 7 дней)
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

```bash
# Делаем скрипт исполняемым
chmod +x /root/backup.sh

# Добавляем в cron (ежедневно в 2:00)
echo "0 2 * * * /root/backup.sh" | crontab -
```

### 3. Автозапуск при перезагрузке

```bash
# Создание systemd сервиса
nano /etc/systemd/system/expense-tracker.service
```

**Содержимое сервиса:**
```ini
[Unit]
Description=Expense Tracker
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/root/expense-tracker
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
# Активация сервиса
systemctl enable expense-tracker.service
systemctl start expense-tracker.service
```

## 🔧 Полезные команды

```bash
# Перезапуск проекта
docker-compose restart

# Обновление проекта
git pull
docker-compose up --build -d

# Просмотр логов
docker-compose logs -f bot
docker-compose logs -f api

# Подключение к базе данных
docker-compose exec db psql -U expense_user -d expense_tracker

# Очистка неиспользуемых образов
docker system prune -a
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs -f`
2. Проверьте статус контейнеров: `docker-compose ps`
3. Проверьте переменные окружения: `cat .env`
4. Проверьте подключение к базе данных
5. Проверьте настройки Telegram бота

## 💰 Примерная стоимость

**Минимальная конфигурация:**
- VPS: 200₽/месяц
- Домен: 200₽/год
- SSL: бесплатно (Let's Encrypt)
- **Итого**: ~220₽/месяц

**Рекомендуемая конфигурация:**
- VPS: 400₽/месяц (2GB RAM)
- Домен: 200₽/год
- SSL: бесплатно
- **Итого**: ~420₽/месяц
