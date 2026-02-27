#!/usr/bin/env bash
set -euo pipefail

MYSQL_CONTAINER="${MYSQL_CONTAINER:-mysql8}"
MYSQL_HOST="${MYSQL_HOST:-localhost}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-Admin168}"
MYSQL_DATABASE="${MYSQL_DATABASE:-dataease_dev}"

SQL_DIR="apps/backend-go/migrations/mysql"

run_sql() {
  local file="$1"
  docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -D "$MYSQL_DATABASE" < "$file"
}

run_sql "$SQL_DIR/20260225_create_sys_role_perm.sql"
run_sql "$SQL_DIR/20260225_seed_sys_perm_baseline.sql"

docker exec "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -D "$MYSQL_DATABASE" -e "SELECT COUNT(*) AS perm_count FROM sys_perm WHERE del_flag=0; SELECT COUNT(*) AS role_perm_count FROM sys_role_perm;"

echo "compat migrations applied"
