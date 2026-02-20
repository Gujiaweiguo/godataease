#!/usr/bin/env bash

set -euo pipefail

DURATION_HOURS="4"
INTERVAL_SECONDS="3600"
OUT_DIR="tmp/shadow-gate/window"
MIGRATION_MODE="go-only"
SKIP_SHADOW_CONTROL="false"
TOOL_BIN_DIR="tmp/shadow-gate/tools/bin"
JAVA_BASE_URL="${JAVA_BASE_URL:-}"
GO_BASE_URL="${GO_BASE_URL:-}"
MISMATCH_RATE=""
SECURITY_INCIDENTS="0"
SEV1_COUNT="0"
SEV2_COUNT="0"

usage() {
  cat <<'EOF'
Run bounded shadow validation window.

USAGE:
  run_shadow_window.sh [options]

OPTIONS:
  --duration-hours <1-4>      Window duration in hours (default 4)
  --interval-seconds <n>      Sleep seconds between hourly checkpoints (default 3600)
  --out-dir <path>            Output directory (default tmp/shadow-gate/window)
  --migration-mode <mode>     compatibility or go-only (default go-only)
  --skip-shadow-control        Only valid in go-only mode
  --tool-bin-dir <path>       Fallback tool path for preflight checks
  --java-base <url>           Java base URL (compatibility mode)
  --go-base <url>             Go base URL
  --mismatch-rate <percent>   Required for go-only mode
  --security-incidents <n>    Critical security incidents count
  --sev1 <n>                  Sev1 regressions count
  --sev2 <n>                  Sev2 regressions count
  -h, --help                  Show help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --duration-hours)
      DURATION_HOURS="$2"
      shift 2
      ;;
    --interval-seconds)
      INTERVAL_SECONDS="$2"
      shift 2
      ;;
    --out-dir)
      OUT_DIR="$2"
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
    --tool-bin-dir)
      TOOL_BIN_DIR="$2"
      shift 2
      ;;
    --java-base)
      JAVA_BASE_URL="$2"
      shift 2
      ;;
    --go-base)
      GO_BASE_URL="$2"
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

if ! [[ "$DURATION_HOURS" =~ ^[0-9]+$ ]]; then
  echo "--duration-hours must be integer" >&2
  exit 2
fi

if (( DURATION_HOURS < 1 || DURATION_HOURS > 4 )); then
  echo "--duration-hours must be between 1 and 4" >&2
  exit 2
fi

if [[ "$MIGRATION_MODE" != "compatibility" && "$MIGRATION_MODE" != "go-only" ]]; then
  echo "--migration-mode must be compatibility or go-only" >&2
  exit 2
fi

if [[ "$SKIP_SHADOW_CONTROL" == "true" && "$MIGRATION_MODE" != "go-only" ]]; then
  echo "--skip-shadow-control is only allowed in go-only mode" >&2
  exit 2
fi

if [[ "$MIGRATION_MODE" == "go-only" && -z "$MISMATCH_RATE" ]]; then
  echo "--mismatch-rate is required in go-only mode" >&2
  exit 2
fi

if [[ "$MIGRATION_MODE" == "compatibility" ]]; then
  if [[ -z "$JAVA_BASE_URL" || -z "$GO_BASE_URL" ]]; then
    echo "--java-base and --go-base are required in compatibility mode" >&2
    exit 2
  fi
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$OUT_DIR"

start_ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
overall_status="PASS"
completed_checkpoints=0

for hour in $(seq 1 "$DURATION_HOURS"); do
  checkpoint_dir="$OUT_DIR/checkpoint-H$(printf "%02d" "$hour")"
  mkdir -p "$checkpoint_dir"

  cmd=(
    "$SCRIPT_DIR/run_shadow_gate.sh"
    --migration-mode "$MIGRATION_MODE"
    --out-dir "$checkpoint_dir"
    --tool-bin-dir "$TOOL_BIN_DIR"
    --security-incidents "$SECURITY_INCIDENTS"
    --sev1 "$SEV1_COUNT"
    --sev2 "$SEV2_COUNT"
  )

  if [[ "$SKIP_SHADOW_CONTROL" == "true" ]]; then
    cmd+=(--skip-shadow-control)
  fi
  if [[ -n "$GO_BASE_URL" ]]; then
    cmd+=(--go-base "$GO_BASE_URL")
  fi
  if [[ -n "$JAVA_BASE_URL" ]]; then
    cmd+=(--java-base "$JAVA_BASE_URL")
  fi
  if [[ -n "$MISMATCH_RATE" ]]; then
    cmd+=(--mismatch-rate "$MISMATCH_RATE")
  fi

  set +e
  "${cmd[@]}"
  gate_rc=$?
  set -e

  completed_checkpoints=$hour

  checkpoint_ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  decision="UNKNOWN"
  if [[ -f "$checkpoint_dir/shadow-gate-decision.md" ]]; then
    decision_line="$(awk '/^- Decision: / { print; exit }' "$checkpoint_dir/shadow-gate-decision.md")"
    decision="${decision_line#- Decision: }"
  fi

  {
    echo "# Checkpoint H$(printf "%02d" "$hour")"
    echo
    echo "- Timestamp: $checkpoint_ts"
    echo "- Exit Code: $gate_rc"
    echo "- Decision: $decision"
  } > "$checkpoint_dir/checkpoint.md"

  if (( hour % 4 == 0 || hour == DURATION_HOURS )); then
    set +e
    "$SCRIPT_DIR/test_alert_policy.sh" \
      --out "$checkpoint_dir/alert-probe.md" \
      --mismatch-rate "$MISMATCH_RATE" \
      --security-incidents "$SECURITY_INCIDENTS" \
      --sev1 "$SEV1_COUNT" \
      --sev2 "$SEV2_COUNT"
    probe_rc=$?
    set -e
    if [[ "$probe_rc" -ne 0 ]]; then
      overall_status="FAIL"
    fi
  fi

  if [[ "$gate_rc" -ne 0 ]]; then
    overall_status="FAIL"
    break
  fi

  if (( hour < DURATION_HOURS )); then
    sleep "$INTERVAL_SECONDS"
  fi
done

end_ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

summary_file="$OUT_DIR/window-summary.md"
{
  echo "# Shadow Window Summary"
  echo
  echo "- Start: $start_ts"
  echo "- End: $end_ts"
  echo "- Duration Hours (requested): $DURATION_HOURS"
  echo "- Duration Hours (completed): $completed_checkpoints"
  echo "- Migration Mode: $MIGRATION_MODE"
  echo "- Skip Shadow Control: $SKIP_SHADOW_CONTROL"
  echo "- Overall Status: $overall_status"
} > "$summary_file"

echo "Window summary: $summary_file"

if [[ "$overall_status" == "PASS" ]]; then
  exit 0
fi

exit 1
