<!-- Skill: qa-owasp-security · Template: vuln-report -->
<!-- Placeholders: {{title}}, {{date}}, {{author}}, {{project}}, {{vuln_id}}, {{severity}}, {{cvss_vector}}, {{cvss_score}}, {{owasp_category}}, {{owasp_surface}}, {{cwe}}, {{component}}, {{version}}, {{steps_to_reproduce}}, {{poc}}, {{impact}}, {{likelihood}}, {{remediation}}, {{references}}, {{discovered_by}}, {{date_discovered}}, {{status}}, {{retest_date}}, {{retest_result}}, {{ticket_ref}} -->

# Vulnerability Report — {{title}}

| Field         | Value                                       |
| ------------- | ------------------------------------------- |
| Vuln ID       | {{vuln_id}} <!-- e.g., VULN-2026-0007 -->   |
| Project       | {{project}}                                 |
| Component     | {{component}} @ {{version}}                 |
| Discovered by | {{discovered_by}}                           |
| Date found    | {{date_discovered}}                         |
| Reported on   | {{date}}                                    |
| Status        | {{status}} <!-- Open / Fixed / Risk-accepted / Won't-fix --> |
| Ticket        | {{ticket_ref}}                              |

## 1. Severity

- Severity: {{severity}} <!-- Critical / High / Medium / Low / Info -->
- CVSS 4.0 vector: {{cvss_vector}} <!-- e.g., CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N -->
- CVSS 4.0 score: {{cvss_score}} <!-- e.g., 9.3 -->

## 2. OWASP Category Mapping

- Surface: {{owasp_surface}} <!-- Web / API / Mobile -->
- Category: {{owasp_category}} <!-- e.g., A03 Injection (Web 2025), API1 BOLA (2023), M4 Input/Output Validation (2024) -->
- CWE: {{cwe}} <!-- e.g., CWE-89, CWE-639 -->

## 3. Affected Component

{{component}} version {{version}}

- Endpoint / file path / module
- Build identifier
- Environment (staging, prod-like, prod)

## 4. Steps to Reproduce

1. {{steps_to_reproduce}}
2. <!-- Continue numbered steps. Be deterministic. -->
3. <!-- Include credentials, headers, payloads. -->
4. <!-- Final step shows the violation observed. -->

## 5. Proof of Concept

```http
{{poc}}
POST /api/v2/orders/42 HTTP/1.1
Host: api.example.com
Authorization: Bearer <attacker-token>

{"customer_id": 99}
```

```text
HTTP/1.1 200 OK
{"order_id": 42, "customer_id": 99, "total": "1234.56"}
<!-- The attacker, authenticated as customer 7, modified an order belonging to customer 42. -->
```

## 6. Impact

{{impact}} <!-- e.g., Cross-tenant data modification, full account takeover, RCE on application server -->

- Confidentiality: High / Low / None
- Integrity: High / Low / None
- Availability: High / Low / None
- Scope of affected users: <!-- all customers / single tenant / authenticated only -->

## 7. Likelihood

{{likelihood}} <!-- High / Medium / Low + reasoning, e.g., "High — exploit requires only a valid low-priv token; no special tooling" -->

## 8. Recommended Remediation

{{remediation}} <!-- e.g., Enforce object-level authorization in the OrderService.update() method by checking that order.customer_id == request.user.customer_id BEFORE persisting. Add a regression test. -->

### Defense in depth
- Validate authorization at the service layer, not only at the controller
- Add structured audit log entries on every mutation
- Add automated test reproducing the abuse case

## 9. References

- OWASP cheat sheet: {{references}}
- CWE: https://cwe.mitre.org/data/definitions/{{cwe}}.html
- Vendor advisory: <!-- link if applicable -->
- Internal ticket: {{ticket_ref}}

## 10. Retest

| Field          | Value             |
| -------------- | ----------------- |
| Retest date    | {{retest_date}}   |
| Retest result  | {{retest_result}} <!-- Fixed / Still vulnerable / Partially fixed --> |
| Retested by    | {{author}}        |
| Evidence path  |                   |

## 11. Status History

| Date     | Status        | Note                            |
| -------- | ------------- | ------------------------------- |
| {{date_discovered}} | Open  | Initial discovery               |
| {{date}} | {{status}}    |                                 |
