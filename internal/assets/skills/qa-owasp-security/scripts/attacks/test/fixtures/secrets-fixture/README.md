# secrets-fixture

Minimal git repo seeded with the AWS public-docs example credentials so
gitleaks reliably finds exactly one secret. Used by
`specs/test-secrets-scan.sh` as a regression guard.

The credentials in `seed.txt` are SYNTHETIC values that match the
gitleaks AWS rule shape but do not authenticate anywhere. We avoid the
canonical AWS docs `EXAMPLE` values because gitleaks 8.x allowlists
them as known-fake test data, which would defeat the regression-guard
purpose of this fixture.

`init-fixture.sh` is idempotent: it deletes any existing `.git/`,
re-initialises the repo, writes `seed.txt` with the seeded credentials,
and commits once. This guarantees a deterministic single-finding scan.
