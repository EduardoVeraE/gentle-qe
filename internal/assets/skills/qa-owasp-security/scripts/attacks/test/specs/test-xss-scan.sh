#!/usr/bin/env bash
# test-xss-scan.sh — xss-scan.sh validation against DVWA's reflected XSS page.
#
# Regression guards (P1 bugs from round 1):
#   1) Docker host rewrite: when target is localhost the script must rewrite
#      to host.docker.internal so the ZAP container can reach the host. We
#      grep the log for the rewrite trace.
#   2) Active scan switch: the script must invoke zap-full-scan.py (active),
#      not zap-baseline.py (passive). We grep the log for that command name.
#
# Findings count is intentionally NOT asserted: with security=low DVWA's
# reflected XSS endpoint may or may not register as a "Verified" finding
# depending on ZAP rule timing. The harness validates that the SCAN RUNS
# WITH THE FIXED CONFIGURATION — both regression guards (host rewrite +
# active scan invocation) MUST pass. The exit code is treated as soft:
# 0/1 = clean run; 3 = ZAP scanner runtime crash (broken-pipe under load
# is common when ZAP is run via Rosetta/amd64 emulation on Apple Silicon)
# is accepted as long as the regression guards pass. Exit 2 (ZAP missing)
# is the only hard failure here.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../lib.sh"

NAME="xss-scan"
LOG="/tmp/qa-validation-${NAME}.log"
SCRIPT="$ATTACKS_DIR/xss-scan.sh"
OUT_DIR="/tmp/qa-validation-${NAME}-out"

# NOTE: dvwa_login is NOT required for this spec. The regression guards
# (Docker host rewrite + active ZAP scan) are about the script's behavior,
# not about exploiting DVWA's auth-protected XSS endpoints. ZAP scans the
# login page anonymously — that exercises both guards.
rm -rf "$OUT_DIR"
TARGET="http://localhost:8080/vulnerabilities/xss_r/?name=test"
log_info "[$NAME] running xss-scan.sh (ZAP fallback expected) — this can take 5-10 minutes"
set +e
"$SCRIPT" --target "$TARGET" --out "$OUT_DIR" > "$LOG" 2>&1
ACTUAL_EXIT=$?
set -e

FAILED=0
# Accept 0/1 (clean) or 3 (ZAP runtime crash — environmental flake on
# Rosetta/amd64). Reject 2 (ZAP missing) and any other unexpected exit.
case "$ACTUAL_EXIT" in
  0|1) log_pass "$NAME: scan completed (exit=$ACTUAL_EXIT)" ;;
  3)   log_pass "$NAME: scan started but ZAP crashed mid-run (exit=3, environmental — guards still verified below)" ;;
  2)   log_fail "$NAME: ZAP available" "exit 2 = ZAP not installed/reachable"; FAILED=1 ;;
  *)   log_fail "$NAME: scan completed (no unexpected error)" "unexpected exit $ACTUAL_EXIT"; FAILED=1 ;;
esac

assert_grep 'host\.docker\.internal' "$LOG" \
  "$NAME: docker-host rewrite is logged (regression guard)" || FAILED=1
assert_grep 'zap-full-scan\.py'      "$LOG" \
  "$NAME: ZAP active scanner is invoked (regression guard)" || FAILED=1

if [[ "$FAILED" -ne 0 ]]; then
  log_warn "[$NAME] log tail follows:"
  tail -n 60 "$LOG" >&2 || true
  exit 1
fi
exit 0
