#!/usr/bin/env bash

set -euo pipefail

ACTION=""
GATEWAY_API="${STAGING_GATEWAY_API:-}"
RULE_ID="${SHADOW_RULE_ID:-}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
RATIO="0.05"
DRY_RUN="false"

usage() {
  cat <<'EOF'
Manage staging shadow rule through gateway API.

USAGE:
  manage_shadow_rule.sh --action <enable|disable> [--gateway-api <url>] [--rule-id <id>] [--auth-token <token>] [--ratio <0..1>] [--dry-run]

OPTIONS:
  --action <enable|disable>  Required action
  --gateway-api <url>        Gateway API base URL (default from STAGING_GATEWAY_API)
  --rule-id <id>             Shadow rule ID (default from SHADOW_RULE_ID)
  --auth-token <token>       Bearer token (default from AUTH_TOKEN)
  --ratio <0..1>             Shadow ratio when enabling (default 0.05)
  --dry-run                  Print request only
  -h, --help                 Show help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --action)
      ACTION="$2"
      shift 2
      ;;
    --gateway-api)
      GATEWAY_API="$2"
      shift 2
      ;;
    --rule-id)
      RULE_ID="$2"
      shift 2
      ;;
    --auth-token)
      AUTH_TOKEN="$2"
      shift 2
      ;;
    --ratio)
      RATIO="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN="true"
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

if [[ -z "$ACTION" || ( "$ACTION" != "enable" && "$ACTION" != "disable" ) ]]; then
  echo "--action must be enable or disable" >&2
  exit 2
fi

if [[ "$DRY_RUN" == "false" ]]; then
  if [[ -z "$GATEWAY_API" || -z "$RULE_ID" || -z "$AUTH_TOKEN" ]]; then
    echo "gateway-api, rule-id, and auth-token are required" >&2
    exit 2
  fi
fi

if [[ -z "$GATEWAY_API" ]]; then
  GATEWAY_API="<missing-gateway-api>"
fi
if [[ -z "$RULE_ID" ]]; then
  RULE_ID="<missing-rule-id>"
fi
if [[ -z "$AUTH_TOKEN" ]]; then
  AUTH_TOKEN="<missing-auth-token>"
fi

payload='{"shadow":{"enabled":false,"ratio":0.0}}'
if [[ "$ACTION" == "enable" ]]; then
  payload="{\"shadow\":{\"enabled\":true,\"ratio\":${RATIO}}}"
fi

url="${GATEWAY_API%/}/rules/${RULE_ID}"

if [[ "$DRY_RUN" == "true" ]]; then
  echo "DRY-RUN request: PATCH $url"
  echo "Payload: $payload"
  exit 0
fi

response="$(curl -sS -X PATCH "$url" -H "Authorization: Bearer $AUTH_TOKEN" -H "Content-Type: application/json" -d "$payload")"
echo "$response"
