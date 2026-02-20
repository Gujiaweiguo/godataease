#!/bin/bash
#
# Contract Diff Runtime Engine - Compare API responses between Java and Go backends
# Part of CI-GATE-002: Go Backend Contract Diff CI Gate
#
# Usage: ./run_contract_diff.sh --whitelist <path> [options]
#
# Exit Codes:
#   0 - All API comparisons passed
#   1 - One or more API comparisons failed
#   2 - Configuration or runtime error
#

set -euo pipefail

#######################################
# Configuration
#######################################

# Default configuration
DEFAULT_JAVA_BASE="http://localhost:8100"
DEFAULT_GO_BASE="http://localhost:8080"
DEFAULT_OUT_DIR="./tmp/contract-diff"
DEFAULT_TIMEOUT=30
DEFAULT_RETRIES=3
SCHEMA_VERSION="1.0.0"

# Runtime configuration (populated from args)
JAVA_BASE=""
GO_BASE=""
OUT_DIR=""
TIMEOUT=""
RETRIES=""
WHITELIST_PATH=""

# Output files
JSON_REPORT=""
MD_REPORT=""

# Counters for summary
TOTAL_APIS=0
PASSED_APIS=0
FAILED_APIS=0

# Track blocking failures
HAS_CRITICAL_FAILURE=false
HAS_HIGH_FAILURE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Results array (stored as temp file for JSON array building)
RESULTS_TEMP_FILE=""

# YAML parser (detected at runtime)
YAML_PARSER=""

#######################################
# Print usage message
#######################################
usage() {
    cat << EOF
Contract Diff Runtime Engine - Compare API responses between Java and Go backends

USAGE:
    $0 --whitelist <path> [OPTIONS]

REQUIRED:
    --whitelist <path>     Path to whitelist YAML file defining APIs to compare

OPTIONS:
    --java-base <url>      Java backend base URL (default: $DEFAULT_JAVA_BASE)
    --go-base <url>        Go backend base URL (default: $DEFAULT_GO_BASE)
    --out-dir <path>       Output directory for reports (default: $DEFAULT_OUT_DIR)
    --timeout <seconds>    Request timeout in seconds (default: $DEFAULT_TIMEOUT)
    --retries <count>      Number of retry attempts (default: $DEFAULT_RETRIES)
    -h, --help             Show this help message

EXAMPLES:
    # Basic usage with defaults
    $0 --whitelist ./whitelist.yaml

    # Custom backends and output
    $0 --whitelist ./whitelist.yaml \\
        --java-base http://java-api:8100 \\
        --go-base http://go-api:8080 \\
        --out-dir ./reports

EXIT CODES:
    0    All API comparisons passed
    1    One or more API comparisons failed
    2    Configuration or runtime error

EOF
}

#######################################
# Log functions
#######################################
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

#######################################
# Parse command line arguments
#######################################
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --whitelist)
                WHITELIST_PATH="$2"
                shift 2
                ;;
            --java-base)
                JAVA_BASE="$2"
                shift 2
                ;;
            --go-base)
                GO_BASE="$2"
                shift 2
                ;;
            --out-dir)
                OUT_DIR="$2"
                shift 2
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --retries)
                RETRIES="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 2
                ;;
        esac
    done

    # Apply defaults
    JAVA_BASE="${JAVA_BASE:-$DEFAULT_JAVA_BASE}"
    GO_BASE="${GO_BASE:-$DEFAULT_GO_BASE}"
    OUT_DIR="${OUT_DIR:-$DEFAULT_OUT_DIR}"
    TIMEOUT="${TIMEOUT:-$DEFAULT_TIMEOUT}"
    RETRIES="${RETRIES:-$DEFAULT_RETRIES}"

    # Validate required args
    if [[ -z "$WHITELIST_PATH" ]]; then
        log_error "--whitelist is required"
        usage
        exit 2
    fi

    if [[ ! -f "$WHITELIST_PATH" ]]; then
        log_error "Whitelist file not found: $WHITELIST_PATH"
        exit 2
    fi
}

#######################################
# Check required tools
#######################################
check_dependencies() {
    local missing=()

    # Check for curl
    if ! command -v curl &> /dev/null; then
        missing+=("curl")
    fi

    # Check for jq (required for JSON processing)
    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    # Check for yq OR python3 with yaml
    if command -v yq &> /dev/null; then
        YAML_PARSER="yq"
    elif command -v python3 &> /dev/null; then
        if python3 -c "import yaml" 2>/dev/null; then
            YAML_PARSER="python"
        else
            missing+=("python3 with PyYAML (pip install pyyaml)")
        fi
    else
        missing+=("yq or python3 with PyYAML")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing[*]}"
        exit 2
    fi

    log_info "Using YAML parser: $YAML_PARSER"
}

#######################################
# Initialize output directory and files
#######################################
init_output() {
    mkdir -p "$OUT_DIR"
    JSON_REPORT="$OUT_DIR/contract-diff.json"
    MD_REPORT="$OUT_DIR/contract-diff.md"
    RESULTS_TEMP_FILE=$(mktemp)

    # Initialize empty results array
    echo "[]" > "$RESULTS_TEMP_FILE"
}

#######################################
# Parse whitelist YAML and validate
# Returns: JSON array of APIs
#######################################
parse_whitelist() {
    local apis_json="[]"

    if [[ "$YAML_PARSER" == "yq" ]]; then
        # Use yq to parse YAML
        apis_json=$(yq eval '
            (.criticalApis // []) + (.highPriorityApis // []) |
            map({
                path: .path,
                method: .method,
                owner: .owner,
                priority: .priority,
                blockingLevel: .blockingLevel,
                notes: .notes
            })
        ' "$WHITELIST_PATH" 2>/dev/null)
    else
        # Use Python to parse YAML
        apis_json=$(python3 << PYTHON_SCRIPT
import yaml
import json
import sys

with open("$WHITELIST_PATH", "r") as f:
    data = yaml.safe_load(f)

apis = []
for section in ["criticalApis", "highPriorityApis"]:
    for api in data.get(section, []):
        apis.append({
            "path": api.get("path"),
            "method": api.get("method"),
            "owner": api.get("owner"),
            "priority": api.get("priority"),
            "blockingLevel": api.get("blockingLevel"),
            "notes": api.get("notes", "")
        })

print(json.dumps(apis))
PYTHON_SCRIPT
        )
    fi

    # Validate required fields
    local validation_errors=""
    local count=0

    while IFS= read -r api; do
        ((count++)) || true
        local path method owner priority blockingLevel
        path=$(echo "$api" | jq -r '.path // empty')
        method=$(echo "$api" | jq -r '.method // empty')
        owner=$(echo "$api" | jq -r '.owner // empty')
        priority=$(echo "$api" | jq -r '.priority // empty')
        blockingLevel=$(echo "$api" | jq -r '.blockingLevel // empty')

        local missing_fields=""
        [[ -z "$path" ]] && missing_fields+="path "
        [[ -z "$method" ]] && missing_fields+="method "
        [[ -z "$owner" ]] && missing_fields+="owner "
        [[ -z "$priority" ]] && missing_fields+="priority "
        [[ -z "$blockingLevel" ]] && missing_fields+="blockingLevel "

        if [[ -n "$missing_fields" ]]; then
            validation_errors+="API #$count ($path): missing fields: $missing_fields\n"
        fi
    done < <(echo "$apis_json" | jq -c '.[]')

    if [[ -n "$validation_errors" ]]; then
        log_error "Whitelist validation failed:" >&2
        echo -e "$validation_errors" >&2
        exit 2
    fi

    echo "$apis_json"
}

#######################################
# Make HTTP request with retry logic
# Arguments:
#   $1 - Base URL
#   $2 - HTTP method
#   $3 - API path
#   $4 - Request body (optional, for POST)
# Returns: JSON with status, body, headers, error
#######################################
make_request() {
    local base_url="$1"
    local method="$2"
    local api_path="$3"
    local body="${4:-}"
    local url="${base_url}${api_path}"

    local attempt=1
    local response_file
    response_file=$(mktemp)
    local http_code=""
    local error_msg=""

    while [[ $attempt -le $RETRIES ]]; do
        # Build curl command
        local curl_args=(
            -s
            -S
            --connect-timeout "$TIMEOUT"
            --max-time "$((TIMEOUT * 2))"
            -w "\n%{http_code}"
            -o "$response_file"
        )

        if [[ "$method" == "POST" ]]; then
            curl_args+=(-X POST)
            if [[ -n "$body" ]]; then
                curl_args+=(-H "Content-Type: application/json")
                curl_args+=(-d "$body")
            fi
        fi

        curl_args+=("$url")

        # Execute request
        if http_code=$(curl "${curl_args[@]}" 2>&1); then
            # Check if we got a valid HTTP code
            if [[ "$http_code" =~ ^[0-9]+$ ]]; then
                # Read response body
                local response_body
                if [[ -s "$response_file" ]]; then
                    response_body=$(cat "$response_file")
                    # Try to parse as JSON
                    if ! echo "$response_body" | jq . > /dev/null 2>&1; then
                        response_body="\"$(echo "$response_body" | sed 's/"/\\"/g' | tr -d '\n')\""
                    else
                        response_body=$(echo "$response_body" | jq -c .)
                    fi
                else
                    response_body="null"
                fi

                rm -f "$response_file"

                # Return JSON result
                echo "{\"status\": $http_code, \"body\": $response_body, \"error\": null}"
                return 0
            fi
        fi

        ((attempt++))
        sleep 1
    done

    # All retries failed
    error_msg="Connection failed after $RETRIES attempts"
    rm -f "$response_file"

    echo "{\"status\": 0, \"body\": null, \"error\": \"$error_msg\"}"
}

#######################################
# Extract response fields for comparison
# Arguments:
#   $1 - Response JSON
# Returns: JSON with code, msg, data fields extracted
#######################################
extract_response_fields() {
    local response="$1"
    local status code msg data error

    status=$(echo "$response" | jq -r '.status')
    error=$(echo "$response" | jq -r '.error // empty')

    local body_type
    body_type=$(echo "$response" | jq -r '.body | type')

    if [[ "$body_type" == "object" ]]; then
        code=$(echo "$response" | jq -r '.body.code // empty')
        msg=$(echo "$response" | jq -r '.body.msg // empty')
        data=$(echo "$response" | jq -c '.body.data // empty')
    else
        code=""
        msg=""
        data=""
    fi

    # Handle empty code/msg/data/error
    [[ -z "$code" || "$code" == "null" ]] && code="null"
    [[ -z "$msg" || "$msg" == "null" ]] && msg="null"
    [[ -z "$error" || "$error" == "null" ]] && error="null"

    local code_json msg_json data_json error_json

    if [[ "$code" == "null" ]]; then
        code_json="null"
    else
        code_json=$(printf '%s' "$code" | jq -R .)
    fi

    if [[ "$msg" == "null" ]]; then
        msg_json="null"
    else
        msg_json=$(printf '%s' "$msg" | jq -R .)
    fi

    if [[ -z "$data" || "$data" == "" || "$data" == "null" ]]; then
        data_json="null"
    elif echo "$data" | jq -e . >/dev/null 2>&1; then
        data_json="$data"
    else
        data_json=$(printf '%s' "$data" | jq -R .)
    fi

    if [[ "$error" == "null" ]]; then
        error_json="null"
    else
        error_json=$(printf '%s' "$error" | jq -R .)
    fi

    echo "{\"status\": $status, \"code\": $code_json, \"msg\": $msg_json, \"data\": $data_json, \"error\": $error_json}"
}

#######################################
# Compare data schemas (field names and types)
# Arguments:
#   $1 - Java data JSON
#   $2 - Go data JSON
# Returns: JSON with schema diff details
#######################################
compare_schema() {
    local java_data="$1"
    local go_data="$2"

    # If either is null or empty, compare directly
    if [[ "$java_data" == "null" || "$go_data" == "null" ]]; then
        if [[ "$java_data" == "$go_data" ]]; then
            echo "{\"diff\": false, \"details\": null}"
        else
            echo "{\"diff\": true, \"details\": \"One data field is null\"}"
        fi
        return
    fi

    # Extract keys and types from both
    local java_schema go_schema
    java_schema=$(echo "$java_data" | jq -c 'if type == "object" then to_entries | map({key: .key, type: (.value | type)}) | sort_by(.key) else [] end' 2>/dev/null || echo "[]")
    go_schema=$(echo "$go_data" | jq -c 'if type == "object" then to_entries | map({key: .key, type: (.value | type)}) | sort_by(.key) else [] end' 2>/dev/null || echo "[]")

    if [[ "$java_schema" == "$go_schema" ]]; then
        echo "{\"diff\": false, \"details\": null}"
    else
        # Find differences
        local java_keys go_keys
        java_keys=$(echo "$java_data" | jq -r 'if type == "object" then keys[] else empty end' 2>/dev/null | sort | tr '\n' ' ')
        go_keys=$(echo "$go_data" | jq -r 'if type == "object" then keys[] else empty end' 2>/dev/null | sort | tr '\n' ' ')

        local details="Schema mismatch"
        [[ -n "$java_keys" ]] && details+=" | Java keys: [$java_keys]"
        [[ -n "$go_keys" ]] && details+=" | Go keys: [$go_keys]"

        echo "{\"diff\": true, \"details\": \"$details\"}"
    fi
}

#######################################
# Compare data values
# Arguments:
#   $1 - Java data JSON
#   $2 - Go data JSON
# Returns: JSON with value diff details
#######################################
compare_values() {
    local java_data="$1"
    local go_data="$2"

    # If schemas differ, skip value comparison
    # If either is null, skip
    if [[ "$java_data" == "null" || "$go_data" == "null" ]]; then
        echo "{\"diff\": false, \"details\": null}"
        return
    fi

    # Direct comparison
    if [[ "$java_data" == "$go_data" ]]; then
        echo "{\"diff\": false, \"details\": null}"
    else
        echo "{\"diff\": true, \"details\": \"Value mismatch\"}"
    fi
}

#######################################
# Classify failure based on diff
# Arguments:
#   $1 - Diff JSON object
#   $2 - Java response
#   $3 - Go response
# Returns: Failure category and severity
#######################################
classify_failure() {
    local diff="$1"
    local java_resp="$2"
    local go_resp="$3"

    local categories="[]"
    local severity="normal"

    # Check for connection error
    local java_error go_error
    java_error=$(echo "$java_resp" | jq -r '.error // empty')
    go_error=$(echo "$go_resp" | jq -r '.error // empty')

    if [[ -n "$java_error" || -n "$go_error" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["CONNECTION_ERROR"]')
        severity="critical"
    fi

    # Check for timeout (status 0 with no connection error means timeout)
    local java_status go_status
    java_status=$(echo "$java_resp" | jq -r '.status')
    go_status=$(echo "$go_resp" | jq -r '.status')

    if [[ "$java_status" == "0" || "$go_status" == "0" ]]; then
        if [[ -z "$java_error" && -z "$go_error" ]]; then
            categories=$(echo "$categories" | jq -c '. + ["TIMEOUT"]')
            [[ "$severity" != "critical" ]] && severity="high"
        fi
    fi

    # Check status diff
    local status_diff
    status_diff=$(echo "$diff" | jq -r '.statusDiff')
    if [[ "$status_diff" == "true" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["STATUS_DIFF"]')
        severity="critical"
    fi

    # Check code diff
    local code_diff
    code_diff=$(echo "$diff" | jq -r '.codeDiff')
    if [[ "$code_diff" == "true" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["CODE_DIFF"]')
        severity="critical"
    fi

    # Check msg diff
    local msg_diff
    msg_diff=$(echo "$diff" | jq -r '.msgDiff')
    if [[ "$msg_diff" == "true" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["MSG_DIFF"]')
        [[ "$severity" == "normal" ]] && severity="high"
    fi

    # Check payload schema diff
    local schema_diff
    schema_diff=$(echo "$diff" | jq -r '.schemaDiff // false')
    if [[ "$schema_diff" == "true" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["PAYLOAD_SCHEMA_DIFF"]')
        severity="critical"
    fi

    # Check payload value diff
    local value_diff
    value_diff=$(echo "$diff" | jq -r '.valueDiff // false')
    if [[ "$value_diff" == "true" ]]; then
        categories=$(echo "$categories" | jq -c '. + ["PAYLOAD_VALUE_DIFF"]')
        # severity stays as is (normal for value diff)
    fi

    echo "{\"categories\": $categories, \"severity\": \"$severity\"}"
}

#######################################
# Compare two API responses
# Arguments:
#   $1 - Java response JSON
#   $2 - Go response JSON
# Returns: JSON with match status and diff details
#######################################
compare_responses() {
    local java_response="$1"
    local go_response="$2"

    # Extract fields
    local java_fields go_fields
    java_fields=$(extract_response_fields "$java_response")
    go_fields=$(extract_response_fields "$go_response")

    local java_status go_status java_code go_code java_msg go_msg java_data go_data java_error go_error
    java_status=$(echo "$java_fields" | jq -r '.status')
    go_status=$(echo "$go_fields" | jq -r '.status')
    java_code=$(echo "$java_fields" | jq -r '.code')
    go_code=$(echo "$go_fields" | jq -r '.code')
    java_msg=$(echo "$java_fields" | jq -r '.msg')
    go_msg=$(echo "$go_fields" | jq -r '.msg')
    java_data=$(echo "$java_fields" | jq -c '.data')
    go_data=$(echo "$go_fields" | jq -c '.data')
    java_error=$(echo "$java_fields" | jq -r '.error')
    go_error=$(echo "$go_fields" | jq -r '.error')

    # Compare status
    local status_diff="false"
    [[ "$java_status" != "$go_status" ]] && status_diff="true"

    # Compare code (as string)
    local code_diff="false"
    [[ "$java_code" != "$go_code" ]] && code_diff="true"

    # Compare msg
    local msg_diff="false"
    [[ "$java_msg" != "$go_msg" ]] && msg_diff="true"

    # Compare schema
    local schema_result
    schema_result=$(compare_schema "$java_data" "$go_data")
    local schema_diff
    schema_diff=$(echo "$schema_result" | jq -r '.diff')

    # Compare values (only if schema matches)
    local value_result
    if [[ "$schema_diff" == "false" ]]; then
        value_result=$(compare_values "$java_data" "$go_data")
    else
        value_result='{"diff": false, "details": null}'
    fi
    local value_diff
    value_diff=$(echo "$value_result" | jq -r '.diff')

    local payload_diff="false"
    [[ "$schema_diff" == "true" || "$value_diff" == "true" ]] && payload_diff="true"

    # Build diff object
    local diff
    diff=$(jq -n \
        --argjson sd "$status_diff" \
        --argjson cd "$code_diff" \
        --argjson md "$msg_diff" \
        --argjson pld "$payload_diff" \
        --argjson scd "$schema_diff" \
        --argjson vld "$value_diff" \
        '{
            statusDiff: $sd,
            codeDiff: $cd,
            msgDiff: $md,
            payloadDiff: $pld,
            schemaDiff: $scd,
            valueDiff: $vld
        }')

    # Determine if passed
    local passed="true"
    if [[ "$status_diff" == "true" || "$code_diff" == "true" || "$msg_diff" == "true" || "$schema_diff" == "true" || "$value_diff" == "true" ]]; then
        passed="false"
    fi

    # Build error message if any
    local error="null"
    if [[ -n "$java_error" && "$java_error" != "null" ]]; then
        error="Java: $java_error"
    fi
    if [[ -n "$go_error" && "$go_error" != "null" ]]; then
        if [[ "$error" == "null" ]]; then
            error="Go: $go_error"
        else
            error="$error; Go: $go_error"
        fi
    fi

    # Build response objects
    local java_resp_obj go_resp_obj error_json
    java_resp_obj=$(jq -n \
        --arg s "$java_status" \
        --arg c "$java_code" \
        --arg m "$java_msg" \
        '{status: ($s | tonumber), code: $c, msg: $m}')
    go_resp_obj=$(jq -n \
        --arg s "$go_status" \
        --arg c "$go_code" \
        --arg m "$go_msg" \
        '{status: ($s | tonumber), code: $c, msg: $m}')

    if [[ "$error" == "null" ]]; then
        error_json="null"
    else
        error_json=$(printf '%s' "$error" | jq -R .)
    fi

    echo "{\"passed\": $passed, \"diff\": $diff, \"javaResponse\": $java_resp_obj, \"goResponse\": $go_resp_obj, \"error\": $error_json}"
}

#######################################
# Add result to results array
#######################################
add_result() {
    local result="$1"

    local current
    current=$(cat "$RESULTS_TEMP_FILE")
    echo "$current" | jq -c ". + [$result]" > "$RESULTS_TEMP_FILE"
}

#######################################
# Run comparison for a single API
#######################################
run_api_comparison() {
    local api="$1"
    local path method owner priority blockingLevel notes
    path=$(echo "$api" | jq -r '.path')
    method=$(echo "$api" | jq -r '.method')
    owner=$(echo "$api" | jq -r '.owner')
    priority=$(echo "$api" | jq -r '.priority')
    blockingLevel=$(echo "$api" | jq -r '.blockingLevel')
    notes=$(echo "$api" | jq -r '.notes // empty')

    log_info "Testing: $method $path [$priority/$blockingLevel]"

    # Make requests
    local java_response go_response
    java_response=$(make_request "$JAVA_BASE" "$method" "$path" "")
    go_response=$(make_request "$GO_BASE" "$method" "$path" "")

    # Compare responses
    local comparison
    comparison=$(compare_responses "$java_response" "$go_response")

    local passed diff java_resp go_resp error
    passed=$(echo "$comparison" | jq -r '.passed')
    diff=$(echo "$comparison" | jq -c '.diff')
    java_resp=$(echo "$comparison" | jq -c '.javaResponse')
    go_resp=$(echo "$comparison" | jq -c '.goResponse')
    error=$(echo "$comparison" | jq -c '.error')

    # Classify failure
    local classification
    classification=$(classify_failure "$diff" "$java_response" "$go_response")
    local categories severity
    categories=$(echo "$classification" | jq -c '.categories')
    severity=$(echo "$classification" | jq -r '.severity')

    # Update counters
    ((TOTAL_APIS++)) || true
    if [[ "$passed" == "true" ]]; then
        ((PASSED_APIS++)) || true
        log_success "$method $path - Parity OK"
    else
        ((FAILED_APIS++)) || true
        log_error "$method $path - Parity FAILED [$severity]"

        # Track blocking failures
        if [[ "$blockingLevel" == "critical" ]]; then
            HAS_CRITICAL_FAILURE=true
        elif [[ "$blockingLevel" == "high" ]]; then
            HAS_HIGH_FAILURE=true
        fi
    fi

    # Build result object
    local result
    result=$(jq -n \
        --arg p "$path" \
        --arg m "$method" \
        --arg o "$owner" \
        --arg pr "$priority" \
        --arg bl "$blockingLevel" \
        --argjson jr "$java_resp" \
        --argjson gr "$go_resp" \
        --argjson d "$diff" \
        --argjson pa "$passed" \
        --argjson e "$error" \
        --argjson c "$categories" \
        --arg s "$severity" \
        --arg n "$notes" \
        '{
            path: $p,
            method: $m,
            owner: $o,
            priority: $pr,
            blockingLevel: $bl,
            javaResponse: $jr,
            goResponse: $gr,
            diff: $d,
            passed: $pa,
            error: $e,
            categories: $c,
            severity: $s,
            notes: $n
        }')

    add_result "$result"
}

#######################################
# Generate JSON report
#######################################
generate_json_report() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Calculate parity
    local parity="0.0"
    if [[ $TOTAL_APIS -gt 0 ]]; then
        parity=$(echo "scale=1; $PASSED_APIS * 100 / $TOTAL_APIS" | bc)
    fi

    local results
    results=$(cat "$RESULTS_TEMP_FILE")

    # Build report
    jq -n \
        --arg ts "$timestamp" \
        --arg v "$SCHEMA_VERSION" \
        --arg wp "$WHITELIST_PATH" \
        --arg jb "$JAVA_BASE" \
        --arg gb "$GO_BASE" \
        --argjson total "$TOTAL_APIS" \
        --argjson passed "$PASSED_APIS" \
        --argjson failed "$FAILED_APIS" \
        --arg parity "$parity" \
        --argjson results "$results" \
        '{
            metadata: {
                timestamp: $ts,
                version: $v,
                whitelistPath: $wp,
                javaBaseUrl: $jb,
                goBaseUrl: $gb
            },
            summary: {
                total: $total,
                passed: $passed,
                failed: $failed,
                parity: ($parity | tonumber)
            },
            results: $results
        }' > "$JSON_REPORT"

    log_info "JSON report saved to: $JSON_REPORT"
}

#######################################
# Generate Markdown report
#######################################
generate_md_report() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Calculate parity
    local parity="0.0"
    if [[ $TOTAL_APIS -gt 0 ]]; then
        parity=$(echo "scale=1; $PASSED_APIS * 100 / $TOTAL_APIS" | bc)
    fi

    {
        echo "# Contract Diff Report"
        echo ""
        echo "**Generated:** $timestamp"
        echo ""
        echo "## Summary"
        echo ""
        echo "| Metric | Value |"
        echo "|--------|-------|"
        echo "| Total APIs | $TOTAL_APIS |"
        echo "| Passed | $PASSED_APIS |"
        echo "| Failed | $FAILED_APIS |"
        echo "| Parity | ${parity}% |"
        echo ""
        echo "## Configuration"
        echo ""
        echo "- **Java Backend:** \`$JAVA_BASE\`"
        echo "- **Go Backend:** \`$GO_BASE\`"
        echo "- **Whitelist:** \`$WHITELIST_PATH\`"
        echo "- **Timeout:** ${TIMEOUT}s"
        echo "- **Retries:** $RETRIES"
        echo ""

        # Failed APIs section
        if [[ $FAILED_APIS -gt 0 ]]; then
            echo "## Failed APIs"
            echo ""

            # Get failed results
            local failed_results
            failed_results=$(cat "$RESULTS_TEMP_FILE" | jq -c '.[] | select(.passed == false)')

            if [[ -n "$failed_results" ]]; then
                echo "| Path | Method | Priority | Blocking | Severity | Categories |"
                echo "|------|--------|----------|----------|----------|------------|"

                while IFS= read -r result; do
                    local path method priority blocking severity categories
                    path=$(echo "$result" | jq -r '.path')
                    method=$(echo "$result" | jq -r '.method')
                    priority=$(echo "$result" | jq -r '.priority')
                    blocking=$(echo "$result" | jq -r '.blockingLevel')
                    severity=$(echo "$result" | jq -r '.severity')
                    categories=$(echo "$result" | jq -r '.categories | join(", ")')

                    echo "| \`$path\` | $method | $priority | $blocking | $severity | $categories |"
                done <<< "$failed_results"

                echo ""

                # Detailed failure information
                echo "### Failure Details"
                echo ""

                while IFS= read -r result; do
                    local path method java_resp go_resp diff error categories notes
                    path=$(echo "$result" | jq -r '.path')
                    method=$(echo "$result" | jq -r '.method')
                    java_resp=$(echo "$result" | jq -c '.javaResponse')
                    go_resp=$(echo "$result" | jq -c '.goResponse')
                    diff=$(echo "$result" | jq -c '.diff')
                    error=$(echo "$result" | jq -r '.error // empty')
                    categories=$(echo "$result" | jq -r '.categories | join(", ")')
                    notes=$(echo "$result" | jq -r '.notes // empty')

                    echo "#### \`$method $path\`"
                    echo ""
                    echo "**Categories:** $categories"
                    [[ -n "$notes" ]] && echo "**Notes:** $notes"
                    [[ -n "$error" && "$error" != "null" ]] && echo "**Error:** $error"
                    echo ""
                    echo "| Field | Java | Go | Match |"
                    echo "|-------|------|-----|-------|"

                    local j_status g_status j_code g_code j_msg g_msg
                    j_status=$(echo "$java_resp" | jq -r '.status')
                    g_status=$(echo "$go_resp" | jq -r '.status')
                    j_code=$(echo "$java_resp" | jq -r '.code')
                    g_code=$(echo "$go_resp" | jq -r '.code')
                    j_msg=$(echo "$java_resp" | jq -r '.msg')
                    g_msg=$(echo "$go_resp" | jq -r '.msg')

                    local status_match="✓"
                    [[ $(echo "$diff" | jq -r '.statusDiff') == "true" ]] && status_match="✗"
                    local code_match="✓"
                    [[ $(echo "$diff" | jq -r '.codeDiff') == "true" ]] && code_match="✗"
                    local msg_match="✓"
                    [[ $(echo "$diff" | jq -r '.msgDiff') == "true" ]] && msg_match="✗"

                    echo "| Status | $j_status | $g_status | $status_match |"
                    echo "| Code | $j_code | $g_code | $code_match |"
                    echo "| Msg | $j_msg | $g_msg | $msg_match |"
                    echo ""
                done <<< "$failed_results"
            fi
        fi

        # Passed APIs section (summary only)
        if [[ $PASSED_APIS -gt 0 ]]; then
            echo "## Passed APIs"
            echo ""
            local passed_results
            passed_results=$(cat "$RESULTS_TEMP_FILE" | jq -c '.[] | select(.passed == true)')

            if [[ -n "$passed_results" ]]; then
                echo "| Path | Method | Priority | Blocking |"
                echo "|------|--------|----------|----------|"

                while IFS= read -r result; do
                    local path method priority blocking
                    path=$(echo "$result" | jq -r '.path')
                    method=$(echo "$result" | jq -r '.method')
                    priority=$(echo "$result" | jq -r '.priority')
                    blocking=$(echo "$result" | jq -r '.blockingLevel')

                    echo "| \`$path\` | $method | $priority | $blocking |"
                done <<< "$passed_results"
            fi
            echo ""
        fi

    } > "$MD_REPORT"

    log_info "Markdown report saved to: $MD_REPORT"
}

#######################################
# Cleanup temp files
#######################################
cleanup() {
    [[ -f "$RESULTS_TEMP_FILE" ]] && rm -f "$RESULTS_TEMP_FILE"
    return 0
}

#######################################
# Main execution
#######################################
main() {
    # Setup cleanup trap
    trap cleanup EXIT

    parse_args "$@"

    echo -e "${GREEN}Contract Diff Runtime Engine${NC}"
    echo "=============================="
    echo "Java Backend: $JAVA_BASE"
    echo "Go Backend:   $GO_BASE"
    echo "Whitelist:    $WHITELIST_PATH"
    echo "Output Dir:   $OUT_DIR"
    echo "Timeout:      ${TIMEOUT}s"
    echo "Retries:      $RETRIES"
    echo ""

    check_dependencies
    init_output

    log_info "Parsing whitelist..."
    local apis
    apis=$(parse_whitelist)
    local api_count
    api_count=$(echo "$apis" | jq 'length')
    log_info "Found $api_count APIs to compare"
    echo ""

    log_info "Running API comparisons..."
    echo ""

    # Process each API
    while IFS= read -r api; do
        run_api_comparison "$api"
    done < <(echo "$apis" | jq -c '.[]')

    echo ""
    log_info "Generating reports..."

    generate_json_report
    generate_md_report

    echo ""
    echo -e "${GREEN}========== Summary ==========${NC}"
    echo "Total:   $TOTAL_APIS"
    echo "Passed:  $PASSED_APIS"
    echo "Failed:  $FAILED_APIS"
    echo ""

    # Determine exit code
    if [[ "$HAS_CRITICAL_FAILURE" == "true" ]]; then
        log_error "Gate FAILED: Critical blocking level API(s) failed"
        exit 1
    elif [[ "$HAS_HIGH_FAILURE" == "true" ]]; then
        log_error "Gate FAILED: High blocking level API(s) failed"
        exit 1
    elif [[ $FAILED_APIS -gt 0 ]]; then
        log_warn "Gate passed with warnings: Some non-blocking APIs failed"
        exit 0
    else
        log_success "Gate PASSED: All APIs have parity"
        exit 0
    fi
}

# Run main function
main "$@"
