#!/bin/bash

# Script to initialize Ollama with qwen2.5:0.5b model for 2GB RAM server
# This script should be run after Ollama container is started

echo "🚀 Initializing Ollama for 2GB RAM server..."

# Wait for Ollama to be ready
echo "⏳ Waiting for Ollama to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "✅ Ollama is ready!"
        break
    fi
    
    echo "⏳ Ollama not ready yet, waiting 10 seconds... (attempt $((attempt + 1))/$max_attempts)"
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Ollama failed to start within 5 minutes"
    exit 1
fi

# Check available models
echo "📋 Checking available models..."
curl -s http://localhost:11434/api/tags | jq -r '.models[]?.name // empty'

# Pull the lightweight model
echo "📥 Pulling qwen2.5:0.5b model (this may take a few minutes)..."
curl -X POST http://localhost:11434/api/pull -d '{
    "name": "qwen2.5:0.5b"
}' --progress-bar

echo ""

# Wait for model to be available
echo "⏳ Waiting for model to be available..."
max_attempts=20
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:11434/api/tags | grep -q "qwen2.5:0.5b"; then
        echo "✅ Model qwen2.5:0.5b is ready!"
        break
    fi
    
    echo "⏳ Model not ready yet, waiting 15 seconds... (attempt $((attempt + 1))/$max_attempts)"
    sleep 15
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Model failed to load within 5 minutes"
    exit 1
fi

# Test the model
echo "🧪 Testing model with a simple prompt..."
response=$(curl -s -X POST http://localhost:11434/api/generate -d '{
    "model": "qwen2.5:0.5b",
    "prompt": "Привет! Как дела?",
    "stream": false
}')

if echo "$response" | grep -q "response"; then
    echo "✅ Model test successful!"
    echo "📝 Response: $(echo "$response" | jq -r '.response // "No response"')"
else
    echo "❌ Model test failed"
    echo "Response: $response"
    exit 1
fi

# Check memory usage
echo "💾 Checking memory usage..."
memory_usage=$(docker stats expense_ollama --no-stream --format "table {{.MemUsage}}" | tail -1)
echo "📊 Ollama memory usage: $memory_usage"

# List available models
echo "📋 Available models:"
curl -s http://localhost:11434/api/tags | jq -r '.models[] | "  - \(.name) (\(.size | . / 1024 / 1024 / 1024 | floor)GB)"'

echo ""
echo "🎉 Ollama initialization completed successfully!"
echo "🔗 Ollama API: http://localhost:11434"
echo "📊 Analytics Service: http://localhost:8081"
echo "💡 Model: qwen2.5:0.5b (optimized for 2GB RAM)"
