#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
REPORT_DIR="${REPORT_DIR:-./tmp/compat-checks}"
REPORT_FILE="$REPORT_DIR/auth-visualization-compat-report.txt"
SUCCESS_CODE="${SUCCESS_CODE:-000000}"
ROLE_TYPE_CODE="${ROLE_TYPE_CODE:-1}"
ADMIN_TOKEN=""

mkdir -p "$REPORT_DIR"
: > "$REPORT_FILE"

echo "[INFO] Base URL: $BASE_URL" | tee -a "$REPORT_FILE"
echo "[INFO] Strict success code: $SUCCESS_CODE" | tee -a "$REPORT_FILE"

TOTAL=0
FAILED=0

# Try login with default admin credentials
login_admin() {
  local tmp
  tmp=$(mktemp)
  local curl_rc=0
  local status

  status=$(curl -sS -o "$tmp" -w "%{http_code}" -X POST "$BASE_URL/login/localLogin" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"DataEase1234"}') || curl_rc=$?

  if [[ "$curl_rc" -ne 0 || "$status" == "000" ]]; then
    rm -f "$tmp"
    echo "[WARN] Admin login request failed, will proceed without auth" | tee -a "$REPORT_FILE"
    return 1
  fi

  if [[ "$status" != "200" ]]; then
    rm -f "$tmp"
    echo "[WARN] Admin login returned status $status, will proceed without auth" | tee -a "$REPORT_FILE"
    return 1
  fi

  # Extract token from response
  ADMIN_TOKEN=$(jq -r '.data.accessToken // .data.token // empty' "$tmp" 2>/dev/null || echo "")
  rm -f "$tmp"

  if [[ -z "$ADMIN_TOKEN" ]]; then
    echo "[WARN] Failed to extract token from login response, will proceed without auth" | tee -a "$REPORT_FILE"
    return 1
  fi

  echo "[INFO] Admin login successful, token obtained" | tee -a "$REPORT_FILE"
  return 0
}

post_json() {
  local path="$1"
  local body="$2"

  local tmp
  tmp=$(mktemp)
  local status
  local curl_rc=0

  # Build auth header if token is available
  local auth_args=()
  if [[ -n "$ADMIN_TOKEN" ]]; then
    auth_args=(-H "Authorization: Bearer $ADMIN_TOKEN")
  fi

  status=$(curl -sS -o "$tmp" -w "%{http_code}" -X POST "$BASE_URL$path" \
    -H "Content-Type: application/json" "${auth_args[@]}" \
    -d "$body") || curl_rc=$?

  if [[ "$curl_rc" -ne 0 || "$status" == "000" ]]; then
    rm -f "$tmp"
    return 1
  fi

  printf '%s|%s\n' "$status" "$tmp"
}

run_case_strict() {
  local name="$1"
  local path="$2"
  local body="$3"

  TOTAL=$((TOTAL + 1))

  local result
  if ! result=$(post_json "$path" "$body"); then
    echo "[FAIL] $name: request failed on $path" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    return
  fi

  local status="${result%%|*}"
  local tmp="${result#*|}"

  if [[ "$status" == "404" ]]; then
    echo "[FAIL] $name: returned 404 on $path" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  if ! jq -e '.code and .msg' "$tmp" >/dev/null 2>&1; then
    echo "[FAIL] $name: missing code/msg envelope on $path (status=$status, body=$(cat "$tmp"))" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  local code
  code=$(jq -r '.code // empty' "$tmp")
  if [[ "$code" != "$SUCCESS_CODE" ]]; then
    echo "[FAIL] $name: expected code=$SUCCESS_CODE got code=${code:-N/A} (status=$status, body=$(cat "$tmp"))" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return
  fi

  echo "[PASS] $name: strict success (status=$status, code=$code)" | tee -a "$REPORT_FILE"
  rm -f "$tmp"
}

create_role_strict() {
  local prefix="$1"
  local label="$2"
  local role_name="compat-check-${label}-$(date +%s)-$RANDOM"
  CREATED_ROLE_ID=""

  TOTAL=$((TOTAL + 1))
  local result
  if ! result=$(post_json "$prefix/system/role/create" "{\"name\":\"$role_name\",\"typeCode\":$ROLE_TYPE_CODE}"); then
    echo "[FAIL] $label system role create: request failed" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    return 1
  fi

  local status="${result%%|*}"
  local tmp="${result#*|}"
  if ! jq -e '.code and .msg' "$tmp" >/dev/null 2>&1; then
    echo "[FAIL] $label system role create: missing envelope (status=$status, body=$(cat "$tmp"))" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return 1
  fi

  local code
  code=$(jq -r '.code // empty' "$tmp")
  if [[ "$status" == "404" || "$code" != "$SUCCESS_CODE" ]]; then
    echo "[FAIL] $label system role create: expected code=$SUCCESS_CODE got status=$status code=${code:-N/A} body=$(cat "$tmp")" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return 1
  fi

  local role_id
  role_id=$(jq -r '.data // empty' "$tmp")
  if [[ -z "$role_id" || ! "$role_id" =~ ^[0-9]+$ || "$role_id" -le 0 ]]; then
    echo "[FAIL] $label system role create: invalid role id (data=$(jq -r '.data' "$tmp"))" | tee -a "$REPORT_FILE"
    FAILED=$((FAILED + 1))
    rm -f "$tmp"
    return 1
  fi

  echo "[PASS] $label system role create: strict success (status=$status, code=$code, roleId=$role_id)" | tee -a "$REPORT_FILE"
  rm -f "$tmp"
  CREATED_ROLE_ID="$role_id"
  return 0
}

run_flow() {
  local prefix="$1"
  local label="$2"

  local role_id
  if ! create_role_strict "$prefix" "$label"; then
    return
  fi
  role_id="$CREATED_ROLE_ID"

  run_case_strict "$label auth menuPermission" "$prefix/auth/menuPermission" "{\"roleId\":$role_id}"
  run_case_strict "$label auth busiPermission" "$prefix/auth/busiPermission" "{\"roleId\":$role_id}"
  run_case_strict "$label auth saveMenuPer" "$prefix/auth/saveMenuPer" "{\"roleId\":$role_id,\"menuIds\":[]}"
  run_case_strict "$label auth saveBusiPer" "$prefix/auth/saveBusiPer" "{\"roleId\":$role_id,\"permIds\":[]}"
  run_case_strict "$label system role permission save" "$prefix/system/role/permission/save" "{\"roleId\":$role_id,\"permIds\":[]}"
  run_case_strict "$label system role update" "$prefix/system/role/update" "{\"id\":$role_id,\"name\":\"compat-check-updated-$role_id\"}"
  run_case_strict "$label visualization tree" "$prefix/dataVisualization/tree" '{"busiFlag":"dashboard-dataV","leaf":false}'
  run_case_strict "$label system role delete" "$prefix/system/role/delete/$role_id" '{}'
}

# Try to login first
login_admin || true

run_flow "/api" "api"
run_flow "/de2api" "de2api"

if [[ "$FAILED" -gt 0 ]]; then
  echo "[FAIL] Compatibility checks failed: $FAILED/$TOTAL" | tee -a "$REPORT_FILE"
  exit 1
fi

echo "[PASS] Compatibility checks passed: $TOTAL/$TOTAL" | tee -a "$REPORT_FILE"
