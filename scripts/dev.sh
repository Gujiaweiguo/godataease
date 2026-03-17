#!/bin/bash
set -e

cd "$(dirname "$0")/.."

build_backend() {
    echo "Building backend (static)..."
    cd apps/backend-go
    make build-static
    cd ../..
}

build_frontend() {
    echo "Building frontend..."
    cd apps/frontend
    npm run build:base
    cd ../..
}

start_dev() {
    echo "Starting dev mode containers..."
    docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.dev.yml up -d
    echo "Waiting for health check..."
    sleep 5
    curl -s http://localhost:8080/health && echo ""
    echo "Dev mode ready: http://localhost:8080"
}

stop_dev() {
    echo "Stopping dev mode containers..."
    docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.dev.yml down
}

case "${1:-}" in
    build)
        build_backend
        build_frontend
        ;;
    build-backend)
        build_backend
        ;;
    build-frontend)
        build_frontend
        ;;
    start)
        start_dev
        ;;
    stop)
        stop_dev
        ;;
    restart)
        stop_dev
        build_backend
        build_frontend
        start_dev
        ;;
    *)
        echo "Usage: $0 {build|build-backend|build-frontend|start|stop|restart}"
        echo ""
        echo "Commands:"
        echo "  build          Build both frontend and backend"
        echo "  build-backend  Build static backend binary"
        echo "  build-frontend Build frontend dist"
        echo "  start          Start dev containers"
        echo "  stop           Stop dev containers"
        echo "  restart        Rebuild and restart"
        exit 1
        ;;
esac
