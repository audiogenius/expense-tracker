# 🚀 Запуск проекта

## Требования
- Docker & Docker Compose
- 2GB RAM
- Ubuntu 22.04+ (для продакшена)

## Локальная разработка

### 1. Клонирование репозитория
```bash
git clone https://github.com/audiogenius/expense-tracker.git
cd expense-tracker
```

### 2. Настройка .env
```bash
cp env.example .env
nano .env
```

### 3. Запуск проекта
```bash
docker-compose up --build -d
```

### 4. Проверка работы
```bash
# Проверить статус
docker-compose ps

# Проверить логи
docker-compose logs -f

# Проверить API
curl http://localhost:8080/health

# Проверить Analytics
curl http://localhost:8081/health

# Проверить Ollama
curl http://localhost:11434/api/tags
```

## Продакшен

### 1. Настройка сервера
```bash
# Подключиться к серверу
ssh root@your-server-ip

# Установить Docker
curl -sSL https://get.docker.com | sh

# Установить Docker Compose
sudo apt install docker-compose-plugin
```

### 2. Настройка Nginx
```bash
# Установить Nginx
sudo apt install nginx

# Настроить конфигурацию
sudo nano /etc/nginx/sites-available/expense-tracker

# Активировать
sudo ln -s /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 3. SSL сертификаты
```bash
# Установить certbot
sudo apt install certbot python3-certbot-nginx

# Получить сертификат
sudo certbot --nginx -d your-domain.com
```

## Конфигурация

### База данных
- PostgreSQL 16
- Оптимизированные индексы
- pg_trgm для поиска

### Telegram бот
1. Создать бота через @BotFather
2. Получить токен
3. Настроить в .env

### Ollama
1. Загрузить модель: `docker exec expense_ollama ollama pull qwen2.5:0.5b`
2. Проверить: `curl http://localhost:11434/api/tags`

## Мониторинг

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

## Решение проблем

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

## Обновления

### Получение обновлений
```bash
git fetch origin
git checkout feature/performance-optimization-2gb-ram
git pull origin feature/performance-optimization-2gb-ram
```

### Применение обновлений
```bash
docker-compose down
docker-compose up --build -d
```

## Поддержка
- **GitHub Issues**: https://github.com/audiogenius/expense-tracker/issues
- **Документация**: docs/
- **Обновления**: docs/updates/
