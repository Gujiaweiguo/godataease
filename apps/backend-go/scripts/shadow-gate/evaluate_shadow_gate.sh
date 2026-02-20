#!/usr/bin/env bash

set -euo pipefail

CONTRACT_REPORT=""
MISMATCH_RATE=""
OUT_FILE=""
MAX_MISMATCH="1"
SECURITY_INCIDENTS="0"
SEV1_COUNT="0"
SEV2_COUNT="0"

usage() {
  cat <<'EOF'
Evaluate Go/No-Go from shadow gate evidence.

USAGE:
  evaluate_shadow_gate.sh (--contract-report <path> | --mismatch-rate <percent>) --out <markdown> [--max-mismatch <percent>] [--security-incidents <n>] [--sev1 <n>] [--sev2 <n>]

RULES:
  - mismatch rate must be < max-mismatch (default 1)
  - security incidents must be 0
  - sev1 and sev2 must be 0

EXIT CODE:
  0  GO
  1  NO-GO
  2  invalid input
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --contract-report)
      CONTRACT_REPORT="$2"
      shift 2
      ;;
    --mismatch-rate)
      MISMATCH_RATE="$2"
      shift 2
      ;;
    --out)
      OUT_FILE="$2"
      shift 2
      ;;
    --max-mismatch)
      MAX_MISMATCH="$2"
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

if [[ -n "$CONTRACT_REPORT" && -n "$MISMATCH_RATE" ]]; then
  echo "use either --contract-report or --mismatch-rate, not both" >&2
  exit 2
fi

if [[ -z "$CONTRACT_REPORT" && -z "$MISMATCH_RATE" ]]; then
  echo "either --contract-report or --mismatch-rate is required" >&2
  exit 2
fi

evidence_source=""
parity="n/a"
mismatch_rate=""

if [[ -n "$CONTRACT_REPORT" ]]; then
  if [[ ! -f "$CONTRACT_REPORT" ]]; then
    echo "contract report not found: $CONTRACT_REPORT" >&2
    exit 2
  fi
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required" >&2
    exit 2
  fi
  parity="$(jq -r '.summary.parity // empty' "$CONTRACT_REPORT")"
  if [[ -z "$parity" ]]; then
    echo "invalid contract report: missing summary.parity" >&2
    exit 2
  fi
  mismatch_rate="$(awk -v p="$parity" 'BEGIN { printf "%.2f", (100 - p) }')"
  evidence_source="contract-report"
else
  if ! awk -v m="$MISMATCH_RATE" 'BEGIN { exit !(m ~ /^[0-9]+(\.[0-9]+)?$/) }'; then
    echo "invalid --mismatch-rate: $MISMATCH_RATE" >&2
    exit 2
  fi
  mismatch_rate="$MISMATCH_RATE"
  parity="$(awk -v m="$mismatch_rate" 'BEGIN { printf "%.2f", (100 - m) }')"
  evidence_source="observability-metric"
fi

decision="GO"
reasons=()

if awk -v m="$mismatch_rate" -v t="$MAX_MISMATCH" 'BEGIN { exit !(m < t) }'; then
  :
else
  decision="NO-GO"
  reasons+=("Mismatch rate ${mismatch_rate}% is not below threshold ${MAX_MISMATCH}%")
fi

if [[ "$SECURITY_INCIDENTS" != "0" ]]; then
  decision="NO-GO"
  reasons+=("Critical security incidents must be 0, got ${SECURITY_INCIDENTS}")
fi

if [[ "$SEV1_COUNT" != "0" || "$SEV2_COUNT" != "0" ]]; then
  decision="NO-GO"
  reasons+=("Sev-1/Sev-2 regressions must be 0, got sev1=${SEV1_COUNT}, sev2=${SEV2_COUNT}")
fi

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

{
  echo "# Shadow Gate Decision"
  echo
  echo "- Timestamp: ${ts}"
  echo "- Decision: ${decision}"
  echo "- Evidence source: ${evidence_source}"
  echo "- Contract parity: ${parity}%"
  echo "- Mismatch rate: ${mismatch_rate}%"
  echo "- Threshold (max mismatch): < ${MAX_MISMATCH}%"
  echo "- Security incidents: ${SECURITY_INCIDENTS}"
  echo "- Sev-1 regressions: ${SEV1_COUNT}"
  echo "- Sev-2 regressions: ${SEV2_COUNT}"
  echo
  echo "## Decision Reasons"
  if [[ ${#reasons[@]} -eq 0 ]]; then
    echo "- All gate criteria satisfied."
  else
    for r in "${reasons[@]}"; do
      echo "- ${r}"
    done
  fi
} > "$OUT_FILE"

echo "Decision report: $OUT_FILE"

if [[ "$decision" == "GO" ]]; then
  exit 0
fi

exit 1
