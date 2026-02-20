#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$COMPOSE_DIR")"

echo "=== DataEase Production Deployment ==="
echo ""

check_requirements() {
    echo "[1/6] Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        echo "ERROR: Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null || ! docker compose version &> /dev/null; then
        echo "ERROR: Docker Compose is not installed"
        exit 1
    fi
    
    echo "  ✓ Docker and Docker Compose are installed"
}

load_env() {
    echo "[2/6] Loading environment configuration..."
    
    if [ -f "$COMPOSE_DIR/.env.prod" ]; then
        export $(grep -v '^#' "$COMPOSE_DIR/.env.prod" | xargs)
        echo "  ✓ Loaded .env.prod"
    elif [ -f "$COMPOSE_DIR/.env.prod.example" ]; then
        echo "  ⚠ .env.prod not found, using .env.prod.example"
        echo "  ⚠ Please copy .env.prod.example to .env.prod and configure"
        export $(grep -v '^#' "$COMPOSE_DIR/.env.prod.example" | xargs)
    else
        echo "  ERROR: No environment file found"
        exit 1
    fi
}

create_directories() {
    echo "[3/6] Creating data directories..."
    
    mkdir -p "${DATA_DIR:-/opt/module/dataease2.0}/data"
    mkdir -p "${DATA_DIR:-/opt/module/dataease2.0}/logs"
    mkdir -p "${DATA_DIR:-/opt/module/dataease2.0}/configs"
    mkdir -p "${MYSQL_DATA_DIR:-/opt/data/mysql/dataease}"
    mkdir -p "${MYSQL_LOG_DIR:-/opt/logs/mysql/dataease}"
    mkdir -p "${REDIS_DATA_DIR:-/opt/data/redis/dataease}"
    
    echo "  ✓ Data directories created"
}

build_frontend() {
    echo "[4/6] Building frontend..."
    
    cd "$PROJECT_ROOT/apps/frontend"
    npm install --silent
    npm run build:base
    
    echo "  ✓ Frontend built"
}

build_backend() {
    echo "[5/6] Building backend image..."
    
    cd "$PROJECT_ROOT"
    docker build -t dataease:${DATAEASE_VERSION:-latest} .
    
    echo "  ✓ Backend image built"
}

deploy() {
    echo "[6/6] Deploying services..."
    
    cd "$COMPOSE_DIR"
    docker compose -f docker-compose.prod.yml up -d
    
    echo ""
    echo "=== Deployment Complete ==="
    echo ""
    echo "Services:"
    echo "  - DataEase App: http://localhost:${HTTP_PORT:-80}"
    echo "  - DataEase API: http://localhost:${SERVER_PORT:-8080}"
    echo "  - MySQL: localhost:${MYSQL_PORT:-3306}"
    echo "  - Redis: localhost:${REDIS_EXTERNAL_PORT:-16379}"
    echo ""
    echo "Logs: docker compose -f docker-compose.prod.yml logs -f"
    echo "Stop:  docker compose -f docker-compose.prod.yml down"
}

case "${1:-deploy}" in
    deploy)
        check_requirements
        load_env
        create_directories
        build_frontend
        build_backend
        deploy
        ;;
    start)
        load_env
        cd "$COMPOSE_DIR"
        docker compose -f docker-compose.prod.yml up -d
        ;;
    stop)
        cd "$COMPOSE_DIR"
        docker compose -f docker-compose.prod.yml down
        ;;
    logs)
        cd "$COMPOSE_DIR"
        docker compose -f docker-compose.prod.yml logs -f "${2:-}"
        ;;
    status)
        cd "$COMPOSE_DIR"
        docker compose -f docker-compose.prod.yml ps
        ;;
    *)
        echo "Usage: $0 {deploy|start|stop|logs|status}"
        exit 1
        ;;
esac
