#!/usr/bin/env bash
# secrets-scan.sh — gitleaks (default) + trufflehog (--deep) repo/dir scan.
# OWASP: A02 Cryptographic / A07 ID&Auth / supply chain. Read-only — safe locally.
set -euo pipefail

SCRIPT_NAME="secrets-scan"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEFAULT_OUT="./security-out/${SCRIPT_NAME}/${TIMESTAMP}"

TARGET=""
OUT_DIR=""
SEVERITY_THRESHOLD="high"
DEEP=0

usage() {
  cat <<EOF
Usage: $0 [--target <path>] [options]

Secrets scanner: gitleaks first, optional trufflehog deep pass.

Options:
  --target <path>            Repo or directory to scan (default: \$PWD)
  --out <dir>                Output directory (default: ${DEFAULT_OUT})
  --severity-threshold <s>   low|medium|high|critical (default: high)
  --deep                     Also run trufflehog filesystem with verification
  -h, --help                 Show help and exit

Examples:
  $0
  $0 --target /path/to/repo --deep
  $0 --target . --severity-threshold critical
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --target) TARGET="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --severity-threshold) SEVERITY_THRESHOLD="$2"; shift 2 ;;
    --deep) DEEP=1; shift ;;
    *) echo "Unknown option: $1" >&2; usage; exit 64 ;;
  esac
done

cat <<'BANNER' >&2
[!] AUTHORIZATION REQUIRED
[!] Secrets scans read repository contents. Verification (--deep) may
[!] make outbound API calls to confirm credentials are live.
BANNER

TARGET="${TARGET:-$PWD}"
OUT_DIR="${OUT_DIR:-$DEFAULT_OUT}"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
SARIF="$OUT_DIR/gitleaks.sarif"
TRUFFLE_JSON="$OUT_DIR/trufflehog.json"

if ! command -v gitleaks >/dev/null 2>&1; then
  cat <<'HINT' >&2
Error: gitleaks is not installed.
Install:
  brew install gitleaks
  # or: docker run --rm -v $PWD:/path zricethezav/gitleaks:latest detect -s /path
HINT
  exit 2
fi

echo "Running gitleaks on $TARGET" | tee "$SUMMARY"
set +e
gitleaks detect --source "$TARGET" --report-format sarif \
  --report-path "$SARIF" --redact 2>&1 | tee -a "$SUMMARY"
GL_EXIT=$?
set -e

GL_FINDINGS=0
if [[ -f "$SARIF" ]]; then
  GL_FINDINGS=$(grep -Eo '"ruleId"' "$SARIF" 2>/dev/null | wc -l | tr -d ' ')
fi

TH_FINDINGS=0
if [[ "$DEEP" -eq 1 ]]; then
  if ! command -v trufflehog >/dev/null 2>&1; then
    cat <<'HINT' >&2
Warning: trufflehog not installed — skipping --deep pass.
Install: brew install trufflehog
HINT
  else
    echo "Running trufflehog filesystem on $TARGET" | tee -a "$SUMMARY"
    set +e
    trufflehog filesystem "$TARGET" --json > "$TRUFFLE_JSON" 2>>"$SUMMARY"
    set -e
    TH_FINDINGS=$(grep -Eo '"DetectorName"' "$TRUFFLE_JSON" 2>/dev/null | wc -l | tr -d ' ')
  fi
fi

TOTAL=$(( GL_FINDINGS + TH_FINDINGS ))

{
  echo "---"
  echo "Target: $TARGET"
  echo "gitleaks findings: $GL_FINDINGS  (exit: $GL_EXIT)"
  echo "trufflehog findings: $TH_FINDINGS"
  echo "Total: $TOTAL"
  echo "Severity threshold: $SEVERITY_THRESHOLD"
  echo "Output: $OUT_DIR"
} | tee -a "$SUMMARY"

case "$SEVERITY_THRESHOLD" in
  low|medium|high|critical) ;;
  *) echo "Invalid --severity-threshold: $SEVERITY_THRESHOLD" >&2; exit 64 ;;
esac

# All secret findings are treated as HIGH severity.
if [[ "$TOTAL" -gt 0 ]]; then
  case "$SEVERITY_THRESHOLD" in
    low|medium|high) echo "Secret findings present."; exit 1 ;;
    critical)        echo "Findings present but below critical."; exit 0 ;;
  esac
fi

echo "No secrets detected."
exit 0
