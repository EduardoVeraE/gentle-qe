<!-- Skill: api-testing · Template: mandatory-headers-checklist -->
<!-- Placeholders: {{api_name}}, {{review_date}}, {{owner}}, {{endpoint_group}}, {{auth_test_ref}}, {{idempotency_test_ref}}, {{tracing_test_ref}}, {{pagination_test_ref}}, {{rate_limit_test_ref}}, {{security_test_ref}}, {{caching_test_ref}}, {{content_test_ref}} -->

# Mandatory headers checklist — {{api_name}}

Reviewed: {{review_date}} · Owner: {{owner}}

One block per header CATEGORY. Within each block, one row per (endpoint group, header). A row is green when the header is asserted in an automated test (the `Test reference` column points to it).

Direction: `req` (client → server) · `res` (server → client) · `both`.
Set by: who is responsible for emitting the header (client SDK, gateway, service, framework middleware).

Sample rows use `{{endpoint_group}}` <!-- e.g., /orders/* --> so you can copy the block and fill it for every group.

## 1. Auth

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `Authorization: Bearer <jwt>` | req | Yes | Client SDK | {{auth_test_ref}} <!-- e.g., tests/api/orders/auth.spec.ts::missing_token_returns_401 --> |
| {{endpoint_group}} | `WWW-Authenticate` | res (on 401) | Yes | Service | {{auth_test_ref}} |

Negative cases that MUST exist: missing token → 401, expired → 401, wrong audience → 401, wrong scope → 403.

## 2. Content

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `Content-Type: application/json` | req (POST/PUT/PATCH) | Yes | Client SDK | {{content_test_ref}} |
| {{endpoint_group}} | `Content-Type: application/json; charset=utf-8` | res | Yes | Framework | {{content_test_ref}} |
| {{endpoint_group}} | `Accept: application/json` | req | Recommended | Client SDK | {{content_test_ref}} |
| {{endpoint_group}} | `Content-Length` | both | Yes | Framework | {{content_test_ref}} |

## 3. Idempotency

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `Idempotency-Key: <uuid-v4>` | req | Yes (POST/PUT/PATCH) | Client SDK | {{idempotency_test_ref}} <!-- e.g., tests/api/orders/idempotency.spec.ts::same_key_returns_same_response --> |

Required behavior: same key + same body within TTL → same response. Same key + different body → 409 / 422.

## 4. Tracing

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `traceparent` (W3C) | both | Yes | OTel middleware | {{tracing_test_ref}} |
| {{endpoint_group}} | `tracestate` | both | Optional | OTel middleware | {{tracing_test_ref}} |
| {{endpoint_group}} | `X-Request-Id` | both | Yes | Gateway | {{tracing_test_ref}} |

The service MUST echo `X-Request-Id` if present, or generate one and return it.

## 5. Pagination

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} (collections) | `Link: <...>; rel="next"` | res | Yes | Service | {{pagination_test_ref}} |
| {{endpoint_group}} (collections) | `Link: <...>; rel="prev"` | res | Conditional | Service | {{pagination_test_ref}} |
| {{endpoint_group}} (collections) | `X-Total-Count` | res | Optional | Service | {{pagination_test_ref}} |

If you use cursor pagination, document the cursor opacity and TTL alongside the test.

## 6. Rate limit

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `X-RateLimit-Limit` | res | Yes | Gateway | {{rate_limit_test_ref}} |
| {{endpoint_group}} | `X-RateLimit-Remaining` | res | Yes | Gateway | {{rate_limit_test_ref}} |
| {{endpoint_group}} | `X-RateLimit-Reset` | res | Yes | Gateway | {{rate_limit_test_ref}} |
| {{endpoint_group}} | `Retry-After` | res (on 429) | Yes | Gateway | {{rate_limit_test_ref}} |

Test must hit the limit deterministically (dedicated test tenant) and assert all four headers on the 429.

## 7. Security

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} | `Strict-Transport-Security: max-age=31536000; includeSubDomains` | res | Yes | Gateway | {{security_test_ref}} |
| {{endpoint_group}} | `X-Content-Type-Options: nosniff` | res | Yes | Gateway | {{security_test_ref}} |
| {{endpoint_group}} | `Referrer-Policy: no-referrer` | res | Yes | Gateway | {{security_test_ref}} |
| {{endpoint_group}} | `Cache-Control: no-store` (sensitive endpoints) | res | Yes | Service | {{security_test_ref}} |
| {{endpoint_group}} | `Access-Control-Allow-Origin` | res (CORS) | Conditional | Gateway | {{security_test_ref}} |

CORS preflight (`OPTIONS`) MUST be tested for every browser-facing group.

## 8. Caching

| Endpoint group | Header | Direction | Required | Set by | Test reference |
| -------------- | ------ | --------- | -------- | ------ | -------------- |
| {{endpoint_group}} (GET) | `ETag` | res | Recommended | Service | {{caching_test_ref}} |
| {{endpoint_group}} (GET) | `If-None-Match` | req | Recommended | Client | {{caching_test_ref}} |
| {{endpoint_group}} (GET) | `Cache-Control` | res | Yes | Service | {{caching_test_ref}} |
| {{endpoint_group}} (GET) | `Last-Modified` | res | Optional | Service | {{caching_test_ref}} |
| {{endpoint_group}} (GET) | `Vary` | res | Conditional | Service | {{caching_test_ref}} |

`304 Not Modified` MUST be tested when `ETag` / `Last-Modified` are advertised.

## How to use

1. Copy the entire file once per API.
2. For every endpoint group, duplicate the relevant rows and replace `{{endpoint_group}}`.
3. Link the `Test reference` to a real test ID — never to a doc.
4. Re-review on every spec change and at every release.
