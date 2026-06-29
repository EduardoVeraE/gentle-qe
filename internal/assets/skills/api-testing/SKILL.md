---
name: api-testing
description: "Trigger: API test, REST, GraphQL, schema validation, OpenAPI, contract test. Functional API testing with Playwright TS and REST Assured."
---

# API Testing (Playwright + REST Assured)

Comprehensive API testing skill covering both Playwright TypeScript (request fixture, Supertest, Zod) and Java (REST Assured, AssertJ, JSON Schema Validator). Provides deep domain expertise for the `api-tester-specialist` agent.

## When to Use This Skill

- Create API tests for REST or GraphQL endpoints
- Validate request/response schemas (Zod, JSON Schema)
- Test authentication flows (OAuth2, JWT, API keys, Bearer tokens)
- Verify error handling (400, 401, 403, 404, 409, 422, 500)
- Test pagination, filtering, sorting edge cases
- Validate idempotency for PUT/DELETE operations
- Contract testing between services
- Rate limiting validation

## ISTQB API Testing Levels

| Level       | Scope                          | Tools                                            | Reference                                                                |
| ----------- | ------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------ |
| Component   | Single endpoint, mocked deps   | Playwright request, Supertest, MSW               | [api-testing-levels-istqb.md](./references/api-testing-levels-istqb.md)  |
| Integration | Multi-service real deps        | REST Assured, Playwright, contract tests         | [api-testing-levels-istqb.md](./references/api-testing-levels-istqb.md)  |
| System      | Full E2E API flow              | Playwright + real env                            | [api-testing-levels-istqb.md](./references/api-testing-levels-istqb.md)  |

## OpenAPI-First Workflow

1. Lint the spec with Spectral.
2. Generate types with `openapi-typescript`.
3. Validate request bodies against the spec (consumer side).
4. Validate response bodies against the spec (provider side).
5. Detect breaking changes with `oasdiff` in CI.
6. Generate examples and edge cases from the spec (Faker + spec).

See [openapi-driven-testing.md](./references/openapi-driven-testing.md) for full workflow details.

## Prerequisites

| Stack      | Requirements                                                          |
| ---------- | --------------------------------------------------------------------- |
| TypeScript | Node.js 18+, `@playwright/test` or `supertest`, `zod`                 |
| Java       | Java 21+, REST Assured 5.x, AssertJ, Jackson, `json-schema-validator` |

## Core Principles

1. **Schema validation on every response** — never trust an unvalidated response
2. **Test all HTTP status codes** — happy path AND error states
3. **Auth testing is mandatory** — verify 401/403 for protected endpoints
4. **Data-driven** — test with valid, invalid, boundary, and empty values
5. **Stateless where possible** — each test cleans up or uses unique data

## Mandatory Headers Convention

Every API test should assert the presence and shape of headers in these eight categories. See [headers-and-contracts.md](./references/headers-and-contracts.md) for the full catalog with examples and validation patterns.

- **Auth** — `Authorization`, API keys, session cookies
- **Content negotiation** — `Accept`, `Content-Type`, `Accept-Language`
- **Idempotency** — `Idempotency-Key` for unsafe-but-repeatable mutations
- **Tracing** — W3C `traceparent`, `tracestate`, correlation IDs
- **Pagination** — `Link` (RFC 5988), `X-Total-Count`, cursor headers
- **Rate limit** — `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`
- **Security** — `Strict-Transport-Security`, `X-Content-Type-Options`, CORS headers
- **Caching** — `ETag`, `Cache-Control`, `Last-Modified`, `Vary`

## Contract Testing

Two complementary paths — choose based on coupling and ownership:

- **Consumer-driven via Pact** — when consumer and provider teams are separate and need to negotiate the contract bidirectionally. Pact files live with the consumer, are verified by the provider in CI. Covered in depth in the `qa-contract-pact` skill (PACT-JS, PACT-JVM, broker, can-i-deploy, message contracts).
- **OpenAPI as contract** — when the provider owns the spec and consumers conform to it. Validate both request and response shapes against the spec on every test run. Covered in [openapi-driven-testing.md](./references/openapi-driven-testing.md).

Use Pact for cross-team contracts where breaking changes need explicit negotiation. Use OpenAPI when the spec is the source of truth and consumers track it.

## Quick Reference — Playwright

```typescript
import { test, expect } from "@playwright/test";

test("GET /api/users returns 200 with valid schema", async ({ request }) => {
  const response = await request.get("/api/users");
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  expect(body).toMatchObject({ data: expect.any(Array) });
});
```

## Quick Reference — REST Assured

```java
import static io.restassured.RestAssured.*;
import static org.hamcrest.Matchers.*;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

@Test
@DisplayName("GET /api/users returns 200 with valid schema")
void getUsers() {
    String token = "test-token";

    given()
        .header("Authorization", "Bearer " + token)
    .when()
        .get("/api/users")
    .then()
        .statusCode(200)
        .body("data", is(instanceOf(List.class)))
        .body("data.size()", greaterThan(0));
}
```

## Common Rationalizations

> Common shortcuts and "good enough" excuses that erode test quality — and the reality behind each.

| Rationalization                                 | Reality                                                                                                      |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| "Schema validation is overkill"                 | Without schema validation, a silent field rename becomes a production incident. Validate every response.     |
| "Happy path testing is enough"                  | Error states (400, 401, 403, 404, 409, 500) are where real failures happen. Test all status codes.           |
| "Auth tests can wait"                           | Unauthenticated access to protected endpoints is a security vulnerability, not a backlog item.               |
| "This endpoint won't change"                    | APIs evolve. Contract tests catch breaking changes before they reach production.                             |
| "Manual API testing with Postman is sufficient" | Manual testing isn't repeatable, can't run in CI, and doesn't scale. Automate API tests.                     |
| "Idempotency doesn't matter"                    | Duplicate requests happen in production. Without idempotency testing, you get duplicate records and charges. |

---

## References

| Document                                                         | Content                                             |
| ---------------------------------------------------------------- | --------------------------------------------------- |
| [REST API Patterns](./references/rest-api-patterns.md)           | CRUD, pagination, filtering, error patterns         |
| [Playwright API Testing](./references/playwright-api-testing.md) | Request fixture, Supertest, TypeScript patterns     |
| [REST Assured Testing](./references/rest-assured-testing.md)     | REST Assured, AssertJ, Java patterns                |
| [Schema Validation](./references/schema-validation.md)           | Zod (TS), JSON Schema (Java), strict vs loose       |
| [Contract Testing](./references/contract-testing.md)             | Request/response contracts, idempotency, versioning |

## Templates

- [Playwright API Spec](./templates/playwright-api-spec.ts) — starter test file for API testing
- [REST Assured Test](./templates/rest-assured-test.java) — starter Java test class

## Scripts

- [API Health Check](./scripts/api-health-check.sh) — validate API endpoints respond correctly

## Troubleshooting

| Issue                          | Solution                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------ |
| 401 on authenticated endpoints | Verify token is fresh; check expiry; re-authenticate                           |
| Flaky API tests                | Add retry logic; check for rate limiting; use unique test data                 |
| Schema validation too strict   | Use `.passthrough()` (Zod) or `additionalProperties: true` for flexible fields |
| Timeout on slow endpoints      | Increase `timeout` in request options; check for server load                   |

---

## Verification

After completing this skill's workflow, confirm:

- [ ] **All CRUD operations tested** — POST, GET, PUT, PATCH, DELETE covered for the resource
- [ ] **Status codes verified** — Success (2xx) AND error codes (4xx, 5xx) tested
- [ ] **Schema validation in place** — Every response validated against a schema (Zod or JSON Schema)
- [ ] **Authentication tested** — 401 returned for protected endpoints without valid credentials
- [ ] **Idempotency verified** — PUT/DELETE produce same result when called multiple times
- [ ] **Edge cases covered** — Empty payloads, invalid types, boundary values, SQL injection attempts
- [ ] **All tests pass** — Playwright API tests or REST Assured tests exit successfully

---

## Exclusions

This skill is scoped to functional, contract, schema, and OpenAPI testing. It does NOT cover:

- NOT for security testing (XSS, SQLi, BOLA, JWT attacks, OWASP API Top 10) — use `qa-owasp-security`
- NOT for performance/load testing — use `k6-load-test`
- NOT for E2E browser flows — use `playwright-e2e-testing` or `selenium-e2e-testing`
- NOT for mobile API testing harness — use `qa-mobile-testing` (the API call patterns from this skill still apply, but device/sim setup belongs there)
