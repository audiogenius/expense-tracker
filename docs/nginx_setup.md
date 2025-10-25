# 🌐 Nginx Setup Guide

## Установка
```bash
sudo apt update
sudo apt install nginx
```

## Конфигурация
1. Создать конфигурацию для проекта:
   ```bash
   sudo nano /etc/nginx/sites-available/expense-tracker
   ```

2. Настроить reverse proxy:
   ```nginx
   server {
       listen 80;
       server_name your-domain.com;
       
       location / {
           proxy_pass http://localhost:3000;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
       
       location /api {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

3. Активировать конфигурацию:
   ```bash
   sudo ln -s /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl reload nginx
   ```

## SSL сертификаты
```bash
# Установить certbot
sudo apt install certbot python3-certbot-nginx

# Получить сертификат
sudo certbot --nginx -d your-domain.com
```

## Мониторинг
```bash
# Проверка статуса
sudo systemctl status nginx

# Логи
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Тест конфигурации
sudo nginx -t
```
