Expense Tracker v1.1 - Monorepo

This repository contains a modular microservices skeleton for the Family Expense Tracker project (MVP).

Services:
- bot-service (Telegram bot)
- api-service (Go REST API)
- ocr-service (OCR: Google Cloud Vision + Tesseract fallback)
- frontend-service (React + TypeScript + Vite)
- db (PostgreSQL)
- proxy-service (Nginx)

Prerequisites (Windows 11):
1. Install Docker Desktop: https://www.docker.com/get-started
2. Install Visual Studio Code: https://code.visualstudio.com/
3. (Optional) Install Git: https://git-scm.com/

Quick start:
1. Copy `.env.example` to `.env` and fill secrets.
2. Build and run:

```powershell
docker compose up --build
```

Check containers:

```powershell
docker ps
```

Open frontend: http://localhost

Notes:
- Follow `docs/` for step-by-step guides: creating Google Cloud API key, Telegram Bot, and Telegram Login Widget.
- This is a skeleton; each service contains a minimal Dockerfile and placeholder code.
