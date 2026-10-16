#!/usr/bin/env bash
#
# Enforce per-module Go test coverage thresholds.
#
# Each backend service is an independent Go module and must independently meet
# the threshold for the statements it owns. Packages that cannot be exercised
# without a live Cassandra — and whose coverage would otherwise dilute the
# signal — are excluded from the denominator:
#
#   * cmd/...                 thin main() entry points
#   * internal/db/...         embedded CQL schema / connection wiring
#   * internal/repositories/  Cassandra CQL queries and row mapping
#
# Usage:
#   ./scripts/coverage/backend.sh packages/backend/auth [threshold]
#   ./scripts/coverage/backend.sh --all [threshold]
#
# The threshold defaults to 80 (percent). Exit code is non-zero if any module
# falls below the threshold or a test fails.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

EXCLUDE_RE='/(cmd|internal/db|internal/repositories)/'
THRESHOLD="${THRESHOLD:-80}"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

check_module() {
  local module_dir="$1"
  local profile="${tmp_dir}/$(basename "${module_dir}").out"
  local filtered="${profile}.filtered"

  if ! ( cd "${REPO_ROOT}/${module_dir}" && go test ./... -coverprofile="${profile}" -covermode=atomic >/dev/null ); then
    printf 'FAIL: go test failed in %s\n' "${module_dir}" >&2
    return 1
  fi

  if [ ! -s "${profile}" ]; then
    printf 'FAIL: no coverage profile produced for %s\n' "${module_dir}" >&2
    return 1
  fi

  { head -n 1 "${profile}"; grep -vE "${EXCLUDE_RE}|^mode:" "${profile}"; } > "${filtered}"

  local total
  total="$(cd "${REPO_ROOT}/${module_dir}" && go tool cover -func="${filtered}" |
    tail -n 1 | grep -oE '[0-9]+\.[0-9]+%' | tr -d '%')"

  printf '%-14s %6s%% (>= %s%%)\n' "$(basename "${module_dir}")" "${total}" "${THRESHOLD}"

  if ! awk -v t="${total}" -v thr="${THRESHOLD}" 'BEGIN { exit !(t + 0 >= thr + 0) }'; then
    printf 'FAIL: %s coverage %s%% is below %s%%\n' "$(basename "${module_dir}")" "${total}" "${THRESHOLD}" >&2
    return 1
  fi
}

status=0

if [ "${1:-}" = "--all" ]; then
  threshold="${2:-}"
  [ -n "${threshold}" ] && THRESHOLD="${threshold}"
  for dir in "${REPO_ROOT}"/packages/backend/*/; do
    check_module "packages/backend/$(basename "${dir}")" || status=1
  done
else
  [ "$#" -ge 1 ] || { echo "usage: $0 <module-dir> [threshold] | --all [threshold]" >&2; exit 2; }
  THRESHOLD="${2:-${THRESHOLD}}"
  check_module "$1" || status=1
fi

exit "${status}"
