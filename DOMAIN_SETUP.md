# 🌐 Настройка домена rd-expense-tracker-bot.ru

## 1️⃣ Telegram Bot Domain

**В Telegram → @BotFather:**

```
/setdomain
→ Выберите: @rd_expense_tracker_bot
→ Введите: rd-expense-tracker-bot.ru
```

✅ Готово! Widget авторизации будет работать.

---

## 2️⃣ DNS настройки

**У регистратора домена (REG.RU / Timeweb):**

```
Тип: A
Имя: @
Значение: ВАШ_IP_СЕРВЕРА
TTL: 3600
```

**Для www (опционально):**
```
Тип: CNAME
Имя: www
Значение: rd-expense-tracker-bot.ru
TTL: 3600
```

⏳ Ждите 5-30 минут распространения DNS.

---

## 3️⃣ Nginx на сервере

**SSH в сервер:**

```bash
# Установка Nginx
apt install -y nginx

# Создание конфига
nano /etc/nginx/sites-available/expense-tracker
```

**Вставьте:**

```nginx
server {
    listen 80;
    server_name rd-expense-tracker-bot.ru www.rd-expense-tracker-bot.ru;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Активация:**

```bash
# Создать symlink
ln -s /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/

# Проверка конфига
nginx -t

# Перезапуск
systemctl restart nginx
```

---

## 4️⃣ SSL сертификат (HTTPS)

```bash
# Установка Certbot
apt install -y certbot python3-certbot-nginx

# Получение сертификата
certbot --nginx -d rd-expense-tracker-bot.ru -d www.rd-expense-tracker-bot.ru

# Email для уведомлений
# Согласитесь с условиями
# Выберите: Redirect HTTP to HTTPS

# Автообновление SSL
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

---

## 5️⃣ Обновление .env на сервере

```bash
cd /root/expense-tracker
nano .env
```

**Ничего менять НЕ НУЖНО!** 

`API_URL=http://api:8080` - это для внутренней сети Docker.

---

## 6️⃣ Проверка

**После настройки DNS (5-30 мин):**

```bash
# Проверка DNS
ping rd-expense-tracker-bot.ru

# Должен показать ваш IP
```

**В браузере:**

- http://rd-expense-tracker-bot.ru → перенаправит на HTTPS
- https://rd-expense-tracker-bot.ru → откроется сайт

---

## 🔥 Firewall (важно!)

**В панели Timeweb откройте порты:**
- 80 (HTTP)
- 443 (HTTPS)
- 22 (SSH)

**Или через ufw на сервере:**

```bash
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

---

## 📋 Чеклист

- [ ] `/setdomain` в @BotFather
- [ ] DNS A-запись настроена
- [ ] Nginx установлен и настроен
- [ ] SSL сертификат получен
- [ ] Порты 80, 443 открыты
- [ ] Сайт доступен по домену
- [ ] HTTPS редирект работает
- [ ] Telegram Login работает

---

## ⚠️ Проблемы

### "Connection refused"
```bash
# Проверьте что контейнеры запущены
docker-compose ps

# Проверьте nginx
systemctl status nginx

# Проверьте логи
docker-compose logs api
docker-compose logs frontend
```

### "502 Bad Gateway"
```bash
# Перезапустите контейнеры
docker-compose restart

# Проверьте порты
netstat -tlnp | grep :3000
netstat -tlnp | grep :8080
```

### Telegram Widget не работает
- Подождите 30 мин после `/setdomain`
- Очистите кеш браузера
- Проверьте HTTPS (должен быть!)

---

**Готово!** Ваш сайт будет на https://rd-expense-tracker-bot.ru 🚀

