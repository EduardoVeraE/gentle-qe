#!/usr/bin/env bash
# xss-scan.sh — XSS scanning via dalfox (primary) with ZAP fallback.
# OWASP: A03 Injection / A05. REQUIRES AUTHORIZATION.
set -euo pipefail

SCRIPT_NAME="xss-scan"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEFAULT_OUT="./security-out/${SCRIPT_NAME}/${TIMESTAMP}"

TARGET=""
OUT_DIR=""
SEVERITY_THRESHOLD="high"
CRAWL=0
USE_ZAP=0

usage() {
  cat <<EOF
Usage: $0 --target <url> [options]

XSS scanner. Default tool: dalfox. Fallback: OWASP ZAP active scan.

Options:
  --target <url>             Target URL (single endpoint by default)
  --out <dir>                Output directory (default: ${DEFAULT_OUT})
  --severity-threshold <s>   low|medium|high|critical (default: high)
  --crawl                    Crawl from the target before fuzzing (off by default)
  --zap                      Force ZAP fallback even if dalfox is installed
  -h, --help                 Show help and exit

Examples:
  $0 --target 'https://staging.example.com/search?q=foo'
  $0 --target 'https://staging.example.com/' --crawl
  $0 --target 'https://staging.example.com/' --zap
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --target) TARGET="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --severity-threshold) SEVERITY_THRESHOLD="$2"; shift 2 ;;
    --crawl) CRAWL=1; shift ;;
    --zap) USE_ZAP=1; shift ;;
    *) echo "Unknown option: $1" >&2; usage; exit 64 ;;
  esac
done

cat <<'BANNER' >&2
[!] AUTHORIZATION REQUIRED
[!] XSS scanning sends active payloads (script/img/svg) to the target.
[!] Run only against systems you own or have written permission to test.
BANNER

if [[ -z "$TARGET" ]]; then
  echo "Error: --target is required." >&2
  usage
  exit 64
fi

OUT_DIR="${OUT_DIR:-$DEFAULT_OUT}"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
JSON_OUT="$OUT_DIR/dalfox.json"

SCAN_EXIT=0

run_dalfox() {
  local mode="url"
  [[ "$CRAWL" -eq 1 ]] && mode="url --deep-domain-xss --skip-bav"
  echo "Running: dalfox $mode $TARGET --format json -o $JSON_OUT" | tee "$SUMMARY"
  set +e
  # shellcheck disable=SC2086
  dalfox $mode "$TARGET" --format json -o "$JSON_OUT" 2>&1 | tee -a "$SUMMARY"
  SCAN_EXIT=${PIPESTATUS[0]}
  set -e
}

run_zap() {
  if ! command -v docker >/dev/null 2>&1; then
    cat <<'HINT' >&2
Error: ZAP fallback requires Docker.
Install Docker Desktop: https://docs.docker.com/get-docker/
Or install dalfox: go install github.com/hahwul/dalfox/v2@latest
HINT
    exit 2
  fi
  # Rewrite localhost/127.0.0.1 → host.docker.internal so the ZAP container
  # can reach the host. User input is preserved in logs above; only the
  # docker invocation uses the rewritten URL.
  local docker_target="$TARGET"
  case "$TARGET" in
    *://localhost*|*://localhost/*|*://localhost:*) docker_target="${TARGET//localhost/host.docker.internal}" ;;
    *://127.0.0.1*) docker_target="${TARGET//127.0.0.1/host.docker.internal}" ;;
  esac
  if [[ "$docker_target" != "$TARGET" ]]; then
    echo "ZAP container target rewrite: $TARGET -> $docker_target" | tee -a "$SUMMARY"
  fi
  cat <<'NOTE' | tee "$SUMMARY"
Running ZAP FULL active scan (fallback mode).
[!] Active scan is INVASIVE: it sends real attack payloads. Written
[!] authorization is mandatory (see banner above).
NOTE
  set +e
  docker run --rm \
    --add-host=host.docker.internal:host-gateway \
    -v "$OUT_DIR":/zap/wrk/:rw zaproxy/zap-stable \
    zap-full-scan.py -t "$docker_target" -J zap.json -r zap.html 2>&1 | tee -a "$SUMMARY"
  SCAN_EXIT=${PIPESTATUS[0]}
  set -e
  JSON_OUT="$OUT_DIR/zap.json"
  # zap-full-scan.py exits 0 clean / 1 warn / 2 fail (findings). Treat
  # any of those as a successful scan run; >2 is a runtime failure.
  if [[ "$SCAN_EXIT" -gt 2 ]]; then
    echo "ZAP runtime error (exit $SCAN_EXIT)." >&2
    exit 3
  fi
  SCAN_EXIT=0
}

if [[ "$USE_ZAP" -eq 1 ]] || ! command -v dalfox >/dev/null 2>&1; then
  if [[ "$USE_ZAP" -ne 1 ]]; then
    cat <<'HINT' >&2
Warning: dalfox not installed — falling back to ZAP active scan.
Install dalfox: go install github.com/hahwul/dalfox/v2@latest
            or: brew install dalfox
HINT
  fi
  run_zap
else
  run_dalfox
  if [[ "$SCAN_EXIT" -ne 0 ]]; then
    echo "dalfox runtime error (exit $SCAN_EXIT)." >&2
    exit 3
  fi
fi

# Distinguish "scan completed with zero findings" from "scan never ran".
if [[ ! -f "$JSON_OUT" ]]; then
  echo "Scanner produced no report file ($JSON_OUT). Treating as runtime error." >&2
  exit 3
fi

FINDINGS=0
if grep -Eq '"severity"\s*:\s*"(High|Medium|Low|Critical)"|"riskcode"\s*:\s*"[1-3]"|"type"\s*:\s*"V"' "$JSON_OUT" 2>/dev/null; then
  FINDINGS=$({ grep -Eo '"severity"\s*:\s*"[^"]+"|"riskcode"\s*:\s*"[1-3]"|"type"\s*:\s*"V"' "$JSON_OUT" 2>/dev/null || true; } | wc -l | tr -d ' ')
fi

{
  echo "---"
  echo "Target: $TARGET"
  echo "Crawl: $CRAWL"
  echo "Findings: $FINDINGS"
  echo "Severity threshold: $SEVERITY_THRESHOLD"
  echo "Report: $JSON_OUT"
} | tee -a "$SUMMARY"

case "$SEVERITY_THRESHOLD" in
  low|medium|high|critical) ;;
  *) echo "Invalid --severity-threshold: $SEVERITY_THRESHOLD" >&2; exit 64 ;;
esac

# Per-finding severity is unreliable across dalfox/ZAP outputs, so we treat
# any finding as ≥ HIGH. Anything except --severity-threshold critical
# triggers exit 1 when findings > 0.
if [[ "$FINDINGS" -gt 0 ]]; then
  case "$SEVERITY_THRESHOLD" in
    low|medium|high) echo "XSS findings detected."; exit 1 ;;
    critical)        echo "Findings present but below 'critical' threshold."; exit 0 ;;
  esac
fi
echo "No XSS findings."
exit 0
