#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="${ROOT_DIR}/.snekproxy.pid"

if [ -f "$PID_FILE" ]; then
  PID="$(cat "$PID_FILE")"
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID"
    sleep 1
    if kill -0 "$PID" 2>/dev/null; then
      kill -9 "$PID"
    fi
    echo "Stopped SnekProxy (PID $PID)"
  else
    echo "No running SnekProxy process found for PID $PID"
  fi
  rm -f "$PID_FILE"
  exit 0
fi

PID="$(pgrep -af "go run ./cmd/snekproxy|cmd/snekproxy" | head -n 1 | awk '{print $1}' || true)"
if [ -n "$PID" ]; then
  kill "$PID"
  echo "Stopped SnekProxy (PID $PID)"
  exit 0
fi

echo "SnekProxy is not running"
