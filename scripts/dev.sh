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

start_redis() {
    echo "Starting Redis container for local development..."
    docker compose -f infra/compose/docker-compose.yml up -d godataease-redis
    echo "Redis ready on 127.0.0.1:16379"
}

stop_redis() {
    echo "Stopping Redis container for local development..."
    docker compose -f infra/compose/docker-compose.yml stop godataease-redis
}

show_local_help() {
    cat <<'EOF'
Local hybrid development

1. Start Redis in Docker
   ./scripts/dev.sh redis-start

2. Start local Go backend
   cd apps/backend-go
   DATABASE_HOST=<external-mysql-host> DATABASE_PORT=3306 DATABASE_NAME=dataease_dev DATABASE_USER=root DATABASE_PASSWORD=<password> REDIS_HOST=127.0.0.1 REDIS_PORT=16379 make run-local

3. Start local frontend
   cd apps/frontend
   npm install
   npm run dev

4. Visit
   Frontend: http://localhost:5173
   Backend health: http://localhost:8080/health
EOF
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
    redis-start)
        start_redis
        ;;
    redis-stop)
        stop_redis
        ;;
    local-help)
        show_local_help
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
        echo "Usage: $0 {build|build-backend|build-frontend|redis-start|redis-stop|local-help|start|stop|restart}"
        echo ""
        echo "Commands:"
        echo "  build          Build both frontend and backend"
        echo "  build-backend  Build static backend binary"
        echo "  build-frontend Build frontend dist"
        echo "  redis-start    Start only Redis for local hybrid development"
        echo "  redis-stop     Stop only Redis for local hybrid development"
        echo "  local-help     Show local frontend + local backend workflow"
        echo "  start          Start dev containers"
        echo "  stop           Stop dev containers"
        echo "  restart        Rebuild and restart"
        exit 1
        ;;
esac
