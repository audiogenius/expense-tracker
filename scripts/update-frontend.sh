#!/bin/bash

# Update frontend service with latest changes
# Run this on the server after pushing changes

echo "🔄 Updating frontend service..."

# Pull latest changes
git pull origin main

# Rebuild only frontend service
docker-compose build frontend

# Restart frontend service
docker-compose up -d frontend

echo "✅ Frontend service updated!"
echo "🌐 Check: https://rd-expense-tracker-bot.ru"

