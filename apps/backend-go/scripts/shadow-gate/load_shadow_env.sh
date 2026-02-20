#!/usr/bin/env bash

set -euo pipefail

ENV_FILE="./.env.shadow-gate"
MIGRATION_MODE="compatibility"
SKIP_SHADOW_CONTROL="false"

usage() {
  cat <<'EOF'
Load and validate shadow gate environment variables.

USAGE:
  load_shadow_env.sh [--env-file <path>] [--migration-mode <compatibility|go-only>] [--skip-shadow-control]

REQUIRED (compatibility):
  JAVA_BASE_URL, GO_BASE_URL, STAGING_GATEWAY_API, SHADOW_RULE_ID, AUTH_TOKEN

REQUIRED (go-only):
  GO_BASE_URL, STAGING_GATEWAY_API, SHADOW_RULE_ID, AUTH_TOKEN

OPTIONAL (go-only + --skip-shadow-control):
  GO_BASE_URL
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env-file)
      ENV_FILE="$2"
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

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 2
fi

set -a
source "$ENV_FILE"
set +a

required=(GO_BASE_URL STAGING_GATEWAY_API SHADOW_RULE_ID AUTH_TOKEN)
if [[ "$MIGRATION_MODE" == "compatibility" ]]; then
  required+=(JAVA_BASE_URL)
fi
if [[ "$MIGRATION_MODE" == "go-only" && "$SKIP_SHADOW_CONTROL" == "true" ]]; then
  required=()
fi

missing=()
invalid=()
for key in "${required[@]}"; do
  value="${!key:-}"
  if [[ -z "$value" ]]; then
    missing+=("$key")
    continue
  fi
  if [[ "$value" == *"replace-with"* || "$value" == *"example.internal"* ]]; then
    invalid+=("$key")
  fi
done

if [[ ${#missing[@]} -gt 0 ]]; then
  echo "missing required variables: ${missing[*]}" >&2
  exit 1
fi

if [[ ${#invalid[@]} -gt 0 ]]; then
  echo "placeholder values are not allowed for: ${invalid[*]}" >&2
  exit 1
fi

echo "Shadow env loaded from $ENV_FILE (mode=$MIGRATION_MODE, skipShadowControl=$SKIP_SHADOW_CONTROL)"
