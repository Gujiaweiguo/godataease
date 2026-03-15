#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${BACKEND_DIR}/.." && pwd)"
FRONTEND_DIR="${REPO_ROOT}/frontend"

: "${SERVER_PORT:=8080}"
: "${FRONTEND_DIST_PATH:=${FRONTEND_DIR}/dist}"
: "${LOG_FILE:=/tmp/godataease-backend-${SERVER_PORT}.log}"

BUILD_FRONTEND="false"
FOREGROUND="false"
REPLACE="false"

usage() {
  cat <<EOF
Usage: $(basename "$0") [--build-frontend] [--replace] [--foreground]

Options:
  --build-frontend  Build apps/frontend/dist before starting backend
  --replace         Stop the current listener on SERVER_PORT before starting
  --foreground      Run in foreground instead of nohup background mode

Environment:
  SERVER_PORT       Backend port (default: 8080)
  FRONTEND_DIST_PATH  Frontend dist path (default: apps/frontend/dist)
  LOG_FILE          Log file for background mode (default: /tmp/godataease-backend-<port>.log)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build-frontend)
      BUILD_FRONTEND="true"
      shift
      ;;
    --replace)
      REPLACE="true"
      shift
      ;;
    --foreground)
      FOREGROUND="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "${BUILD_FRONTEND}" == "true" ]]; then
  echo "[INFO] Building frontend assets..."
  (cd "${FRONTEND_DIR}" && npm run build:base)
fi

if [[ ! -f "${FRONTEND_DIST_PATH}/index.html" ]]; then
  echo "frontend dist not found: ${FRONTEND_DIST_PATH}/index.html" >&2
  echo "Run with --build-frontend or set FRONTEND_DIST_PATH correctly." >&2
  exit 1
fi

if command -v lsof >/dev/null 2>&1; then
  LISTENER_PIDS="$(lsof -ti TCP:${SERVER_PORT} -sTCP:LISTEN || true)"
else
  LISTENER_PIDS=""
fi

if [[ -n "${LISTENER_PIDS}" ]]; then
  if [[ "${REPLACE}" != "true" ]]; then
    echo "port ${SERVER_PORT} is already in use by: ${LISTENER_PIDS}" >&2
    echo "Use --replace to stop the existing listener first." >&2
    exit 1
  fi

  echo "[INFO] Stopping existing listener(s) on port ${SERVER_PORT}: ${LISTENER_PIDS}"
  kill ${LISTENER_PIDS}
  sleep 1
fi

START_CMD="SERVER_PORT=${SERVER_PORT} FRONTEND_DIST_PATH=${FRONTEND_DIST_PATH} make run"

if [[ "${FOREGROUND}" == "true" ]]; then
  echo "[INFO] Starting backend in foreground on port ${SERVER_PORT}"
  cd "${BACKEND_DIR}"
  exec sh -c "${START_CMD}"
fi

echo "[INFO] Starting backend in background on port ${SERVER_PORT}"
nohup sh -c "cd '${BACKEND_DIR}' && ${START_CMD}" > "${LOG_FILE}" 2>&1 &
BG_PID=$!

for _ in $(seq 1 30); do
  if curl -fsS "http://localhost:${SERVER_PORT}/api/ping" >/dev/null 2>&1; then
    echo "[INFO] Backend is ready on http://localhost:${SERVER_PORT}"
    echo "[INFO] PID: ${BG_PID}"
    echo "[INFO] Log: ${LOG_FILE}"
    exit 0
  fi
  sleep 1
done

echo "backend did not become ready in time; check log: ${LOG_FILE}" >&2
exit 1
