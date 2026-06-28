#!/usr/bin/env bash
# test-sqli-test.sh — sqli-test.sh validation against DVWA's classic GET-based
# SQLi endpoint. Requires DVWA up and a logged-in PHPSESSID with security=low.
#
# Findings count is intentionally NOT asserted (DVWA injection success can
# be flaky depending on session state). The harness validates that the
# SCRIPT RUNS CORRECTLY — exit 3 (runtime error / sqlmap missing) is the
# failure mode we guard against. Both exit 0 and exit 1 are accepted.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../lib.sh"

NAME="sqli-test"
LOG="/tmp/qa-validation-${NAME}.log"
SCRIPT="$ATTACKS_DIR/sqli-test.sh"
OUT_DIR="/tmp/qa-validation-${NAME}-out"

log_info "[$NAME] logging in to DVWA"
PHPSESSID=$(dvwa_login)

rm -rf "$OUT_DIR"
TARGET="http://localhost:8080/vulnerabilities/sqli/?id=1&Submit=Submit"
COOKIE="PHPSESSID=$PHPSESSID; security=low"
log_info "[$NAME] running sqli-test.sh against $TARGET"
set +e
"$SCRIPT" --target "$TARGET" --cookie "$COOKIE" --risk 3 --level 3 \
  --out "$OUT_DIR" > "$LOG" 2>&1
ACTUAL_EXIT=$?
set -e

FAILED=0
case "$ACTUAL_EXIT" in
  0|1) log_pass "$NAME: scan completed (exit=$ACTUAL_EXIT)" ;;
  2)   log_fail "$NAME: scan completed (no tool-missing error)" "exit 2 = sqlmap not installed"; FAILED=1 ;;
  *)   log_fail "$NAME: scan completed (no runtime error)" "unexpected exit $ACTUAL_EXIT"; FAILED=1 ;;
esac

assert_grep 'sqlmap exit code:' "$LOG" \
  "$NAME: sqlmap was invoked (regression guard)" || FAILED=1

if [[ "$FAILED" -ne 0 ]]; then
  log_warn "[$NAME] log tail follows:"
  tail -n 40 "$LOG" >&2 || true
  exit 1
fi
exit 0
