#!/bin/bash
set -euo pipefail

VERIFY_ONLY="false"
if [[ "${1:-}" == "--verify-only" ]]; then
  VERIFY_ONLY="true"
fi

: "${DB_CONTAINER:=mysql8}"
: "${DB_NAME:=dataease_dev}"
: "${DB_USER:=root}"
: "${DB_PASSWORD:=Admin168}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATION_FILE="${SCRIPT_DIR}/../migrations/mysql/20260315_fix_official_example_payload_json_valid.sql"

if [[ ! -f "${MIGRATION_FILE}" ]]; then
  echo "migration file not found: ${MIGRATION_FILE}" >&2
  exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
  echo "mysql container '${DB_CONTAINER}' is not running" >&2
  exit 1
fi

if [[ "${VERIFY_ONLY}" != "true" ]]; then
  docker exec -i "${DB_CONTAINER}" mysql -u"${DB_USER}" -p"${DB_PASSWORD}" -D "${DB_NAME}" < "${MIGRATION_FILE}"
fi

RESULT="$({
  docker exec "${DB_CONTAINER}" mysql -N -B -u"${DB_USER}" -p"${DB_PASSWORD}" -D "${DB_NAME}" -e "
SELECT id, CHAR_LENGTH(COALESCE(canvas_style_data,'')) AS style_len, CHAR_LENGTH(COALESCE(component_data,'')) AS comp_len
FROM data_visualization_info
WHERE id IN ('985188400292302870','985188400292302871')
ORDER BY id;
"
} | tr -d '\r')"

echo "${RESULT}"

if ! awk 'BEGIN{ok=1} {if ($2<=2 || $3<=2) ok=0} END{exit ok?0:1}' <<< "${RESULT}"; then
  echo "official example payload verify failed (style/component payload still empty)" >&2
  exit 1
fi

echo "official example payload verified"
