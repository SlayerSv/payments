#!/bin/sh

# Ждем, когда OpenBao станет доступен
until wget -qO- $BAO_ADDR/v1/sys/health | grep -q '"initialized":true'; do
  echo "Waiting for OpenBao to be ready..."
  sleep 2
done

echo "OpenBao is up! Configuring Transit Engine..."

# Логинимся (токен берем из переменной окружения)
export BAO_TOKEN="myroot"

# 1. Включаем Transit секреты (если еще не включены)
bao secrets enable transit || echo "Transit already enabled"

# 2. Создаем ключ для подписи JWT (типа ed25519 или rsa-2048)
# Назовем ключ 'jwt-key'
bao write -f transit/keys/jwt_key type=ed25519

echo "✅ OpenBao configured: Transit engine enabled and 'jwt_key' created."