#!/bin/bash
set -e

# --- Wait for Redis ---
REDIS_HOST="${REDIS_HOST:-redis}" # Default to 'redis' if env var not set
REDIS_PORT="${REDIS_PORT:-6379}" # Default to 6379 if env var not set

echo "Waiting for Redis at ${REDIS_HOST}:${REDIS_PORT}..."
until nc -z "${REDIS_HOST}" "${REDIS_PORT}"; do
  echo "Redis is unavailable - sleeping"
  sleep 1
done
echo "Redis is ready for crawler!"

echo "Setup complete. Starting crawler binary."
