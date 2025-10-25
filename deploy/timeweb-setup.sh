#!/bin/bash

# Скрипт для настройки expense tracker на Timeweb Cloud MSK 30
# 1 x 3.3 ГГц CPU • 2 ГБ RAM • 30 ГБ NVMe • Ubuntu 22.04

echo "🚀 Настройка Expense Tracker на Timeweb Cloud MSK 30..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Проверка прав root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}❌ Запустите скрипт с правами root: sudo ./timeweb-setup.sh${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Система: Timeweb Cloud MSK 30${NC}"
echo -e "${BLUE}💾 RAM: 2GB${NC}"
echo -e "${BLUE}💽 CPU: 1 x 3.3 ГГц${NC}"
echo -e "${BLUE}💿 Диск: 30 ГБ NVMe${NC}"
echo ""

# Обновление системы
echo -e "${YELLOW}⏳ Обновление системы...${NC}"
apt update && apt upgrade -y

# Установка необходимых пакетов
echo -e "${YELLOW}⏳ Установка необходимых пакетов...${NC}"
apt install -y \
    curl \
    wget \
    git \
    unzip \
    htop \
    nano \
    ufw \
    fail2ban \
    nginx \
    certbot \
    python3-certbot-nginx

# Установка Docker
echo -e "${YELLOW}⏳ Установка Docker...${NC}"
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Установка Docker Compose
echo -e "${YELLOW}⏳ Установка Docker Compose...${NC}"
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Настройка firewall
echo -e "${YELLOW}⏳ Настройка firewall...${NC}"
ufw allow ssh
ufw allow 80
ufw allow 443
ufw --force enable

# Настройка fail2ban
echo -e "${YELLOW}⏳ Настройка fail2ban...${NC}"
systemctl enable fail2ban
systemctl start fail2ban

# Создание пользователя для приложения
echo -e "${YELLOW}⏳ Создание пользователя app...${NC}"
useradd -m -s /bin/bash app
usermod -aG docker app

# Создание директории проекта
echo -e "${YELLOW}⏳ Создание директории проекта...${NC}"
mkdir -p /root/expense-tracker
chown app:app /root/expense-tracker

# Клонирование репозитория
echo -e "${YELLOW}⏳ Клонирование репозитория...${NC}"
cd /root/expense-tracker
sudo -u app git clone https://github.com/audiogenius/expense-tracker.git .

# Переключение на ветку с оптимизациями
sudo -u app git checkout feature/performance-optimization-2gb-ram

# Создание .env файла
echo -e "${YELLOW}⏳ Создание конфигурации...${NC}"
cat > .env << 'EOF'
# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=expense_tracker_2024_secure
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

# API Configuration
API_PORT=8080
JWT_SECRET=your_jwt_secret_key_here_change_this_in_production
JWT_EXPIRES_IN=24h

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_CHAT_IDS=your_telegram_chat_id_here

# OCR Configuration
USE_LOCAL_OCR=false

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
EOF

# Настройка Nginx
echo -e "${YELLOW}⏳ Настройка Nginx...${NC}"
cat > /etc/nginx/sites-available/expense-tracker << 'EOF'
server {
    listen 80;
    server_name your-domain.com;  # Замените на ваш домен
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;  # Замените на ваш домен
    
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

# Проверка конфигурации Nginx
nginx -t

# Создание systemd service для автозапуска
echo -e "${YELLOW}⏳ Создание systemd service...${NC}"
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

# Перезагрузка systemd
systemctl daemon-reload
systemctl enable expense-tracker

# Настройка мониторинга памяти
echo -e "${YELLOW}⏳ Настройка мониторинга памяти...${NC}"
cat > /root/monitor-memory.sh << 'EOF'
#!/bin/bash
# Мониторинг памяти для 2GB сервера

MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
MEMORY_LIMIT=85  # 85% лимит для 2GB RAM

if (( $(echo "$MEMORY_USAGE > $MEMORY_LIMIT" | bc -l) )); then
    echo "⚠️ High memory usage: ${MEMORY_USAGE}%"
    
    # Перезапуск контейнеров при высокой нагрузке
    cd /root/expense-tracker
    docker-compose restart ollama analytics
fi
EOF

chmod +x /root/monitor-memory.sh

# Добавление в crontab
echo "*/5 * * * * /root/monitor-memory.sh" | crontab -u root -

# Создание скрипта инициализации Ollama
echo -e "${YELLOW}⏳ Создание скрипта инициализации Ollama...${NC}"
cat > /root/expense-tracker/init-ollama-timeweb.sh << 'EOF'
#!/bin/bash

echo "🚀 Инициализация Ollama на Timeweb сервере..."

# Ожидание запуска Ollama
echo "⏳ Ожидание запуска Ollama..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "✅ Ollama готов!"
        break
    fi
    
    echo "⏳ Ollama не готов, ожидание 10 секунд... (попытка $((attempt + 1))/$max_attempts)"
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Ollama не запустился за 5 минут"
    exit 1
fi

# Загрузка модели qwen2.5:0.5b
echo "📥 Загрузка модели qwen2.5:0.5b..."
docker exec expense_ollama ollama pull qwen2.5:0.5b

# Тест модели
echo "🧪 Тестирование модели..."
response=$(curl -s -X POST http://localhost:11434/api/generate -d '{
    "model": "qwen2.5:0.5b",
    "prompt": "Привет! Как дела?",
    "stream": false
}')

if echo "$response" | grep -q "response"; then
    echo "✅ Модель работает!"
    echo "📝 Ответ: $(echo "$response" | jq -r '.response // "Нет ответа"')"
else
    echo "❌ Тест модели не прошел"
    exit 1
fi

echo "🎉 Ollama инициализирован успешно!"
EOF

chmod +x /root/expense-tracker/init-ollama-timeweb.sh

# Создание скрипта развертывания
echo -e "${YELLOW}⏳ Создание скрипта развертывания...${NC}"
cat > /root/expense-tracker/deploy-timeweb.sh << 'EOF'
#!/bin/bash

echo "🚀 Развертывание Expense Tracker на Timeweb..."

# Остановка сервисов
echo "⏳ Остановка сервисов..."
systemctl stop expense-tracker

# Обновление кода
echo "⏳ Обновление кода..."
cd /root/expense-tracker
sudo -u app git pull origin feature/performance-optimization-2gb-ram

# Сборка и запуск
echo "⏳ Сборка и запуск контейнеров..."
sudo -u app docker-compose down
sudo -u app docker-compose build --no-cache
sudo -u app docker-compose up -d

# Ожидание запуска
echo "⏳ Ожидание запуска сервисов..."
sleep 30

# Инициализация Ollama
echo "⏳ Инициализация Ollama..."
./init-ollama-timeweb.sh

# Проверка здоровья
echo "⏳ Проверка здоровья сервисов..."
curl -f http://localhost:8080/api/health || echo "❌ API не отвечает"
curl -f http://localhost:8081/health || echo "❌ Analytics не отвечает"
curl -f http://localhost:11434/api/tags || echo "❌ Ollama не отвечает"

# Запуск сервиса
echo "⏳ Запуск сервиса..."
systemctl start expense-tracker

echo "✅ Развертывание завершено!"
echo "🌐 Приложение доступно по адресу: https://your-domain.com"
EOF

chmod +x /root/expense-tracker/deploy-timeweb.sh

# Создание README для Timeweb
cat > /root/expense-tracker/TIMEWEB_README.md << 'EOF'
# Expense Tracker на Timeweb Cloud MSK 30

## 🚀 Быстрый старт

1. **Настройка домена:**
   ```bash
   # Замените your-domain.com на ваш домен в файлах:
   nano /etc/nginx/sites-available/expense-tracker
   nano /root/expense-tracker/.env
   ```

2. **Получение SSL сертификата:**
   ```bash
   certbot --nginx -d your-domain.com
   ```

3. **Запуск приложения:**
   ```bash
   cd /root/expense-tracker
   ./deploy-timeweb.sh
   ```

## 📊 Мониторинг

- **Логи:** `docker-compose logs -f`
- **Память:** `htop` или `free -h`
- **Контейнеры:** `docker ps`
- **Сервис:** `systemctl status expense-tracker`

## 🔧 Управление

- **Перезапуск:** `systemctl restart expense-tracker`
- **Остановка:** `systemctl stop expense-tracker`
- **Обновление:** `./deploy-timeweb.sh`

## ⚡ Оптимизации для 2GB RAM

- Ollama: 1.5GB лимит
- API: кэширование 5 минут
- Frontend: виртуализация
- Database: оптимизированные индексы
- Мониторинг памяти каждые 5 минут
EOF

echo ""
echo -e "${GREEN}✅ Настройка завершена!${NC}"
echo ""
echo -e "${BLUE}📋 Следующие шаги:${NC}"
echo "1. Настройте домен в /etc/nginx/sites-available/expense-tracker"
echo "2. Обновите .env файл с вашими токенами"
echo "3. Получите SSL сертификат: certbot --nginx -d your-domain.com"
echo "4. Запустите приложение: cd /root/expense-tracker && ./deploy-timeweb.sh"
echo ""
echo -e "${BLUE}📁 Файлы созданы:${NC}"
echo "- /root/expense-tracker/ - директория проекта"
echo "- /root/expense-tracker/.env - конфигурация"
echo "- /root/expense-tracker/deploy-timeweb.sh - скрипт развертывания"
echo "- /root/expense-tracker/init-ollama-timeweb.sh - инициализация Ollama"
echo "- /etc/systemd/system/expense-tracker.service - автозапуск"
echo ""
echo -e "${GREEN}🎉 Готово к развертыванию!${NC}"
