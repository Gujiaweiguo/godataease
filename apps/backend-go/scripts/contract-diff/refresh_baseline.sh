#!/bin/bash
#
# Baseline Refresh Script - Update baseline fixtures from Java backend responses
# Part of BASELINE-003: Baseline Refresh Script
#
# Usage: ./refresh_baseline.sh --whitelist <path> --apply [options]
#        ./refresh_baseline.sh --whitelist <path> --dry-run [options]
#
# Exit Codes:
#   0 - Success (dry-run preview or apply completed)
#   1 - Runtime error
#   2 - Configuration error
#

set -euo pipefail

#######################################
# Configuration
#######################################

# Default configuration
DEFAULT_JAVA_BASE="http://localhost:8100"
DEFAULT_WHITELIST="./whitelist.yaml"
DEFAULT_BASELINES_DIR="../testdata/contract-diff/baselines"
DEFAULT_TIMEOUT=30
DEFAULT_RETRIES=3
SCHEMA_VERSION="1.0.0"

# Runtime configuration
JAVA_BASE=""
WHITELIST_PATH=""
BASELINES_DIR=""
TIMEOUT=""
RETRIES=""
API_FILTER=""
DRY_RUN=false
APPLY_MODE=false

# Counters
TOTAL_APIS=0
UPDATED_APIS=0
NEW_APIS=0
UNCHANGED_APIS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# YAML parser
YAML_PARSER=""

#######################################
# Print usage message
#######################################
usage() {
    cat << EOF
Baseline Refresh Script - Update baseline fixtures from Java backend responses

USAGE:
    $0 --whitelist <path> --dry-run [OPTIONS]
    $0 --whitelist <path> --apply [OPTIONS]

MODES (exactly one required):
    --dry-run           Preview changes without modifying files
    --apply             Actually write/update baseline files

REQUIRED:
    --whitelist <path>  Path to whitelist YAML file

OPTIONS:
    --java-base <url>       Java backend base URL (default: $DEFAULT_JAVA_BASE)
    --baselines-dir <path>  Output directory for baseline fixtures (default: $DEFAULT_BASELINES_DIR)
    --api-filter <regex>    Filter APIs to refresh by path regex (default: all)
    --timeout <seconds>     Request timeout in seconds (default: $DEFAULT_TIMEOUT)
    --retries <count>       Number of retry attempts (default: $DEFAULT_RETRIES)
    -h, --help              Show this help message

EXAMPLES:
    # Preview all baseline changes
    $0 --whitelist ./whitelist.yaml --dry-run

    # Update all baselines
    $0 --whitelist ./whitelist.yaml --apply

    # Update only dataset APIs
    $0 --whitelist ./whitelist.yaml --apply --api-filter "/api/dataset/.*"

    # Custom Java backend
    $0 --whitelist ./whitelist.yaml --apply \\
        --java-base http://java-api:8100 \\
        --baselines-dir ./baselines

EXIT CODES:
    0    Success
    1    Runtime error
    2    Configuration error

EOF
}

#######################################
# Log functions
#######################################
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_diff() {
    echo -e "${CYAN}[DIFF]${NC} $1"
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
            --baselines-dir)
                BASELINES_DIR="$2"
                shift 2
                ;;
            --api-filter)
                API_FILTER="$2"
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
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --apply)
                APPLY_MODE=true
                shift
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
    BASELINES_DIR="${BASELINES_DIR:-$DEFAULT_BASELINES_DIR}"
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

    # Validate mode (exactly one required)
    if [[ "$DRY_RUN" == "true" && "$APPLY_MODE" == "true" ]]; then
        log_error "Cannot use both --dry-run and --apply"
        usage
        exit 2
    fi

    if [[ "$DRY_RUN" == "false" && "$APPLY_MODE" == "false" ]]; then
        log_error "Must specify either --dry-run or --apply"
        usage
        exit 2
    fi
}

#######################################
# Check required tools
#######################################
check_dependencies() {
    local missing=()

    if ! command -v curl &> /dev/null; then
        missing+=("curl")
    fi

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if command -v yq &> /dev/null; then
        YAML_PARSER="yq"
    elif command -v python3 &> /dev/null; then
        if python3 -c "import yaml" 2>/dev/null; then
            YAML_PARSER="python"
        else
            missing+=("python3 with PyYAML")
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
# Parse whitelist YAML
#######################################
parse_whitelist() {
    local apis_json="[]"

    if [[ "$YAML_PARSER" == "yq" ]]; then
        apis_json=$(yq eval '
            (.criticalApis // []) + (.highPriorityApis // []) |
            map({
                path: .path,
                method: .method,
                owner: .owner,
                priority: .priority,
                blockingLevel: .blockingLevel
            })
        ' "$WHITELIST_PATH" 2>/dev/null)
    else
        apis_json=$(python3 << PYTHON_SCRIPT
import yaml
import json

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
            "blockingLevel": api.get("blockingLevel")
        })

print(json.dumps(apis))
PYTHON_SCRIPT
        )
    fi

    echo "$apis_json"
}

#######################################
# Make HTTP request with retry
#######################################
make_request() {
    local base_url="$1"
    local method="$2"
    local api_path="$3"
    local url="${base_url}${api_path}"

    local attempt=1
    local response_file
    response_file=$(mktemp)
    local http_code=""

    while [[ $attempt -le $RETRIES ]]; do
        local curl_args=(
            -s -S
            --connect-timeout "$TIMEOUT"
            --max-time "$((TIMEOUT * 2))"
            -w "\n%{http_code}"
            -o "$response_file"
        )

        [[ "$method" == "POST" ]] && curl_args+=(-X POST -H "Content-Type: application/json")

        curl_args+=("$url")

        if http_code=$(curl "${curl_args[@]}" 2>&1); then
            if [[ "$http_code" =~ ^[0-9]+$ ]]; then
                local response_body
                if [[ -s "$response_file" ]]; then
                    response_body=$(cat "$response_file")
                else
                    response_body=""
                fi
                rm -f "$response_file"
                echo "{\"status\": $http_code, \"body\": $(echo "$response_body" | jq -c . 2>/dev/null || echo 'null')}"
                return 0
            fi
        fi

        ((attempt++))
        sleep 1
    done

    rm -f "$response_file"
    echo "{\"status\": 0, \"body\": null, \"error\": \"Connection failed after $RETRIES attempts\"}"
}

#######################################
# Get capability from API path
#######################################
get_capability() {
    local path="$1"
    # Extract second path segment as capability
    local segment
    segment=$(echo "$path" | sed 's|^/api/||' | cut -d'/' -f1)
    echo "$segment"
}

#######################################
# Generate baseline filename from API path
#######################################
get_baseline_filename() {
    local path="$1"
    local method="$2"

    # Remove /api/ prefix
    local stripped="${path#/api/}"
    # Replace / with _
    local name="${stripped//\//_}"
    # Handle path params :param -> _param
    name="${name//:/_}"
    # Append method
    echo "${name}_${method}.json"
}

#######################################
# Load existing baseline (if exists)
#######################################
load_existing_baseline() {
    local filepath="$1"
    if [[ -f "$filepath" ]]; then
        cat "$filepath"
    else
        echo ""
    fi
}

#######################################
# Generate baseline fixture JSON
#######################################
generate_baseline() {
    local api="$1"
    local response="$2"
    local commit_sha
    commit_sha=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    local path method owner priority blockingLevel
    path=$(echo "$api" | jq -r '.path')
    method=$(echo "$api" | jq -r '.method')
    owner=$(echo "$api" | jq -r '.owner')
    priority=$(echo "$api" | jq -r '.priority')
    blockingLevel=$(echo "$api" | jq -r '.blockingLevel')

    local status body code msg data
    status=$(echo "$response" | jq -r '.status')
    body=$(echo "$response" | jq -c '.body')
    code=$(echo "$body" | jq -r '.code // "null"' 2>/dev/null || echo "null")
    msg=$(echo "$body" | jq -r '.msg // "null"' 2>/dev/null || echo "null")
    data=$(echo "$body" | jq -c '.data // null' 2>/dev/null || echo "null")

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg p "$path" \
        --arg m "$method" \
        --arg o "$owner" \
        --arg pr "$priority" \
        --arg bl "$blockingLevel" \
        --argjson st "$status" \
        --arg c "$code" \
        --arg msg "$msg" \
        --argjson d "$data" \
        --arg ts "$timestamp" \
        --arg cs "$commit_sha" \
        --arg v "$SCHEMA_VERSION" \
        '{
            apiIdentity: {
                path: $p,
                method: $m,
                owner: $o,
                priority: $pr,
                blockingLevel: $bl
            },
            standardResponse: {
                status: $st,
                code: $c,
                msg: $msg,
                data: $d
            },
            versionMetadata: {
                captured_at: $ts,
                java_version: "21",
                baseline_version: $v,
                commit_sha: $cs
            }
        }'
}

#######################################
# Compare two baselines
#######################################
compare_baselines() {
    local existing="$1"
    local new="$2"

    if [[ -z "$existing" ]]; then
        echo "NEW"
        return
    fi

    # Compare standardResponse sections
    local existing_resp new_resp
    existing_resp=$(echo "$existing" | jq -c '.standardResponse' 2>/dev/null)
    new_resp=$(echo "$new" | jq -c '.standardResponse' 2>/dev/null)

    if [[ "$existing_resp" == "$new_resp" ]]; then
        echo "UNCHANGED"
    else
        echo "CHANGED"
    fi
}

#######################################
# Process single API
#######################################
process_api() {
    local api="$1"
    local path method
    path=$(echo "$api" | jq -r '.path')
    method=$(echo "$api" | jq -r '.method')

    ((TOTAL_APIS++)) || true

    # Apply filter if specified
    if [[ -n "$API_FILTER" ]]; then
        if ! echo "$path" | grep -qE "$API_FILTER"; then
            return
        fi
    fi

    local capability filename filepath
    capability=$(get_capability "$path")
    filename=$(get_baseline_filename "$path" "$method")
    filepath="${BASELINES_DIR}/${capability}/${filename}"

    log_info "Fetching: $method $path"

    # Fetch response from Java backend
    local response
    response=$(make_request "$JAVA_BASE" "$method" "$path")

    local status
    status=$(echo "$response" | jq -r '.status')

    if [[ "$status" == "0" ]]; then
        log_error "Failed to fetch: $method $path"
        return
    fi

    # Generate new baseline
    local new_baseline
    new_baseline=$(generate_baseline "$api" "$response")

    # Load existing baseline
    local existing_baseline
    existing_baseline=$(load_existing_baseline "$filepath")

    # Compare
    local cmp_result
    cmp_result=$(compare_baselines "$existing_baseline" "$new_baseline")

    case "$cmp_result" in
        "NEW")
            ((NEW_APIS++)) || true
            if [[ "$DRY_RUN" == "true" ]]; then
                log_diff "[NEW] $filepath"
            else
                mkdir -p "$(dirname "$filepath")"
                echo "$new_baseline" | jq '.' > "$filepath"
                log_success "[NEW] $filepath"
            fi
            ;;
        "CHANGED")
            ((UPDATED_APIS++)) || true
            if [[ "$DRY_RUN" == "true" ]]; then
                log_diff "[CHANGED] $filepath"
                # Show diff preview
                local existing_code new_code existing_msg new_msg
                existing_code=$(echo "$existing_baseline" | jq -r '.standardResponse.code')
                new_code=$(echo "$new_baseline" | jq -r '.standardResponse.code')
                existing_msg=$(echo "$existing_baseline" | jq -r '.standardResponse.msg')
                new_msg=$(echo "$new_baseline" | jq -r '.standardResponse.msg')
                [[ "$existing_code" != "$new_code" ]] && echo "    code: $existing_code -> $new_code"
                [[ "$existing_msg" != "$new_msg" ]] && echo "    msg: $existing_msg -> $new_msg"
            else
                mkdir -p "$(dirname "$filepath")"
                echo "$new_baseline" | jq '.' > "$filepath"
                log_success "[UPDATED] $filepath"
            fi
            ;;
        "UNCHANGED")
            ((UNCHANGED_APIS++)) || true
            log_info "[UNCHANGED] $filepath"
            ;;
    esac
}

#######################################
# Main execution
#######################################
main() {
    parse_args "$@"
    check_dependencies

    local mode_str
    if [[ "$DRY_RUN" == "true" ]]; then
        mode_str="DRY-RUN (preview only)"
    else
        mode_str="APPLY (will update files)"
    fi

    echo -e "${GREEN}Baseline Refresh Script${NC}"
    echo "========================"
    echo "Mode:        $mode_str"
    echo "Java Base:   $JAVA_BASE"
    echo "Whitelist:   $WHITELIST_PATH"
    echo "Baselines:   $BASELINES_DIR"
    [[ -n "$API_FILTER" ]] && echo "API Filter:  $API_FILTER"
    echo ""

    log_info "Parsing whitelist..."
    local apis
    apis=$(parse_whitelist)
    local api_count
    api_count=$(echo "$apis" | jq 'length')
    log_info "Found $api_count APIs in whitelist"
    echo ""

    if [[ "$APPLY_MODE" == "true" ]]; then
        mkdir -p "$BASELINES_DIR"
    fi

    log_info "Processing APIs..."
    echo ""

    while IFS= read -r api; do
        process_api "$api"
    done < <(echo "$apis" | jq -c '.[]')

    echo ""
    echo -e "${GREEN}========== Summary ==========${NC}"
    echo "Total APIs processed: $TOTAL_APIS"
    echo "New baselines:        $NEW_APIS"
    echo "Updated baselines:    $UPDATED_APIS"
    echo "Unchanged baselines:  $UNCHANGED_APIS"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo ""
        log_warn "This was a dry-run. No files were modified."
        log_info "Run with --apply to update baseline files."
    fi

    exit 0
}

main "$@"
