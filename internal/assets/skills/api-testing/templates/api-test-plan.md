<!-- Skill: api-testing · Template: api-test-plan -->
<!-- Placeholders: {{plan_title}}, {{project_name}}, {{plan_date}}, {{author}}, {{component_scope}}, {{integration_scope}}, {{system_scope}}, {{openapi_spec_url}}, {{openapi_version}}, {{dev_base_url}}, {{stage_base_url}}, {{prod_mirror_base_url}}, {{auth_strategy}}, {{token_source}}, {{test_data_strategy}}, {{tooling_stack}}, {{contract_approach}}, {{pact_broker_url}}, {{schedule_start}}, {{schedule_end}}, {{entry_criteria}}, {{exit_criteria}}, {{reporting_channel}}, {{stakeholders}}, {{run_id}} -->

# {{plan_title}} <!-- e.g., Orders API v3 — Q2 Test Plan -->

| Field | Value |
| ----- | ----- |
| Project | {{project_name}} |
| Date | {{plan_date}} |
| Author | {{author}} |
| OpenAPI spec | {{openapi_spec_url}} (version `{{openapi_version}}`) |

## 1. Scope (ISTQB levels)

We split scope by ISTQB test level. An item appears in EXACTLY ONE level — duplication across levels is a smell.

### 1.1 Component (unit + slice)
- {{component_scope}} <!-- e.g., Validation rules in OrderValidator, money rounding in PriceCalculator -->
- Schema validators (Zod / Pydantic / Bean Validation) in isolation
- Mappers (DTO ↔ domain) with table-driven cases

### 1.2 Integration
- {{integration_scope}} <!-- e.g., OrderService against in-memory Postgres, Stripe webhook handler against recorded fixtures -->
- DB migrations applied on a fresh schema
- Outbound HTTP clients with WireMock / MSW
- Message handlers with embedded broker

### 1.3 System (end-to-end through the deployed API)
- {{system_scope}} <!-- e.g., Full /orders happy path through gateway → service → DB → outbox -->
- Auth flows hitting the real IdP
- Rate-limiting, pagination, idempotency on real infra
- Contract verification against staging

## 2. API surface

- OpenAPI spec: {{openapi_spec_url}}
- Spec version under test: `{{openapi_version}}`
- Generated client: regenerate on every spec change; commit lockfile.
- Operations covered: every operation in the spec MUST appear in `openapi-test-checklist.md`.

## 3. Environments

| Env | Base URL | Purpose | Data |
| --- | -------- | ------- | ---- |
| dev | {{dev_base_url}} | Devs and component/integration runs | Synthetic, reset on demand |
| stage | {{stage_base_url}} | System tests, contract verification, pre-deploy gate | Synthetic + anonymized prod subset |
| prod-mirror | {{prod_mirror_base_url}} | Smoke + read-only canaries | Anonymized prod snapshot |

Production is NOT a test environment. Read-only canaries with opaque tokens only.

## 4. Auth strategy

- {{auth_strategy}} <!-- e.g., OAuth2 Bearer + JWT validated against /.well-known/jwks.json -->
- Tokens fetched via {{token_source}} <!-- e.g., client_credentials against {{idp_url}}/oauth/token -->
- Test tokens are short-lived (≤ 15 min), scoped per test, never reused across users.
- Negative auth cases (no token, expired token, wrong audience, wrong scope) are MANDATORY for every protected endpoint.

## 5. Test data strategy

{{test_data_strategy}} <!-- e.g., Synthetic via factory functions per resource; anonymized prod snapshot for read-heavy reports; opaque IDs everywhere — no PII in fixtures -->

Rules:
- Synthetic by default. Use factories, not raw JSON, so schema drift breaks tests early.
- Anonymized data only when synthetic cannot reproduce shape (e.g., long-tail catalogs).
- Opaque identifiers (UUIDs / ULIDs) — never reuse prod IDs.
- Test data is namespaced (e.g., `qa-{{run_id}}-...`) and cleaned up post-run.

## 6. Tooling

{{tooling_stack}} <!-- e.g., Playwright (system), REST Assured (integration), Spectral (lint), oasdiff (breaking-change diff), Pact (consumer-driven contracts) -->

| Layer | Tool |
| ----- | ---- |
| Spec lint | Spectral |
| Breaking-change diff | oasdiff |
| Component | Vitest / JUnit / pytest |
| Integration | REST Assured / Supertest |
| System | Playwright API |
| Contract | Pact + Pact Broker / OpenAPI verification |

## 7. Mandatory headers per endpoint group

Tracked in `mandatory-headers-checklist.md`. Plan-level summary:

| Group | Auth | Idempotency-Key | Trace (W3C `traceparent`) | `X-Request-Id` |
| ----- | ---- | --------------- | ------------------------- | -------------- |
| `/orders/*` write | Required | Required (POST/PUT/PATCH) | Required | Required |
| `/orders/*` read | Required | N/A | Required | Required |
| `/users/*` | Required | Required (POST/PATCH) | Required | Required |
| `/health`, `/metrics` | Optional | N/A | Optional | Optional |

## 8. Contract testing approach

{{contract_approach}} <!-- e.g., OpenAPI (provider) + Pact (consumer-driven) — both run on every PR -->

- OpenAPI: spec is the source of truth. Every operation has a contract test that validates request and response against the schema.
- Pact: each consumer publishes a contract; provider verifies on every PR.
- Pact broker: {{pact_broker_url}}
- can-i-deploy gates promotion to stage and prod.

## 9. Risk matrix

| Risk | Likelihood | Impact | Mitigation |
| ---- | ---------- | ------ | ---------- |
| Spec drift (code ≠ OpenAPI) | High | High | Spectral + contract tests on every PR |
| Breaking change shipped | Medium | High | oasdiff gate; semver discipline |
| Flaky auth tests | Medium | Medium | Token cache per worker, retry on 401 once |
| Test data bleed across runs | Medium | Medium | Namespaced IDs; teardown hooks |
| Rate limits in CI | Low | Medium | Dedicated CI tenant; backoff |

## 10. Schedule

- Start: {{schedule_start}}
- End: {{schedule_end}}
- Milestones: spec freeze → contract green → system green → exit review.

## 11. Entry criteria

{{entry_criteria}}
- OpenAPI spec linted and merged.
- Test environments reachable and seeded.
- Auth credentials provisioned.
- CI pipeline green on the target branch.

## 12. Exit criteria

{{exit_criteria}}
- 100% of in-scope operations have at least one test at each applicable level.
- Spectral: 0 errors, ≤ documented warnings.
- oasdiff: 0 breaking changes vs. previous release (or explicitly accepted).
- Contract tests: 100% pass; can-i-deploy returns OK.
- No P1/P2 bugs open.

## 13. Reporting

- Channel: {{reporting_channel}} <!-- e.g., #qa-orders Slack + weekly written summary -->
- Dashboards: contract pass rate, flake rate, p95 latency per operation.
- Defects logged with operation ID, request/response, trace ID.

## 14. Stakeholders

{{stakeholders}} <!-- e.g., API owner: @alice; QA lead: @bob; SRE on-call: rotation; Product: @carol -->
