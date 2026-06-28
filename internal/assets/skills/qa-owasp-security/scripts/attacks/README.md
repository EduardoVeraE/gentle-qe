# Attack scripts

Hands-on probes used by the `qa-owasp-security` skill. Each script is a
thin wrapper around an established tool (or a focused, dependency-free
Node script) and follows a shared CLI contract so they compose cleanly
into pipelines.

> **AUTHORIZATION REQUIRED.** These scripts send active payloads at the
> target. Run them only against systems you own or have written
> permission to test (engagement letter, bug bounty scope, internal
> pentest agreement). All scripts print an authorization banner at start.

## Shared interface

Every script accepts the following flags:

| Flag | Description | Default |
|---|---|---|
| `--help`, `-h` | Print usage and exit 0 | — |
| `--target <url-or-path>` | URL for web/API targets, directory for filesystem scans | required |
| `--out <dir>` | Output directory for results | `./security-out/<script>/<timestamp>/` |
| `--severity-threshold <low\|medium\|high\|critical>` | Exit non-zero when a finding ≥ threshold is found | `high` |

Exit codes:

- `0` — clean run (no findings ≥ threshold)
- `1` — findings at or above threshold
- `2` — required tool not installed (so callers can distinguish missing
  tooling from real vulnerabilities)
- `3` — unexpected runtime error
- `64` — invalid CLI usage

## Script index

| Script | Tool used | Install | OWASP cat | Example |
|---|---|---|---|---|
| `sqli-test.sh` | sqlmap | `brew install sqlmap` | A03 Injection | `./sqli-test.sh --target 'https://staging/api?id=1'` |
| `xss-scan.sh` | dalfox (fallback: ZAP) | `go install github.com/hahwul/dalfox/v2@latest` | A03 Injection | `./xss-scan.sh --target 'https://staging/search?q=foo'` |
| `secrets-scan.sh` | gitleaks (+ trufflehog --deep) | `brew install gitleaks trufflehog` | A02/A07 supply chain | `./secrets-scan.sh --target . --deep` |
| `deps-scan.sh` | npm audit / pip-audit / govulncheck / cargo-audit + trivy | `brew install trivy` (+ ecosystem tool) | A06 Vulnerable Components | `./deps-scan.sh --target ./web` |
| `jwt-test.mjs` | pure Node | Node 18+ | A07 / API2 | `./jwt-test.mjs --target https://api/me --token eyJhbG...` |
| `ssrf-test.mjs` | pure Node | Node 18+ | A10 / API7 SSRF | `./ssrf-test.mjs --target 'https://app/fetch?url={{INJECT}}'` |
| `bola-test.mjs` | pure Node | Node 18+ | API1 BOLA | `./bola-test.mjs --target 'https://api/users/{{ID}}' --auth-token $T --id-range 1-100` |

## Per-script notes

- **sqli-test.sh** — runs `sqlmap --batch --risk 2 --level 3` by default,
  writes the session to `<out>/session/`, and treats any sqlmap-confirmed
  injection as HIGH severity.
- **xss-scan.sh** — primary path is dalfox (fast, focused). Use `--zap`
  to force a ZAP baseline scan when dalfox is unavailable. `--crawl`
  enables deep-domain XSS mode.
- **secrets-scan.sh** — gitleaks always runs and emits SARIF. `--deep`
  adds a trufflehog filesystem pass with credential verification (note:
  this makes outbound API calls).
- **deps-scan.sh** — auto-detects ecosystem from `package.json`,
  `requirements.txt` / `pyproject.toml`, `go.mod`, or `Cargo.toml`.
  Always also runs `trivy fs --scanners vuln,license` as a second
  opinion. Aggregated counters land in `aggregate.json`.
- **jwt-test.mjs** — implements the four most useful JWT attacks:
  `alg=none`, weak HMAC secret dictionary, `kid` header injection, and
  RS256→HS256 algorithm confusion (with `--public-key`). Pure Node, no
  external deps.
- **ssrf-test.mjs** — substitutes `{{INJECT}}` in the target URL with
  cloud-metadata payloads (AWS IMDSv1, GCP, Azure), localhost ports,
  `file://`, and a DNS-rebinding placeholder. Compares response size
  and body markers against a benign baseline.
- **bola-test.mjs** — iterates a numeric range or explicit id list,
  classifies responses, and (with `--secondary-token`) cross-checks
  horizontal privilege escalation between two real accounts. Built-in
  `--rate` rate limiter to avoid DoS-ing the target.

## Output layout

```
security-out/
  <script-name>/
    <YYYYMMDDTHHMMSSZ>/
      summary.txt        # human-readable summary
      report.json        # machine-readable details
      <tool>.sarif       # when the wrapped tool emits SARIF
      matrix.csv         # bola-test only
```

## CI integration

Each script returns a non-zero exit code when findings meet the
threshold. Wire them into your pipeline as quality gates and upload
the SARIF outputs (gitleaks, trivy) to GitHub Code Scanning or your
Security Dashboard for triage in one place.
