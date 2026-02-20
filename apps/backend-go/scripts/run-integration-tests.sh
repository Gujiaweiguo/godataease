#!/bin/bash
set -e

: "${TEST_DB_HOST:=localhost}"
: "${TEST_DB_PORT:=3306}"
: "${TEST_DB_USER:=root}"
: "${TEST_DB_PASSWORD:=Admin168}"
: "${TEST_DB_NAME:=dataease_test}"

echo "Running integration tests with MySQL 8..."
echo "Database: ${TEST_DB_HOST}:${TEST_DB_PORT}/${TEST_DB_NAME}"

export TEST_DB_HOST
export TEST_DB_PORT
export TEST_DB_USER
export TEST_DB_PASSWORD
export TEST_DB_NAME

cd "$(dirname "$0")/.."

go test -tags=integration -v ./internal/repository/... -count=1
