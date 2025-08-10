#!/usr/bin/env bash
set -euo pipefail

echo "[postCreate] Installing frontend deps…"
cd frontend
if [ -f package.json ]; then
  npm install
else
  echo "No frontend/package.json found; skipping npm install."
fi

echo "[postCreate] Tidying Go modules…"
cd ../backend
if [ -f go.mod ]; then
  go mod tidy
else
  echo "No backend/go.mod found; skipping go mod tidy."
fi

echo "[postCreate] Done."
