# 🐳 Docker Setup Guide

## Установка Docker
```bash
# Обновить систему
sudo apt update

# Установить Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавить пользователя в группу docker
sudo usermod -aG docker $USER

# Установить Docker Compose
sudo apt install docker-compose-plugin
```

## Работа с проектом
```bash
# Запуск проекта
docker-compose up -d

# Остановка
docker-compose down

# Пересборка
docker-compose up --build -d

# Логи
docker-compose logs -f

# Статус
docker-compose ps
```

## Мониторинг ресурсов
```bash
# Использование памяти
docker stats

# Диск
docker system df

# Очистка
docker system prune -a
```
