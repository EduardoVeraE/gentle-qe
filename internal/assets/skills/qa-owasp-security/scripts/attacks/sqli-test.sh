#!/usr/bin/env bash
# sqli-test.sh — sqlmap wrapper for SQL injection probing.
# OWASP: A03 Injection. REQUIRES AUTHORIZATION.
set -euo pipefail

SCRIPT_NAME="sqli-test"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEFAULT_OUT="./security-out/${SCRIPT_NAME}/${TIMESTAMP}"

TARGET=""
OUT_DIR=""
SEVERITY_THRESHOLD="high"
RISK="2"
LEVEL="3"
DATA=""
COOKIE=""

usage() {
  cat <<EOF
Usage: $0 --target <url> [options]

SQL injection probing wrapper around sqlmap (--batch).

Options:
  --target <url>             Target URL (with parameter, e.g. https://app/api?id=1)
  --out <dir>                Output directory (default: ${DEFAULT_OUT})
  --severity-threshold <s>   low|medium|high|critical — exit non-zero if any
                             finding is at or above this level (default: high)
  --risk <1-3>               sqlmap --risk (default: 2)
  --level <1-5>              sqlmap --level (default: 3)
  --data <body>              POST body (forwarded to --data)
  --cookie <cookies>         Cookie header (forwarded to --cookie)
  -h, --help                 Show this help and exit

Examples:
  $0 --target 'https://staging.example.com/api/items?id=1'
  $0 --target 'https://staging.example.com/login' --data 'user=a&pass=b' --risk 3 --level 5
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --target) TARGET="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --severity-threshold) SEVERITY_THRESHOLD="$2"; shift 2 ;;
    --risk) RISK="$2"; shift 2 ;;
    --level) LEVEL="$2"; shift 2 ;;
    --data) DATA="$2"; shift 2 ;;
    --cookie) COOKIE="$2"; shift 2 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 64 ;;
  esac
done

cat <<'BANNER' >&2
[!] AUTHORIZATION REQUIRED
[!] sqlmap sends active SQL injection payloads to the target.
[!] Run only against systems you own or have written permission to test.
BANNER

if [[ -z "$TARGET" ]]; then
  echo "Error: --target is required." >&2
  usage
  exit 64
fi

if ! command -v sqlmap >/dev/null 2>&1; then
  cat <<'HINT' >&2
Error: sqlmap is not installed.
Install one of:
  brew install sqlmap            # macOS (Homebrew)
  apt-get install -y sqlmap      # Debian/Ubuntu
  pip install sqlmap             # any platform with Python
HINT
  exit 2
fi

OUT_DIR="${OUT_DIR:-$DEFAULT_OUT}"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
JSON_LOG="$OUT_DIR/sqlmap.log"
SESSION_DIR="$OUT_DIR/session"

ARGS=(--batch --risk "$RISK" --level "$LEVEL" -u "$TARGET"
      --output-dir="$SESSION_DIR" --flush-session --disable-coloring)
[[ -n "$DATA" ]]   && ARGS+=(--data "$DATA")
[[ -n "$COOKIE" ]] && ARGS+=(--cookie "$COOKIE")

echo "Running: sqlmap ${ARGS[*]}" | tee "$SUMMARY"
set +e
sqlmap "${ARGS[@]}" 2>&1 | tee "$JSON_LOG"
SQLMAP_EXIT=$?
set -e

FINDINGS=0
if grep -Ei "is vulnerable|sqlmap identified the following injection point" "$JSON_LOG" >/dev/null 2>&1; then
  FINDINGS=1
fi

{
  echo "---"
  echo "Target: $TARGET"
  echo "Risk: $RISK  Level: $LEVEL"
  echo "Findings detected: $FINDINGS"
  echo "sqlmap exit code: $SQLMAP_EXIT"
  echo "Severity threshold: $SEVERITY_THRESHOLD"
  echo "Output: $OUT_DIR"
} | tee -a "$SUMMARY"

# Severity gate: any vuln from sqlmap is treated as HIGH by default.
case "$SEVERITY_THRESHOLD" in
  low|medium|high|critical) ;;
  *) echo "Invalid --severity-threshold: $SEVERITY_THRESHOLD" >&2; exit 64 ;;
esac

if [[ "$FINDINGS" -gt 0 ]]; then
  case "$SEVERITY_THRESHOLD" in
    low|medium|high) echo "Vulnerabilities found at or above threshold."; exit 1 ;;
    critical)        echo "Findings present but below 'critical' threshold."; exit 0 ;;
  esac
fi

echo "No SQL injection vulnerabilities detected by sqlmap."
exit 0
