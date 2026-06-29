#!/usr/bin/env bash
# run-validation.sh — top-level harness for the qa-owasp-security attack
# scripts. Boots DVWA, runs every spec under specs/, tallies pass/fail,
# tears DVWA down on exit (including Ctrl+C).
set -euo pipefail

HARNESS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck disable=SC1091
source "$HARNESS_DIR/lib.sh"

# ---------- pre-flight ----------------------------------------------------

REQUIRED=(docker gitleaks sqlmap trivy node)
MISSING=()
for tool in "${REQUIRED[@]}"; do
  command -v "$tool" >/dev/null 2>&1 || MISSING+=("$tool")
done
if [[ "${#MISSING[@]}" -gt 0 ]]; then
  log_fail "preflight" "missing required tools: ${MISSING[*]}"
  cat <<HINT >&2
Install hints (macOS / Homebrew):
  brew install gitleaks sqlmap trivy node
  brew install --cask docker            # Docker Desktop
HINT
  exit 2
fi

if ! command -v dalfox >/dev/null 2>&1; then
  log_warn "dalfox not installed — xss-scan.sh will exercise the ZAP Docker fallback (slower)"
fi

if ! docker info >/dev/null 2>&1; then
  log_fail "preflight" "docker daemon is not running"
  exit 2
fi

# ---------- lifecycle -----------------------------------------------------

trap dvwa_down EXIT

dvwa_up

# ---------- run specs -----------------------------------------------------

tally_init

shopt -s nullglob
SPECS=("$HARNESS_DIR"/specs/test-*.sh)
shopt -u nullglob
IFS=$'\n' SPECS=($(printf "%s\n" "${SPECS[@]}" | sort))
unset IFS

if [[ "${#SPECS[@]}" -eq 0 ]]; then
  log_fail "harness" "no specs found under $HARNESS_DIR/specs"
  exit 1
fi

for spec in "${SPECS[@]}"; do
  name=$(basename "$spec" .sh)
  printf "\n%s>>>>%s running %s\n" "$C_CYAN" "$C_RESET" "$name"
  set +e
  bash "$spec"
  rc=$?
  set -e
  if [[ "$rc" -eq 0 ]]; then
    tally_record pass
  else
    tally_record fail
  fi
done

# ---------- report --------------------------------------------------------

if tally_report; then
  exit 0
fi
exit 1
