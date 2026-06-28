#!/usr/bin/env bash
# deps-scan.sh — multi-ecosystem dependency audit + trivy fs.
# OWASP: A06 Vulnerable & Outdated Components / supply chain. Read-only.
set -euo pipefail

SCRIPT_NAME="deps-scan"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEFAULT_OUT="./security-out/${SCRIPT_NAME}/${TIMESTAMP}"

TARGET=""
OUT_DIR=""
SEVERITY_THRESHOLD="high"

usage() {
  cat <<EOF
Usage: $0 [--target <dir>] [options]

Dependency audit. Detects ecosystem (Node, Python, Go, Rust) and runs
the matching auditor. Always runs trivy fs as a second opinion.

Options:
  --target <dir>             Project directory to scan (default: \$PWD)
  --out <dir>                Output directory (default: ${DEFAULT_OUT})
  --severity-threshold <s>   low|medium|high|critical (default: high)
  -h, --help                 Show help and exit

Examples:
  $0
  $0 --target ./web --severity-threshold critical
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --target) TARGET="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --severity-threshold) SEVERITY_THRESHOLD="$2"; shift 2 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 64 ;;
  esac
done

cat <<'BANNER' >&2
[!] AUTHORIZATION REQUIRED
[!] Dependency scans read the project tree and may query public CVE feeds.
BANNER

TARGET="${TARGET:-$PWD}"
OUT_DIR="${OUT_DIR:-$DEFAULT_OUT}"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
AGG="$OUT_DIR/aggregate.json"

echo "{}" > "$AGG"
echo "Scanning $TARGET" | tee "$SUMMARY"

eco_findings=0
trivy_findings=0

run_eco() {
  local kind="$1" cmd="$2" outfile="$3" pattern="$4" hint="$5"
  echo "Detected $kind" | tee -a "$SUMMARY"
  if [[ -n "$hint" ]] && ! command -v "${cmd%% *}" >/dev/null 2>&1; then
    echo "Hint: $hint" >&2; return 0
  fi
  set +e
  (cd "$TARGET" && eval "$cmd") > "$OUT_DIR/$outfile" 2>>"$SUMMARY"
  local eco_exit=$?
  eco_findings=$({ grep -Eo "$pattern" "$OUT_DIR/$outfile" 2>/dev/null || true; } | wc -l | tr -d ' ')
  set -e
  return "$eco_exit"
}

ECOSYSTEM="none"
if [[ -f "$TARGET/package.json" ]]; then
  ECOSYSTEM="node"
  # npm audit fails with ENOLOCK on repos without package-lock.json. Generate
  # a read-only lockfile (no install, no scripts) so audit can proceed.
  if [[ ! -f "$TARGET/package-lock.json" && ! -f "$TARGET/npm-shrinkwrap.json" ]]; then
    cat <<'WARN' >&2
Warning: no package-lock.json found. Generating read-only lockfile via
         'npm i --package-lock-only --ignore-scripts' so npm audit can run.
WARN
    set +e
    (cd "$TARGET" && npm i --package-lock-only --ignore-scripts) >>"$SUMMARY" 2>&1
    npm_lock_exit=$?
    set -e
    if [[ "$npm_lock_exit" -ne 0 ]]; then
      echo "Warning: npm i --package-lock-only failed (exit $npm_lock_exit). Skipping npm audit; trivy will still run." >&2
    fi
  fi
  if [[ -f "$TARGET/package-lock.json" || -f "$TARGET/npm-shrinkwrap.json" ]]; then
    run_eco "Node (package.json)" "npm audit --json" "npm-audit.json" \
      '"severity"\s*:\s*"(high|critical)"' "" || true
  fi
elif [[ -f "$TARGET/requirements.txt" || -f "$TARGET/pyproject.toml" ]]; then
  ECOSYSTEM="python"
  run_eco "Python (requirements/pyproject)" "pip-audit --format json" "pip-audit.json" \
    '"id"\s*:' "pip install pip-audit" || true
elif [[ -f "$TARGET/go.mod" ]]; then
  ECOSYSTEM="go"
  run_eco "Go (go.mod)" "govulncheck -json ./..." "govulncheck.json" \
    '"OSV"' "go install golang.org/x/vuln/cmd/govulncheck@latest" || true
elif [[ -f "$TARGET/Cargo.toml" ]]; then
  ECOSYSTEM="rust"
  run_eco "Rust (Cargo.toml)" "cargo audit --json" "cargo-audit.json" \
    '"id"\s*:' "cargo install cargo-audit" || true
else
  echo "No supported manifest detected — skipping ecosystem audit." | tee -a "$SUMMARY"
fi

if command -v trivy >/dev/null 2>&1; then
  echo "Running trivy fs --scanners vuln,license" | tee -a "$SUMMARY"
  set +e
  trivy fs --scanners vuln,license --format json --output "$OUT_DIR/trivy.json" \
    --severity HIGH,CRITICAL "$TARGET" 2>>"$SUMMARY"
  set -e
  trivy_findings=$({ grep -Eo '"VulnerabilityID"' "$OUT_DIR/trivy.json" 2>/dev/null || true; } | wc -l | tr -d ' ')
else
  cat <<'HINT' >&2
Warning: trivy not installed — skipping fs scanner.
Install: brew install trivy   |   https://aquasecurity.github.io/trivy/
HINT
fi

TOTAL=$(( eco_findings + trivy_findings ))

{
  echo "---"
  echo "Ecosystem: $ECOSYSTEM"
  echo "Ecosystem findings (high/crit): $eco_findings"
  echo "Trivy findings (high/crit):     $trivy_findings"
  echo "Total: $TOTAL"
  echo "Severity threshold: $SEVERITY_THRESHOLD"
  echo "Output: $OUT_DIR"
} | tee -a "$SUMMARY"

# Aggregate JSON (lightweight; full reports remain per-tool)
cat > "$AGG" <<JSON
{
  "target": "$TARGET",
  "ecosystem": "$ECOSYSTEM",
  "findings": { "ecosystem": $eco_findings, "trivy": $trivy_findings, "total": $TOTAL },
  "severity_threshold": "$SEVERITY_THRESHOLD",
  "timestamp": "$TIMESTAMP"
}
JSON

case "$SEVERITY_THRESHOLD" in
  low|medium|high|critical) ;;
  *) echo "Invalid --severity-threshold: $SEVERITY_THRESHOLD" >&2; exit 64 ;;
esac

if [[ "$TOTAL" -gt 0 ]]; then
  echo "Vulnerable dependencies detected." >&2
  exit 1
fi
echo "No vulnerable dependencies detected at threshold $SEVERITY_THRESHOLD."
exit 0
