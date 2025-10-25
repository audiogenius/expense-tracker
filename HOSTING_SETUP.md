# 🚀 Полная настройка хостинга Expense Tracker

## 📋 Содержание
1. [Первоначальная настройка сервера](#первоначальная-настройка-сервера)
2. [Установка проекта с GitHub](#установка-проекта-с-github)
3. [Настройка домена и SSL](#настройка-домена-и-ssl)
4. [Полная очистка сервера](#полная-очистка-сервера)
5. [Обновление проекта](#обновление-проекта)
6. [Мониторинг и управление](#мониторинг-и-управление)

---

## 🛠 Первоначальная настройка сервера

### Подключение к серверу

```bash
# Подключение по SSH
ssh root@YOUR_SERVER_IP

# Или для Timeweb Cloud MSK 30:
ssh root@147.45.246.210
```

### Обновление системы

```bash
# Обновление пакетов
apt update && apt upgrade -y

# Установка необходимых пакетов
apt install -y curl wget git unzip htop nano ufw fail2ban nginx certbot python3-certbot-nginx
```

### Установка Docker

```bash
# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Установка Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Проверка установки
docker --version
docker-compose --version
```

### Настройка безопасности

```bash
# Настройка firewall
ufw allow ssh
ufw allow 80
ufw allow 443
ufw --force enable

# Настройка fail2ban
systemctl enable fail2ban
systemctl start fail2ban

# Создание пользователя для приложения
useradd -m -s /bin/bash app
usermod -aG docker app
```

---

## 📥 Установка проекта с GitHub

### Клонирование репозитория

```bash
# Создание директории проекта
mkdir -p /root/expense-tracker
cd /root/expense-tracker

# Клонирование с GitHub
git clone https://github.com/audiogenius/expense-tracker.git .

# Переключение на актуальную ветку
git checkout feature/performance-optimization-2gb-ram
```

### Настройка конфигурации

```bash
# Копирование оптимизированного docker-compose для 2GB RAM
cp deploy/docker-compose.timeweb.yml docker-compose.yml

# Создание .env файла
cp env.example .env
nano .env
```

### Заполнение .env файла

```env
# Database Configuration
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=YourStrongPassword123!
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_WHITELIST=your_telegram_id_here
TELEGRAM_CHAT_IDS=123456789,987654321

# API Configuration
BOT_API_KEY=your_secure_bot_api_key_here
JWT_SECRET=your_jwt_secret_key_here

# API URLs
API_URL=http://api:8080

# Analytics Service Configuration
ANALYTICS_PORT=8081
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=qwen2.5:0.5b

# Ollama Configuration (оптимизировано для 2GB RAM)
OLLAMA_NUM_PARALLEL=1
OLLAMA_MAX_LOADED_MODELS=1
OLLAMA_HOST=0.0.0.0:11434

# Google Cloud Vision (optional for OCR)
USE_LOCAL_OCR=false
```

### Запуск проекта

```bash
# Установка прав
chown -R app:app /root/expense-tracker

# Запуск контейнеров
sudo -u app docker-compose up -d

# Ожидание запуска
sleep 60

# Инициализация Ollama
sudo -u app ./deploy/init-ollama-timeweb.sh
```

### Проверка работы

```bash
# Проверка статуса контейнеров
docker-compose ps

# Проверка здоровья сервисов
curl http://localhost:8080/api/health
curl http://localhost:8081/health
curl http://localhost:11434/api/tags

# Просмотр логов
docker-compose logs -f
```

---

## 🌐 Настройка домена и SSL

### Настройка Nginx

```bash
# Создание конфигурации сайта
cat > /etc/nginx/sites-available/expense-tracker << 'EOF'
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com www.your-domain.com;
    
    # SSL configuration (будет настроен certbot)
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Proxy to frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Proxy to API
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Proxy to Analytics
    location /analytics/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Активация сайта
ln -sf /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Проверка конфигурации
nginx -t
systemctl reload nginx
```

### Получение SSL сертификата

```bash
# Получение SSL сертификата
certbot --nginx -d your-domain.com -d www.your-domain.com

# Автообновление SSL
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### Настройка Telegram Bot Domain

1. Найдите бота **@BotFather** в Telegram
2. Отправьте команду: `/setdomain`
3. Выберите вашего бота
4. Введите домен: `your-domain.com`

---

## 🗑 Полная очистка сервера

### Остановка всех сервисов

```bash
# Остановка проекта
cd /opt/expense-tracker
docker-compose down -v

# Остановка Nginx
systemctl stop nginx
```

### Удаление Docker контейнеров и образов

```bash
# Удаление всех контейнеров
docker rm -f $(docker ps -aq)

# Удаление всех образов
docker rmi -f $(docker images -q)

# Удаление всех томов
docker volume rm $(docker volume ls -q)

# Очистка системы Docker
docker system prune -a --volumes
```

### Удаление проекта

```bash
# Удаление директории проекта
rm -rf /root/expense-tracker

# Удаление пользователя
userdel -r app
```

### Удаление Nginx конфигурации

```bash
# Удаление конфигурации сайта
rm -f /etc/nginx/sites-enabled/expense-tracker
rm -f /etc/nginx/sites-available/expense-tracker

# Перезапуск Nginx
systemctl restart nginx
```

### Удаление SSL сертификатов

```bash
# Удаление сертификатов
rm -rf /etc/letsencrypt/live/your-domain.com
rm -rf /etc/letsencrypt/archive/your-domain.com
rm -rf /etc/letsencrypt/renewal/your-domain.com.conf

# Очистка cron задач
crontab -r
```

### Полная очистка системы (ОПЦИОНАЛЬНО)

```bash
# Удаление Docker
apt remove -y docker.io docker-compose
rm -rf /var/lib/docker

# Удаление Nginx
apt remove -y nginx

# Удаление Certbot
apt remove -y certbot python3-certbot-nginx

# Очистка пакетов
apt autoremove -y
apt autoclean
```

---

## 🔄 Обновление проекта

### Автоматическое обновление

```bash
# Переход в директорию проекта
cd /root/expense-tracker

# Остановка сервисов
docker-compose down

# Обновление кода
git pull origin feature/performance-optimization-2gb-ram

# Пересборка и запуск
sudo -u app docker-compose up --build -d

# Проверка статуса
docker-compose ps
```

### Обновление до новой версии

```bash
# Проверка доступных веток
git branch -r

# Переключение на новую ветку
git checkout new-feature-branch

# Обновление
git pull origin new-feature-branch

# Пересборка
sudo -u app docker-compose up --build -d
```

---

## 📊 Мониторинг и управление

### Systemd сервис для автозапуска

```bash
# Создание systemd сервиса
cat > /etc/systemd/system/expense-tracker.service << 'EOF'
[Unit]
Description=Expense Tracker Application
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/root/expense-tracker
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
User=app
Group=app

[Install]
WantedBy=multi-user.target
EOF

# Активация сервиса
systemctl daemon-reload
systemctl enable expense-tracker
systemctl start expense-tracker
```

### Мониторинг ресурсов

```bash
# Мониторинг памяти
free -h

# Мониторинг диска
df -h

# Мониторинг Docker
docker stats --no-stream

# Мониторинг процессов
htop
```

### Полезные команды

```bash
# Перезапуск проекта
systemctl restart expense-tracker

# Просмотр логов
docker-compose logs -f

# Проверка здоровья
curl http://localhost:8080/api/health
curl http://localhost:8081/health

# Подключение к базе данных
docker-compose exec db psql -U expense_user -d expense_tracker
```

---

## 🆘 Решение проблем

### Сайт не открывается

```bash
# Проверка контейнеров
docker-compose ps

# Проверка Nginx
systemctl status nginx

# Проверка портов
netstat -tlnp | grep :80
netstat -tlnp | grep :443
```

### Высокое потребление памяти

```bash
# Проверка использования памяти
free -h
docker stats

# Перезапуск тяжелых сервисов
docker-compose restart ollama analytics
```

### SSL проблемы

```bash
# Проверка сертификатов
certbot certificates

# Обновление сертификатов
certbot renew --dry-run
```

---

## 📞 Поддержка

- **GitHub Issues**: https://github.com/audiogenius/expense-tracker/issues
- **Документация обновлений**: `docs/updates/`
- **Telegram Support**: @timeweb_support_bot

---

**Готово! Ваш Expense Tracker развернут и готов к использованию! 🚀**
