#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting service"

echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Go app on port 8080"
/app/bin/app &
APP_PID=$!

# Ждем, пока Go приложение запустится
sleep 2

echo "[run.sh] Starting Caddy on port ${PORT:-80}"
exec caddy run --config /etc/caddy/Caddyfile

