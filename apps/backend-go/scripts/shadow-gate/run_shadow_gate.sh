#!/usr/bin/env bash

set -euo pipefail

DEFAULT_OUT_DIR="./tmp/shadow-gate"
DEFAULT_WHITELIST="./testdata/contract-diff/critical-whitelist.yaml"

OUT_DIR="$DEFAULT_OUT_DIR"
WHITELIST="$DEFAULT_WHITELIST"
JAVA_BASE_URL="${JAVA_BASE_URL:-}"
GO_BASE_URL="${GO_BASE_URL:-}"
MIGRATION_MODE="compatibility"
MISMATCH_RATE=""
TOOL_BIN_DIR="./tmp/shadow-gate/tools/bin"
SKIP_SHADOW_CONTROL="false"
SECURITY_INCIDENTS="0"
SEV1_COUNT="0"
SEV2_COUNT="0"
AUTO_SHADOW_RULE="false"

usage() {
  cat <<'EOF'
Run Shadow Gate (preflight + contract diff + decision)

USAGE:
  run_shadow_gate.sh [options]

OPTIONS:
  --out-dir <path>            Output directory (default ./tmp/shadow-gate)
  --whitelist <path>          Whitelist file (default ./testdata/contract-diff/critical-whitelist.yaml)
  --migration-mode <mode>     compatibility (default) or go-only
  --mismatch-rate <percent>   Use observed mismatch directly (for go-only mode)
  --tool-bin-dir <path>       Fallback tooling dir for kubectl/helm checks
  --skip-shadow-control       In go-only mode, skip control-plane/env checks and use offline rehearsal path
  --java-base <url>           Java baseline URL (or env JAVA_BASE_URL)
  --go-base <url>             Go candidate URL (or env GO_BASE_URL)
  --security-incidents <n>    Critical security incidents count (default 0)
  --sev1 <n>                  Sev-1 regression count (default 0)
  --sev2 <n>                  Sev-2 regression count (default 0)
  --auto-shadow-rule          Enable shadow before run and disable on exit (requires gateway env vars)
  -h, --help                  Show help

NOTES:
  - This script does not implement 48h loop collection.
  - It provides command-level gate execution for preflight and Go/No-Go decision.
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
    --migration-mode)
      MIGRATION_MODE="$2"
      shift 2
      ;;
    --mismatch-rate)
      MISMATCH_RATE="$2"
      shift 2
      ;;
    --tool-bin-dir)
      TOOL_BIN_DIR="$2"
      shift 2
      ;;
    --skip-shadow-control)
      SKIP_SHADOW_CONTROL="true"
      shift
      ;;
    --java-base)
      JAVA_BASE_URL="$2"
      shift 2
      ;;
    --go-base)
      GO_BASE_URL="$2"
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
    --auto-shadow-rule)
      AUTO_SHADOW_RULE="true"
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

if [[ "$MIGRATION_MODE" == "compatibility" && -z "$JAVA_BASE_URL" ]]; then
  echo "java base URL is required in compatibility mode (--java-base or env JAVA_BASE_URL)" >&2
  exit 2
fi

if [[ "$MIGRATION_MODE" == "go-only" && -z "$MISMATCH_RATE" ]]; then
  echo "--mismatch-rate is required in go-only mode" >&2
  exit 2
fi

if [[ "$SKIP_SHADOW_CONTROL" == "true" && "$MIGRATION_MODE" != "go-only" ]]; then
  echo "--skip-shadow-control is only allowed in go-only mode" >&2
  exit 2
fi

if [[ "$SKIP_SHADOW_CONTROL" == "true" && "$AUTO_SHADOW_RULE" == "true" ]]; then
  echo "--skip-shadow-control cannot be used with --auto-shadow-rule" >&2
  exit 2
fi

if [[ "$MIGRATION_MODE" == "compatibility" || "$SKIP_SHADOW_CONTROL" != "true" ]]; then
  if [[ -z "$GO_BASE_URL" ]]; then
    echo "go base URL is required (--go-base or env GO_BASE_URL)" >&2
    exit 2
  fi
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

mkdir -p "$OUT_DIR"
CONTRACT_DIR="$OUT_DIR/contract-diff"
mkdir -p "$CONTRACT_DIR"

if [[ "$AUTO_SHADOW_RULE" == "true" ]]; then
  "$SCRIPT_DIR/manage_shadow_rule.sh" --action enable
fi

cleanup() {
  if [[ "$AUTO_SHADOW_RULE" == "true" ]]; then
    "$SCRIPT_DIR/manage_shadow_rule.sh" --action disable || true
  fi
}
trap cleanup EXIT

set +e
preflight_args=(
  --out-dir "$OUT_DIR"
  --whitelist "$WHITELIST"
  --tool-bin-dir "$TOOL_BIN_DIR"
  --migration-mode "$MIGRATION_MODE"
)
if [[ "$SKIP_SHADOW_CONTROL" == "true" ]]; then
  preflight_args+=(--skip-shadow-control)
fi
"$SCRIPT_DIR/preflight_check.sh" "${preflight_args[@]}"
preflight_rc=$?
set -e

if [[ "$preflight_rc" -ne 0 ]]; then
  echo "Preflight blocked. See $OUT_DIR/preflight.md"
  exit 1
fi

contract_rc=0
if [[ "$MIGRATION_MODE" == "compatibility" ]]; then
  set +e
  "$PROJECT_ROOT/scripts/contract-diff/run_contract_diff.sh" \
    --java-base "$JAVA_BASE_URL" \
    --go-base "$GO_BASE_URL" \
    --out-dir "$CONTRACT_DIR" \
    --whitelist "$WHITELIST"
  contract_rc=$?
  set -e

  if [[ ! -f "$CONTRACT_DIR/contract-diff.json" ]]; then
    echo "Contract diff report missing: $CONTRACT_DIR/contract-diff.json" >&2
    exit 1
  fi

  "$SCRIPT_DIR/evaluate_shadow_gate.sh" \
    --contract-report "$CONTRACT_DIR/contract-diff.json" \
    --out "$OUT_DIR/shadow-gate-decision.md" \
    --security-incidents "$SECURITY_INCIDENTS" \
    --sev1 "$SEV1_COUNT" \
    --sev2 "$SEV2_COUNT"
else
  "$SCRIPT_DIR/evaluate_shadow_gate.sh" \
    --mismatch-rate "$MISMATCH_RATE" \
    --out "$OUT_DIR/shadow-gate-decision.md" \
    --security-incidents "$SECURITY_INCIDENTS" \
    --sev1 "$SEV1_COUNT" \
    --sev2 "$SEV2_COUNT"
fi

echo "Shadow gate decision report: $OUT_DIR/shadow-gate-decision.md"

if [[ "$contract_rc" -ne 0 ]]; then
  exit 1
fi
