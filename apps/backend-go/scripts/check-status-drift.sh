#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${BACKEND_ROOT}/../.." && pwd)"

WHITELIST="${BACKEND_ROOT}/testdata/contract-diff/critical-whitelist.yaml"
INVENTORY="${REPO_ROOT}/openspec/changes/update-api-compatibility-parity-governance/endpoint-inventory.md"
BRIDGE_HANDLER="${BACKEND_ROOT}/internal/transport/http/handler/compatibility_bridge_handler.go"

if [[ ! -f "$WHITELIST" ]]; then
  echo "ERROR: whitelist not found: $WHITELIST"
  exit 1
fi

if [[ ! -f "$INVENTORY" ]]; then
  ARCHIVED_INVENTORY=$(ls -1 "${REPO_ROOT}/openspec/changes/archive"/*-update-api-compatibility-parity-governance/endpoint-inventory.md 2>/dev/null | tail -n 1 || true)
  if [[ -n "${ARCHIVED_INVENTORY}" && -f "${ARCHIVED_INVENTORY}" ]]; then
    INVENTORY="${ARCHIVED_INVENTORY}"
  fi
fi

if [[ ! -f "$INVENTORY" ]]; then
  echo "ERROR: endpoint inventory not found in active/archived changes"
  exit 1
fi

python3 - "$WHITELIST" "$INVENTORY" "$BRIDGE_HANDLER" <<'PY'
import re
import sys
from pathlib import Path

whitelist_path = Path(sys.argv[1])
inventory_path = Path(sys.argv[2])
bridge_path = Path(sys.argv[3])

def strip_quotes(v: str) -> str:
    v = v.strip()
    if len(v) >= 2 and ((v[0] == '"' and v[-1] == '"') or (v[0] == "'" and v[-1] == "'")):
        return v[1:-1]
    return v

inventory_map = {}
for raw in inventory_path.read_text(encoding="utf-8").splitlines():
    m = re.match(r"^\| `(?P<path>/[^`]+)` \| (?P<method>[A-Z]+) \| (?P<status>\*\*)?(full|partial|stub|missing)(\*\*)? \|", raw.strip())
    if not m:
        continue
    cells = [c.strip() for c in raw.strip().split("|") if c.strip()]
    if len(cells) < 3:
        continue
    path = cells[0].strip("`")
    method = cells[1]
    status = cells[2].replace("*", "").lower()
    inventory_map[(method, path)] = status

entries = []
section = ""
current = None
in_gaps = False
for raw in whitelist_path.read_text(encoding="utf-8").splitlines():
    line = raw.rstrip("\n")
    s = line.strip()
    if not s:
        continue

    top = re.match(r"^([A-Za-z][A-Za-z0-9_]*):\s*$", s)
    if top and not line.startswith(" "):
        if current:
            entries.append(current)
            current = None
        section = top.group(1)
        in_gaps = False
        continue

    if s.startswith("- path:"):
        if current:
            entries.append(current)
        current = {
            "section": section,
            "path": strip_quotes(s.split(":", 1)[1]),
            "method": "",
            "go_status": "",
            "has_gaps": False,
        }
        in_gaps = False
        continue

    if not current:
        continue

    if s.startswith("method:"):
        current["method"] = strip_quotes(s.split(":", 1)[1]).upper()
        continue

    if s.startswith("goStatus:"):
        current["go_status"] = strip_quotes(s.split(":", 1)[1]).lower()
        continue

    if s.startswith("gaps:"):
        current["has_gaps"] = True
        in_gaps = True
        continue

    if in_gaps and s.startswith("-"):
        current["has_gaps"] = True
        continue

    if in_gaps and not s.startswith("-"):
        in_gaps = False

if current:
    entries.append(current)

bridge_source = bridge_path.read_text(encoding="utf-8") if bridge_path.exists() else ""
placeholder_patterns = [
    r"response\.Success\(c,\s*nil\)\s*//\s*TODO",
    r"response\.Success\(c,\s*\[\]interface\{\}\{\}\)",
    r"//\s*(placeholder|stub|not implemented)",
]

placeholder_hits = []
for p in placeholder_patterns:
    if re.search(p, bridge_source, flags=re.IGNORECASE):
        placeholder_hits.append(p)

drifts = []
warnings = []
checked = 0
for e in entries:
    if e["section"] == "nativeGoRoutes":
        continue
    if not e["method"] or not e["path"] or not e["go_status"]:
        continue
    key = (e["method"], e["path"])
    if key not in inventory_map:
        continue
    checked += 1
    inv_status = inventory_map[key]
    if inv_status != e["go_status"]:
        drifts.append((e["method"], e["path"], e["go_status"], inv_status))
    if e["go_status"] == "partial" and not e["has_gaps"]:
        warnings.append(f"partial endpoint missing gaps: {e['method']} {e['path']}")

print("=== Compatibility Endpoint Status Drift Check ===")
print(f"compared endpoints: {checked}")

if placeholder_hits:
    for p in placeholder_hits:
        warnings.append(f"potential placeholder pattern: {p}")

for w in warnings:
    print(f"WARNING: {w}")

if drifts:
    for method, path, wl, inv in drifts:
        print(f"DRIFT: {method} {path} whitelist={wl} inventory={inv}")
    sys.exit(1)

print("PASSED: no status drift detected")
sys.exit(0)
PY
