#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
REPORT_DIR="${REPORT_DIR:-./tmp/compat-checks}"
REPORT_FILE="$REPORT_DIR/auth-visualization-compat-report.txt"

mkdir -p "$REPORT_DIR"
: > "$REPORT_FILE"

echo "[INFO] Base URL: $BASE_URL" | tee -a "$REPORT_FILE"

TOTAL=0
FAILED=0

run_case() {
  local name="$1"
  local path="$2"
  local body="$3"

  TOTAL=$((TOTAL + 1))

  local tmp
  tmp=$(mktemp)
  local status
  local curl_rc=0
  status=$(curl -sS -o "$tmp" -w "%{http_code}" -X POST "$BASE_URL$path" -H "Content-Type: application/json" -d "$body") || curl_rc=$?

  if [[ "$curl_rc" -ne 0 || "$status" == "000" ]]; then
    echo "[FAIL] $name: request failed on $path (curl_rc=$curl_rc, status=$status)" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  if [[ "$status" == "404" ]]; then
    echo "[FAIL] $name: returned 404 on $path" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  if ! jq -e '.code and .msg' "$tmp" >/dev/null 2>&1; then
    echo "[FAIL] $name: missing code/msg envelope on $path (status=$status)" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  echo "[PASS] $name: non-404 and envelope ok (status=$status, code=$(jq -r '.code' "$tmp"))" | tee -a "$REPORT_FILE"
  rm -f "$tmp"
}

run_case "api auth menuPermission" "/api/auth/menuPermission" '{"roleId":1}'
run_case "api auth busiPermission" "/api/auth/busiPermission" '{"roleId":1}'
run_case "api auth saveMenuPer" "/api/auth/saveMenuPer" '{"roleId":1,"menuIds":[]}'
run_case "api auth saveBusiPer" "/api/auth/saveBusiPer" '{"roleId":1,"permIds":[]}'
run_case "api system role permission save" "/api/system/role/permission/save" '{"roleId":1,"permIds":[]}'
run_case "api system role create" "/api/system/role/create" '{"name":"compat-check-role"}'
run_case "api system role update" "/api/system/role/update" '{"id":1,"name":"compat-check-role-updated"}'
run_case "api system role delete" "/api/system/role/delete/1" '{}'
run_case "api visualization tree" "/api/dataVisualization/tree" '{"busiFlag":"dashboard-dataV","leaf":false}'

run_case "de2api auth menuPermission" "/de2api/auth/menuPermission" '{"roleId":1}'
run_case "de2api auth busiPermission" "/de2api/auth/busiPermission" '{"roleId":1}'
run_case "de2api auth saveMenuPer" "/de2api/auth/saveMenuPer" '{"roleId":1,"menuIds":[]}'
run_case "de2api auth saveBusiPer" "/de2api/auth/saveBusiPer" '{"roleId":1,"permIds":[]}'
run_case "de2api system role permission save" "/de2api/system/role/permission/save" '{"roleId":1,"permIds":[]}'
run_case "de2api system role create" "/de2api/system/role/create" '{"name":"compat-check-role"}'
run_case "de2api system role update" "/de2api/system/role/update" '{"id":1,"name":"compat-check-role-updated"}'
run_case "de2api system role delete" "/de2api/system/role/delete/1" '{}'
run_case "de2api visualization tree" "/de2api/dataVisualization/tree" '{"busiFlag":"dashboard-dataV","leaf":false}'

if [[ "$FAILED" -gt 0 ]]; then
  echo "[FAIL] Compatibility checks failed: $FAILED/$TOTAL" | tee -a "$REPORT_FILE"
  exit 1
fi

echo "[PASS] Compatibility checks passed: $TOTAL/$TOTAL" | tee -a "$REPORT_FILE"
