<!-- Skill: qa-owasp-security · Template: security-test-plan -->
<!-- Placeholders: {{title}}, {{date}}, {{author}}, {{project}}, {{target}}, {{version}}, {{scope_in}}, {{scope_out}}, {{authorization_ref}}, {{threat_actors}}, {{assets}}, {{environment}}, {{tools}}, {{test_data_policy}}, {{entry_criteria}}, {{exit_criteria}}, {{risk_matrix}}, {{schedule_start}}, {{schedule_end}}, {{reporting_cadence}}, {{stakeholders}}, {{methodology}} -->

# Security Test Plan — {{title}}

| Field    | Value                                          |
| -------- | ---------------------------------------------- |
| Project  | {{project}}                                    |
| Target   | {{target}} <!-- e.g., api.example.com v2.3 --> |
| Version  | {{version}}                                    |
| Author   | {{author}}                                     |
| Date     | {{date}} <!-- YYYY-MM-DD -->                   |
| Status   | Draft / Approved / In-Execution / Closed       |

> ISTQB-aligned test plan, security-flavored. Read alongside the project-level master test plan.

## 1. Scope

### In scope
- {{scope_in}} <!-- e.g., REST API endpoints under /api/v2/*, web SPA, OAuth flow -->

### Out of scope
- {{scope_out}} <!-- e.g., third-party SaaS, physical security, social engineering -->

## 2. Authorization

Written authorization to test is recorded in: {{authorization_ref}} <!-- e.g., signed RoE document, ticket ID -->

- [ ] Rules of Engagement signed
- [ ] Emergency contact list shared
- [ ] Allowed source IPs whitelisted

## 3. Threat Actors

- {{threat_actors}} <!-- e.g., unauthenticated external attacker, malicious tenant, insider with read-only role -->

## 4. Assets

| Asset           | Classification | Owner      |
| --------------- | -------------- | ---------- |
| {{assets}}      | Confidential   | {{author}} |

## 5. OWASP Categories in Scope

### Web — OWASP Top 10 (2025)
- [ ] A01 Broken Access Control
- [ ] A02 Cryptographic Failures
- [ ] A03 Injection
- [ ] A04 Insecure Design
- [ ] A05 Security Misconfiguration
- [ ] A06 Vulnerable & Outdated Components
- [ ] A07 Identification & Authentication Failures
- [ ] A08 Software & Data Integrity Failures
- [ ] A09 Security Logging & Monitoring Failures
- [ ] A10 Server-Side Request Forgery (SSRF)

### API — OWASP API Top 10 (2023)
- [ ] API1 Broken Object Level Authorization
- [ ] API2 Broken Authentication
- [ ] API3 Broken Object Property Level Authorization
- [ ] API4 Unrestricted Resource Consumption
- [ ] API5 Broken Function Level Authorization
- [ ] API6 Unrestricted Access to Sensitive Business Flows
- [ ] API7 Server-Side Request Forgery
- [ ] API8 Security Misconfiguration
- [ ] API9 Improper Inventory Management
- [ ] API10 Unsafe Consumption of APIs

### Mobile — OWASP Mobile Top 10 (2024)
- [ ] M1 Improper Credential Usage
- [ ] M2 Inadequate Supply Chain Security
- [ ] M3 Insecure Authentication/Authorization
- [ ] M4 Insufficient Input/Output Validation
- [ ] M5 Insecure Communication
- [ ] M6 Inadequate Privacy Controls
- [ ] M7 Insufficient Binary Protections
- [ ] M8 Security Misconfiguration
- [ ] M9 Insecure Data Storage
- [ ] M10 Insufficient Cryptography

## 6. Test Environment

{{environment}} <!-- e.g., dedicated staging segregated from prod, prod-like data shape, no real PII -->

- Segregated from production: yes / no
- Prod-like data: yes / no (synthetic only)
- Reset cadence: daily / on-demand

## 7. Tooling

{{tools}} <!-- e.g., Burp Suite Pro, OWASP ZAP, sqlmap, nuclei, semgrep, mobsf -->

## 8. Test Data

{{test_data_policy}} <!-- e.g., synthetic only, no production dumps, generated via factory_boy -->

## 9. Methodology

{{methodology}} <!-- e.g., OWASP ASVS L2 + OWASP Testing Guide v4.2 -->

## 10. Entry Criteria

- {{entry_criteria}} <!-- e.g., build deployed to staging, smoke tests green, RoE signed -->

## 11. Exit Criteria

- {{exit_criteria}} <!-- e.g., zero open Critical findings, all High remediated or risk-accepted -->

## 12. Risk Matrix

{{risk_matrix}}

|              | Low impact | Medium impact | High impact | Critical impact |
| ------------ | ---------- | ------------- | ----------- | --------------- |
| Rare         | Low        | Low           | Medium      | High            |
| Unlikely     | Low        | Medium        | Medium      | High            |
| Likely       | Medium     | Medium        | High        | Critical        |
| Almost cert. | Medium     | High          | Critical    | Critical        |

## 13. Schedule

- Start: {{schedule_start}} <!-- YYYY-MM-DD -->
- End: {{schedule_end}}
- Milestones: kickoff, mid-engagement check-in, draft report, retest

## 14. Reporting Cadence

{{reporting_cadence}} <!-- e.g., daily Slack standup, weekly written summary, immediate notification on Critical -->

## 15. Stakeholders

| Role             | Name              | Contact |
| ---------------- | ----------------- | ------- |
| Engagement lead  | {{author}}        |         |
| Security owner   | {{stakeholders}}  |         |
| Product owner    |                   |         |
| Incident contact |                   |         |
