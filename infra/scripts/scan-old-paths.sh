#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

OLD_PATH_PATTERNS=(
    "^backend-go/"
    "/backend-go/"
    "\"backend-go/"
    "'backend-go/"
    "core/core-frontend"
    "core/core-backend"
)

ALLOWLIST=(
    "openspec/changes/archive/"
    ".sisyphus/"
    "legacy/README-READONLY.md"
    "infra/scripts/scan-old-paths.sh"
)

STRICT_MODE=false
if [[ "$1" == "--strict" ]]; then
    STRICT_MODE=true
fi

echo "=========================================="
echo "旧路径残留扫描"
echo "=========================================="
echo ""

cd "$PROJECT_ROOT"

total_issues=0
blocking_issues=0

for pattern in "${OLD_PATH_PATTERNS[@]}"; do
    echo -e "${YELLOW}扫描模式:${NC} $pattern"
    
    matches=$(grep -rnE --include="*.md" --include="*.yml" --include="*.yaml" --include="*.json" --include="*.xml" --include="*.sh" --include="*.ts" --include="*.js" --include="*.go" --include="*.java" \
        "$pattern" . 2>/dev/null | grep -v "node_modules" | grep -v ".git" || true)
    
    if [[ -n "$matches" ]]; then
        filtered_matches=""
        blocking_count=0
        while IFS= read -r line; do
            is_allowed=false
            for allow in "${ALLOWLIST[@]}"; do
                if [[ "$line" == *"$allow"* ]]; then
                    is_allowed=true
                    break
                fi
            done
            
            if [[ "$is_allowed" == false ]]; then
                filtered_matches="$filtered_matches$line\n"
                ((blocking_count++))
            fi
        done <<< "$matches"
        
        if [[ $blocking_count -gt 0 ]]; then
            echo -e "  ${RED}发现 $blocking_count 个匹配${NC}"
            echo -e "$filtered_matches" | head -5
            ((total_issues += blocking_count))
            ((blocking_issues += blocking_count))
        else
            echo -e "  ${GREEN}✓ 无阻塞级残留${NC}"
        fi
    else
        echo -e "  ${GREEN}✓ 无匹配${NC}"
    fi
    echo ""
done

echo "检查旧目录是否存在..."
if [[ -d "backend-go" ]]; then
    echo -e "${RED}✗ 旧目录仍存在: backend-go/${NC}"
    ((blocking_issues++))
fi
if [[ -d "core/core-frontend" ]]; then
    echo -e "${RED}✗ 旧目录仍存在: core/core-frontend/${NC}"
    ((blocking_issues++))
fi
if [[ -d "core/core-backend" ]]; then
    echo -e "${RED}✗ 旧目录仍存在: core/core-backend/${NC}"
    ((blocking_issues++))
fi
if [[ -d "core" ]]; then
    echo -e "${RED}✗ 旧目录仍存在: core/${NC}"
    ((blocking_issues++))
fi

echo ""
echo "=========================================="
echo "扫描结果汇总"
echo "=========================================="
echo -e "总问题数: $total_issues"
echo -e "阻塞级问题: $blocking_issues"
echo ""

if [[ $blocking_issues -gt 0 ]]; then
    echo -e "${RED}✗ 扫描失败: 发现 $blocking_issues 个阻塞级旧路径残留${NC}"
    if [[ "$STRICT_MODE" == true ]]; then
        exit 1
    else
        echo -e "${YELLOW}提示: 使用 --strict 模式会导致构建失败${NC}"
        exit 0
    fi
else
    echo -e "${GREEN}✓ 扫描通过: 无阻塞级旧路径残留${NC}"
    exit 0
fi
