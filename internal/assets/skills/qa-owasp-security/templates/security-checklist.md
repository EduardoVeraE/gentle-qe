<!-- Skill: qa-owasp-security · Template: security-checklist -->
<!-- Placeholders: {{title}}, {{date}}, {{author}}, {{project}}, {{target}}, {{tool_web}}, {{tool_api}}, {{tool_mobile}}, {{evidence_ref}}, {{result}}, {{notes}} -->

# Security Checklist — {{title}}

| Field   | Value         |
| ------- | ------------- |
| Project | {{project}}   |
| Target  | {{target}}    |
| Author  | {{author}}    |
| Date    | {{date}}      |

> Per-OWASP-category gate. Every row must be Pass, Fail, or N/A with evidence reference. Empty rows fail the gate.

Result legend: `Pass` / `Fail` / `N/A` (must justify N/A in notes).

## Web — OWASP Top 10 (2025)

| Category                                            | Test performed                                                  | Tool          | Result       | Evidence ref      |
| --------------------------------------------------- | --------------------------------------------------------------- | ------------- | ------------ | ----------------- |
| A01 Broken Access Control                           | Force-browse and IDOR sweep across all authenticated routes     | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A02 Cryptographic Failures                          | TLS config audit, secret-in-transit, weak hash detection        | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A03 Injection                                       | SQLi, NoSQLi, OS command, LDAP, template injection sweep        | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A04 Insecure Design                                 | Threat-model review of new features; abuse-case tests           | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A05 Security Misconfiguration                       | Default creds, verbose errors, header audit, dir listing        | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A06 Vulnerable & Outdated Components                | SCA scan of dependencies and Docker base images                 | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A07 Identification & Authentication Failures        | Session fixation, brute force, credential stuffing, weak MFA    | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A08 Software & Data Integrity Failures              | CI/CD signing review, deserialization, untrusted update sources | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A09 Security Logging & Monitoring Failures          | Log coverage of auth, authz, admin actions; alert path verified | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |
| A10 Server-Side Request Forgery (SSRF)              | URL fetchers, webhook callbacks, image-resize endpoints         | {{tool_web}}  | {{result}}   | {{evidence_ref}}  |

## API — OWASP API Top 10 (2023)

| Category                                                   | Test performed                                                       | Tool          | Result       | Evidence ref      |
| ---------------------------------------------------------- | -------------------------------------------------------------------- | ------------- | ------------ | ----------------- |
| API1 Broken Object Level Authorization                     | Cross-tenant object access on every `/:id` route                     | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API2 Broken Authentication                                 | Token forgery, JWT alg-none, refresh-token replay                    | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API3 Broken Object Property Level Authorization            | Mass-assignment, hidden field exposure (excessive data)              | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API4 Unrestricted Resource Consumption                     | Rate-limit bypass, expensive query, regex DoS                        | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API5 Broken Function Level Authorization                   | Privileged endpoint access from low-priv role                        | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API6 Unrestricted Access to Sensitive Business Flows       | Workflow abuse (mass-signup, bulk-buy, scraping)                     | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API7 Server-Side Request Forgery                           | URL params accepting external schemes; metadata service reach        | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API8 Security Misconfiguration                             | CORS, headers, default debug routes, verbose stack traces            | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API9 Improper Inventory Management                         | Shadow / deprecated / staging endpoints exposed in prod              | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |
| API10 Unsafe Consumption of APIs                           | Trust validation of upstream third-party API responses               | {{tool_api}}  | {{result}}   | {{evidence_ref}}  |

## Mobile — OWASP Mobile Top 10 (2024)

| Category                                       | Test performed                                                        | Tool             | Result       | Evidence ref      |
| ---------------------------------------------- | --------------------------------------------------------------------- | ---------------- | ------------ | ----------------- |
| M1 Improper Credential Usage                   | Hardcoded secrets, key reuse, credentials in logs                     | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M2 Inadequate Supply Chain Security            | SBOM review, third-party SDK audit, build pipeline integrity          | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M3 Insecure Authentication/Authorization       | Biometric bypass, token storage, server-side authz                    | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M4 Insufficient Input/Output Validation        | Deep-link injection, IPC validation, WebView XSS                      | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M5 Insecure Communication                      | TLS pinning, fallback to HTTP, weak ciphers                           | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M6 Inadequate Privacy Controls                 | PII collection review, consent flow, data minimization                | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M7 Insufficient Binary Protections             | Anti-tamper, root/jailbreak detection, obfuscation                    | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M8 Security Misconfiguration                   | Exported components, debuggable build, backup flag                    | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M9 Insecure Data Storage                       | Keychain/Keystore use, plaintext SQLite, shared prefs                 | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |
| M10 Insufficient Cryptography                  | Algorithm strength, IV reuse, custom crypto detection                 | {{tool_mobile}}  | {{result}}   | {{evidence_ref}}  |

## Gate decision

- [ ] All Critical and High items Pass
- [ ] Every N/A has written justification
- [ ] Evidence stored under `{{evidence_ref}}`

Notes: {{notes}}
