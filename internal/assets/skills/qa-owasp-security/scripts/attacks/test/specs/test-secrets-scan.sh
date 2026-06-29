#!/usr/bin/env bash
# test-secrets-scan.sh — secrets-scan.sh validation against a deterministic
# fixture repo seeded with one fake AWS credential pair.
#
# Asserts: exit 1 (one finding ≥ high) AND log shows non-zero gitleaks count.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../lib.sh"

NAME="secrets-scan"
LOG="/tmp/qa-validation-${NAME}.log"
FIXTURE="$HARNESS_DIR/fixtures/secrets-fixture"
SCRIPT="$ATTACKS_DIR/secrets-scan.sh"
OUT_DIR="/tmp/qa-validation-${NAME}-out"

rm -rf "$OUT_DIR"
log_info "[$NAME] initialising fixture"
bash "$FIXTURE/init-fixture.sh" >/dev/null

log_info "[$NAME] running $SCRIPT --target $FIXTURE"
set +e
"$SCRIPT" --target "$FIXTURE" --out "$OUT_DIR" > "$LOG" 2>&1
ACTUAL_EXIT=$?
set -e

FAILED=0
assert_exit_code 1 "$ACTUAL_EXIT" "$NAME: exits 1 on findings" || FAILED=1
assert_grep 'gitleaks findings: [1-9]' "$LOG" "$NAME: gitleaks reports >= 1 finding" || FAILED=1

if [[ "$FAILED" -ne 0 ]]; then
  log_warn "[$NAME] log tail follows:"
  tail -n 30 "$LOG" >&2 || true
  exit 1
fi
exit 0
