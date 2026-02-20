#!/usr/bin/env bash

set -euo pipefail

HEALTH_URL="${HEALTH_URL:-}"
DRY_RUN="false"

usage() {
  cat <<'EOF'
Rollback rehearsal for shadow gate.

USAGE:
  rollback_rehearsal.sh [--health-url <url>] [--dry-run]

NOTES:
  - Calls manage_shadow_rule.sh --action disable.
  - Optionally checks Java stable health endpoint.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --health-url)
      HEALTH_URL="$2"
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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$DRY_RUN" == "true" ]]; then
  "$SCRIPT_DIR/manage_shadow_rule.sh" --action disable --dry-run
  if [[ -n "$HEALTH_URL" ]]; then
    echo "DRY-RUN health check: GET $HEALTH_URL"
  fi
  exit 0
fi

"$SCRIPT_DIR/manage_shadow_rule.sh" --action disable

if [[ -n "$HEALTH_URL" ]]; then
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required for health check" >&2
    exit 2
  fi
  code="$(curl -sS -o /tmp/shadow_rehearsal_health.out -w '%{http_code}' "$HEALTH_URL")"
  if [[ "$code" != "200" ]]; then
    echo "Health check failed with status $code" >&2
    exit 1
  fi
fi

echo "Rollback rehearsal completed successfully"
