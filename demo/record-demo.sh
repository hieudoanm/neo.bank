#!/usr/bin/env bash
#
# Record the NEO Bank end-to-end demo into demo/demo.mp4.
#
# The script boots the full stack with docker compose, waits until the backend
# really works, then hands the UI to Playwright. Playwright registers two
# customers, transfers money and verifies both balances; the resulting video is
# converted to demo/demo.mp4.
#
# Requirements: Docker Compose v2, pnpm, Node, and ffmpeg
#   (system ffmpeg is preferred; Playwright's bundled copy is used as fallback).
#
# Usage:
#   demo/record-demo.sh                  # record, then tear the stack down
#   KEEP_STACK=1 demo/record-demo.sh     # leave the stack running afterwards
#   DEMO_SLOWMO=600 demo/record-demo.sh  # slower actions -> longer, clearer video
#   FRONTEND_PORT=3001 demo/record-demo.sh  # if :3000 is taken (also set
#                                           # CORS_ORIGINS=http://localhost:3001)
#   DEMO_BASE_URL=http://host:3000 demo/record-demo.sh  # run against another stack
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEMO_DIR="$ROOT/demo"
APP_DIR="$ROOT/packages/frontend/app"
COMPOSE_FILE="$ROOT/docker-compose.yaml"
OUTPUT="$DEMO_DIR/demo.mp4"

# The frontend host port is configurable (docker-compose.yaml maps
# ${FRONTEND_PORT:-3000}); Playwright drives that same URL.
: "${FRONTEND_PORT:=3000}"
export FRONTEND_PORT
: "${DEMO_BASE_URL:=http://localhost:${FRONTEND_PORT}}"
export DEMO_BASE_URL

log() { printf '\n\033[1m==>\033[0m %s\n' "$*"; }

find_ffmpeg() {
  if command -v ffmpeg >/dev/null 2>&1; then
    echo "ffmpeg"
    return 0
  fi
  local cache="${PLAYWRIGHT_BROWSERS_PATH:-$HOME/Library/Caches/ms-playwright}"
  local bin
  bin="$(find "$cache" -maxdepth 3 -type f \
    \( -name 'ffmpeg-mac' -o -name 'ffmpeg-linux' -o -name 'ffmpeg-win64.exe' \) \
    2>/dev/null | head -n 1 || true)"
  if [ -n "$bin" ]; then
    echo "$bin"
    return 0
  fi
  return 1
}

# Wait until the backend is actually functional: a register attempt must return
# 201 (created) or 409 (already registered). 5xx/000 mean the services are
# still booting (Cassandra boot + migrations + Kafka), so we keep waiting.
wait_for_backend() {
  local base="$DEMO_BASE_URL"
  local gateway="${DEMO_GATEWAY_URL:-http://localhost:8080}"
  local email="readiness-check@example.com"
  local body code i

  log "Waiting for the frontend on $base"
  for i in $(seq 1 90); do
    if curl -fsS -o /dev/null "$base/" 2>/dev/null; then
      break
    fi
    printf '  still starting the frontend (%s/90)\n' "$i"
    sleep 3
  done

  log "Waiting for the backend on $gateway (register probe)"
  body="$(printf '{"email":"%s","password":"Readiness-Passw0rd!"}' "$email")"
  for i in $(seq 1 120); do
    code="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$gateway/api/v1/register" \
      -H 'Content-Type: application/json' -d "$body" || echo 000)"
    case "$code" in
      201 | 409)
        printf '  backend is ready (HTTP %s)\n' "$code"
        return 0
        ;;
    esac
    printf '  still starting the backend (HTTP %s, %s/120)\n' "$code" "$i"
    sleep 5
  done

  printf 'Backend never became ready; check "docker compose -f %s logs"\n' "$COMPOSE_FILE" >&2
  return 1
}

log "Starting the NEO Bank stack (docker compose up -d --build)"
docker compose -f "$COMPOSE_FILE" up -d --build

wait_for_backend

if [ ! -x "$APP_DIR/node_modules/.bin/playwright" ]; then
  log "Installing frontend dependencies"
  (cd "$APP_DIR" && pnpm install --no-frozen-lockfile)
fi

log "Installing Playwright's Chromium (first run only)"
(cd "$APP_DIR" && pnpm e2e:install >/dev/null)

log "Recording the Playwright demo (DEMO_SLOWMO=${DEMO_SLOWMO:-300})"
set +e
(cd "$APP_DIR" && pnpm e2e)
STATUS=$?
set -e

VIDEO="$(find "$APP_DIR/test-results" -type f -name '*.webm' -print -quit 2>/dev/null || true)"
FFMPEG="$(find_ffmpeg || true)"

if [ -n "$VIDEO" ] && [ -n "$FFMPEG" ]; then
  log "Converting the recording to demo/demo.mp4"
  "$FFMPEG" -y -loglevel error -i "$VIDEO" \
    -c:v libx264 -pix_fmt yuv420p -movflags +faststart "$OUTPUT"
  printf '  %s\n' "$OUTPUT"
elif [ -n "$VIDEO" ]; then
  printf '\n\033[33mwarning:\033[0m ffmpeg not found; left the raw recording at %s\n' "$VIDEO" >&2
else
  printf '\n\033[33mwarning:\033[0m no recording found under %s/test-results\n' "$APP_DIR" >&2
fi

if [ "${KEEP_STACK:-0}" = "1" ]; then
  log "Leaving the stack running (KEEP_STACK=1)"
else
  log "Tearing the stack down (set KEEP_STACK=1 to keep it)"
  docker compose -f "$COMPOSE_FILE" down
fi

exit "$STATUS"
