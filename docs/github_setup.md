# 📚 GitHub Setup Guide

## Создание репозитория
1. Создать новый репозиторий на GitHub
2. Клонировать локально:
   ```bash
   git clone https://github.com/username/expense-tracker.git
   cd expense-tracker
   ```
3. Настроить remote:
   ```bash
   git remote -v
   ```

## Работа с ветками
- `main` - основная ветка
- `feature/*` - новые функции
- `hotfix/*` - исправления

## Коммиты
- Использовать conventional commits
- Писать понятные сообщения
- Делать атомарные коммиты

## Pull Request
1. Создать ветку для новой функции
2. Внести изменения
3. Создать PR в main
4. Провести code review
5. Слить изменения

## Синхронизация
```bash
# Получить изменения
git fetch origin

# Обновить локальную ветку
git pull origin main

# Отправить изменения
git push origin feature/new-feature
```
