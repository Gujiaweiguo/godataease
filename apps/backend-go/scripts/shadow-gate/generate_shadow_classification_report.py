#!/usr/bin/env python3

import argparse
import os
import re
from datetime import datetime, timezone


def normalize_path(path: str) -> str:
    return re.sub(r"\{([a-zA-Z0-9_]+)\}", r":\1", path)


def parse_registered_routes(handler_root: str):
    routes = set()
    group_re = re.compile(
        r"([A-Za-z_][A-Za-z0-9_]*)\s*:=\s*[A-Za-z0-9_\.]+\s*\.Group\(\"([^\"]+)\"\)"
    )
    route_re = re.compile(
        r"([A-Za-z_][A-Za-z0-9_]*)\.(GET|POST|PUT|DELETE|PATCH|Any)\(\"([^\"]+)\""
    )

    for root, _, files in os.walk(handler_root):
        for name in files:
            if not name.endswith(".go"):
                continue
            file_path = os.path.join(root, name)
            with open(file_path, "r", encoding="utf-8") as f:
                lines = f.readlines()

            groups = {}
            for raw in lines:
                line = raw.strip()
                if line.startswith("//"):
                    continue

                gm = group_re.search(line)
                if gm:
                    groups[gm.group(1)] = gm.group(2)
                    continue

                rm = route_re.search(line)
                if not rm:
                    continue

                var = rm.group(1)
                method = rm.group(2).upper()
                route_path = rm.group(3)

                if var in groups:
                    full = normalize_path(groups[var] + route_path)
                    routes.add((method, full))
                elif route_path.startswith("/"):
                    routes.add((method, normalize_path(route_path)))

    return routes


def parse_whitelist(path: str):
    with open(path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    section = None
    current = None
    routes = []

    def flush_current():
        nonlocal current
        if current and "path" in current and "method" in current:
            routes.append(current)
        current = None

    for raw in lines:
        line = raw.strip()
        if not line:
            continue

        if line.startswith("criticalApis:"):
            flush_current()
            section = "critical"
            continue
        if line.startswith("highPriorityApis:"):
            flush_current()
            section = "high"
            continue
        if line.startswith("nativeGoRoutes:"):
            flush_current()
            section = "native"
            continue

        if section not in ("critical", "high"):
            continue

        m = re.match(r"-\s+path:\s+\"([^\"]+)\"", line)
        if m:
            flush_current()
            current = {"path": m.group(1), "section": section}
            continue

        if current is None:
            continue

        kv = re.match(r"([a-zA-Z]+):\s+\"([^\"]+)\"", line)
        if kv:
            current[kv.group(1)] = kv.group(2)

    flush_current()
    return routes


def read_window(window_dir: str):
    summary = os.path.join(window_dir, "window-summary.md")
    if not os.path.isfile(summary):
        raise SystemExit(f"window summary not found: {summary}")

    with open(summary, "r", encoding="utf-8") as f:
        text = f.read()

    completed = extract_number(text, r"Duration Hours \(completed\):\s*(\d+)")
    requested = extract_number(text, r"Duration Hours \(requested\):\s*(\d+)")
    status = extract_text(text, r"Overall Status:\s*(\w+)")

    checkpoint_dir = os.path.join(window_dir, f"checkpoint-H{completed:02d}")
    decision_file = os.path.join(checkpoint_dir, "shadow-gate-decision.md")
    alert_file = os.path.join(checkpoint_dir, "alert-probe.md")

    if not os.path.isfile(decision_file):
        raise SystemExit(f"decision report not found: {decision_file}")

    with open(decision_file, "r", encoding="utf-8") as f:
        decision_text = f.read()

    mismatch_rate = extract_decimal(
        decision_text, r"Mismatch rate:\s*([0-9]+(?:\.[0-9]+)?)%"
    )
    security_incidents = extract_number(decision_text, r"Security incidents:\s*(\d+)")
    sev1 = extract_number(decision_text, r"Sev-1 regressions:\s*(\d+)")
    sev2 = extract_number(decision_text, r"Sev-2 regressions:\s*(\d+)")
    gate_decision = extract_text(decision_text, r"Decision:\s*([A-Z\-]+)")

    alert_triggered = "unknown"
    if os.path.isfile(alert_file):
        with open(alert_file, "r", encoding="utf-8") as f:
            alert_text = f.read()
        alert_triggered = "none"
        if "- none" not in alert_text:
            alert_triggered = "triggered"

    return {
        "requested": requested,
        "completed": completed,
        "status": status,
        "decision": gate_decision,
        "mismatch_rate": mismatch_rate,
        "security_incidents": security_incidents,
        "sev1": sev1,
        "sev2": sev2,
        "alert_triggered": alert_triggered,
        "decision_file": decision_file,
        "summary_file": summary,
        "alert_file": alert_file if os.path.isfile(alert_file) else "n/a",
    }


def extract_number(text: str, pattern: str) -> int:
    m = re.search(pattern, text)
    if not m:
        raise SystemExit(f"pattern not found: {pattern}")
    return int(m.group(1))


def extract_decimal(text: str, pattern: str) -> float:
    m = re.search(pattern, text)
    if not m:
        raise SystemExit(f"pattern not found: {pattern}")
    return float(m.group(1))


def extract_text(text: str, pattern: str) -> str:
    m = re.search(pattern, text)
    if not m:
        raise SystemExit(f"pattern not found: {pattern}")
    return m.group(1)


def classify_routes(routes, registered_routes):
    blocking = []
    non_blocking = []

    alias_map = {
        ("GET", "/templateMarket/searchTemplate"): ("GET", "/templateMarket/search"),
    }

    for r in routes:
        if r.get("section") != "critical":
            continue
        owner = r.get("owner", "unknown")
        method = r.get("method", "GET").upper()
        path = normalize_path(r.get("path", ""))
        route = f"{method} {path}"
        go_status = r.get("goStatus", "unknown")
        notes = r.get("notes", "")
        exists = (method, path) in registered_routes

        if not exists and (method, path) in alias_map:
            exists = alias_map[(method, path)] in registered_routes

        item = {
            "route": route,
            "owner": owner,
            "go_status": go_status,
            "notes": notes,
        }

        if not exists:
            item["go_status"] = "route-missing"
            blocking.append(item)
        elif go_status in ("missing", "stub"):
            item["go_status"] = "metadata-stale"
            non_blocking.append(item)
        elif go_status == "partial":
            non_blocking.append(item)

    return blocking, non_blocking


def write_report(path: str, window, blocking, non_blocking):
    now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    mismatch_category = "low"
    if window["mismatch_rate"] >= 1:
        mismatch_category = "blocking"
    elif window["mismatch_rate"] >= 0.5:
        mismatch_category = "warning"

    security_summary = "none"
    root_cause = "N/A"
    mitigation = "N/A"
    status = "closed"
    if window["security_incidents"] > 0 or window["sev1"] > 0 or window["sev2"] > 0:
        security_summary = "incidents-detected"
        root_cause = "To be determined by incident review"
        mitigation = "Block cutover and complete remediation"
        status = "open"

    lines = [
        "# SHADOW-004 Mismatch and Security Classification Report",
        "",
        f"- Generated At: {now}",
        f"- Window Summary: `{window['summary_file']}`",
        f"- Decision Source: `{window['decision_file']}`",
        f"- Alert Probe: `{window['alert_file']}`",
        "",
        "## Execution Overview",
        "",
        f"- Requested Hours: {window['requested']}",
        f"- Completed Hours: {window['completed']}",
        f"- Window Status: {window['status']}",
        f"- Gate Decision: {window['decision']}",
        "",
        "## Mismatch Classification",
        "",
        f"- Measured mismatch rate: {window['mismatch_rate']:.2f}%",
        f"- Category: {mismatch_category}",
        "- Threshold: blocking if >= 1.00%",
        "",
        "## Security Incident Summary",
        "",
        f"- Critical security incidents: {window['security_incidents']}",
        f"- Sev-1 regressions: {window['sev1']}",
        f"- Sev-2 regressions: {window['sev2']}",
        f"- Summary: {security_summary}",
        f"- Root Cause: {root_cause}",
        f"- Mitigation: {mitigation}",
        f"- Mitigation Status: {status}",
        "",
        "## Route-level Blocking Defects",
        "",
        "| Route | Owner | Basis | Notes |",
        "|------|-------|-------|-------|",
    ]

    if not blocking:
        lines.append("| none | n/a | none-observed | n/a |")
    else:
        for item in blocking:
            lines.append(
                f"| {item['route']} | {item['owner']} | goStatus={item['go_status']} | {item['notes']} |"
            )

    lines.extend(
        [
            "",
            "## Route-level Non-blocking Defects",
            "",
            "| Route | Owner | Basis | Notes |",
            "|------|-------|-------|-------|",
        ]
    )

    if not non_blocking:
        lines.append("| none | n/a | none-observed | n/a |")
    else:
        for item in non_blocking:
            lines.append(
                f"| {item['route']} | {item['owner']} | goStatus={item['go_status']} | {item['notes']} |"
            )

    lines.extend(
        [
            "",
            "## Notes",
            "",
            "- This classification combines go-only shadow window evidence with observed route registration in Go handlers.",
            "- metadata-stale means whitelist goStatus does not match registered route presence and needs business-level semantics confirmation.",
        ]
    )

    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")


def main():
    parser = argparse.ArgumentParser(
        description="Generate SHADOW-004 mismatch/security classification report."
    )
    parser.add_argument("--whitelist", required=True)
    parser.add_argument("--window-dir", required=True)
    parser.add_argument("--handler-root", default="internal/transport/http/handler")
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    routes = parse_whitelist(args.whitelist)
    if not routes:
        raise SystemExit("no routes parsed from whitelist")

    registered_routes = parse_registered_routes(args.handler_root)
    if not registered_routes:
        raise SystemExit("no registered routes parsed from handler source")

    window = read_window(args.window_dir)
    blocking, non_blocking = classify_routes(routes, registered_routes)

    out_path = os.path.abspath(args.out)
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    write_report(out_path, window, blocking, non_blocking)
    print(out_path)


if __name__ == "__main__":
    main()
