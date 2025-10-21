# 🚀 POLETELI - Полная инструкция деплоя для новичка

## 📋 Содержание
1. [Выгрузка на GitHub](#1️⃣-выгрузка-на-github)
2. [Подключение к серверу](#2️⃣-подключение-к-серверу-timeweb)
3. [Загрузка проекта на сервер](#3️⃣-загрузка-проекта-на-сервер)
4. [Создание и настройка .env](#4️⃣-создание-и-настройка-env)
5. [Проверка файлов](#5️⃣-проверка-файлов)
6. [Запуск сайта](#6️⃣-запуск-сайта)
7. [Настройка домена](#7️⃣-настройка-домена)
8. [Подключение VS Code к серверу](#8️⃣-подключение-vs-code-к-серверу-бонус)

---

## 1️⃣ Выгрузка на GitHub

### Способ 1: Через GitHub Desktop (САМЫЙ ПРОСТОЙ)

1. **Скачайте GitHub Desktop:**
   - Перейдите: https://desktop.github.com/
   - Нажмите **Download for Windows**
   - Установите программу

2. **Войдите в GitHub:**
   - Откройте GitHub Desktop
   - File → Options → Accounts → Sign in
   - Введите логин/пароль от GitHub

3. **Добавьте проект:**
   - File → Add Local Repository
   - Нажмите **Choose...**
   - Выберите папку: `C:\expense-tracker\expense-tracker_1_0_cursor`
   - Нажмите **Add Repository**

4. **Создайте репозиторий на GitHub:**
   - Repository → Repository Settings
   - Нажмите **Publish repository**
   - Имя: `expense-tracker`
   - Снимите галочку **Keep this code private** (если хотите публичный)
   - Нажмите **Publish Repository**

5. **Готово!** Проект на GitHub: `https://github.com/audiogenius/expense-tracker`

---

### Способ 2: Через Cursor (если настроен Git)

1. **В Cursor:**
   - Откройте Source Control (Ctrl+Shift+G)
   - Нажмите **+** рядом с измененными файлами (Stage)
   - Введите сообщение: `Initial commit`
   - Нажмите **✓ Commit**

2. **Push на GitHub:**
   - Нажмите **...** (три точки)
   - Push
   - Выберите **origin**

---

### Способ 3: Через PowerShell (для продвинутых)

1. **Откройте PowerShell в папке проекта:**
   - Shift + Правая кнопка в папке `expense-tracker_1_0_cursor`
   - "Открыть окно PowerShell здесь"

2. **Выполните команды:**

```powershell
# Проверка статуса
git status

# Добавить все файлы
git add .

# Сделать коммит
git commit -m "Initial commit: expense tracker v1.1"

# Подключить GitHub репозиторий
git remote add origin https://github.com/audiogenius/expense-tracker.git

# Отправить на GitHub
git push -u origin main
```

3. **Если попросит логин:**
   - Введите username от GitHub
   - Пароль: используйте **Personal Access Token** (не обычный пароль!)
   - Как создать токен: https://github.com/settings/tokens → Generate new token

---

## 2️⃣ Подключение к серверу Timeweb

### Что вам нужно знать:
- **IP адрес:** `147.45.246.210`
- **Логин:** `root`
- **Пароль:** (из письма от Timeweb или в панели управления)

---

### Вариант А: Через PuTTY (для Windows, РЕКОМЕНДУЕТСЯ)

1. **Скачайте PuTTY:**
   - https://www.putty.org/
   - Скачайте файл **putty.exe**
   - Запустите (установка не нужна)

2. **Подключитесь:**
   - **Host Name:** `147.45.246.210`
   - **Port:** `22`
   - **Connection type:** SSH
   - Нажмите **Open**

3. **Первое подключение:**
   - Появится предупреждение о ключе
   - Нажмите **Yes** (Accept)

4. **Вход:**
   - **login as:** `root`
   - Нажмите Enter
   - **Password:** вставьте пароль (правая кнопка мыши = вставить)
   - Пароль НЕ ОТОБРАЖАЕТСЯ при вводе - это нормально!
   - Нажмите Enter

5. **Вы на сервере!** Видите строку:
   ```
   root@server:~#
   ```

---

### Вариант Б: Через PowerShell (встроенный в Windows)

1. **Откройте PowerShell:**
   - Win + X → Windows PowerShell

2. **Подключитесь:**
   ```powershell
   ssh root@147.45.246.210
   ```

3. **Первое подключение:**
   - Напишите `yes` и нажмите Enter

4. **Введите пароль:**
   - Вставьте пароль (Ctrl+V или правая кнопка)
   - Нажмите Enter

---

## 3️⃣ Загрузка проекта на сервер

**ВЫ ДОЛЖНЫ БЫТЬ ПОДКЛЮЧЕНЫ К СЕРВЕРУ (см. Шаг 2)**

### Шаг 1: Установка необходимых программ

```bash
# Обновление системы
apt update && apt upgrade -y

# Установка Git
apt install -y git

# Установка Docker
curl -fsSL https://get.docker.com | sh

# Установка Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Проверка установки
docker --version
docker-compose --version
```

**Копируйте команды по одной и нажимайте Enter!**

Процесс займет 2-5 минут.

---

### Шаг 2: Клонирование проекта с GitHub

```bash
# Перейти в домашнюю папку
cd /root

# Склонировать проект
git clone https://github.com/audiogenius/expense-tracker.git

# Перейти в папку проекта
cd expense-tracker

# Проверить что все скачалось
ls -la
```

**Вы должны увидеть папки:**
- api-service
- bot-service
- db
- frontend-service
- и другие

---

## 4️⃣ Создание и настройка .env

### Шаг 1: Создать файл .env

```bash
# Убедитесь что вы в папке проекта
pwd
# Должно показать: /root/expense-tracker

# Создать .env из примера
cp env.example .env

# Открыть редактор nano
nano .env
```

---

### Шаг 2: Заполнить .env

**Внутри редактора nano:**

1. **Удалите все** (Ctrl+K несколько раз)

2. **Вставьте это** (правая кнопка мыши или Shift+Insert):

```env
# Database Configuration
POSTGRES_USER=expense_user
POSTGRES_PASSWORD=YourStrongPassword123!
POSTGRES_DB=expense_tracker
TZ=Europe/Moscow

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=8334084074:AAGSS8tyvrRSzH1XwNdDaDNpy52KlkdBc6E
TELEGRAM_WHITELIST=260144148,1063957118

# API Configuration
BOT_API_KEY=Yoq6zEpXJB4f8KmVnLtD9hRw3NsM7u2g
JWT_SECRET=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwOTA4MTk5MCIsIm5hbWUiOiJBYmxlZXYgRGluaXNsYW0iLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyfQ.Y8tbSPLoDha4QYRqGurvb5HEt67eLPJsoSDzvJrreq0

# API URLs
API_URL=http://api:8080

# Google Cloud Vision (optional for OCR)
USE_LOCAL_OCR=false
```

3. **Сохранить файл:**
   - Нажмите **Ctrl + O** (буква O, не ноль)
   - Появится: `File Name to Write: .env`
   - Нажмите **Enter**
   - Внизу появится: `[ Wrote 18 lines ]`

4. **Выйти из редактора:**
   - Нажмите **Ctrl + X**

---

### Памятка по nano (редактор текста):

```
Ctrl + O  → Сохранить (Write Out)
Enter     → Подтвердить имя файла
Ctrl + X  → Выйти (Exit)
Ctrl + K  → Удалить строку
Ctrl + U  → Вставить
```

**ВАЖНО:** При сохранении нажимайте английскую букву **O**, не цифру 0!

---

## 5️⃣ Проверка файлов

### Посмотреть список всех файлов:

```bash
# Подробный список
ls -lah

# Только имена файлов
ls
```

---

### Проверить есть ли .env:

```bash
# Проверить существование
ls -la | grep .env

# Если файл есть, увидите строку типа:
# -rw-r--r-- 1 root root  512 Oct 21 15:30 .env
```

---

### Посмотреть содержимое .env:

```bash
# Показать весь файл
cat .env

# Показать только важные строки
grep TELEGRAM .env
grep BOT_API .env
grep JWT .env
```

**Проверьте что все токены на месте!**

---

### Проверить ключевые параметры:

```bash
# Проверить что Telegram токен есть
grep TELEGRAM_BOT_TOKEN .env

# Должно показать:
# TELEGRAM_BOT_TOKEN=8334084074:...
```

Если видите ваш токен - **все OK!**

---

## 6️⃣ Запуск сайта

### Шаг 1: Скопировать package-lock.json

**Вариант А: С вашего компьютера (Windows PowerShell):**

```powershell
# Откройте PowerShell на Windows
# Перейдите в папку проекта
cd C:\expense-tracker\expense-tracker_1_0_cursor

# Скопируйте файл на сервер
scp frontend-service/package-lock.json root@147.45.246.210:/root/expense-tracker/frontend-service/
```

Введите пароль от сервера.

---

**Вариант Б: Сгенерировать на сервере:**

```bash
# На сервере
cd /root/expense-tracker/frontend-service

# Установить зависимости
npm install

# Вернуться в корень
cd /root/expense-tracker
```

---

### Шаг 2: Запустить Docker

```bash
# Убедитесь что вы в папке проекта
cd /root/expense-tracker

# Запустить все сервисы
docker-compose up -d
```

**Что происходит:**
- Docker скачивает образы (PostgreSQL, Node, Go)
- Собирает ваши сервисы
- Запускает контейнеры

**Время:** 5-10 минут при первом запуске.

---

### Шаг 3: Проверить статус

```bash
# Проверить статус контейнеров
docker-compose ps
```

**Вы должны увидеть:**
```
NAME                 STATUS              PORTS
expense_db          Up (healthy)        5432/tcp
expense_api         Up (healthy)        8080/tcp
expense_bot         Up                  
expense_frontend    Up                  80/tcp
expense_proxy       Up                  80/tcp, 443/tcp
```

**Все должны быть `Up` и `healthy`!**

---

### Шаг 4: Посмотреть логи

```bash
# Логи всех сервисов
docker-compose logs -f

# Только бот
docker-compose logs bot

# Только API
docker-compose logs api
```

**В логах бота должно быть:**
```
✅ Bot is running! Waiting for messages...
```

**Выход из логов:** Нажмите `Ctrl + C`

---

### Шаг 5: Проверить работу

**В браузере на вашем компьютере:**

```
http://147.45.246.210
```

Должен открыться сайт!

**Проверка API:**
```
http://147.45.246.210:8080/health
```

Должно показать: `{"status":"ok","service":"api"}`

---

## 7️⃣ Настройка домена

### Шаг 1: Настроить DNS

**Зайдите в панель управления доменом (где купили домен):**

1. Найдите раздел **DNS записи** или **DNS управление**
2. Создайте **A-запись:**
   ```
   Тип: A
   Имя: @ (или оставьте пустым)
   Значение: 147.45.246.210
   TTL: 3600
   ```
3. Создайте **A-запись для www:**
   ```
   Тип: A
   Имя: www
   Значение: 147.45.246.210
   TTL: 3600
   ```
4. Сохраните изменения

**Ожидание:** 5-30 минут пока DNS обновится.

---

### Шаг 2: Настроить Telegram Bot Domain

**В Telegram на телефоне:**

1. Найдите бота **@BotFather**
2. Отправьте команду: `/setdomain`
3. Выберите вашего бота: `@rd_expense_tracker_bot`
4. Введите домен: `rd-expense-tracker-bot.ru`
5. Готово!

---

### Шаг 3: Установить Nginx

**На сервере (в PuTTY/PowerShell SSH):**

```bash
# Установка Nginx
apt install -y nginx

# Проверка
nginx -v
```

---

### Шаг 4: Настроить Nginx

```bash
# Создать конфиг
nano /etc/nginx/sites-available/expense-tracker
```

**Вставьте в редактор:**

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

**Сохраните:** Ctrl+O → Enter → Ctrl+X

---

### Шаг 5: Активировать конфиг

```bash
# Создать ссылку
ln -s /etc/nginx/sites-available/expense-tracker /etc/nginx/sites-enabled/

# Проверить конфиг
nginx -t

# Должно показать:
# nginx: configuration file /etc/nginx/nginx.conf test is successful

# Перезапустить Nginx
systemctl restart nginx

# Проверить статус
systemctl status nginx
```

**Должно быть:** `active (running)` зеленым цветом

---

### Шаг 6: Открыть порты в Firewall

**В панели Timeweb:**

1. Зайдите в панель управления сервером
2. Найдите раздел **Firewall** или **Безопасность**
3. Откройте порты:
   - **80** (HTTP)
   - **443** (HTTPS)
   - **22** (SSH - должен быть уже открыт)

**Или через командную строку на сервере:**

```bash
# Установить ufw
apt install -y ufw

# Разрешить SSH (ВАЖНО сделать первым!)
ufw allow 22/tcp

# Разрешить HTTP
ufw allow 80/tcp

# Разрешить HTTPS
ufw allow 443/tcp

# Включить firewall
ufw enable

# Проверить статус
ufw status
```

---

### Шаг 7: Установить SSL сертификат (HTTPS)

```bash
# Установка Certbot
apt install -y certbot python3-certbot-nginx

# Получить сертификат
certbot --nginx -d rd-expense-tracker-bot.ru -d www.rd-expense-tracker-bot.ru
```

**Вам зададут вопросы:**

1. **Email:** введите ваш email (для уведомлений)
2. **Agree to Terms:** `Y` (согласиться)
3. **Share email:** `N` (не делиться)
4. **Redirect HTTP to HTTPS:** `2` (перенаправлять)

**Готово!** Теперь сайт доступен по HTTPS.

---

### Шаг 8: Автообновление SSL

```bash
# Добавить в cron
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -

# Проверить
crontab -l
```

---

### Шаг 9: Проверка

**В браузере:**

```
http://rd-expense-tracker-bot.ru
```

Должно перенаправить на:

```
https://rd-expense-tracker-bot.ru
```

**Telegram Login Widget должен работать!**

---

## 8️⃣ Подключение VS Code к серверу (БОНУС)

### Способ 1: SSH через VS Code (Remote SSH)

1. **Установите расширение:**
   - Откройте VS Code
   - Extensions (Ctrl+Shift+X)
   - Найдите: **Remote - SSH**
   - Нажмите **Install**

2. **Подключитесь:**
   - Нажмите F1
   - Введите: `Remote-SSH: Connect to Host`
   - Выберите **+ Add New SSH Host**
   - Введите: `ssh root@147.45.246.210`
   - Выберите файл конфига (первый вариант)
   - Нажмите **Connect**

3. **Введите пароль:**
   - В открывшемся окне введите пароль
   - Готово! Вы подключены к серверу

4. **Открыть проект:**
   - File → Open Folder
   - Выберите: `/root/expense-tracker`
   - Нажмите OK

**Теперь вы можете редактировать файлы прямо на сервере!**

---

### Способ 2: SFTP (для файлов)

1. **Установите FileZilla:**
   - https://filezilla-project.org/
   - Download FileZilla Client
   - Установите

2. **Подключитесь:**
   - Хост: `sftp://147.45.246.210`
   - Имя пользователя: `root`
   - Пароль: ваш пароль
   - Порт: `22`
   - Нажмите **Быстрое соединение**

3. **Работа с файлами:**
   - Слева - ваш компьютер
   - Справа - сервер
   - Перетаскивайте файлы между ними

---

## 🎉 ГОТОВО!

### Что теперь работает:

- ✅ **Проект на GitHub:** https://github.com/audiogenius/expense-tracker
- ✅ **Сервер настроен:** Ubuntu 22.04, Docker, Nginx
- ✅ **Сайт работает:** https://rd-expense-tracker-bot.ru
- ✅ **Telegram бот:** @rd_expense_tracker_bot
- ✅ **SSL сертификат:** HTTPS включен
- ✅ **Доступ к серверу:** через SSH, VS Code

---

## 📝 Полезные команды для работы

### На сервере:

```bash
# Перейти в проект
cd /root/expense-tracker

# Посмотреть статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f

# Перезапустить
docker-compose restart

# Остановить
docker-compose down

# Запустить заново
docker-compose up -d

# Обновить проект с GitHub
git pull
docker-compose up --build -d

# Посмотреть использование ресурсов
docker stats

# Посмотреть диск
df -h

# Посмотреть RAM
free -h
```

---

### На компьютере:

```powershell
# Подключиться к серверу
ssh root@147.45.246.210

# Скопировать файл на сервер
scp C:\path\to\file.txt root@147.45.246.210:/root/

# Скопировать файл с сервера
scp root@147.45.246.210:/root/file.txt C:\Downloads\
```

---

## 🆘 Решение проблем

### Сайт не открывается

```bash
# Проверить контейнеры
docker-compose ps

# Если не healthy - посмотреть логи
docker-compose logs api
docker-compose logs frontend

# Перезапустить
docker-compose restart
```

---

### Бот не отвечает

```bash
# Проверить логи бота
docker-compose logs bot

# Должно быть:
# ✅ Bot is running! Waiting for messages...

# Если нет - проверить .env
cat .env | grep TELEGRAM

# Перезапустить бота
docker-compose restart bot
```

---

### 502 Bad Gateway

```bash
# Проверить что Docker запущен
docker-compose ps

# Перезапустить Nginx
systemctl restart nginx

# Проверить порты
netstat -tlnp | grep :3000
netstat -tlnp | grep :8080
```

---

### Забыли пароль от сервера

**В панели Timeweb:**
- Перейдите в управление сервером
- Нажмите **Сброс пароля**
- Новый пароль придет на email

---

## 📞 Контакты и помощь

- **GitHub Issues:** https://github.com/audiogenius/expense-tracker/issues
- **Документация Timeweb:** https://timeweb.cloud/help
- **Telegram Support Bot:** @timeweb_support_bot

---

## 🎓 Что вы изучили:

1. ✅ Git и GitHub
2. ✅ SSH подключение
3. ✅ Linux (Ubuntu) командная строка
4. ✅ Docker и Docker Compose
5. ✅ Nginx веб-сервер
6. ✅ SSL сертификаты
7. ✅ DNS настройки
8. ✅ Firewall
9. ✅ VS Code Remote SSH

**Поздравляю! Вы настоящий DevOps! 🚀**

---

_Создано с ❤️ для вашего проекта Expense Tracker_

