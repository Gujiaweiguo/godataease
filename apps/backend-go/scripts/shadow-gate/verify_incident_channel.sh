#!/usr/bin/env bash

set -euo pipefail

WEBHOOK_URL=""
OUT_FILE=""
DRY_RUN="false"

usage() {
  cat <<'EOF'
Verify real incident channel visibility for SHADOW-002.

USAGE:
  verify_incident_channel.sh --webhook-url <url> --out <path> [--dry-run]

EXIT CODE:
  0  message delivered (or dry-run)
  1  delivery failed
  2  invalid input
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --webhook-url)
      WEBHOOK_URL="$2"
      shift 2
      ;;
    --out)
      OUT_FILE="$2"
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

if [[ -z "$OUT_FILE" ]]; then
  echo "--out is required" >&2
  exit 2
fi

if [[ "$DRY_RUN" == "false" && -z "$WEBHOOK_URL" ]]; then
  echo "--webhook-url is required unless --dry-run is used" >&2
  exit 2
fi

mkdir -p "$(dirname "$OUT_FILE")"

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
payload="{\"event\":\"shadow-alert-test\",\"timestamp\":\"${ts}\",\"severity\":\"critical\",\"message\":\"SHADOW-002 visibility test\"}"

status="dry-run"
http_code="000"
result_rc=0

if [[ "$DRY_RUN" == "false" ]]; then
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required" >&2
    exit 2
  fi
  set +e
  http_code="$(curl -sS -o /tmp/shadow_incident_channel.out -w '%{http_code}' -X POST "$WEBHOOK_URL" -H 'Content-Type: application/json' -d "$payload")"
  curl_rc=$?
  set -e
  if [[ "$curl_rc" -ne 0 ]]; then
    status="failed"
    result_rc=1
  elif [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
    status="delivered"
  else
    status="failed"
    result_rc=1
  fi
fi

{
  echo "# Incident Channel Visibility Test"
  echo
  echo "- Timestamp: $ts"
  echo "- Mode: $([[ "$DRY_RUN" == "true" ]] && echo dry-run || echo live)"
  echo "- Delivery Status: $status"
  echo "- HTTP Status: $http_code"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "- Webhook URL: n/a"
  else
    echo "- Webhook URL: $WEBHOOK_URL"
  fi
} > "$OUT_FILE"

echo "Incident visibility report: $OUT_FILE"
exit "$result_rc"
