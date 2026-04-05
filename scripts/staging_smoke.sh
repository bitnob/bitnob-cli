#!/bin/bash
set -euo pipefail

BIN="${BIN:-./bitnob}"
PROFILE="${PROFILE:-staging}"
RUN_WRITE="${RUN_WRITE:-0}"

CLIENT_ID="${BITNOB_CLIENT_ID:-}"
SECRET_KEY="${BITNOB_SECRET_KEY:-}"

if [[ -z "${CLIENT_ID}" || -z "${SECRET_KEY}" ]]; then
  echo "BITNOB_CLIENT_ID and BITNOB_SECRET_KEY are required"
  exit 1
fi

if [[ ! -x "${BIN}" ]]; then
  echo "binary not found or not executable: ${BIN}"
  exit 1
fi

WORKDIR="$(mktemp -d /tmp/bitnob-staging-smoke.XXXXXX)"
trap 'rm -rf "${WORKDIR}"' EXIT

export BITNOB_CONFIG_PATH="${WORKDIR}/config.json"
export BITNOB_STATE_DIR="${WORKDIR}/state"
export BITNOB_SECRET_BACKEND="file"
export BITNOB_CLIENT_ID="${CLIENT_ID}"
export BITNOB_SECRET_KEY="${SECRET_KEY}"

run() {
  local label="$1"
  shift
  echo
  echo "==> ${label}"
  "$@"
}

run "version" "${BIN}" version
run "login" "${BIN}" login --profile "${PROFILE}"
run "whoami" "${BIN}" whoami
run "balances" "${BIN}" balances
run "wallets" "${BIN}" wallets
run "trading prices" "${BIN}" trading prices
run "payout limits" "${BIN}" payouts limits

if [[ "${RUN_WRITE}" == "1" ]]; then
  TS="$(date +%s)"
  EMAIL="smoke+${TS}@example.com"
  run "customers create (write probe)" \
    "${BIN}" customers create \
    --customer-type individual \
    --email "${EMAIL}" \
    --first-name Smoke \
    --last-name Test
else
  echo
  echo "==> write probe skipped (set RUN_WRITE=1 to enable)"
fi

run "logout" "${BIN}" logout

echo
echo "staging smoke suite: PASS"
