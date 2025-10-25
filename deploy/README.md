# 🚀 Deploy скрипты для Expense Tracker

## 📁 Содержимое папки

### 🐳 `docker-compose.timeweb.yml`
**Оптимизированная конфигурация для Timeweb Cloud MSK 30 (2GB RAM)**

**Особенности:**
- ✅ Ограничения памяти для всех сервисов
- ✅ Ollama: 1.2GB RAM (вместо 1.5GB)
- ✅ PostgreSQL: 512MB RAM
- ✅ API: 256MB RAM
- ✅ Analytics: 128MB RAM
- ✅ Остальные сервисы: 64-128MB RAM

**Использование:**
```bash
# Копирование оптимизированной конфигурации
cp deploy/docker-compose.timeweb.yml docker-compose.yml

# Запуск
docker-compose up --build -d

# Проверка статуса
docker-compose ps
```

### 🔧 `timeweb-setup.sh`
**Полная автоматическая настройка сервера**

**Что делает:**
- ✅ Обновление системы
- ✅ Установка Docker и Docker Compose
- ✅ Установка Nginx, SSL, безопасности
- ✅ Клонирование проекта
- ✅ Настройка systemd сервиса
- ✅ Создание скриптов мониторинга

**Использование:**
```bash
# Автоматическая установка
curl -sSL https://raw.githubusercontent.com/audiogenius/expense-tracker/main/deploy/timeweb-setup.sh | bash

# Или локально
chmod +x deploy/timeweb-setup.sh
./deploy/timeweb-setup.sh
```

### 🚀 `deploy-to-timeweb.sh`
**Быстрое развертывание с доменом**

**Что делает:**
- ✅ Клонирование проекта
- ✅ Настройка .env
- ✅ Настройка Nginx с SSL
- ✅ Создание systemd сервиса
- ✅ Инициализация Ollama

**Использование:**
```bash
# Развертывание с доменом
chmod +x deploy/deploy-to-timeweb.sh
./deploy/deploy-to-timeweb.sh your-domain.com
```

---

## 🎯 Какой скрипт использовать?

### 🆕 **Первый раз (новая установка)**
```bash
# Полная автоматическая установка
curl -sSL https://raw.githubusercontent.com/audiogenius/expense-tracker/main/deploy/timeweb-setup.sh | bash
```

### 🔄 **Обновление существующего**
```bash
# Переход в проект
cd /root/expense-tracker

# Обновление кода
git pull origin feature/performance-optimization-2gb-ram

# Пересборка
docker-compose up --build -d
```

### 🌐 **Развертывание с доменом**
```bash
# Быстрое развертывание
./deploy/deploy-to-timeweb.sh your-domain.com
```

---

## 📊 Мониторинг ресурсов

### Проверка использования памяти
```bash
# Общее использование
free -h

# Docker статистика
docker stats --no-stream

# Конкретный контейнер
docker stats expense_ollama
```

### Проверка статуса сервисов
```bash
# Статус контейнеров
docker-compose ps

# Логи
docker-compose logs -f

# Health checks
curl http://localhost:8080/api/health
curl http://localhost:8081/health
curl http://localhost:11434/api/tags
```

---

## ⚠️ Важные ограничения для 2GB RAM

### 🐳 **Ограничения памяти:**
- **Ollama**: 1.2GB (основной потребитель)
- **PostgreSQL**: 512MB
- **API**: 256MB
- **Analytics**: 128MB
- **Frontend**: 128MB
- **Bot**: 128MB
- **OCR**: 128MB
- **Proxy**: 64MB

### 🔧 **Оптимизации:**
- Используется Alpine Linux для экономии места
- Ограниченное количество параллельных процессов Ollama
- Минимальные резервы памяти
- Health checks для автоматического перезапуска

---

## 🆘 Решение проблем

### Высокое потребление памяти
```bash
# Перезапуск тяжелых сервисов
docker-compose restart ollama analytics

# Проверка использования
docker stats
```

### Медленная загрузка Ollama
```bash
# Проверка статуса
curl http://localhost:11434/api/tags

# Перезапуск
docker-compose restart ollama
```

### Проблемы с базой данных
```bash
# Проверка логов
docker-compose logs db

# Перезапуск
docker-compose restart db
```

---

## 📞 Поддержка

- **GitHub Issues**: https://github.com/audiogenius/expense-tracker/issues
- **Документация**: [HOSTING_SETUP.md](../HOSTING_SETUP.md)
- **Обновления**: [docs/updates/](../docs/updates/)

---

**Готово к развертыванию на Timeweb Cloud MSK 30! 🚀**
