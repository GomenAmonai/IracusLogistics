#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_BASE="${API_BASE:-http://127.0.0.1:8080/api}"
FRONTEND_BASE="${FRONTEND_BASE:-http://127.0.0.1:5173}"

for command in curl jq go; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Требуется команда: $command" >&2
    exit 1
  fi
done

expect_jq() {
  local json="$1"
  local filter="$2"
  local message="$3"

  if ! jq -e "$filter" >/dev/null <<<"$json"; then
    echo "FAIL: $message" >&2
    exit 1
  fi
  echo "PASS: $message"
}

seed_output="$(cd "$ROOT_DIR/backend" && go run ./cmd/seed-demo -frontend-url "$FRONTEND_BASE")"
client_url="$(sed -n 's/^Клиент: //p' <<<"$seed_output")"
empty_client_url="$(sed -n 's/^Клиент без грузов: //p' <<<"$seed_output")"
client_token="${client_url#*token=}"
empty_client_token="${empty_client_url#*token=}"

if [[ -z "$client_token" || "$client_token" == "$client_url" ]]; then
  echo "FAIL: seed-demo не вернул client-токен" >&2
  exit 1
fi

if [[ -z "$empty_client_token" || "$empty_client_token" == "$empty_client_url" ]]; then
  echo "FAIL: seed-demo не вернул токен клиента без грузов" >&2
  exit 1
fi

health="$(curl --fail --silent --show-error "$API_BASE/health")"
expect_jq "$health" '.status == "ok" and .database == "ok"' "API и PostgreSQL доступны"

for page in / /manager.html /webapp.html; do
  curl --fail --silent --show-error --output /dev/null "$FRONTEND_BASE$page"
done
echo "PASS: три frontend entrypoint доступны"

login="$(curl --fail --silent --show-error \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@icaris.local","password":"demo-local-only"}' \
  "$API_BASE/auth/login")"
manager_token="$(jq -er '.token' <<<"$login")"
echo "PASS: demo-менеджер авторизован"

manager_shipments="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $manager_token" \
  "$API_BASE/shipments")"
shipment_id="$(jq -er '.[] | select(.tracking_key == "ICR-DEM0000000") | .id' <<<"$manager_shipments")"
echo "PASS: менеджер видит demo-груз"

manager_messages="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $manager_token" \
  "$API_BASE/shipments/$shipment_id/messages")"
expect_jq "$manager_messages" 'any(.[]; .text | contains("[icaris-local-demo]"))' "менеджер видит demo-чат"

client_shipments="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $client_token" \
  "$API_BASE/app/shipments")"
expect_jq "$client_shipments" 'any(.[]; .tracking_key == "ICR-DEM0000000")' "клиент видит только свой demo-груз"

empty_client_shipments="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $empty_client_token" \
  "$API_BASE/app/shipments")"
expect_jq "$empty_client_shipments" 'length == 0' "новый demo-клиент видит корректное пустое состояние"

client_detail="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $client_token" \
  "$API_BASE/app/shipments/$shipment_id")"
expect_jq "$client_detail" '.shipment.tracking_key == "ICR-DEM0000000" and (.history | length >= 2)' "клиент получает детали и историю статусов"

client_messages="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $client_token" \
  "$API_BASE/app/shipments/$shipment_id/messages")"
expect_jq "$client_messages" 'any(.[]; .from_role == "client") and any(.[]; .from_role == "manager")' "клиент получает обе стороны переписки"

client_payments="$(curl --fail --silent --show-error \
  -H "Authorization: Bearer $client_token" \
  "$API_BASE/app/shipments/$shipment_id/payments")"
expect_jq "$client_payments" 'any(.[]; .amount == "1200" and .channel == "bank_transfer")' "клиент получает demo-платёж"

echo "Smoke demo завершён успешно."
