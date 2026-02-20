#!/usr/bin/env bash

set -euo pipefail

OUT_FILE=""
MISMATCH_RATE="0"
SECURITY_INCIDENTS="0"
SEV1_COUNT="0"
SEV2_COUNT="0"

usage() {
  cat <<'EOF'
Test shadow alert policy with synthetic inputs.

USAGE:
  test_alert_policy.sh --out <path> [--mismatch-rate <percent>] [--security-incidents <n>] [--sev1 <n>] [--sev2 <n>]

EXIT CODE:
  0  no critical alerts triggered
  1  one or more critical alerts triggered
  2  invalid input
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)
      OUT_FILE="$2"
      shift 2
      ;;
    --mismatch-rate)
      MISMATCH_RATE="$2"
      shift 2
      ;;
    --security-incidents)
      SECURITY_INCIDENTS="$2"
      shift 2
      ;;
    --sev1)
      SEV1_COUNT="$2"
      shift 2
      ;;
    --sev2)
      SEV2_COUNT="$2"
      shift 2
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

alerts=()

if awk -v m="$MISMATCH_RATE" 'BEGIN { exit !(m >= 1.0) }'; then
  alerts+=("shadow_mismatch_rate_block")
fi

if [[ "$SECURITY_INCIDENTS" != "0" ]]; then
  alerts+=("shadow_security_incident_block")
fi

if [[ "$SEV1_COUNT" != "0" || "$SEV2_COUNT" != "0" ]]; then
  alerts+=("shadow_sev12_regression_block")
fi

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

{
  echo "# Shadow Alert Test Result"
  echo
  echo "- Timestamp: $ts"
  echo "- Input mismatch rate: $MISMATCH_RATE"
  echo "- Input security incidents: $SECURITY_INCIDENTS"
  echo "- Input sev1: $SEV1_COUNT"
  echo "- Input sev2: $SEV2_COUNT"
  echo
  echo "## Triggered Alerts"
  if [[ ${#alerts[@]} -eq 0 ]]; then
    echo "- none"
  else
    for a in "${alerts[@]}"; do
      echo "- $a"
    done
  fi
} > "$OUT_FILE"

echo "Alert test report: $OUT_FILE"

if [[ ${#alerts[@]} -eq 0 ]]; then
  exit 0
fi

exit 1
