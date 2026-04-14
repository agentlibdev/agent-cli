#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BASE_URL="${AGENTLIB_BASE_URL:-http://127.0.0.1:8787}"
REF="${AGENTLIB_SMOKE_REF:-raul/code-reviewer@0.4.0}"
QUERY="${AGENTLIB_SMOKE_QUERY:-${REF#*/}}"
QUERY="${QUERY%@*}"
TARGET="${AGENTLIB_SMOKE_TARGET:-codex}"

WORKDIR="$(mktemp -d)"
cleanup() {
  rm -rf "$WORKDIR"
}
trap cleanup EXIT INT TERM

export HOME="$WORKDIR/home"
mkdir -p "$HOME"

run() {
  AGENTLIB_BASE_URL="$BASE_URL" HOME="$HOME" go run ./cmd/agentlib "$@"
}

require_contains() {
  local haystack="$1"
  local needle="$2"
  if ! printf '%s\n' "$haystack" | grep -F -- "$needle" >/dev/null; then
    printf 'expected to find %q in output\n' "$needle" >&2
    return 1
  fi
}

echo "search: $QUERY"
search_output="$(run search "$QUERY")"
printf '%s\n' "$search_output"
require_contains "$search_output" "$REF"

echo "install: $REF"
install_output="$(run install --runtime "$TARGET" "$REF")"
printf '%s\n' "$install_output"
require_contains "$install_output" "installed: ${REF}"
require_contains "$install_output" "activated: $TARGET"

echo "status: $REF"
status_output="$(run status "$REF")"
printf '%s\n' "$status_output"
require_contains "$status_output" "installed: yes"
require_contains "$status_output" "active targets: 1"
require_contains "$status_output" "$TARGET"

echo "activations: list"
activations_output="$(run activations list)"
printf '%s\n' "$activations_output"
require_contains "$activations_output" "$TARGET"
require_contains "$activations_output" "$REF"

echo "deactivate: $TARGET"
deactivate_output="$(run deactivate --target "$TARGET" "$REF")"
printf '%s\n' "$deactivate_output"
require_contains "$deactivate_output" "deactivated: ${REF} -> ${TARGET}"

echo "status after deactivate: $REF"
status_after_deactivate="$(run status "$REF")"
printf '%s\n' "$status_after_deactivate"
require_contains "$status_after_deactivate" "active targets: 0"

echo "remove: $REF"
remove_output="$(run remove "$REF")"
printf '%s\n' "$remove_output"
require_contains "$remove_output" "removed: ${REF}"

echo "status after remove: $REF"
status_after_remove="$(run status "$REF")"
printf '%s\n' "$status_after_remove"
require_contains "$status_after_remove" "installed: no"
require_contains "$status_after_remove" "active targets: 0"

echo "smoke ok"
