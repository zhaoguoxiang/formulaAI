#!/usr/bin/env bash
# =============================================================================
# Formula AI System - End-to-End Integration Smoke Test
# =============================================================================
# This script verifies the full system is operational:
#   1. Docker compose services are healthy
#   2. Seed data is loaded (3 formulas)
#   3. CRUD operations work via API
#   4. Matrix API returns structured data
#   5. Analysis endpoints return statistics
#   6. Frontend serves HTML
#
# Usage:
#   ./scripts/e2e-test.sh [BASE_URL_API] [BASE_URL_FRONTEND]
#   Default: API=http://localhost:8080  Frontend=http://localhost:8081
# =============================================================================

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
API_BASE="${1:-http://localhost:8080}"
FRONTEND_BASE="${2:-http://localhost:8081}"
EVIDENCE_DIR=".omo/evidence/e2e"
TIMESTAMP="$(date -u +%Y%m%d-%H%M%S)"
LOG_FILE="${EVIDENCE_DIR}/e2e-test-${TIMESTAMP}.log"

PASS=0
FAIL=0

mkdir -p "$EVIDENCE_DIR"

# ── Helpers ───────────────────────────────────────────────────────────────────
green()  { printf '\033[32m%s\033[0m\n' "$*"; }
red()    { printf '\033[31m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
log()    { printf '%s\n' "$*" | tee -a "$LOG_FILE"; }
log_n()  { printf '%s' "$*" | tee -a "$LOG_FILE"; }

pass() { PASS=$((PASS + 1)); green "  ✓ PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); red   "  ✗ FAIL: $1"; }

check_http() {
    local desc="$1"; local method="$2"; local url="$3"
    local expected="${4:-200}"
    local data="${5:-}"
    local extra_opts=""

    if [ -n "$data" ]; then
        local http_code
        http_code=$(curl -s -o /tmp/e2e_resp_body -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" -d "$data" "$url" 2>/tmp/e2e_curl_err)
    else
        local http_code
        http_code=$(curl -s -o /tmp/e2e_resp_body -w "%{http_code}" -X "$method" "$url" 2>/tmp/e2e_curl_err)
    fi

    if [ "$http_code" = "$expected" ]; then
        pass "$desc (HTTP $http_code)"
        return 0
    else
        local err_body
        err_body=$(head -c 200 /tmp/e2e_resp_body)
        fail "$desc (expected $expected, got $http_code: $err_body)"
        return 1
    fi
}

# ── Phase 1: Health Checks ────────────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 1: Health Checks"
log "══════════════════════════════════════════════════════════════════"

# Docker compose health
log ""
log "--- Docker compose services ---"
docker ps --filter "name=formula" --format "  {{.Names}}: {{.Status}}" 2>/dev/null | tee -a "$LOG_FILE"

if docker ps --filter "name=formula-postgres" --filter "health=healthy" --format "ok" 2>/dev/null | grep -q ok; then
    pass "PostgreSQL is healthy"
else
    fail "PostgreSQL not healthy"
fi

check_http "Backend health endpoint" GET "${API_BASE}/api/health"

check_http "Frontend health endpoint" GET "${FRONTEND_BASE}/health"

# ── Phase 2: Seed Data Verification ───────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 2: Seed Data Verification"
log "══════════════════════════════════════════════════════════════════"

LIST_RESP=$(curl -s "${API_BASE}/api/formulas")
FORMULA_COUNT=$(echo "$LIST_RESP" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")

log_n "  Formula count: ${FORMULA_COUNT}"
if [ "$FORMULA_COUNT" -ge 3 ]; then
    green " (≥3 expected)"
    pass "Seed data: at least 3 formulas found ($FORMULA_COUNT)"
else
    red " (<3 expected)"
    fail "Seed data: expected ≥3 formulas, got $FORMULA_COUNT"
fi

# Verify specific seed codes exist
SEED_CODES=$(echo "$LIST_RESP" | python3 -c "
import sys,json
codes=[f['code'] for f in json.load(sys.stdin)]
print('|'.join(codes))
" 2>/dev/null)

for code in "FML-SL-001" "FML-EP-001" "FML-CT-001"; do
    if echo "$SEED_CODES" | grep -q "$code"; then
        pass "Seed formula $code present"
    else
        fail "Seed formula $code missing"
    fi
done

# ── Phase 3: CRUD Operations ──────────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 3: CRUD Operations"
log "══════════════════════════════════════════════════════════════════"

# CREATE
log "--- CREATE 4th formula ---"
CREATE_PAYLOAD='{
    "name": "UV-Curable Inkjet Ink",
    "code": "FML-UV-001",
    "component_mode": "single",
    "status": "draft",
    "parts": [
      {
        "name": "PartMain",
        "mix_ratio": 100,
        "sort_order": 1,
        "ingredients": [
          {"sort_order": 1, "material": "UV Monomer Blend", "percentage": 40,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000001", "dosing_order": 1, "use_ratio": 100, "dosing_method": "metered"}]},
          {"sort_order": 2, "material": "Photoinitiator", "percentage": 8,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000001", "dosing_order": 2, "use_ratio": 100, "dosing_method": "manual"}]},
          {"sort_order": 3, "material": "Pigment Dispersion", "percentage": 30,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000002", "dosing_order": 1, "use_ratio": 100, "dosing_method": "gravimetric"}]},
          {"sort_order": 4, "material": "Acrylate Oligomer", "percentage": 15,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000002", "dosing_order": 2, "use_ratio": 100, "dosing_method": "metered"}]},
          {"sort_order": 5, "material": "Adhesion Promoter", "percentage": 5,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000001", "dosing_order": 3, "use_ratio": 100, "dosing_method": "manual"}]},
          {"sort_order": 6, "material": "Wetting Agent", "percentage": 2,
            "dosing_actions": [{"step_id": "00000000-0000-0000-0000-e2e000000003", "dosing_order": 1, "use_ratio": 100, "dosing_method": "manual"}]}
        ]
      }
    ],
    "steps": [
      {"id": "00000000-0000-0000-0000-e2e000000001", "step_no": 1, "name": "Raw Material Charging", "temperature": "25°C", "duration": "30 min"},
      {"id": "00000000-0000-0000-0000-e2e000000002", "step_no": 2, "name": "High-Speed Dispersion", "temperature": "30°C", "duration": "45 min"},
      {"id": "00000000-0000-0000-0000-e2e000000003", "step_no": 3, "name": "Filtration & QC", "temperature": "25°C", "duration": "15 min"}
    ]
}'

check_http "CREATE formula (POST)" POST "${API_BASE}/api/formulas" "201" "$CREATE_PAYLOAD"

# Extract the new formula ID from the response body saved by check_http
NEW_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" < /tmp/e2e_resp_body 2>/dev/null || echo "")

if [ -z "$NEW_ID" ]; then
    fail "Could not extract formula ID from CREATE response"
else
    pass "Extracted created formula ID: $NEW_ID"
fi

# READ single
if [ -n "$NEW_ID" ]; then
    check_http "READ formula by ID (GET)" GET "${API_BASE}/api/formulas/${NEW_ID}"
fi

# READ list (verify count increased)
NEW_COUNT=$(curl -s "${API_BASE}/api/formulas" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
if [ "$NEW_COUNT" -eq 4 ]; then
    pass "READ list: 4 formulas after create (was 3)"
else
    fail "READ list: expected 4, got $NEW_COUNT"
fi

# UPDATE
if [ -n "$NEW_ID" ]; then
    log "--- UPDATE formula ---"
    UPDATE_PAYLOAD=$(python3 -c "
import json
data = json.load(open('/tmp/e2e_resp_body'))
data['status'] = 'active'
data['name'] = 'UV-Curable Inkjet Ink (Updated)'
print(json.dumps(data))
" 2>/dev/null)

    check_http "UPDATE formula (PUT)" PUT "${API_BASE}/api/formulas/${NEW_ID}" "200" "$UPDATE_PAYLOAD"

    # Verify update
    UPDATED_STATUS=$(curl -s "${API_BASE}/api/formulas/${NEW_ID}" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null)
    if [ "$UPDATED_STATUS" = "active" ]; then
        pass "UPDATE verified: status changed to active"
    else
        fail "UPDATE verification: expected status=active, got=$UPDATED_STATUS"
    fi
fi

# DELETE
if [ -n "$NEW_ID" ]; then
    log "--- DELETE formula ---"
    check_http "DELETE formula" DELETE "${API_BASE}/api/formulas/${NEW_ID}" "204"

    # Verify gone
    check_http "DELETE verified (404 on re-fetch)" GET "${API_BASE}/api/formulas/${NEW_ID}" "404"

    # Verify count back to 3
    FINAL_COUNT=$(curl -s "${API_BASE}/api/formulas" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    if [ "$FINAL_COUNT" -eq 3 ]; then
        pass "DELETE cleanup: back to 3 formulas"
    else
        fail "DELETE cleanup: expected 3, got $FINAL_COUNT"
    fi
fi

# ── Phase 4: Matrix API ───────────────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 4: Matrix API"
log "══════════════════════════════════════════════════════════════════"

check_http "Matrix API (GET /api/formulas/matrix)" GET "${API_BASE}/api/formulas/matrix"

MATRIX_COUNT=$(curl -s "${API_BASE}/api/formulas/matrix" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('formulas',[])))" 2>/dev/null || echo "0")
if [ "$MATRIX_COUNT" -eq 3 ]; then
    pass "Matrix API: 3 formulas with restructured data"
else
    fail "Matrix API: expected 3, got $MATRIX_COUNT"
fi

# Check matrix structure has parts and steps
MATRIX_STRUCT=$(curl -s "${API_BASE}/api/formulas/matrix" | python3 -c "
import sys,json
data = json.load(sys.stdin)
f = data['formulas'][0]
print(f'parts:{len(f.get(\"parts\",[]))}_steps:{len(f.get(\"steps\",[]))}')
" 2>/dev/null)
log "  Matrix structure sample: $MATRIX_STRUCT"

# Filter test
check_http "Matrix filter (component_mode=single)" GET "${API_BASE}/api/formulas/matrix?component_mode=single"
SINGLE_COUNT=$(curl -s "${API_BASE}/api/formulas/matrix?component_mode=single" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('formulas',[])))" 2>/dev/null || echo "0")
if [ "$SINGLE_COUNT" -eq 2 ]; then
    pass "Matrix filter: 2 single-component formulas"
else
    fail "Matrix filter: expected 2 single, got $SINGLE_COUNT"
fi

# ── Phase 5: Analysis Endpoints ───────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 5: Analysis Endpoints"
log "══════════════════════════════════════════════════════════════════"

check_http "Analysis: ingredient-distribution" GET "${API_BASE}/api/analysis/ingredient-distribution"
check_http "Analysis: component-mode-ratio" GET "${API_BASE}/api/analysis/component-mode-ratio"
check_http "Analysis: step-count-distribution" GET "${API_BASE}/api/analysis/step-count-distribution"
check_http "Analysis: dosing-method-stats" GET "${API_BASE}/api/analysis/dosing-method-stats"

# Verify non-empty responses
for ep in "ingredient-distribution" "component-mode-ratio" "step-count-distribution" "dosing-method-stats"; do
    COUNT=$(curl -s "${API_BASE}/api/analysis/${ep}" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    if [ "$COUNT" -gt 0 ]; then
        pass "Analysis $ep: $COUNT result(s)"
    else
        fail "Analysis $ep: empty result"
    fi
done

# ── Phase 6: Frontend Verification ────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "Phase 6: Frontend Verification"
log "══════════════════════════════════════════════════════════════════"

# Check frontend serves index.html
HTTP_FE=$(curl -s -o /tmp/e2e_frontend -w "%{http_code}" "${FRONTEND_BASE}/" 2>/dev/null)
if [ "$HTTP_FE" = "200" ]; then
    pass "Frontend index.html loads (HTTP 200)"
else
    fail "Frontend index.html returned $HTTP_FE"
fi

# Check for Angular app-root
if grep -q 'app-root' /tmp/e2e_frontend 2>/dev/null; then
    pass "Frontend contains Angular <app-root>"
else
    fail "Frontend missing <app-root> element"
fi

# Check for title
if grep -q 'FormulAI' /tmp/e2e_frontend 2>/dev/null; then
    pass "Frontend has FormulAI title"
else
    fail "Frontend missing FormulAI title"
fi

# Check frontend health endpoint
check_http "Frontend health endpoint" GET "${FRONTEND_BASE}/health"

# ── Summary ───────────────────────────────────────────────────────────────────
log ""
log "══════════════════════════════════════════════════════════════════"
log "E2E TEST SUMMARY"
log "══════════════════════════════════════════════════════════════════"
log "  PASSED: ${PASS}"
log "  FAILED: ${FAIL}"
log ""
log "  Log:    ${LOG_FILE}"
log "══════════════════════════════════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
    red "E2E TESTS FAILED: ${FAIL} failure(s)"
    exit 1
else
    green "E2E TESTS ALL PASSED: ${PASS} tests"
    exit 0
fi
