#!/bin/bash

# Script to update expense tracker on server
# Run this on the server: root@147.45.246.210

echo "🚀 Updating Expense Tracker on server..."

# Navigate to project directory
cd /root/expense-tracker

# Pull latest changes from GitHub
echo "📥 Pulling latest changes from GitHub..."
git pull origin main

# Stop all services
echo "🛑 Stopping all services..."
docker-compose down

# Rebuild and start services
echo "🔨 Rebuilding and starting services..."
docker-compose up --build -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 30

# Check service status
echo "📊 Checking service status..."
docker-compose ps

# Test API health
echo "🔍 Testing API health..."
curl -f http://localhost:8080/api/health || echo "❌ API health check failed"

# Test Analytics health
echo "🔍 Testing Analytics health..."
curl -f http://localhost:8081/health || echo "❌ Analytics health check failed"

# Test Ollama
echo "🔍 Testing Ollama..."
curl -f http://localhost:11434/api/tags || echo "❌ Ollama health check failed"

echo "✅ Update completed!"
echo "🌐 Frontend: http://147.45.246.210:3000"
echo "🔧 API: http://147.45.246.210:8080"
echo "📊 Analytics: http://147.45.246.210:8081"
