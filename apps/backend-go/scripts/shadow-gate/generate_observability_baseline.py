#!/usr/bin/env python3

import argparse
import json
import os
import re
from datetime import datetime, timezone


def parse_whitelist(path: str):
    with open(path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    section = None
    current = None
    routes = []

    def flush_current():
        nonlocal current
        if not current:
            return
        if "path" in current and "method" in current:
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

        path_match = re.match(r"-\s+path:\s+\"([^\"]+)\"", line)
        if path_match:
            flush_current()
            current = {
                "path": path_match.group(1),
                "section": section,
            }
            continue

        if current is None:
            continue

        kv_match = re.match(r"([a-zA-Z]+):\s+\"([^\"]+)\"", line)
        if kv_match:
            key, value = kv_match.group(1), kv_match.group(2)
            current[key] = value

    flush_current()
    return routes


def build_dashboard(routes, whitelist_path: str):
    critical_routes = [r for r in routes if r.get("section") == "critical"]
    generated_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    coverage_rows = []
    for r in critical_routes:
        coverage_rows.append(
            {
                "route": f"{r.get('method', 'GET')} {r.get('path', '')}",
                "owner": r.get("owner", "unknown"),
                "priority": r.get("priority", "P0"),
                "blockingLevel": r.get("blockingLevel", "critical"),
                "goStatus": r.get("goStatus", "unknown"),
            }
        )

    return {
        "name": "shadow-gate-observability",
        "version": "1.0.0",
        "generatedAt": generated_at,
        "sourceWhitelist": whitelist_path,
        "coverage": {
            "criticalRouteCount": len(critical_routes),
            "highPriorityRouteCount": len(
                [r for r in routes if r.get("section") == "high"]
            ),
        },
        "panels": [
            {
                "id": "shadow_mismatch_rate",
                "title": "Shadow Critical Route Mismatch Rate",
                "type": "timeseries",
                "query": 'shadow_mismatch_rate_percent{scope="critical"}',
                "thresholds": {"warning": 0.5, "critical": 1.0},
            },
            {
                "id": "shadow_security_incidents",
                "title": "Shadow Critical Security Incidents",
                "type": "stat",
                "query": 'shadow_security_incidents_total{severity=~"critical|high"}',
                "thresholds": {"critical": 0},
            },
            {
                "id": "shadow_sev12_regressions",
                "title": "Shadow Sev-1/Sev-2 Regressions",
                "type": "stat",
                "query": 'shadow_regressions_total{severity=~"sev1|sev2"}',
                "thresholds": {"critical": 0},
            },
            {
                "id": "shadow_error_distribution",
                "title": "Shadow Route Error Distribution",
                "type": "timeseries",
                "query": "sum by (route, error_class) (rate(shadow_route_errors_total[5m]))",
            },
            {
                "id": "shadow_critical_route_coverage",
                "title": "Critical Route Coverage",
                "type": "table",
                "rows": coverage_rows,
            },
        ],
    }


def build_alert_policy(whitelist_path: str):
    generated_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    return {
        "version": "1.0.0",
        "generatedAt": generated_at,
        "sourceWhitelist": whitelist_path,
        "severityModel": {
            "critical": "block cutover immediately",
            "high": "investigate within 30m",
            "warning": "review within 4h",
        },
        "rules": [
            {
                "name": "shadow_mismatch_rate_block",
                "expr": 'shadow_mismatch_rate_percent{scope="critical"} >= 1',
                "for": "5m",
                "severity": "critical",
                "action": "no-go",
            },
            {
                "name": "shadow_security_incident_block",
                "expr": 'shadow_security_incidents_total{severity=~"critical|high"} > 0',
                "for": "1m",
                "severity": "critical",
                "action": "no-go",
            },
            {
                "name": "shadow_sev12_regression_block",
                "expr": 'shadow_regressions_total{severity=~"sev1|sev2"} > 0',
                "for": "1m",
                "severity": "critical",
                "action": "no-go",
            },
            {
                "name": "shadow_error_distribution_warn",
                "expr": "sum(rate(shadow_route_errors_total[5m])) > 0",
                "for": "10m",
                "severity": "warning",
                "action": "investigate",
            },
        ],
        "escalation": {
            "critical": [
                "Observability Engineer",
                "Release Manager",
                "Engineering Manager",
            ],
            "high": ["Observability Engineer", "API Compatibility Owner"],
            "warning": ["Observability Engineer"],
        },
    }


def write_yaml_like(path: str, data):
    def emit(obj, indent=0):
        space = " " * indent
        lines = []
        if isinstance(obj, dict):
            for key, value in obj.items():
                if isinstance(value, (dict, list)):
                    lines.append(f"{space}{key}:")
                    lines.extend(emit(value, indent + 2))
                else:
                    if isinstance(value, str):
                        lines.append(f"{space}{key}: {json.dumps(value)}")
                    else:
                        lines.append(f"{space}{key}: {value}")
        elif isinstance(obj, list):
            for item in obj:
                if isinstance(item, (dict, list)):
                    lines.append(f"{space}-")
                    lines.extend(emit(item, indent + 2))
                else:
                    if isinstance(item, str):
                        lines.append(f"{space}- {json.dumps(item)}")
                    else:
                        lines.append(f"{space}- {item}")
        return lines

    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(emit(data)) + "\n")


def main():
    parser = argparse.ArgumentParser(
        description="Generate shadow observability dashboard and alert policy from whitelist."
    )
    parser.add_argument("--whitelist", required=True)
    parser.add_argument("--out-dir", required=True)
    args = parser.parse_args()

    routes = parse_whitelist(args.whitelist)
    if not routes:
        raise SystemExit("no critical/high routes parsed from whitelist")

    out_dir = os.path.abspath(args.out_dir)
    os.makedirs(out_dir, exist_ok=True)

    dashboard = build_dashboard(routes, args.whitelist)
    policy = build_alert_policy(args.whitelist)

    dashboard_path = os.path.join(out_dir, "shadow-dashboard.json")
    policy_path = os.path.join(out_dir, "shadow-alert-policy.yaml")

    with open(dashboard_path, "w", encoding="utf-8") as f:
        json.dump(dashboard, f, indent=2)
        f.write("\n")

    write_yaml_like(policy_path, policy)

    print(dashboard_path)
    print(policy_path)


if __name__ == "__main__":
    main()
