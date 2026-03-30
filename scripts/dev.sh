#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

build_backend() {
    echo "构建后端 (static)..."
    cd apps/backend-go
    make build-static
    cd ../..
}

build_frontend() {
    echo "构建前端..."
    cd apps/frontend
    npm run build:base
    cd ../..
}

build_all() {
    build_backend
    build_frontend
}

# ==========================================
# 开发环境命令
# ==========================================

dev_infra_start() {
    echo -e "${GREEN}启动开发环境基础设施 (Redis)...${NC}"
    docker compose -f infra/compose/docker-compose.dev.yml --env-file infra/compose/.env.dev up -d
    echo "等待 Redis 就绪..."
    sleep 3
    echo -e "${GREEN}开发基础设施就绪：${NC}"
    echo "  MySQL: 共用 my-net 中的 mysql8 容器 (127.0.0.1:3306)"
    echo "  Redis: 127.0.0.1:16379"
}

dev_infra_stop() {
    echo "停止开发环境基础设施..."
    docker compose -f infra/compose/docker-compose.dev.yml --env-file infra/compose/.env.dev down
}

dev_backend() {
    echo -e "${GREEN}启动本地 Go 后端...${NC}"
    cd apps/backend-go
    DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 DATABASE_NAME=dataease_dev \
    DATABASE_USER=root DATABASE_PASSWORD=Admin168 \
    REDIS_HOST=127.0.0.1 REDIS_PORT=16379 \
    make run-local
}

dev_frontend() {
    echo -e "${GREEN}启动本地前端...${NC}"
    cd apps/frontend
    npm run dev
}

dev_start() {
    dev_infra_start
    echo ""
    echo -e "${YELLOW}请在新终端中分别执行：${NC}"
    echo "  后端: ./scripts/dev.sh dev-backend"
    echo "  前端: ./scripts/dev.sh dev-frontend"
    echo ""
    echo "访问地址："
    echo "  前端: http://localhost:5173"
    echo "  后端: http://localhost:8080/health"
    echo "  API文档: http://localhost:8080/doc.html"
}

dev_stop() {
    dev_infra_stop
}

# ==========================================
# 正式环境命令
# ==========================================

prod_build() {
    echo -e "${GREEN}构建生产镜像...${NC}"
    build_all
    docker build -t godataease:latest .
    echo -e "${GREEN}镜像构建完成: godataease:latest${NC}"
}

prod_start() {
    echo -e "${GREEN}启动正式环境...${NC}"
    docker compose -f infra/compose/docker-compose.prod.yml --env-file infra/compose/.env.prod up -d
    echo "等待服务就绪..."
    sleep 10
    echo -e "${GREEN}正式环境就绪：${NC}"
    echo "  应用: http://localhost:8080"
    echo "  Nginx: http://localhost:80"
}

prod_stop() {
    echo "停止正式环境..."
    docker compose -f infra/compose/docker-compose.prod.yml --env-file infra/compose/.env.prod down
}

# ==========================================
# 帮助
# ==========================================

show_help() {
    cat <<'EOF'
DataEase 开发与部署脚本

用法: ./scripts/dev.sh <command>

开发环境命令：
  dev-infra          启动开发基础设施（Redis + 共用 MySQL）
  dev-infra-stop     停止开发基础设施
  dev-backend        启动本地 Go 后端
  dev-frontend       启动本地前端
  dev-start          启动开发环境（基础设施，提示手动启动前后端）
  dev-stop           停止开发环境

正式环境命令：
  prod-build         构建生产镜像
  prod-start         启动正式环境
  prod-stop          停止正式环境

通用命令：
  build              构建前后端
  build-backend      仅构建后端
  build-frontend     仅构建前端

配置文件：
  开发环境: infra/compose/.env.dev
  正式环境: infra/compose/.env.prod

默认访问地址（开发环境）：
  前端: http://localhost:5173
  后端: http://localhost:8080/health
  API文档: http://localhost:8080/doc.html

默认登录凭据：
  用户名: admin
  密码: admin123
EOF
}

case "${1:-}" in
    dev-infra|dev-infra-start)
        dev_infra_start
        ;;
    dev-infra-stop)
        dev_infra_stop
        ;;
    dev-backend)
        dev_backend
        ;;
    dev-frontend)
        dev_frontend
        ;;
    dev-start)
        dev_start
        ;;
    dev-stop)
        dev_stop
        ;;
    prod-build)
        prod_build
        ;;
    prod-start)
        prod_start
        ;;
    prod-stop)
        prod_stop
        ;;
    build)
        build_all
        ;;
    build-backend)
        build_backend
        ;;
    build-frontend)
        build_frontend
        ;;
    *)
        show_help
        exit 1
        ;;
esac
