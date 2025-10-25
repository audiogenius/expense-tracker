#!/bin/bash

echo "=== Очистка лишних скриптов ==="

echo "1. Удаляем дублирующие и ненужные скрипты:"
rm -f scripts/debug-403.sh
rm -f scripts/debug-403-final.sh
rm -f scripts/fix-nginx.sh
rm -f scripts/fix-nginx-manual.sh
rm -f scripts/fix-nginx-post.sh
rm -f scripts/fix-nginx-cors.sh
rm -f scripts/fix-nginx-final.sh
rm -f scripts/clean-nginx.sh
rm -f scripts/recreate-nginx.sh
rm -f scripts/disable-https-redirect.sh
rm -f scripts/disable-https-completely.sh
rm -f scripts/test-api.sh
rm -f scripts/test-api-simple.sh
rm -f scripts/test-frontend-api.sh
rm -f scripts/test-auth.sh
rm -f scripts/debug-auth.sh
rm -f scripts/debug-login.sh
rm -f scripts/check-nginx-config.sh

echo "2. Оставляем только нужные скрипты:"
echo "✅ fix-nginx-simple.sh - основной скрипт исправления nginx"
echo "✅ test-full-flow.sh - полное тестирование системы"
echo "✅ init-ollama.sh - инициализация Ollama"
echo "✅ init-ollama.ps1 - инициализация Ollama для Windows"

echo "3. Проверяем оставшиеся скрипты:"
ls -la scripts/

echo ""
echo "=== Очистка завершена ==="
