# Expense Tracker Makefile

.PHONY: help build up down logs clean test

# Default target
help:
	@echo "Expense Tracker - Available commands:"
	@echo ""
	@echo "  make build     - Build all Docker images"
	@echo "  make up        - Start all services"
	@echo "  make down      - Stop all services"
	@echo "  make logs      - Show logs from all services"
	@echo "  make clean     - Clean up Docker resources"
	@echo "  make test      - Run tests"
	@echo "  make dev       - Start development environment"
	@echo "  make prod      - Start production environment"

# Build all services
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Start with build
up-build:
	docker-compose up --build -d

# Stop all services
down:
	docker-compose down

# Show logs
logs:
	docker-compose logs -f

# Show logs for specific service
logs-bot:
	docker-compose logs -f bot

logs-api:
	docker-compose logs -f api

logs-db:
	docker-compose logs -f db

# Clean up Docker resources
clean:
	docker-compose down -v
	docker system prune -f

# Run tests
test:
	docker-compose exec api go test ./...
	docker-compose exec bot go test ./...

# Development environment
dev:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Production environment
prod:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Database operations
db-shell:
	docker-compose exec db psql -U expense_user -d expense_tracker

db-backup:
	docker-compose exec db pg_dump -U expense_user expense_tracker > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore:
	docker-compose exec -T db psql -U expense_user -d expense_tracker < $(FILE)

# Status check
status:
	docker-compose ps

# Restart services
restart:
	docker-compose restart

# Update and restart
update:
	git pull
	docker-compose up --build -d
