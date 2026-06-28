<!-- Skill: qa-owasp-security · Template: threat-model -->
<!-- Placeholders: {{title}}, {{date}}, {{author}}, {{project}}, {{system_overview}}, {{trust_boundaries}}, {{dfd}}, {{assets}}, {{element_name}}, {{stride_s}}, {{stride_t}}, {{stride_r}}, {{stride_i}}, {{stride_d}}, {{stride_e}}, {{mitigation}}, {{linked_test}}, {{threat_id}}, {{dread_d}}, {{dread_r}}, {{dread_e}}, {{dread_a}}, {{dread_di}}, {{dread_score}}, {{owner}}, {{review_cycle}} -->

# Threat Model — {{title}}

| Field   | Value         |
| ------- | ------------- |
| Project | {{project}}   |
| Author  | {{author}}    |
| Date    | {{date}}      |
| Review  | {{review_cycle}} <!-- e.g., quarterly --> |

## 1. System Overview

{{system_overview}} <!-- e.g., Multi-tenant SaaS with web SPA, REST API, Postgres, Redis, S3 -->

## 2. Assets

| ID  | Asset       | Sensitivity                | Owner     |
| --- | ----------- | -------------------------- | --------- |
| A1  | {{assets}}  | Confidential / Public / PII | {{owner}} |

## 3. Trust Boundaries

{{trust_boundaries}} <!-- e.g., Internet ↔ Edge, Edge ↔ App tier, App ↔ Data tier, App ↔ 3rd-party APIs -->

## 4. Data Flow Diagram (DFD)

```
{{dfd}}
+----------+       +--------+       +----------+       +-----------+
|  User    | --->  |  CDN   | --->  |   API    | --->  |  Database |
|(browser) |       | (edge) |       | (svc)    |       |  (RDS)    |
+----------+       +--------+       +----------+       +-----------+
     |                                   |
     v                                   v
[Trust boundary]                  [Trust boundary]
```

Replace the placeholder block with the real DFD. Mark every trust boundary crossing.

## 5. Per-Element STRIDE Walkthrough

Legend: S=Spoofing, T=Tampering, R=Repudiation, I=Information disclosure, D=Denial of service, E=Elevation of privilege.
Mark Y if the threat applies to the element, N if not.

| Element            | S            | T            | R            | I            | D            | E            | Mitigation         | Test            |
| ------------------ | ------------ | ------------ | ------------ | ------------ | ------------ | ------------ | ------------------ | --------------- |
| {{element_name}}   | {{stride_s}} | {{stride_t}} | {{stride_r}} | {{stride_i}} | {{stride_d}} | {{stride_e}} | {{mitigation}}     | {{linked_test}} |
| User → CDN         | Y            | Y            | N            | Y            | Y            | N            | TLS 1.3, HSTS      | TC-SEC-001      |
| CDN → API          | Y            | Y            | Y            | Y            | Y            | Y            | mTLS, WAF          | TC-SEC-002      |
| API → Database     | Y            | Y            | Y            | Y            | Y            | Y            | IAM, encryption    | TC-SEC-003      |

## 6. Risk-Rated Threats (DREAD)

DREAD scores each dimension 1-10. Average = risk score. Score >= 7 is High.

| ID            | Threat description    | D            | R            | E            | A            | DI            | Score          | Owner     |
| ------------- | --------------------- | ------------ | ------------ | ------------ | ------------ | ------------- | -------------- | --------- |
| {{threat_id}} | <!-- describe -->     | {{dread_d}}  | {{dread_r}}  | {{dread_e}}  | {{dread_a}}  | {{dread_di}}  | {{dread_score}} | {{owner}} |
| TM-001        | JWT replay across tenants | 8 | 7 | 6 | 9 | 7 | 7.4 | sec-team |
| TM-002        | SSRF via image-fetch worker | 9 | 6 | 7 | 8 | 6 | 7.2 | platform |

## 7. Mitigations

| Threat ID     | Mitigation                                | Status                  | Owner     | Due        |
| ------------- | ----------------------------------------- | ----------------------- | --------- | ---------- |
| {{threat_id}} | {{mitigation}}                            | Planned / In progress / Done | {{owner}} | {{date}} |

## 8. Linked Tests

| Threat ID     | Test ID         | Test type               | Result            |
| ------------- | --------------- | ----------------------- | ----------------- |
| {{threat_id}} | {{linked_test}} | Unit / Integration / Pentest | Pass / Fail / N/A |

## 9. Review Log

| Date      | Reviewer    | Change                        |
| --------- | ----------- | ----------------------------- |
| {{date}}  | {{author}}  | Initial threat model created  |

## 10. Notes

- Threat model is a living document. Re-run STRIDE on every architecture change crossing a trust boundary.
- A missing mitigation column is a finding in itself — never leave blank without justification.
- Cross-reference every threat with at least one automated or manual test.
