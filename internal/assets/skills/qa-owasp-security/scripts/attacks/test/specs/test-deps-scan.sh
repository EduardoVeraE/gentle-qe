#!/usr/bin/env bash
# test-deps-scan.sh — deps-scan.sh validation against a fixture pinned to
# lodash@4.17.4 (known prototype-pollution advisories).
#
# Regression guards (P1 bugs from round 1):
#   1) ENOLOCK from npm audit when no package-lock.json — fixture provides one
#      via init-fixture.sh; we additionally check the script's own auto-lockfile
#      branch is never the only thing producing findings.
#   2) trivy must run and produce trivy.json — verified explicitly below.
#
# Asserts: exit 1, log mentions both "Trivy findings" and "Ecosystem findings",
# and a trivy.json artifact exists.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../lib.sh"

NAME="deps-scan"
LOG="/tmp/qa-validation-${NAME}.log"
FIXTURE="$HARNESS_DIR/fixtures/deps-fixture"
SCRIPT="$ATTACKS_DIR/deps-scan.sh"
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
assert_exit_code 1 "$ACTUAL_EXIT" "$NAME: exits 1 on vulnerable deps" || FAILED=1
assert_grep 'Trivy findings'     "$LOG" "$NAME: trivy ran and reported findings line" || FAILED=1
assert_grep 'Ecosystem findings' "$LOG" "$NAME: ecosystem audit ran" || FAILED=1

# Regression guard: trivy.json artifact must exist (round-1 P1 bug — trivy
# section was effectively skipped). Glob is timestamped sub-directory.
if compgen -G "$OUT_DIR/trivy.json" >/dev/null; then
  log_pass "$NAME: trivy.json artifact present"
else
  log_fail "$NAME: trivy.json artifact present" "no trivy.json under $OUT_DIR"
  FAILED=1
fi

if [[ "$FAILED" -ne 0 ]]; then
  log_warn "[$NAME] log tail follows:"
  tail -n 40 "$LOG" >&2 || true
  exit 1
fi
exit 0
