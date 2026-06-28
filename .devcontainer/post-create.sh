#!/bin/bash
set -e

echo "=== FormulAI Dev Container Setup ==="

# ── Go backend ──
echo "--- Installing Go modules ---"
cd /workspace/backend && go mod download

# ── Angular frontend ──
echo "--- Installing npm packages ---"
cd /workspace/frontend && npm ci 2>/dev/null || npm install

# ── Python Agent ──
echo "--- Installing Python packages ---"
/opt/agent-venv/bin/pip install --no-cache-dir /workspace/agent-service/

echo ""
echo "=== Setup complete ==="
echo ""
echo "Run these in separate terminals:"
echo "  1. cd backend && air                    # Go API on :8080"
echo "  2. cd frontend && npx ng serve --host 0.0.0.0  # Angular on :4200"
echo "  3. cd agent-service && uvicorn src.main:app --reload --host 0.0.0.0 --port 5050"
