# deps-fixture

Minimal Node project pinned to `lodash@4.17.4` so trivy/npm audit
deterministically report at least one HIGH-severity advisory. Used by
`specs/test-deps-scan.sh` as a regression guard for the
`npm audit ENOLOCK` + missing-trivy-invocation P1 bugs.

`init-fixture.sh` regenerates `package-lock.json` via
`npm install --package-lock-only --ignore-scripts` — no postinstall
scripts run, no node_modules/ is created.
