# 🚀 Команды для обновления сервера

## Подключение к серверу
```bash
ssh root@147.45.246.210
```

## Обновление проекта
```bash
# Перейти в директорию проекта
cd /root/expense-tracker

# Получить последние изменения
git pull origin main

# Остановить все сервисы
docker-compose down

# Пересобрать и запустить сервисы
docker-compose up --build -d

# Проверить статус сервисов
docker-compose ps
```

## Проверка работоспособности
```bash
# Проверить API
curl http://localhost:8080/api/health

# Проверить Analytics
curl http://localhost:8081/health

# Проверить Ollama
curl http://localhost:11434/api/tags

# Проверить логи
docker-compose logs -f
```

## Быстрое обновление (одной командой)
```bash
cd /root/expense-tracker && git pull origin main && docker-compose down && docker-compose up --build -d
```

## Что было обновлено:
✅ **Исправлено разделение расходов** - бот теперь может находить пользователей по @username
✅ **Реализован Lazy Loading** - компоненты загружаются по требованию
✅ **Улучшена Ollama** - добавлена память и контекст разговора
✅ **Исправлены все ошибки** - TypeScript и Go ошибки устранены
✅ **Модульная архитектура** - код разбит на модули согласно принципам чистой архитектуры

## Доступные сервисы:
- 🌐 **Frontend**: http://147.45.246.210:3000
- 🔧 **API**: http://147.45.246.210:8080  
- 📊 **Analytics**: http://147.45.246.210:8081
- 🤖 **Ollama**: http://147.45.246.210:11434
