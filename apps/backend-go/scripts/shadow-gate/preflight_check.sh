#!/usr/bin/env bash

set -euo pipefail

DEFAULT_OUT_DIR="./tmp/shadow-gate"
DEFAULT_WHITELIST="./testdata/contract-diff/critical-whitelist.yaml"
DEFAULT_TOOL_BIN_DIR="./tmp/shadow-gate/tools/bin"
DEFAULT_MIGRATION_MODE="compatibility"

OUT_DIR="${DEFAULT_OUT_DIR}"
WHITELIST="${DEFAULT_WHITELIST}"
TOOL_BIN_DIR="${DEFAULT_TOOL_BIN_DIR}"
MIGRATION_MODE="${DEFAULT_MIGRATION_MODE}"
SKIP_SHADOW_CONTROL="false"

usage() {
  cat <<'EOF'
Shadow Gate Preflight Check

USAGE:
  preflight_check.sh [--out-dir <path>] [--whitelist <path>] [--tool-bin-dir <path>] [--migration-mode <compatibility|go-only>] [--skip-shadow-control]

OPTIONS:
  --out-dir <path>                 Output directory for reports
  --whitelist <path>               Critical API whitelist file path
  --tool-bin-dir <path>            Fallback tooling directory for kubectl/helm
  --migration-mode <mode>          compatibility (default) or go-only
  --skip-shadow-control            In go-only mode, mark gateway/rule/token as n/a
  -h, --help                       Show help

OUTPUT:
  - preflight.json
  - preflight.md

EXIT CODES:
  0  Preflight ready (no blocking items)
  1  Preflight blocked
  2  Invalid arguments
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out-dir)
      OUT_DIR="$2"
      shift 2
      ;;
    --whitelist)
      WHITELIST="$2"
      shift 2
      ;;
    --tool-bin-dir)
      TOOL_BIN_DIR="$2"
      shift 2
      ;;
    --migration-mode)
      MIGRATION_MODE="$2"
      shift 2
      ;;
    --skip-shadow-control)
      SKIP_SHADOW_CONTROL="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ "$MIGRATION_MODE" != "compatibility" && "$MIGRATION_MODE" != "go-only" ]]; then
  echo "--migration-mode must be compatibility or go-only" >&2
  exit 2
fi

if [[ "$SKIP_SHADOW_CONTROL" == "true" && "$MIGRATION_MODE" != "go-only" ]]; then
  echo "--skip-shadow-control is only allowed in go-only mode" >&2
  exit 2
fi

mkdir -p "$OUT_DIR"

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
json_file="$OUT_DIR/preflight.json"
md_file="$OUT_DIR/preflight.md"

is_set() {
  [[ -n "${!1:-}" ]]
}

command_state() {
  local name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    echo "ready"
  else
    echo "blocked"
  fi
}

command_or_fallback_state() {
  local name="$1"
  local fallback_bin="$2"
  if command -v "$name" >/dev/null 2>&1; then
    echo "ready"
    return
  fi
  if [[ -x "$fallback_bin" ]]; then
    echo "ready"
    return
  fi
  echo "blocked"
}

env_state() {
  local name="$1"
  if is_set "$name"; then
    local value="${!name}"
    if [[ "$value" == *"replace-with"* || "$value" == *"example.internal"* ]]; then
      echo "blocked"
      return
    fi
    echo "ready"
  else
    echo "blocked"
  fi
}

file_state() {
  local path="$1"
  if [[ -f "$path" ]]; then
    echo "ready"
  else
    echo "blocked"
  fi
}

jq_state="$(command_state jq)"
curl_state="$(command_state curl)"
kubectl_state="$(command_or_fallback_state kubectl "$TOOL_BIN_DIR/kubectl")"
helm_state="$(command_or_fallback_state helm "$TOOL_BIN_DIR/helm")"

java_base_state="$(env_state JAVA_BASE_URL)"
go_base_state="$(env_state GO_BASE_URL)"
gateway_state="$(env_state STAGING_GATEWAY_API)"
rule_state="$(env_state SHADOW_RULE_ID)"
token_state="$(env_state AUTH_TOKEN)"

if [[ "$MIGRATION_MODE" == "go-only" ]]; then
  java_base_state="n/a"
  if [[ "$SKIP_SHADOW_CONTROL" == "true" ]]; then
    go_base_state="n/a"
    gateway_state="n/a"
    rule_state="n/a"
    token_state="n/a"
  fi
fi

whitelist_state="$(file_state "$WHITELIST")"

total_blockers=0
for state in \
  "$jq_state" "$curl_state" "$kubectl_state" "$helm_state" \
  "$go_base_state" "$gateway_state" "$rule_state" "$token_state" \
  "$whitelist_state"; do
  if [[ "$state" == "blocked" ]]; then
    total_blockers=$((total_blockers + 1))
  fi
done

if [[ "$MIGRATION_MODE" == "compatibility" && "$java_base_state" == "blocked" ]]; then
  total_blockers=$((total_blockers + 1))
fi

overall_status="ready"
if [[ "$total_blockers" -gt 0 ]]; then
  overall_status="blocked"
fi

jq -n \
  --arg timestamp "$timestamp" \
  --arg status "$overall_status" \
  --arg migration_mode "$MIGRATION_MODE" \
  --arg skip_shadow_control "$SKIP_SHADOW_CONTROL" \
  --arg whitelist "$WHITELIST" \
  --arg tool_bin_dir "$TOOL_BIN_DIR" \
  --arg jq_state "$jq_state" \
  --arg curl_state "$curl_state" \
  --arg kubectl_state "$kubectl_state" \
  --arg helm_state "$helm_state" \
  --arg java_base_state "$java_base_state" \
  --arg go_base_state "$go_base_state" \
  --arg gateway_state "$gateway_state" \
  --arg rule_state "$rule_state" \
  --arg token_state "$token_state" \
  --arg whitelist_state "$whitelist_state" \
  --argjson blockers "$total_blockers" \
  '{
    timestamp: $timestamp,
    status: $status,
    migrationMode: $migration_mode,
    skipShadowControl: ($skip_shadow_control == "true"),
    blockers: $blockers,
    checks: {
      commands: {
        jq: $jq_state,
        curl: $curl_state,
        kubectl: $kubectl_state,
        helm: $helm_state,
        fallbackToolBinDir: $tool_bin_dir
      },
      environment: {
        JAVA_BASE_URL: $java_base_state,
        GO_BASE_URL: $go_base_state,
        STAGING_GATEWAY_API: $gateway_state,
        SHADOW_RULE_ID: $rule_state,
        AUTH_TOKEN: $token_state
      },
      files: {
        whitelist: {
          path: $whitelist,
          state: $whitelist_state
        }
      }
    }
  }' >"$json_file"

{
  echo "# Shadow Gate Preflight Report"
  echo
  echo "- Timestamp: $timestamp"
  echo "- Overall Status: $overall_status"
  echo "- Migration Mode: $MIGRATION_MODE"
  echo "- Skip Shadow Control: $SKIP_SHADOW_CONTROL"
  echo "- Blocking Items: $total_blockers"
  echo
  echo "## Command Checks"
  echo
  echo "| Item | State |"
  echo "|------|-------|"
  echo "| jq | $jq_state |"
  echo "| curl | $curl_state |"
  echo "| kubectl | $kubectl_state |"
  echo "| helm | $helm_state |"
  echo
  echo "## Environment Checks"
  echo
  echo "| Variable | State |"
  echo "|----------|-------|"
  echo "| JAVA_BASE_URL | $java_base_state |"
  echo "| GO_BASE_URL | $go_base_state |"
  echo "| STAGING_GATEWAY_API | $gateway_state |"
  echo "| SHADOW_RULE_ID | $rule_state |"
  echo "| AUTH_TOKEN | $token_state |"
  echo
  echo "## File Checks"
  echo
  echo "| File | State |"
  echo "|------|-------|"
  echo "| $WHITELIST | $whitelist_state |"
} >"$md_file"

echo "Preflight JSON: $json_file"
echo "Preflight Markdown: $md_file"

if [[ "$overall_status" == "blocked" ]]; then
  exit 1
fi

exit 0
