#!/bin/bash

# Скрипт быстрого развертывания Expense Tracker на Timeweb Cloud MSK 30
# Использование: ./deploy-to-timeweb.sh your-domain.com

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Проверка аргументов
if [ $# -eq 0 ]; then
    echo -e "${RED}❌ Укажите домен: ./deploy-to-timeweb.sh your-domain.com${NC}"
    exit 1
fi

DOMAIN=$1
echo -e "${BLUE}🚀 Развертывание Expense Tracker на домен: $DOMAIN${NC}"

# Проверка прав root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}❌ Запустите скрипт с правами root: sudo ./deploy-to-timeweb.sh $DOMAIN${NC}"
    exit 1
fi

# Создание директории проекта
echo -e "${YELLOW}⏳ Создание директории проекта...${NC}"
mkdir -p /root/expense-tracker
cd /root/expense-tracker

# Клонирование репозитория
echo -e "${YELLOW}⏳ Клонирование репозитория...${NC}"
if [ ! -d ".git" ]; then
    git clone https://github.com/audiogenius/expense-tracker.git .
fi

# Переключение на ветку с оптимизациями
git checkout feature/performance-optimization-2gb-ram
git pull origin feature/performance-optimization-2gb-ram

# Копирование оптимизированного docker-compose
echo -e "${YELLOW}⏳ Настройка Docker Compose для Timeweb...${NC}"
cp deploy/docker-compose.timeweb.yml docker-compose.yml

# Создание .env файла
echo -e "${YELLOW}⏳ Создание конфигурации...${NC}"
cat > .env << EOF
# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=expense_tracker_$(date +%s)_secure
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

# API Configuration
API_PORT=8080
JWT_SECRET=$(openssl rand -base64 32)
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
cat > /etc/nginx/sites-available/expense-tracker << EOF
server {
    listen 80;
    server_name $DOMAIN;
    
    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name $DOMAIN;
    
    # SSL configuration (будет настроен certbot)
    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Proxy to frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Proxy to API
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Proxy to Analytics
    location /analytics/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Активация сайта
ln -sf /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Проверка конфигурации Nginx
nginx -t

# Создание systemd service
echo -e "${YELLOW}⏳ Создание systemd service...${NC}"
cat > /etc/systemd/system/expense-tracker.service << EOF
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

# Создание пользователя app
useradd -m -s /bin/bash app 2>/dev/null || true
usermod -aG docker app
chown -R app:app /root/expense-tracker

# Перезагрузка systemd
systemctl daemon-reload
systemctl enable expense-tracker

# Запуск приложения
echo -e "${YELLOW}⏳ Запуск приложения...${NC}"
sudo -u app docker-compose down 2>/dev/null || true
sudo -u app docker-compose build --no-cache
sudo -u app docker-compose up -d

# Ожидание запуска
echo -e "${YELLOW}⏳ Ожидание запуска сервисов...${NC}"
sleep 60

# Инициализация Ollama
echo -e "${YELLOW}⏳ Инициализация Ollama...${NC}"
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Ollama готов!${NC}"
        break
    fi
    
    echo -e "${YELLOW}⏳ Ollama не готов, ожидание 10 секунд... (попытка $((attempt + 1))/$max_attempts)${NC}"
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo -e "${RED}❌ Ollama не запустился за 5 минут${NC}"
    exit 1
fi

# Загрузка модели
echo -e "${YELLOW}⏳ Загрузка модели qwen2.5:0.5b...${NC}"
docker exec expense_ollama ollama pull qwen2.5:0.5b

# Тест модели
echo -e "${YELLOW}⏳ Тестирование модели...${NC}"
response=$(curl -s -X POST http://localhost:11434/api/generate -d '{
    "model": "qwen2.5:0.5b",
    "prompt": "Привет! Как дела?",
    "stream": false
}')

if echo "$response" | grep -q "response"; then
    echo -e "${GREEN}✅ Модель работает!${NC}"
else
    echo -e "${YELLOW}⚠️ Тест модели не прошел, но это может быть нормально${NC}"
fi

# Получение SSL сертификата
echo -e "${YELLOW}⏳ Получение SSL сертификата...${NC}"
certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN

# Перезапуск Nginx
systemctl reload nginx

# Запуск сервиса
systemctl start expense-tracker

# Проверка здоровья
echo -e "${YELLOW}⏳ Проверка здоровья сервисов...${NC}"
curl -f http://localhost:8080/api/health && echo -e "${GREEN}✅ API работает${NC}" || echo -e "${RED}❌ API не отвечает${NC}"
curl -f http://localhost:8081/health && echo -e "${GREEN}✅ Analytics работает${NC}" || echo -e "${RED}❌ Analytics не отвечает${NC}"
curl -f http://localhost:11434/api/tags && echo -e "${GREEN}✅ Ollama работает${NC}" || echo -e "${RED}❌ Ollama не отвечает${NC}"

# Проверка памяти
echo -e "${YELLOW}⏳ Проверка использования памяти...${NC}"
free -h
docker stats --no-stream

echo ""
echo -e "${GREEN}🎉 Развертывание завершено!${NC}"
echo ""
echo -e "${BLUE}📋 Информация:${NC}"
echo "🌐 Приложение: https://$DOMAIN"
echo "📊 API: https://$DOMAIN/api/health"
echo "🤖 Analytics: https://$DOMAIN/analytics/health"
echo "💾 Использование памяти: $(free -h | grep Mem | awk '{print $3 "/" $2}')"
echo ""
echo -e "${BLUE}🔧 Управление:${NC}"
echo "• Перезапуск: systemctl restart expense-tracker"
echo "• Логи: docker-compose logs -f"
echo "• Статус: systemctl status expense-tracker"
echo "• Обновление: cd /root/expense-tracker && git pull && docker-compose up -d --build"
echo ""
echo -e "${BLUE}⚠️ Важно:${NC}"
echo "1. Обновите TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_IDS в .env"
echo "2. Проверьте настройки безопасности"
echo "3. Настройте мониторинг памяти"
echo ""
echo -e "${GREEN}✅ Готово к использованию!${NC}"
