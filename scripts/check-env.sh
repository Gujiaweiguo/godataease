#!/bin/bash
set -e

echo "Checking development environment..."

errors=0

check_go() {
    if command -v go &> /dev/null; then
        version=$(go version | grep -oP 'go\d+\.\d+' | head -1)
        echo "✅ Go: $version"
    else
        echo "❌ Go: not found (require Go 1.24+)"
        ((errors++))
    fi
}

check_node() {
    if command -v node &> /dev/null; then
        version=$(node --version)
        echo "✅ Node.js: $version"
    else
        echo "❌ Node.js: not found (require Node 18+)"
        ((errors++))
    fi
}

check_npm() {
    if command -v npm &> /dev/null; then
        version=$(npm --version)
        echo "✅ npm: $version"
    else
        echo "❌ npm: not found"
        ((errors++))
    fi
}

check_docker() {
    if command -v docker &> /dev/null; then
        version=$(docker --version | cut -d' ' -f3 | tr -d ',')
        echo "✅ Docker: $version"
    else
        echo "⚠️  Docker: not found (optional for container mode)"
    fi
}

check_go
check_node
check_npm
check_docker

if [ $errors -gt 0 ]; then
    echo ""
    echo "❌ Environment check failed with $errors error(s)"
    exit 1
else
    echo ""
    echo "✅ Environment check passed"
fi
