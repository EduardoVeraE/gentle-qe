# lib.sh — shared helpers for the qa-owasp-security validation harness.
# Source it; do not execute it. Callers must `set -euo pipefail` themselves.
#
# CRITICAL DESIGN RULE
# --------------------
# NEVER pipe a script's output into another command and capture $? — the
# pipe captures the LAST command's exit status, not the script under test.
# This bug burned us once already (the entire reason this harness exists).
#
# Always do:
#     set +e
#     <script_under_test> > "$LOG" 2>&1
#     ACTUAL_EXIT=$?
#     set -e
#
# `assert_exit_code` below assumes the caller already captured $? this way.

# Resolve repo paths once, regardless of caller cwd.
HARNESS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ATTACKS_DIR="$(cd "$HARNESS_DIR/.." && pwd)"
COOKIE_FILE="${COOKIE_FILE:-/tmp/qa-owasp-validation-cookie.txt}"

# Colours when stdout is a TTY; plain otherwise.
if [[ -t 1 ]]; then
  C_RED=$'\033[0;31m'; C_GREEN=$'\033[0;32m'; C_CYAN=$'\033[0;36m'
  C_YELLOW=$'\033[0;33m'; C_RESET=$'\033[0m'
else
  C_RED=""; C_GREEN=""; C_CYAN=""; C_YELLOW=""; C_RESET=""
fi

log_info() { printf "%s[i]%s %s\n" "$C_CYAN" "$C_RESET" "$*"; }
log_warn() { printf "%s[!]%s %s\n" "$C_YELLOW" "$C_RESET" "$*" >&2; }
log_pass() { printf "%s[PASS]%s %s\n" "$C_GREEN" "$C_RESET" "$*"; }
log_fail() {
  local name="$1"; shift
  printf "%s[FAIL]%s %s — %s\n" "$C_RED" "$C_RESET" "$name" "$*" >&2
}

# DVWA lifecycle ------------------------------------------------------------

dvwa_up() {
  log_info "Bringing up DVWA via docker-compose"
  (cd "$HARNESS_DIR" && docker compose up -d) >/dev/null
  local i
  for i in $(seq 1 60); do
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login.php || true)
    if [[ "$code" == "200" ]]; then
      log_info "DVWA ready (login.php returned 200 after ${i}s)"
      return 0
    fi
    sleep 1
  done
  log_fail "dvwa_up" "DVWA did not become ready within 60s"
  return 1
}

dvwa_down() {
  log_info "Tearing down DVWA"
  (cd "$HARNESS_DIR" && docker compose down -v) >/dev/null 2>&1 || true
  rm -f "$COOKIE_FILE"
}

# DVWA login: create the SQLite DB, log in as admin/password, and lower the
# security level to "low" so injections actually trigger. Writes the session
# cookie to $COOKIE_FILE and prints the PHPSESSID to stdout.
dvwa_login() {
  rm -f "$COOKIE_FILE"
  # Setup page builds the DB on first hit.
  curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    "http://localhost:8080/setup.php" >/dev/null
  local token
  token=$(curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    "http://localhost:8080/setup.php" \
    | grep -Eo 'name="user_token" value="[a-f0-9]+"' \
    | head -n1 | sed -E 's/.*value="([a-f0-9]+)".*/\1/')
  curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    --data "create_db=Create+%2F+Reset+Database&user_token=${token}" \
    "http://localhost:8080/setup.php" >/dev/null
  # Login form has its own user_token.
  token=$(curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    "http://localhost:8080/login.php" \
    | grep -Eo 'name="user_token" value="[a-f0-9]+"' \
    | head -n1 | sed -E 's/.*value="([a-f0-9]+)".*/\1/')
  curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    --data "username=admin&password=password&Login=Login&user_token=${token}" \
    "http://localhost:8080/login.php" >/dev/null
  # DVWA requires a hit to index.php after login to register the user
  # session before security level changes are accepted.
  curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    "http://localhost:8080/index.php" >/dev/null
  # Lower security to "low" so SQLi/XSS payloads actually fire.
  token=$(curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    "http://localhost:8080/security.php" \
    | grep -Eo 'name="user_token" value="[a-f0-9]+"' \
    | head -n1 | sed -E 's/.*value="([a-f0-9]+)".*/\1/')
  curl -s -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
    --data "security=low&seclev_submit=Submit&user_token=${token}" \
    "http://localhost:8080/security.php" >/dev/null
  local sid
  sid=$(grep -E '\sPHPSESSID\s' "$COOKIE_FILE" | awk '{print $7}' | tail -n1)
  if [[ -z "$sid" ]]; then
    log_fail "dvwa_login" "no PHPSESSID captured"
    return 1
  fi
  printf "%s" "$sid"
}

# Assertions ----------------------------------------------------------------

# assert_exit_code <expected> <actual> <test_name>
assert_exit_code() {
  local expected="$1" actual="$2" name="$3"
  if [[ "$expected" == "$actual" ]]; then
    log_pass "$name (exit=$actual)"
    return 0
  fi
  log_fail "$name" "expected exit $expected, got $actual"
  return 1
}

# assert_grep <pattern> <file> <test_name>
assert_grep() {
  local pattern="$1" file="$2" name="$3"
  if [[ ! -f "$file" ]]; then
    log_fail "$name" "log file missing: $file"
    return 1
  fi
  if grep -Eq "$pattern" "$file"; then
    log_pass "$name (matched /$pattern/)"
    return 0
  fi
  log_fail "$name" "pattern /$pattern/ not found in $file"
  return 1
}

# Tally ---------------------------------------------------------------------

tally_init() { PASS=0; FAIL=0; }

tally_record() {
  case "$1" in
    pass) PASS=$((PASS + 1)) ;;
    fail) FAIL=$((FAIL + 1)) ;;
    *)    log_warn "tally_record: unknown result '$1'" ;;
  esac
}

tally_report() {
  printf "\n==================== VALIDATION SUMMARY ====================\n"
  printf "PASS: %s%d%s   FAIL: %s%d%s\n" \
    "$C_GREEN" "$PASS" "$C_RESET" "$C_RED" "$FAIL" "$C_RESET"
  printf "============================================================\n"
  [[ "$FAIL" -eq 0 ]]
}
