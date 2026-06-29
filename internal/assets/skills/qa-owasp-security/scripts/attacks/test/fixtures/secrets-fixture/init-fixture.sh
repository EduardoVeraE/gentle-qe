#!/usr/bin/env bash
# init-fixture.sh — recreate the secrets-fixture git repo from scratch.
# Idempotent: every run produces the same single-commit history with one
# seeded AWS-style credential pair (canonical AWS docs example values —
# NOT real). gitleaks reliably flags this on every run.
set -euo pipefail

FIXTURE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$FIXTURE_DIR"

# Wipe any previous repo state so every run is identical.
rm -rf .git

# Synthetic AWS-style credentials that match gitleaks rules but are
# obviously fake. We deliberately avoid the AWS docs "EXAMPLE" values
# because gitleaks 8.x allowlists them as known-fake test data, which
# defeats the regression-guard purpose of this fixture.
cat > seed.txt <<'EOF'
# SYNTHETIC test credentials — NOT REAL. Pattern matches gitleaks AWS
# rule but the values are random and do not authenticate anywhere.
aws_access_key_id     = AKIAZ4QW3FT8K9PD2XR5
aws_secret_access_key = a8Hb4Tc2Ld6Mf0Ng9Ph3Qj5Rk7Sm1Tn3Uo5Vp7Wq
EOF

git init -q -b main
git -c user.email=fixture@example.com \
    -c user.name=fixture \
    -c commit.gpgsign=false \
    add .
git -c user.email=fixture@example.com \
    -c user.name=fixture \
    -c commit.gpgsign=false \
    commit -q -m "seed: fake AWS credentials for gitleaks fixture"

echo "secrets-fixture initialised at $FIXTURE_DIR"
