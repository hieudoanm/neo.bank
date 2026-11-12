#!/usr/bin/env bash
#
# install.sh — One-command local setup for neo.bank.
#
# Creates .env from .env.example with a generated development RSA key pair,
# then starts the full stack with Docker Compose.
#
# Usage:
#   ./scripts/install.sh
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
    echo "Creating .env with a generated development RSA key pair..."

    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT

    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$TMP_DIR/private.pem" 2>/dev/null
    openssl rsa -in "$TMP_DIR/private.pem" -pubout -out "$TMP_DIR/public.pem" 2>/dev/null

    PRIV_PEM="$(awk 'BEGIN{ORS="\\n"}{print}' "$TMP_DIR/private.pem")"
    PUB_PEM="$(awk 'BEGIN{ORS="\\n"}{print}' "$TMP_DIR/public.pem")"

    awk -v priv="$PRIV_PEM" -v pub="$PUB_PEM" '
        /^JWT_PRIVATE_KEY_PEM=/ { print "JWT_PRIVATE_KEY_PEM=\"" priv "\""; next }
        /^JWT_PUBLIC_KEY_PEM=/ { print "JWT_PUBLIC_KEY_PEM=\"" pub "\""; next }
        { print }
    ' .env.example > .env

    echo "Wrote .env (development keys only — never use for real deployments)."
else
    echo ".env already exists — leaving it untouched."
fi

echo "Starting the stack with Docker Compose..."
docker compose up --build -d

echo
echo "neo.bank is running:"
echo "  Frontend : http://localhost:3000"
echo "  API      : http://localhost:8080 (/healthz, /api/v1)"
echo
echo "Manage with: make compose-logs / make compose-down"
