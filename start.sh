#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="${ROOT_DIR}/.snekproxy.pid"
LOG_FILE="${ROOT_DIR}/.snekproxy.log"

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "SnekProxy already running with PID $(cat "$PID_FILE")"
  exit 0
fi

cd "$ROOT_DIR"
nohup go run ./cmd/snekproxy --enable --config ./config.yaml > "$LOG_FILE" 2>&1 &
PID=$!
printf '%s\n' "$PID" > "$PID_FILE"

sleep 1
if kill -0 "$PID" 2>/dev/null; then
  echo "Started SnekProxy (PID $PID)"
  echo "Logs: $LOG_FILE"
else
  echo "Failed to start SnekProxy" >&2
  rm -f "$PID_FILE"
  exit 1
fi
