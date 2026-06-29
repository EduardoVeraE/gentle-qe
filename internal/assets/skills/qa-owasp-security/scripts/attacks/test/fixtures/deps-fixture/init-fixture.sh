#!/usr/bin/env bash
# init-fixture.sh — generate package-lock.json for the deps-fixture.
# We pin lodash@4.17.4 (known prototype-pollution CVEs). The lockfile is
# generated read-only (no install, no scripts) so npm audit + trivy can
# both run deterministically.
set -euo pipefail

FIXTURE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$FIXTURE_DIR"

# Always regenerate so the lockfile matches package.json exactly.
rm -f package-lock.json
npm install --package-lock-only --ignore-scripts --silent >/dev/null

echo "deps-fixture lockfile generated at $FIXTURE_DIR/package-lock.json"
