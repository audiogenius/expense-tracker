#!/bin/bash

# Script to initialize Ollama with qwen2.5:0.5b model
# This script should be run after Ollama container is started

echo "Initializing Ollama with qwen2.5:0.5b model..."

# Wait for Ollama to be ready
echo "Waiting for Ollama to be ready..."
until curl -s http://ollama:11434/api/tags > /dev/null 2>&1; do
    echo "Ollama not ready yet, waiting 5 seconds..."
    sleep 5
done

echo "Ollama is ready, pulling qwen2.5:0.5b model..."

# Pull the model
curl -X POST http://ollama:11434/api/pull -d '{
    "name": "qwen2.5:0.5b"
}'

echo "Model pull initiated. This may take a few minutes..."

# Wait for model to be available
echo "Waiting for model to be available..."
until curl -s http://ollama:11434/api/tags | grep -q "qwen2.5:0.5b"; do
    echo "Model not ready yet, waiting 10 seconds..."
    sleep 10
done

echo "Ollama initialization completed successfully!"
echo "Model qwen2.5:0.5b is ready for use."
