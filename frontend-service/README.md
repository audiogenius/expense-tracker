# 🌐 Frontend Service

## Описание
React приложение для управления расходами с оптимизацией производительности.

## Функции
- React SPA
- Виртуализация списков
- Мемоизация компонентов
- Lazy loading
- API кэширование

## Архитектура
- React + TypeScript
- Vite
- Chart.js
- react-window

## Запуск
```bash
docker-compose up frontend
```

## Конфигурация
- API_URL
- PORT

## Компоненты
- BalanceCard - карточка баланса
- ExpenseLineChart - график расходов
- CategoryPieChart - диаграмма категорий
- VirtualizedTransactionList - виртуализированный список
- CategoryAutocomplete - автодополнение

## Логи
```bash
docker-compose logs -f frontend
```
