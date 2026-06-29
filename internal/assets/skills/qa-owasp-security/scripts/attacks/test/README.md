# qa-owasp-security — Validation Harness

Automated, deterministic validation that prevents regressions on the
attack scripts under `internal/assets/skills/qa-owasp-security/scripts/attacks/`.

The harness boots a local DVWA container, runs each script against a
pinned target or fixture, and asserts both exit codes and log
signatures. Findings counts that are nondeterministic are NOT asserted —
the harness validates that scans **run correctly**, not that they
discover any specific number of vulns.

## Why this exists

Round 1 shipped two P1 bugs that only surfaced when running the scripts
end-to-end:

| Beads ID         | Bug                                                | Spec that guards it          |
| ---------------- | -------------------------------------------------- | ---------------------------- |
| `gentle-qa-bt5`  | (H1 scaffold)                                      | n/a — infrastructure          |
| `gentle-qa-b7n`  | (H2 specs)                                         | n/a — infrastructure          |
| `gentle-qa-c9u`  | (H3 runner)                                        | n/a — infrastructure          |
| deps-scan ENOLOCK + missing trivy invocation       | `specs/test-deps-scan.sh`     |
| xss-scan Docker host rewrite + active-scan switch  | `specs/test-xss-scan.sh`      |

The lesson encoded throughout `lib.sh`: **never** capture an exit code
from the right side of a pipe. `cmd | tail; echo $?` returns `tail`'s
status, not `cmd`'s. Use `cmd > log 2>&1; echo $?` directly.

## Prerequisites

Already required by the attack scripts:

- `docker` (Docker Desktop on macOS, with the daemon running)
- `gitleaks`
- `sqlmap`
- `trivy`
- `node` + `npm`

Optional: `dalfox`. If absent, `xss-scan.sh` exercises the ZAP Docker
fallback (slower but the path we most need to validate, since the round-1
P1 bugs lived there).

The DVWA image (`vulnerables/web-dvwa`) is amd64-only. On Apple Silicon
it runs through Rosetta — `docker-compose.yml` pins
`platform: linux/amd64` to keep behaviour explicit.

## Usage

From the repo root:

```sh
bash internal/assets/skills/qa-owasp-security/scripts/attacks/test/run-validation.sh
```

The runner:

1. Verifies prereqs (exits 2 if any required tool is missing).
2. Boots DVWA on `http://localhost:8080` via docker-compose.
3. Runs every `specs/test-*.sh` in lexicographic order.
4. Tears DVWA down (also on Ctrl+C, via `trap`).
5. Exits 0 if every spec passed; 1 otherwise.

Expect ~10-15 minutes wall-clock when `dalfox` is missing — ZAP active
scans dominate runtime.

## What each spec asserts

| Spec                            | Exit | Log signatures                                              |
| ------------------------------- | ---- | ----------------------------------------------------------- |
| `specs/test-secrets-scan.sh`    | `1`  | `gitleaks findings: [1-9]`                                  |
| `specs/test-deps-scan.sh`       | `1`  | `Trivy findings`, `Ecosystem findings`, trivy.json artifact |
| `specs/test-sqli-test.sh`       | `1`  | `Findings detected: [1-9]`                                  |
| `specs/test-xss-scan.sh`        | 0/1  | `host.docker.internal`, `zap-full-scan.py`; rejects exit 3  |

xss-scan tolerates exit 0 (no findings) and 1 (findings) because ZAP's
verification of reflected XSS on DVWA-low is timing-sensitive; the
genuine failure mode we guard against is exit 3 (runtime error) plus a
missing rewrite/active-scan trace in the log.

## Adding a new spec

1. Drop `specs/test-<name>.sh` next to the existing files, copy any of
   them as a starting point.
2. `set -euo pipefail`, source `../lib.sh` via the standard
   `SCRIPT_DIR` idiom, run the script under test with output redirected
   to `/tmp/qa-validation-<name>.log` and capture `$?` directly into
   `ACTUAL_EXIT`. Never pipe.
3. Use `assert_exit_code` and `assert_grep`. Tail the log on failure for
   debuggability.
4. The runner picks it up automatically — no registry to update.

## CI snippet (GitHub Actions)

```yaml
jobs:
  qa-owasp-validation:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4
      - run: brew install gitleaks sqlmap trivy
      - run: bash internal/assets/skills/qa-owasp-security/scripts/attacks/test/run-validation.sh
```

(Linux runners need a Docker-capable host; the harness itself is
platform-agnostic but DVWA + ZAP both require Docker.)
