#!/usr/bin/env bash
set -euo pipefail

# Defaults if not set in env
BACKEND_PORT="${BACKEND_PORT:-8081}"
FRONTEND_PORT="${FRONTEND_PORT:-5174}"

echo "Starting backend on :$BACKEND_PORT"
( cd backend && BACKEND_PORT="$BACKEND_PORT" go run gen-library/backend ) &

# Give the backend a second to start
sleep 1

echo "Starting frontend on :$FRONTEND_PORT"
( cd frontend && npm run dev -- --port "$FRONTEND_PORT" )
