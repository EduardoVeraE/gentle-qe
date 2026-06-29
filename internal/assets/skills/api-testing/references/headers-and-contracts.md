# Headers and Contracts

Two complementary references in one document:

- **Part 1** — Catalog of mandatory and recommended HTTP headers grouped by category, with TS Playwright assertions for each.
- **Part 2** — Contract testing principles, when to use OpenAPI vs Pact, and a Pact-in-3-steps skeleton.

HTTP/1.1 header names are case-insensitive (RFC 7230 section 3.2). Examples here use conventional casing (`Content-Type`, `X-Request-ID`). Playwright's `response.headers()` lowercases all keys, so assertions read `response.headers()['content-type']`. When SETTING a header via `request.post({ headers })`, casing is preserved on the wire but treated as equivalent by HTTP servers.

---

# Part 1 — Mandatory Headers Catalog

## Quick Reference Table

| Header                        | Category            | Direction      | Required when                                       | Tested by                              |
| ----------------------------- | ------------------- | -------------- | --------------------------------------------------- | -------------------------------------- |
| `Authorization`               | Auth                | Request        | Endpoint requires auth                              | Auth assertion + 401/403 path          |
| `X-Api-Key`                   | Auth                | Request        | API-key style auth (no JWT)                         | Auth assertion + 401 missing-key       |
| `X-Auth-Token`                | Auth                | Request        | Custom token auth                                   | Auth assertion + 401 invalid           |
| `Content-Type`                | Content negotiation | Request        | Request has body                                    | 415 when wrong, 200 when right         |
| `Accept`                      | Content negotiation | Request        | Server supports multiple media types                | 406 when unsupported                   |
| `Content-Encoding`            | Content negotiation | Both           | Body is compressed                                  | Round-trip with gzip/br                |
| `Accept-Encoding`             | Content negotiation | Request        | Client supports compression                         | Server returns matching `Content-Encoding` |
| `Accept-Language`             | Content negotiation | Request        | Server supports i18n                                | Localized response body                |
| `Idempotency-Key`             | Idempotency         | Request        | Non-idempotent verb (POST, PATCH) needs safe retry  | Replay → same response; conflict → 422 |
| `traceparent`                 | Tracing             | Both           | W3C tracing enabled                                 | Header echoed/propagated in response   |
| `tracestate`                  | Tracing             | Both           | Vendor-specific trace metadata present              | Round-trip preserved                   |
| `X-Request-ID`                | Tracing             | Both           | Legacy trace correlation                            | Echoed in response, present in logs    |
| `X-Correlation-ID`            | Tracing             | Both           | Cross-service correlation (legacy)                  | Echoed downstream                      |
| `Link`                        | Pagination          | Response       | List endpoint with more pages                       | Parse `rel="next"`, follow, validate   |
| `X-Total-Count`               | Pagination          | Response       | UI needs total record count                         | Sum of pages == total                  |
| `X-RateLimit-Limit`           | Rate limit          | Response       | Endpoint is rate-limited                            | Header present on every response       |
| `X-RateLimit-Remaining`       | Rate limit          | Response       | Endpoint is rate-limited                            | Decreases per call                     |
| `X-RateLimit-Reset`           | Rate limit          | Response       | Endpoint is rate-limited                            | Epoch seconds in future                |
| `Retry-After`                 | Rate limit          | Response       | 429 or 503 returned                                 | Numeric seconds or HTTP-date           |
| `Strict-Transport-Security`   | Security            | Response       | HTTPS endpoint                                      | Header present + `max-age` >= 31536000 |
| `X-Frame-Options`             | Security            | Response       | HTML response                                       | `DENY` or `SAMEORIGIN`                 |
| `X-Content-Type-Options`      | Security            | Response       | Always (recommended)                                | Equals `nosniff`                       |
| `Content-Security-Policy`     | Security            | Response       | HTML response                                       | Policy string matches expected         |
| `Referrer-Policy`             | Security            | Response       | Always (recommended)                                | One of allowed values                  |
| `Permissions-Policy`          | Security            | Response       | Browser-facing endpoint                             | Restricts sensitive features           |
| `ETag`                        | Caching             | Response       | Cacheable resource                                  | Round-trip with `If-None-Match` → 304  |
| `If-None-Match`               | Caching             | Request        | Conditional GET                                     | Server returns 304 on match            |
| `Cache-Control`               | Caching             | Both           | Always (recommended)                                | `max-age` or `no-store` per endpoint   |
| `Last-Modified`               | Caching             | Response       | Resource has modification timestamp                 | Round-trip with `If-Modified-Since`    |
| `If-Modified-Since`           | Caching             | Request        | Conditional GET (date-based)                        | Server returns 304 on match            |

---

## 1. Auth

Authentication credentials presented by the client. The server enforces them and returns 401 (unauthenticated) or 403 (authenticated but unauthorized).

### `Authorization: Bearer <jwt>`

- **Direction:** request only.
- **Required when:** endpoint sits behind JWT-based auth (OAuth2 access tokens, OpenID Connect, custom JWT).
- **Set by:** client.
- **Mandatory:** yes for protected endpoints. Missing → 401.

```typescript
import { test, expect } from "@playwright/test";

test("GET /api/me with Bearer token returns 200", async ({ request }) => {
  const token = process.env.JWT_TOKEN!;
  const response = await request.get("/api/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(response.status()).toBe(200);
});

test("GET /api/me without token returns 401", async ({ request }) => {
  const response = await request.get("/api/me");
  expect(response.status()).toBe(401);
  expect(response.headers()["www-authenticate"]).toMatch(/Bearer/);
});
```

### `Authorization: Basic <base64(user:pass)>`

- **Direction:** request only.
- **Required when:** legacy or service-to-service basic auth.
- **Set by:** client.
- **Mandatory:** yes for protected endpoints.

```typescript
test("GET /admin with Basic auth returns 200", async ({ request }) => {
  const credentials = Buffer.from("admin:s3cret").toString("base64");
  const response = await request.get("/admin", {
    headers: { Authorization: `Basic ${credentials}` },
  });
  expect(response.status()).toBe(200);
});
```

### `X-Api-Key`

- **Direction:** request only.
- **Required when:** API-key auth schemes (no JWT, no OAuth).
- **Set by:** client.
- **Mandatory:** yes for protected endpoints. Missing → 401.

```typescript
test("GET /api/data with API key returns 200", async ({ request }) => {
  const response = await request.get("/api/data", {
    headers: { "X-Api-Key": process.env.API_KEY! },
  });
  expect(response.status()).toBe(200);
});

test("GET /api/data with invalid key returns 401", async ({ request }) => {
  const response = await request.get("/api/data", {
    headers: { "X-Api-Key": "invalid-key" },
  });
  expect(response.status()).toBe(401);
});
```

### `X-Auth-Token`

- **Direction:** request only.
- **Required when:** custom session-token schemes.
- **Set by:** client.
- **Mandatory:** yes for protected endpoints.

```typescript
test("GET /api/session with X-Auth-Token returns 200", async ({ request }) => {
  const response = await request.get("/api/session", {
    headers: { "X-Auth-Token": process.env.SESSION_TOKEN! },
  });
  expect(response.status()).toBe(200);
});
```

### OAuth2 Bearer scopes assertion

OAuth2 tokens carry SCOPES that limit what the bearer can do. Tests should verify both:

1. A token WITH the required scope succeeds.
2. A token WITHOUT it returns 403.

```typescript
test("DELETE /api/users/1 requires admin:write scope", async ({ request }) => {
  const readOnlyToken = process.env.READ_ONLY_TOKEN!;
  const adminToken = process.env.ADMIN_TOKEN!;

  // Read-only token: 403 forbidden
  const forbidden = await request.delete("/api/users/1", {
    headers: { Authorization: `Bearer ${readOnlyToken}` },
  });
  expect(forbidden.status()).toBe(403);

  // Admin token: 204 success
  const allowed = await request.delete("/api/users/1", {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  expect(allowed.status()).toBe(204);
});
```

---

## 2. Content Negotiation

Negotiate media type, language, and compression between client and server.

### `Content-Type` (request)

- **Direction:** request when there is a body; response always.
- **Required when:** request has a body.
- **Set by:** client (request); server (response).
- **Mandatory:** yes when body present. Missing or wrong → 415 Unsupported Media Type.

```typescript
test("POST /api/users with wrong Content-Type returns 415", async ({ request }) => {
  const response = await request.post("/api/users", {
    headers: { "Content-Type": "text/plain" },
    data: "name=Test",
  });
  expect(response.status()).toBe(415);
});

test("POST /api/users with application/json returns 201", async ({ request }) => {
  const response = await request.post("/api/users", {
    headers: { "Content-Type": "application/json" },
    data: { name: "Test", email: "test@example.com" },
  });
  expect(response.status()).toBe(201);
  expect(response.headers()["content-type"]).toMatch(/application\/json/);
});
```

### `Accept` (request)

- **Direction:** request only.
- **Required when:** server supports multiple representations (JSON, XML, CSV).
- **Set by:** client.
- **Mandatory:** recommended. Wrong/unsupported → 406 Not Acceptable.

```typescript
test("GET /api/users with Accept: application/xml returns 406 when only JSON is supported", async ({ request }) => {
  const response = await request.get("/api/users", {
    headers: { Accept: "application/xml" },
  });
  expect(response.status()).toBe(406);
});

test("GET /api/users with Accept: application/json returns 200", async ({ request }) => {
  const response = await request.get("/api/users", {
    headers: { Accept: "application/json" },
  });
  expect(response.status()).toBe(200);
  expect(response.headers()["content-type"]).toMatch(/application\/json/);
});
```

### `Content-Encoding` (gzip / br)

- **Direction:** both — request when body is compressed; response when server compresses.
- **Required when:** body is compressed.
- **Set by:** whoever compresses.
- **Mandatory:** required if body is compressed.

```typescript
test("GET /api/large returns gzip-compressed body", async ({ request }) => {
  const response = await request.get("/api/large", {
    headers: { "Accept-Encoding": "gzip, br" },
  });
  expect(response.status()).toBe(200);
  expect(response.headers()["content-encoding"]).toMatch(/^(gzip|br)$/);
});
```

### `Accept-Encoding`

- **Direction:** request only.
- **Required when:** client supports decompression.
- **Set by:** client.
- **Mandatory:** recommended for performance.

```typescript
test("GET /api/data honors Accept-Encoding: br", async ({ request }) => {
  const response = await request.get("/api/data", {
    headers: { "Accept-Encoding": "br" },
  });
  expect(response.headers()["content-encoding"]).toBe("br");
});
```

### `Accept-Language`

- **Direction:** request only.
- **Required when:** server supports i18n.
- **Set by:** client.
- **Mandatory:** recommended.

```typescript
test("GET /api/messages returns Spanish when Accept-Language is es", async ({ request }) => {
  const response = await request.get("/api/messages", {
    headers: { "Accept-Language": "es-ES" },
  });
  expect(response.status()).toBe(200);
  expect(response.headers()["content-language"]).toMatch(/^es/);
  const body = await response.json();
  expect(body.greeting).toMatch(/hola/i);
});
```

---

## 3. Idempotency

`Idempotency-Key` lets clients retry non-idempotent operations safely. The server stores the key and the response for a window (e.g., 24h). Same key → returns cached response. Different request body with the same key → 422.

This follows the [IETF draft-ietf-httpapi-idempotency-key-header](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/) and Stripe's convention.

- **Direction:** request only.
- **Required when:** safe-retry desired on POST/PATCH/DELETE that creates resources or charges money.
- **Set by:** client (UUID v4 recommended).
- **Mandatory:** depends on API. Stripe makes it OPTIONAL; some APIs make it MANDATORY for write operations.

```typescript
import { randomUUID } from "node:crypto";

test("POST /api/payments with same Idempotency-Key returns same response", async ({ request }) => {
  const key = randomUUID();
  const payload = { amount: 1000, currency: "USD" };

  const first = await request.post("/api/payments", {
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": key,
    },
    data: payload,
  });
  expect(first.status()).toBe(201);
  const firstBody = await first.json();

  // Replay with same key → same response, no second charge
  const second = await request.post("/api/payments", {
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": key,
    },
    data: payload,
  });
  expect(second.status()).toBe(201);
  const secondBody = await second.json();
  expect(secondBody.id).toBe(firstBody.id);
});

test("POST /api/payments with same key but different body returns 422", async ({ request }) => {
  const key = randomUUID();
  await request.post("/api/payments", {
    headers: { "Content-Type": "application/json", "Idempotency-Key": key },
    data: { amount: 1000, currency: "USD" },
  });

  const conflict = await request.post("/api/payments", {
    headers: { "Content-Type": "application/json", "Idempotency-Key": key },
    data: { amount: 9999, currency: "USD" }, // Different body, same key
  });
  expect(conflict.status()).toBe(422);
});
```

---

## 4. Tracing (W3C)

`traceparent` and `tracestate` (W3C Trace Context, [W3C Recommendation](https://www.w3.org/TR/trace-context/)) propagate distributed traces. Legacy headers (`X-Request-ID`, `X-Correlation-ID`) serve the same purpose pre-W3C.

### `traceparent`

- **Direction:** both — client sends, server propagates to downstream and echoes in response.
- **Required when:** distributed tracing enabled.
- **Set by:** client (or first hop). Format: `version-trace-id-parent-id-flags`.
- **Mandatory:** recommended for any service in a traced call graph.

```typescript
test("GET /api/orders propagates traceparent", async ({ request }) => {
  const traceparent = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01";
  const response = await request.get("/api/orders", {
    headers: { traceparent },
  });
  expect(response.status()).toBe(200);
  // Response must carry a traceparent (echoed or new span under same trace-id)
  const responseTrace = response.headers()["traceparent"];
  expect(responseTrace).toMatch(/^00-0af7651916cd43dd8448eb211c80319c-[0-9a-f]{16}-0[01]$/);
});
```

### `tracestate`

- **Direction:** both.
- **Required when:** vendor-specific trace metadata accompanies traceparent.
- **Set by:** any tracing vendor in the chain.
- **Mandatory:** optional companion to traceparent.

```typescript
test("GET /api/orders preserves tracestate", async ({ request }) => {
  const response = await request.get("/api/orders", {
    headers: {
      traceparent: "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
      tracestate: "vendor1=value1,vendor2=value2",
    },
  });
  expect(response.headers()["tracestate"]).toContain("vendor1=value1");
});
```

### `X-Request-ID` (legacy)

- **Direction:** both.
- **Required when:** legacy correlation, no W3C tracing.
- **Set by:** client (or generated by edge proxy).
- **Mandatory:** recommended for correlation.

```typescript
test("GET /api/health echoes X-Request-ID", async ({ request }) => {
  const reqId = "req-12345";
  const response = await request.get("/api/health", {
    headers: { "X-Request-ID": reqId },
  });
  expect(response.headers()["x-request-id"]).toBe(reqId);
});
```

### `X-Correlation-ID` (legacy)

- **Direction:** both.
- **Required when:** cross-service correlation in older systems.
- **Set by:** edge service.
- **Mandatory:** recommended in the absence of W3C tracing.

```typescript
test("GET /api/orders propagates X-Correlation-ID", async ({ request }) => {
  const correlationId = "corr-abcdef";
  const response = await request.get("/api/orders", {
    headers: { "X-Correlation-ID": correlationId },
  });
  expect(response.headers()["x-correlation-id"]).toBe(correlationId);
});
```

---

## 5. Pagination

### `Link` header (RFC 5988)

Standard way to expose pagination links. Format: `<URL>; rel="next", <URL>; rel="prev"`.

- **Direction:** response.
- **Required when:** list endpoint has more results.
- **Set by:** server.
- **Mandatory:** recommended (REST mature level).

```typescript
function parseLinkHeader(header: string): Record<string, string> {
  const links: Record<string, string> = {};
  for (const part of header.split(",")) {
    const match = part.trim().match(/^<([^>]+)>;\s*rel="([^"]+)"$/);
    if (match) links[match[2]] = match[1];
  }
  return links;
}

test("GET /api/users?page=1 returns Link with rel=next", async ({ request }) => {
  const response = await request.get("/api/users?page=1&limit=10");
  expect(response.status()).toBe(200);
  const linkHeader = response.headers()["link"];
  expect(linkHeader).toBeDefined();
  const links = parseLinkHeader(linkHeader!);
  expect(links.next).toMatch(/page=2/);

  // Follow next, validate it works
  const next = await request.get(links.next);
  expect(next.status()).toBe(200);
});
```

### `X-Total-Count`

- **Direction:** response.
- **Required when:** UI needs total record count.
- **Set by:** server.
- **Mandatory:** recommended.

```typescript
test("GET /api/users returns X-Total-Count consistent with pagination", async ({ request }) => {
  const page1 = await request.get("/api/users?page=1&limit=10");
  const total = parseInt(page1.headers()["x-total-count"]!, 10);
  expect(total).toBeGreaterThan(0);

  const page1Body = await page1.json();
  expect(page1Body.length).toBeLessThanOrEqual(10);

  // Sum of all pages should equal total
  const lastPage = Math.ceil(total / 10);
  const last = await request.get(`/api/users?page=${lastPage}&limit=10`);
  const lastBody = await last.json();
  expect((lastPage - 1) * 10 + lastBody.length).toBe(total);
});
```

### Cursor-based via response body

Some APIs (GitHub GraphQL, AWS) expose cursors in body, not headers.

```typescript
test("GET /api/feed returns cursor in body for next page", async ({ request }) => {
  const first = await request.get("/api/feed");
  const firstBody = await first.json();
  expect(firstBody).toHaveProperty("nextCursor");

  const second = await request.get(`/api/feed?cursor=${firstBody.nextCursor}`);
  expect(second.status()).toBe(200);
});
```

---

## 6. Rate Limit

Rate-limit headers tell clients how much budget they have left and when it resets. The de-facto convention is `X-RateLimit-*`. The IETF draft `RateLimit` (without prefix) is an emerging standard.

- **Direction:** response.
- **Required when:** endpoint is rate-limited.
- **Set by:** server.
- **Mandatory:** recommended on every response from a rate-limited endpoint.

```typescript
test("GET /api/search exposes rate-limit headers", async ({ request }) => {
  const response = await request.get("/api/search?q=test");
  expect(response.headers()["x-ratelimit-limit"]).toMatch(/^\d+$/);
  expect(response.headers()["x-ratelimit-remaining"]).toMatch(/^\d+$/);
  expect(response.headers()["x-ratelimit-reset"]).toMatch(/^\d+$/);

  const remaining = parseInt(response.headers()["x-ratelimit-remaining"]!, 10);
  const limit = parseInt(response.headers()["x-ratelimit-limit"]!, 10);
  expect(remaining).toBeLessThanOrEqual(limit);
});

test("Exceeding rate limit returns 429 with Retry-After", async ({ request }) => {
  const limit = 100;
  let lastResponse;
  for (let i = 0; i < limit + 5; i++) {
    lastResponse = await request.get("/api/search?q=test");
    if (lastResponse.status() === 429) break;
  }
  expect(lastResponse!.status()).toBe(429);
  const retryAfter = lastResponse!.headers()["retry-after"];
  expect(retryAfter).toBeDefined();
  // Retry-After is either seconds (integer) or HTTP-date
  expect(retryAfter).toMatch(/^(\d+|.+GMT)$/);
});
```

`Retry-After` also appears on 503 Service Unavailable:

```typescript
test("503 includes Retry-After", async ({ request }) => {
  const response = await request.get("/api/maintenance-test");
  if (response.status() === 503) {
    expect(response.headers()["retry-after"]).toBeDefined();
  }
});
```

---

## 7. Security Headers (Response)

These are FUNCTIONAL tests — assert the header is present with the correct value. The deeper SECURITY analysis (threat model, CSP policy review, HSTS preload eligibility) lives in the `qa-owasp-security` skill. Keep this catalog focused on "header present + correct value".

- **Direction:** response.
- **Set by:** server (or edge proxy / CDN).

### `Strict-Transport-Security` (HSTS)

- **Required when:** HTTPS endpoint.
- **Mandatory:** mandatory for production HTTPS.

```typescript
test("HTTPS response includes HSTS with max-age >= 1 year", async ({ request }) => {
  const response = await request.get("https://api.example.com/health");
  const hsts = response.headers()["strict-transport-security"];
  expect(hsts).toMatch(/max-age=(\d+)/);
  const maxAge = parseInt(hsts!.match(/max-age=(\d+)/)![1], 10);
  expect(maxAge).toBeGreaterThanOrEqual(31536000); // 1 year
});
```

### `X-Frame-Options`

- **Required when:** HTML response (clickjacking protection).
- **Mandatory:** required for HTML; redundant if CSP `frame-ancestors` is set.

```typescript
test("HTML response sets X-Frame-Options", async ({ request }) => {
  const response = await request.get("/dashboard");
  const xfo = response.headers()["x-frame-options"];
  expect(["DENY", "SAMEORIGIN"]).toContain(xfo);
});
```

### `X-Content-Type-Options: nosniff`

- **Required when:** always (cheap, broadly applicable).
- **Mandatory:** recommended on all responses.

```typescript
test("All responses set X-Content-Type-Options: nosniff", async ({ request }) => {
  const response = await request.get("/api/health");
  expect(response.headers()["x-content-type-options"]).toBe("nosniff");
});
```

### `Content-Security-Policy`

- **Required when:** HTML response.
- **Mandatory:** required for HTML pages serving user content.

```typescript
test("HTML response sets a restrictive CSP", async ({ request }) => {
  const response = await request.get("/dashboard");
  const csp = response.headers()["content-security-policy"];
  expect(csp).toBeDefined();
  expect(csp).toMatch(/default-src/);
  expect(csp).not.toMatch(/unsafe-inline/); // Functional check; deeper audit lives in OWASP skill
});
```

### `Referrer-Policy`

- **Required when:** always (recommended).
- **Mandatory:** recommended.

```typescript
test("Response sets Referrer-Policy", async ({ request }) => {
  const response = await request.get("/api/health");
  const policy = response.headers()["referrer-policy"];
  expect([
    "no-referrer",
    "no-referrer-when-downgrade",
    "same-origin",
    "strict-origin",
    "strict-origin-when-cross-origin",
  ]).toContain(policy);
});
```

### `Permissions-Policy`

- **Required when:** browser-facing endpoint.
- **Mandatory:** recommended.

```typescript
test("HTML response restricts sensitive features via Permissions-Policy", async ({ request }) => {
  const response = await request.get("/dashboard");
  const policy = response.headers()["permissions-policy"];
  expect(policy).toBeDefined();
  expect(policy).toMatch(/camera=\(\)/); // Camera disabled
  expect(policy).toMatch(/microphone=\(\)/); // Microphone disabled
});
```

---

## 8. Caching

HTTP caching headers prevent unnecessary work for both client and server. The two main mechanisms are validators (`ETag`, `Last-Modified`) and freshness (`Cache-Control`, `Expires`).

### `ETag` + `If-None-Match` → 304

- **Direction:** response (`ETag`), request (`If-None-Match`).
- **Required when:** resource is cacheable.
- **Set by:** server (`ETag`), client (`If-None-Match`).
- **Mandatory:** recommended on cacheable GETs.

```typescript
test("GET /api/users/1 round-trip ETag returns 304", async ({ request }) => {
  const first = await request.get("/api/users/1");
  expect(first.status()).toBe(200);
  const etag = first.headers()["etag"];
  expect(etag).toBeDefined();

  // Second request with If-None-Match → 304 Not Modified
  const second = await request.get("/api/users/1", {
    headers: { "If-None-Match": etag! },
  });
  expect(second.status()).toBe(304);
});
```

### `Cache-Control`

- **Direction:** both — request hints (`no-cache`, `max-stale`); response directives (`max-age`, `no-store`, `private`).
- **Required when:** always (recommended).
- **Set by:** server primarily.
- **Mandatory:** recommended on every response.

```typescript
test("GET /api/static-asset has Cache-Control with max-age", async ({ request }) => {
  const response = await request.get("/api/static-asset");
  const cacheControl = response.headers()["cache-control"];
  expect(cacheControl).toMatch(/max-age=\d+/);
});

test("GET /api/me sets Cache-Control: no-store for sensitive data", async ({ request }) => {
  const response = await request.get("/api/me", {
    headers: { Authorization: `Bearer ${process.env.JWT_TOKEN}` },
  });
  expect(response.headers()["cache-control"]).toMatch(/no-store/);
});
```

### `Last-Modified` + `If-Modified-Since`

- **Direction:** response (`Last-Modified`), request (`If-Modified-Since`).
- **Required when:** resource has a meaningful modification timestamp.
- **Set by:** server (`Last-Modified`), client (`If-Modified-Since`).
- **Mandatory:** recommended where ETag is not available.

```typescript
test("GET /api/feed round-trip Last-Modified returns 304", async ({ request }) => {
  const first = await request.get("/api/feed");
  const lastModified = first.headers()["last-modified"];
  expect(lastModified).toBeDefined();

  const second = await request.get("/api/feed", {
    headers: { "If-Modified-Since": lastModified! },
  });
  expect(second.status()).toBe(304);
});
```

---

# Part 2 — Contract Testing Principles

## 1. Why Contract Tests

API contracts drift. A backend team changes a field name; a frontend that consumed the old shape breaks in production weeks later. Contract tests catch drift at CI time, BEFORE deploy.

OpenAPI/AsyncAPI lint catches STATIC drift (spec vs spec). Contract tests catch RUNTIME drift (real producer response vs consumer expectation). They are complementary — lint your spec AND verify the runtime.

Two failure modes contract tests prevent:

- Producer changes shape → all consumers break silently in production.
- Consumer assumes a field that producer never promised → breaks the moment producer reshapes the response.

## 2. Two Paths

### Consumer-Driven (Pact)

Consumer tests describe expected interactions ("when I send X, I expect response Y"). The test run produces a **pact file** (JSON). The provider then runs a **verification** step that replays the pact against its real implementation.

- **Best for:** external consumers, many consumers, polyglot stacks.
- **Strength:** consumer's actual usage drives the contract — no field gets dropped because "no one uses it" when in fact someone does.
- **Weakness:** state setup (provider must be seeded before each interaction); pact files can drift if not gated by a broker.
- **Bridges to skill:** `qa-contract-pact` — that skill covers Pact in depth (PACT-JS, PACT-JVM, broker, can-i-deploy, matchers, message pact, troubleshooting).

### OpenAPI as Contract

Provider publishes the OpenAPI spec. Consumer generates types/clients from it. Both sides validate runtime traffic against the spec.

- **Best for:** one provider with internal consumers, monorepo, single-language stacks.
- **Strength:** single source of truth, generated types eliminate manual mapping.
- **Weakness:** spec must be kept current; consumers may use endpoints/fields the spec does not describe.
- **Already covered in:** `references/openapi-driven-testing.md` (A3 of this skill).

### Decision Tree — When to Choose Each

```
Is the API consumed by external/third-party clients?
├── Yes → Pact (consumer-driven). External consumers cannot wait on the provider.
└── No → Are consumers in the same monorepo / same release train?
        ├── Yes → OpenAPI as contract. Lower setup cost.
        └── No (polyglot org, separate release cadence)?
                ├── Yes → BOTH. OpenAPI for type safety, Pact for runtime drift.
                └── Hybrid (some external, some internal) → BOTH.
```

Heuristic:

- 1 provider, 1-3 internal consumers → **OpenAPI**.
- 1 provider, N external consumers → **Pact**.
- N providers, M consumers, separate teams → **Pact + OpenAPI**.

## 3. Pact in 3 Steps

Using `@pact-foundation/pact` v12+. Install with `npm i -D @pact-foundation/pact`.

### Step 1 — Consumer expectation

```typescript
// consumer/user-service.pact.test.ts
import path from "node:path";
import { PactV3, MatchersV3 } from "@pact-foundation/pact";
import { fetchUser } from "./user-client";

const { like, integer, string } = MatchersV3;

const provider = new PactV3({
  consumer: "WebApp",
  provider: "UserService",
  dir: path.resolve(process.cwd(), "pacts"),
});

describe("User service consumer", () => {
  it("fetches a user by id", async () => {
    provider
      .given("user 1 exists")
      .uponReceiving("a request for user 1")
      .withRequest({
        method: "GET",
        path: "/api/users/1",
        headers: { Accept: "application/json" },
      })
      .willRespondWith({
        status: 200,
        headers: { "Content-Type": "application/json" },
        body: like({
          id: integer(1),
          name: string("Alice"),
          email: string("alice@example.com"),
        }),
      });

    await provider.executeTest(async (mockServer) => {
      const user = await fetchUser(mockServer.url, 1);
      expect(user.id).toBe(1);
      expect(user.name).toBe("Alice");
    });
  });
});
```

### Step 2 — Run consumer test (publishes pact JSON)

```bash
npm test -- consumer/user-service.pact.test.ts
# Generates pacts/WebApp-UserService.json
```

In CI, publish to the broker:

```bash
npx pact-broker publish ./pacts \
  --consumer-app-version=$GIT_SHA \
  --branch=$GIT_BRANCH \
  --broker-base-url=$PACT_BROKER_URL \
  --broker-token=$PACT_BROKER_TOKEN
```

### Step 3 — Provider verifies pact

```typescript
// provider/user-service.verify.test.ts
import { Verifier } from "@pact-foundation/pact";
import { startServer } from "../src/server";

describe("User service provider verification", () => {
  let server: { close: () => Promise<void>; url: string };

  beforeAll(async () => {
    server = await startServer({ port: 0 });
  });

  afterAll(async () => {
    await server.close();
  });

  it("verifies all consumer pacts", async () => {
    await new Verifier({
      provider: "UserService",
      providerBaseUrl: server.url,
      pactBrokerUrl: process.env.PACT_BROKER_URL!,
      pactBrokerToken: process.env.PACT_BROKER_TOKEN!,
      publishVerificationResult: true,
      providerVersion: process.env.GIT_SHA!,
      providerVersionBranch: process.env.GIT_BRANCH!,
      consumerVersionSelectors: [
        { mainBranch: true },
        { deployedOrReleased: true },
      ],
      stateHandlers: {
        "user 1 exists": async () => {
          // Seed DB so GET /api/users/1 returns expected data
          await seedUser({ id: 1, name: "Alice", email: "alice@example.com" });
        },
      },
    }).verifyProvider();
  });
});
```

## 4. Pact Broker

The broker is the central registry between consumer and provider. It does three things:

1. **Stores pacts** — published from consumer CI.
2. **Publishes verification results** — provider verifications post pass/fail back.
3. **Gates deploys via `can-i-deploy`** — the deploy script asks the broker "given consumer X at version A and provider Y at version B, are they compatible?" and only proceeds on green.

```bash
# Before deploying consumer
npx pact-broker can-i-deploy \
  --pacticipant WebApp \
  --version $GIT_SHA \
  --to-environment production \
  --broker-base-url $PACT_BROKER_URL \
  --broker-token $PACT_BROKER_TOKEN
```

Tag versions by environment (`staging`, `production`, `main`) so the broker knows what is deployed where.

## 5. Pitfalls

- **Flaky pacts from state setup.** If `stateHandlers` are not deterministic (DB not reset, race conditions), provider verification fails intermittently. Fix: each state handler resets and seeds idempotently.
- **Over-specifying interactions.** Asserting on every field, including ones the consumer never reads, makes the contract brittle. Fix: only assert on fields the consumer ACTUALLY uses; use loose matchers (`like`, `integer`, `string`) instead of exact values.
- **Forgetting `can-i-deploy`.** Without the gate, you publish pacts and verify, but still deploy a consumer that depends on an un-verified provider version. Fix: make `can-i-deploy` a required CI step before any deploy.
- **Missing matchers cause exact-match brittleness.** `body: { id: 1 }` asserts the exact value `1`. `body: like({ id: integer(1) })` asserts "any integer, sample is 1". Always wrap dynamic fields in matchers.
- **No consumer version selectors.** Verifying every pact ever published is slow. Fix: use `consumerVersionSelectors` (`{ mainBranch: true }`, `{ deployedOrReleased: true }`) to verify only what matters.
- **Pact for everything.** Pact is for SHAPE contracts, not business logic. Do not test "creating a user sends a welcome email" with Pact — that is integration territory.

---

## See Also

- `references/openapi-driven-testing.md` — OpenAPI as contract, runtime validation against spec.
- `references/contract-testing.md` — broader contract testing patterns (request/response shape, status codes).
- `references/playwright-api-testing.md` — Playwright `request` fixture for HTTP testing.
- `qa-owasp-security` skill — deeper security analysis of the response headers in section 7.
- `qa-contract-pact` skill — Pact-specific patterns: matchers, message pact, broker workflows, can-i-deploy, troubleshooting.
