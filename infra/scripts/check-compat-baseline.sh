#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ROLE_ID="${ROLE_ID:-1}"

PASS_COUNT=0
FAIL_COUNT=0

print_header() {
  printf "\n== Go Compatibility Baseline Check ==\n"
  printf "BASE_URL: %s\n" "$BASE_URL"
  printf "ROLE_ID : %s\n\n" "$ROLE_ID"
}

extract_code() {
  local body="$1"
  local code

  code=$(printf '%s' "$body" | sed -n 's/.*"code":"\([^"]*\)".*/\1/p')
  if [[ -z "$code" ]]; then
    code=$(printf '%s' "$body" | sed -n 's/.*"code":\([0-9][0-9]*\).*/\1/p')
  fi
  printf '%s' "$code"
}

check_endpoint() {
  local name="$1"
  local method="$2"
  local url="$3"
  local data="${4:-}"

  local body
  local http_code
  local api_code
  local tmp_body
  tmp_body="$(mktemp)"

  if [[ -n "$data" ]]; then
    http_code="$(curl -sS -X "$method" "$url" -H "Content-Type: application/json" -d "$data" -o "$tmp_body" -w "%{http_code}")"
  else
    http_code="$(curl -sS -X "$method" "$url" -o "$tmp_body" -w "%{http_code}")"
  fi

  body="$(cat "$tmp_body")"
  rm -f "$tmp_body"
  api_code="$(extract_code "$body")"

  if [[ "$http_code" == "200" && "$api_code" == "000000" ]]; then
    PASS_COUNT=$((PASS_COUNT + 1))
    printf "[PASS] %-36s http=%s code=%s\n" "$name" "$http_code" "$api_code"
  else
    FAIL_COUNT=$((FAIL_COUNT + 1))
    printf "[FAIL] %-36s http=%s code=%s\n" "$name" "$http_code" "${api_code:-N/A}"
    printf "       url=%s\n" "$url"
    printf "       body=%s\n" "$body"
  fi
}

print_summary() {
  printf "\n== Summary ==\n"
  printf "PASS: %d\n" "$PASS_COUNT"
  printf "FAIL: %d\n" "$FAIL_COUNT"
  if [[ "$FAIL_COUNT" -gt 0 ]]; then
    exit 1
  fi
}

print_header

check_endpoint "role.byCurOrg" "POST" "$BASE_URL/de2api/role/byCurOrg" '{}'
check_endpoint "auth.menuPermission" "GET" "$BASE_URL/de2api/auth/menuPermission"
check_endpoint "auth.busiPermission" "GET" "$BASE_URL/de2api/auth/busiPermission"
check_endpoint "role.permission.save" "POST" "$BASE_URL/de2api/role/permission/save" "{\"roleId\":$ROLE_ID,\"permIds\":[]}"
check_endpoint "system.role.permission.save" "POST" "$BASE_URL/de2api/system/role/permission/save" "{\"roleId\":$ROLE_ID,\"permIds\":[]}"
check_endpoint "dataVisualization.tree" "POST" "$BASE_URL/de2api/dataVisualization/tree" '{"busiFlag":"dashboard-dataV"}'

print_summary
