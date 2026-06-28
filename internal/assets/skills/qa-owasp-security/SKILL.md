---
name: qa-owasp-security
description: "Trigger: OWASP Top 10, security testing, XSS, SQLi, CSRF, SSRF, threat modeling, secrets scan. Apply web, API, and mobile security testing patterns."
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
---

# OWASP Security Testing Toolkit

ISTQB Non-functional security testing aligned with the OWASP family of standards (Web Top 10 2025, API Top 10 2023, Mobile Top 10 2024) plus threat modeling (STRIDE) and supporting tooling (ZAP, Burp, MobSF, semgrep, gitleaks, trivy).

**Core principle**: Security testing is risk-based and authorized. Every test MUST be tied to an explicit threat (OWASP category, STRIDE element, or documented risk) and run only against systems you have written permission to test.

## When to Use This Skill

- Performing an **OWASP Top 10** baseline scan against a web application
- Auditing a REST/GraphQL API against the **OWASP API Top 10**
- Reviewing a mobile app against the **OWASP Mobile Top 10**
- Building a **STRIDE threat model** for a feature, service, or system
- Validating specific attack vectors: **XSS, SQLi, CSRF, SSRF, BOLA, BFLA, auth bypass, JWT attacks, IDOR**
- Scanning a repository for **leaked secrets** (gitleaks, trufflehog)
- Auditing dependencies for **known CVEs** (npm audit, trivy, OSV)
- Triaging security findings and writing **vulnerability reports** with CVSS scoring
- Producing a **pentest report** at the end of an engagement

## ISTQB Layer

Layer 4 — Non-functional testing by type → **Security**.

This skill complements (does not replace) functional layers:

| Layer | Coverage |
| ----- | -------- |
| 1. Foundation | `qa-manual-istqb` |
| 2. Strategy | `qa-manual-istqb`, `playwright-regression-strategy` |
| 3. Functional by level | `api-testing`, `playwright-e2e-testing`, `selenium-e2e-testing` |
| 4. Non-functional by type | **`qa-owasp-security`** (this skill), `k6-load-test`, `a11y-playwright-testing` |
| 5. Tooling | `playwright-cli`, `playwright-mcp-inspect` |

## Scope

| Domain | OWASP Source | Reference file |
| ------ | ------------ | -------------- |
| Web | OWASP Top 10 — 2025 | `references/owasp-top10-2025-web.md` |
| API | OWASP API Top 10 — 2023 | `references/owasp-api-top10-2023.md` |
| Mobile | OWASP Mobile Top 10 — 2024 | `references/owasp-mobile-top10-2024.md` |
| Cross-cutting | STRIDE threat modeling | `references/threat-modeling-stride.md` |
| Cross-cutting | Tooling catalog (ZAP, Burp, MobSF, semgrep, gitleaks, trivy) | `references/security-tooling.md` |

## Prerequisites

| Requirement | Notes |
| ----------- | ----- |
| Node.js 18+ | Required for the artifact CLI (`scripts/security_artifacts.mjs`) |
| Docker | Required to run OWASP ZAP, MobSF, and similar containerized scanners |
| Git | For repo-level scans (gitleaks, trufflehog, trivy fs) |
| Bash 5+ | For helper scripts under `scripts/attacks/` |
| **Written authorization** | Mandatory — never test a system you do not own or have explicit permission to test |
| Target metadata | URL/host, app package, scope boundaries, allowed test windows, contact for incidents |

## Quick Start

Generate security artifacts from templates (CLI implemented in `scripts/security_artifacts.mjs`):

```bash
# List available templates
node scripts/security_artifacts.mjs list

# Create a vulnerability report
node scripts/security_artifacts.mjs create vuln-report --out reports --title "Reflected XSS in /search"

# Create a pentest engagement report
node scripts/security_artifacts.mjs create pentest-report --out reports --target "api.example.com"

# Create a STRIDE threat model
node scripts/security_artifacts.mjs create threat-model --out specs --feature "Checkout"
```

Run per-vector attack helpers (implemented in `scripts/attacks/`):

```bash
scripts/attacks/xss-test.sh            https://target.example.com/search
scripts/attacks/sqli-test.sh           https://target.example.com/api/users
scripts/attacks/ssrf-test.sh           https://target.example.com/fetch
scripts/attacks/jwt-attack-test.sh     https://api.example.com/login
scripts/attacks/secrets-scan.sh        .
```

## Workflows

### 1) Run a baseline OWASP Top 10 web scan

1. Confirm written authorization and scope (URL, paths in/out of scope, time windows).
2. Run an automated baseline (OWASP ZAP baseline scan in Docker) against the target.
3. For each finding, manually verify and map to the **OWASP Top 10 2025** category.
4. Triage: severity (CVSS), exploitability, business impact.
5. Produce a vulnerability report per confirmed issue.

See `references/owasp-top10-2025-web.md` for category-by-category test guidance.

### 2) Run an API security audit (OWASP API Top 10)

1. Collect the API specification (OpenAPI/Swagger, GraphQL schema, Postman collection).
2. Enumerate endpoints, authentication models, and authorization rules.
3. Test each **OWASP API Top 10 2023** category — focus on BOLA, BFLA, broken auth, mass assignment.
4. Validate rate limiting, input validation, and resource consumption.
5. Document findings with reproduction steps and request/response evidence.

See `references/owasp-api-top10-2023.md`.

### 3) Run a mobile app security review (OWASP Mobile Top 10)

1. Obtain the app binary (APK/IPA) and any reverse-engineering authorization.
2. Run static analysis with MobSF (`docker run mobsf/mobsf`).
3. Run dynamic analysis on a rooted/jailbroken test device.
4. Map findings to **OWASP Mobile Top 10 2024** categories.
5. Verify platform-specific concerns (insecure storage, weak crypto, insecure IPC).

See `references/owasp-mobile-top10-2024.md`.

### 4) Build a threat model (STRIDE)

1. Diagram the system: trust boundaries, data flows, components, external actors.
2. For each component/data flow, enumerate threats by **STRIDE** category (Spoofing, Tampering, Repudiation, Information disclosure, Denial of service, Elevation of privilege).
3. Rate each threat (likelihood × impact) and propose mitigations.
4. Track residual risk explicitly.

See `references/threat-modeling-stride.md` and `templates/threat-model.md`.

### 5) Triage findings and write vulnerability reports

1. Reproduce reliably; reduce to minimal request/payload.
2. Score with **CVSS v3.1** (base + environmental).
3. Map to OWASP category and CWE ID.
4. Document: summary, impact, reproduction, evidence, remediation, references.
5. Track lifecycle through retest and closure.

Use `templates/vuln-report.md` and `references/security-tooling.md` for tool-specific evidence capture.

## Inputs to Collect

- **Target**: URL(s), API base, mobile package, repo path. One concrete artifact per engagement.
- **Authorization**: written scope, contact, allowed windows, exclusions (no DoS, no production data exfiltration, etc.).
- **Threat actors**: who you are simulating (anonymous internet, authenticated user, insider, compromised dependency).
- **Data sensitivity**: PII, PHI, PCI, secrets, regulated data classes in scope.
- **Compliance constraints**: SOC2, ISO 27001, PCI-DSS, HIPAA, GDPR — affects evidence retention and disclosure.
- **Existing controls**: WAF, rate limiting, IDS/IPS — both to test their effectiveness and to avoid false positives.

## Outputs

| Artifact | When produced | Template |
| -------- | ------------- | -------- |
| Vulnerability report | Per confirmed finding | `templates/vuln-report.md` |
| Pentest engagement report | End of engagement | `templates/pentest-report.md` |
| Threat model | Design phase or audit | `templates/threat-model.md` |
| OWASP checklist results | Per scan domain (web/API/mobile) | `templates/owasp-checklist.md` |
| Attack evidence (requests, payloads, screenshots) | Per finding | Captured by `scripts/attacks/*` |

## Exclusions

This skill is deliberately scoped. Do NOT use it for:

- **General API functional testing** — use `api-testing`.
- **General mobile functional testing** — use `qa-mobile-testing`.
- **Accessibility testing** (WCAG, ARIA) — use `a11y-playwright-testing`.
- **Performance / load / stress testing** — use `k6-load-test`.
- **Functional E2E flows** — use `playwright-e2e-testing` or `selenium-e2e-testing`.
- **Test planning, test cases, traceability** — use `qa-manual-istqb`.

If a request mixes concerns (e.g., "load test the login and check for auth bypass"), split it: load with `k6-load-test`, auth bypass with this skill.
