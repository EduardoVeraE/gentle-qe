# OpenAPI-Driven Testing

> Treating the OpenAPI document as the single source of truth for the API surface, and deriving lints, types, validators, mocks, breaking-change checks, and test data directly from it.

---

## 1. Why OpenAPI-First Testing Wins

When the OpenAPI document is authoritative, every downstream artifact becomes a derivative — and derivatives stay in sync automatically.

- **Generated types**: The contract is compiled into TypeScript types. A spec change that breaks a consumer breaks the build, not production.
- **Generated validators**: Request and response validators are produced from the same schemas the spec exposes. There is no second place where the rules live.
- **Automated drift detection**: `oasdiff` compares two versions of the spec and tells you which changes are breaking, before the PR merges.
- **Mock servers and fixtures for free**: Tools like Prism or `@stoplight/prism-cli` serve a working mock from the spec, useful for consumer-side contract tests.
- **Docs that cannot lie**: docs rendered from the spec are correct by construction when tests pass against the spec.

**The alternative** — handwritten types, validators, and docs maintained in parallel — is a coordination problem that grows quadratically with the number of endpoints. Every team that has tried it ends up with three "sources of truth" that disagree, and bugs that sit in the gaps between them.

**The mental model**: the spec is the contract. Producers prove they implement it. Consumers prove they consume it. The spec itself must be linted and version-checked — if it drifts from reality, every downstream guarantee evaporates.

---

## 2. Linting the Spec with Spectral

Before anything else, the spec must be valid OpenAPI **and** conform to your house style.

### Install

```bash
npm install --save-dev @stoplight/spectral-cli@^6.11.0
# or globally
npm install -g @stoplight/spectral-cli@^6.11.0
```

Pin to 6.x. The 6.x line is stable and ships the `spectral:oas` ruleset that understands OpenAPI 3.0 and 3.1.

### Default rulesets

Spectral ships two built-in rulesets:

- `spectral:oas` — structural OpenAPI rules (valid `$ref`, required fields, type correctness, etc.)
- `spectral:asyncapi` — for AsyncAPI documents, ignore here

### Running it

```bash
npx spectral lint openapi.yaml
npx spectral lint openapi.yaml --ruleset .spectral.yaml
npx spectral lint openapi.yaml --format junit --output spectral-report.xml
```

### Custom rulesets

Most teams need to encode conventions the default ruleset does not enforce — for example, `operationId` naming, mandatory tags, response schema requirements. Define a `.spectral.yaml` at repo root:

```yaml
extends: [[spectral:oas, recommended]]

rules:
  # Every operation must have an operationId
  operation-operationId:
    description: Operations must define an operationId
    given: $.paths[*][get,post,put,patch,delete]
    severity: error
    then:
      field: operationId
      function: truthy

  # operationId must be camelCase and start with a verb
  operation-operationId-naming:
    description: operationId must be camelCase verb-first
    given: $.paths[*][get,post,put,patch,delete].operationId
    severity: error
    then:
      function: pattern
      functionOptions:
        match: '^(get|list|create|update|delete|replace|search)[A-Z][a-zA-Z0-9]*$'

  # Every operation must declare at least one tag
  operation-tags-required:
    description: Operations must have at least one tag
    given: $.paths[*][get,post,put,patch,delete]
    severity: error
    then:
      field: tags
      function: schema
      functionOptions:
        schema:
          type: array
          minItems: 1

  # All non-204 responses must reference a schema
  response-schema-required:
    description: 2xx/4xx/5xx responses must reference a schema
    given: $.paths[*][*].responses[?(@property != '204')].content[*]
    severity: error
    then:
      field: schema
      function: truthy

  # Error responses must use the shared Error schema
  error-uses-error-schema:
    description: 4xx/5xx must reference components/schemas/Error
    given: $.paths[*][*].responses[?(@property =~ /^[45]/)].content.application/json.schema
    severity: warn
    then:
      function: schema
      functionOptions:
        schema:
          properties:
            $ref:
              const: '#/components/schemas/Error'
```

### CI integration

A failing lint must block merge. GitHub Actions step:

```yaml
- name: Lint OpenAPI spec
  run: npx spectral lint openapi.yaml --fail-severity=warn
```

The `--fail-severity=warn` flag treats warnings as failures. Use `error` if your team is still ramping up.

---

## 3. Generating Types from the Spec (TypeScript)

The simplest, highest-leverage transformation: turn the spec into TypeScript types.

### `openapi-typescript`

```bash
npm install --save-dev openapi-typescript@^7.4.0
```

Generate:

```bash
npx openapi-typescript openapi.yaml -o src/generated/api.ts
```

This produces a file with two top-level types:

- `paths` — a map of every URL + method to its request and response shapes
- `components` — every named schema, parameter, response, and request body from the spec

### Using generated types in tests

```typescript
import type { paths, components } from "../src/generated/api";

// Pull request and response types out of paths
type CreateUserRequest =
  paths["/users"]["post"]["requestBody"]["content"]["application/json"];
type CreateUserResponse =
  paths["/users"]["post"]["responses"]["201"]["content"]["application/json"];

// Or pull a named component schema directly
type User = components["schemas"]["User"];

const newUser: CreateUserRequest = {
  email: "ada@example.com",
  name: "Ada Lovelace",
};
```

If the spec changes — say, `email` becomes required and a new `role` field is added — `tsc` will fail at every call site that does not match. The tests stop compiling before they run, which is exactly what you want.

### Regenerating in CI

Treat `src/generated/` as committed but verify it is in sync:

```yaml
- name: Regenerate API types
  run: npx openapi-typescript openapi.yaml -o src/generated/api.ts

- name: Verify no drift
  run: git diff --exit-code src/generated/api.ts
```

If a spec change ships without regenerated types, this step fails the PR.

### Java equivalent (brief)

For Java/Kotlin services, `openapi-generator-cli` produces typed clients and server stubs, and `swagger-codegen` is the older alternative. They follow the same philosophy: the spec compiles into typed code.

---

## 4. Request Validation Against the Spec

Generated types catch shape errors at compile time. Runtime validation catches them when test data, fixtures, or external inputs slip through.

### Approach: Ajv + the spec's schemas

```bash
npm install --save-dev ajv@^8.12.0 ajv-formats@^3.0.1 js-yaml@^4.1.0
```

Compile validators once, reuse forever:

```typescript
import Ajv from "ajv";
import addFormats from "ajv-formats";
import yaml from "js-yaml";
import fs from "fs";

const spec = yaml.load(fs.readFileSync("openapi.yaml", "utf8")) as any;

const ajv = new Ajv({ allErrors: true, strict: false });
addFormats(ajv);

// Register every component schema by its $ref
for (const [name, schema] of Object.entries<any>(spec.components.schemas)) {
  ajv.addSchema(schema, `#/components/schemas/${name}`);
}

export function validateRequest(schemaRef: string, payload: unknown) {
  const validate = ajv.getSchema(schemaRef);
  if (!validate) throw new Error(`Schema not found: ${schemaRef}`);
  const valid = validate(payload);
  if (!valid) {
    throw new Error(
      `Request validation failed: ${ajv.errorsText(validate.errors)}`
    );
  }
  return payload;
}
```

### Asserting in tests

```typescript
import { validateRequest } from "./openapi-validator";

test("POST /users request body matches the spec", async () => {
  const body = {
    email: "ada@example.com",
    name: "Ada Lovelace",
  };

  // This throws if the body diverges from the spec
  validateRequest("#/components/schemas/CreateUserRequest", body);

  const res = await fetch("http://localhost:3000/users", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });

  expect(res.status).toBe(201);
});
```

### Why pre-send validation matters

It catches **test bugs**. A common failure mode: the test sends an incorrect payload, the server rejects it with 400, the test asserts on a 400, and a real consumer-side bug ships. Pre-send validation forces the test to send a spec-conformant payload, which means a 400 from the server is now genuinely a server bug.

### Java equivalent (brief)

Atlassian's `swagger-request-validator` (`com.atlassian.oai:swagger-request-validator-core`) does this for Java/Kotlin and integrates with REST Assured, MockMvc, and WireMock.

---

## 5. Response Validation Against the Spec

Just as important as validating what you send is validating what you receive.

### Approach A: Ajv (same as request side)

Reuse the validator from section 4, just point it at response schemas:

```typescript
const validate = ajv.getSchema("#/components/schemas/User");
if (!validate(await res.json())) {
  throw new Error(ajv.errorsText(validate.errors));
}
```

### Approach B: Zod from OpenAPI

If your codebase already uses Zod, derive Zod schemas from the spec.

```bash
npm install --save-dev @anatine/zod-openapi@^2.2.7 zod@^3.23.8
```

Or use `openapi-zod-client` to generate Zod schemas + a typed client at build time. Either way, you get runtime validation in tests:

```typescript
import { z } from "zod";

const UserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  name: z.string().min(1),
  role: z.enum(["admin", "user"]),
  createdAt: z.string().datetime(),
});

test("GET /users/:id returns a User", async () => {
  const res = await fetch("http://localhost:3000/users/abc-123");
  const json = await res.json();

  // Throws with detailed path on mismatch
  const user = UserSchema.parse(json);

  expect(user.email).toBe("ada@example.com");
});
```

A small custom matcher gives nicer messages:

```typescript
import { expect } from "vitest";
import type { ZodSchema } from "zod";

expect.extend({
  toMatchSchema(received: unknown, schema: ZodSchema) {
    const result = schema.safeParse(received);
    return result.success
      ? { pass: true, message: () => "schema matched" }
      : {
          pass: false,
          message: () =>
            `schema mismatch:\n${result.error.issues
              .map((i) => `  ${i.path.join(".")}: ${i.message}`)
              .join("\n")}`,
        };
  },
});

declare module "vitest" {
  interface Assertion<T> {
    toMatchSchema(schema: ZodSchema): T;
  }
}

// Usage
expect(await res.json()).toMatchSchema(UserSchema);
```

### Why response validation must run in BOTH layers

- **Provider tests** (server-side) prove the implementation produces what the spec promises.
- **Consumer tests** (client-side, often against a mock) prove the consumer can parse what the spec promises.

If you only validate on one side, drift can hide. A server adding a non-nullable field is invisible to consumer tests if those tests use stale mocks. A consumer expecting an extra field is invisible to provider tests. Both must validate against the same spec to catch both directions.

This is also where Pact-style consumer-driven contracts complement OpenAPI: see `contract-testing.md` for the broader picture.

---

## 6. Breaking-Change Detection (oasdiff)

`oasdiff` is the standard tool for diffing two OpenAPI documents and classifying the differences.

### Install

```bash
# macOS
brew tap oasdiff/homebrew-oasdiff
brew install oasdiff

# Or via Go
go install github.com/oasdiff/oasdiff@latest
```

Pin to 1.10+ for reliable breaking-change detection on OpenAPI 3.1.

### Basic usage

```bash
oasdiff diff openapi.old.yaml openapi.new.yaml
oasdiff breaking openapi.old.yaml openapi.new.yaml
oasdiff changelog openapi.old.yaml openapi.new.yaml
```

### Severity levels

`oasdiff breaking` classifies each change as:

- **error** — a hard breaking change (removed endpoint, required parameter added, response shape narrowed)
- **warn** — a likely breaking change depending on consumer behavior (default value changed, format tightened)
- **info** — non-breaking but notable (description changed)

### CI gating

```yaml
- name: Check for breaking changes
  run: |
    oasdiff breaking \
      --fail-on ERR \
      origin/main:openapi.yaml \
      openapi.yaml
```

The `origin/main:openapi.yaml` syntax pulls the spec from the base branch directly, no checkout dance required.

### Allowlisting expected breaks

Sometimes you genuinely need a breaking change — a deprecated endpoint is finally being removed. Use an allowlist:

```bash
oasdiff breaking \
  --fail-on ERR \
  --warn-ignore .breaking-allowlist.yaml \
  openapi.old.yaml openapi.new.yaml
```

`.breaking-allowlist.yaml`:

```yaml
- id: api-removed-without-deprecation
  text: "DELETE /v1/legacy-users removed in 2.0.0 release"
  ignored-until: "2026-12-31"
```

The `ignored-until` date forces you to revisit the allowlist. Stale allowlists are themselves a smell.

---

## 7. Generating Examples and Edge Cases from the Spec

The spec encodes more than shape — `format`, `pattern`, `enum`, `minimum`, `maximum`, `minLength`, `maxLength` all describe boundaries. A good test harness mines these.

### Faker + spec metadata

```bash
npm install --save-dev @faker-js/faker@^9.0.0
```

Build values that respect the constraints:

```typescript
import { faker } from "@faker-js/faker";

interface FieldSpec {
  type: string;
  format?: string;
  pattern?: string;
  enum?: unknown[];
  minimum?: number;
  maximum?: number;
  minLength?: number;
  maxLength?: number;
}

function generateValue(spec: FieldSpec): unknown {
  if (spec.enum) return faker.helpers.arrayElement(spec.enum);
  if (spec.type === "string") {
    if (spec.format === "email") return faker.internet.email();
    if (spec.format === "uuid") return faker.string.uuid();
    if (spec.format === "date-time") return faker.date.recent().toISOString();
    if (spec.pattern) return faker.helpers.fromRegExp(spec.pattern);
    return faker.string.alpha({
      length: { min: spec.minLength ?? 1, max: spec.maxLength ?? 20 },
    });
  }
  if (spec.type === "integer") {
    return faker.number.int({ min: spec.minimum ?? 0, max: spec.maximum ?? 100 });
  }
  if (spec.type === "number") {
    return faker.number.float({ min: spec.minimum ?? 0, max: spec.maximum ?? 100 });
  }
  return null;
}
```

### Boundary value analysis

The spec gives you the boundaries; tests must hit them. For numeric fields, generate `min`, `min-1`, `min+1`, `max`, `max-1`, `max+1`. For strings, generate `"a".repeat(minLength)`, `"a".repeat(minLength-1)`, `"a".repeat(maxLength)`, `"a".repeat(maxLength+1)`. The values one below the minimum and one above the maximum are the off-by-one tests — the ones the implementation gets wrong most often.

### Property-based testing with fast-check

```bash
npm install --save-dev fast-check@^3.22.0
```

Drive arbitraries from the spec rather than hand-rolling them:

```typescript
import fc from "fast-check";

function arbitraryFromSpec(spec: FieldSpec): fc.Arbitrary<unknown> {
  if (spec.enum) return fc.constantFrom(...spec.enum);
  if (spec.type === "string" && spec.format === "email") return fc.emailAddress();
  if (spec.type === "integer") return fc.integer({ min: spec.minimum, max: spec.maximum });
  if (spec.type === "string") {
    return fc.string({ minLength: spec.minLength, maxLength: spec.maxLength });
  }
  return fc.anything();
}

test("POST /users accepts any spec-valid email", () => {
  fc.assert(
    fc.asyncProperty(fc.emailAddress(), async (email) => {
      const res = await fetch("http://localhost:3000/users", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ email, name: "Test" }),
      });
      return res.status === 201 || res.status === 409;
    }),
    { numRuns: 50 }
  );
});
```

Property tests find inputs you would never write by hand. Combined with the spec, they explore the entire region of "valid" without the human having to enumerate it.

---

## 8. Contract Diff in CI

The full pipeline: lint, regenerate, validate, diff. A GitHub Actions example:

```yaml
name: API Contract

on:
  pull_request:
    paths:
      - "openapi.yaml"
      - "src/**"

jobs:
  contract:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Lint OpenAPI spec
        run: npx spectral lint openapi.yaml --fail-severity=warn

      - name: Regenerate types
        run: npx openapi-typescript openapi.yaml -o src/generated/api.ts

      - name: Verify generated types are committed
        run: git diff --exit-code src/generated/api.ts

      - name: Install oasdiff
        run: |
          curl -sSL https://github.com/oasdiff/oasdiff/releases/download/v1.10.25/oasdiff_1.10.25_linux_amd64.tar.gz \
            | tar xz -C /usr/local/bin oasdiff

      - name: Get base spec
        run: git show origin/${{ github.base_ref }}:openapi.yaml > openapi.base.yaml

      - name: Check for breaking changes
        id: breaking
        run: |
          oasdiff breaking openapi.base.yaml openapi.yaml \
            --fail-on ERR \
            --format markdown > breaking-report.md || echo "BREAKING=true" >> $GITHUB_ENV

      - name: Comment breaking changes on PR
        if: env.BREAKING == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const body = fs.readFileSync('breaking-report.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Breaking changes detected\n\n' + body,
            });

      - name: Run contract tests
        run: npm test -- tests/api/openapi.spec.ts
```

The PR comment makes the breakage visible to reviewers without forcing them to dig into CI logs.

---

## 9. Code Examples (TypeScript)

A consolidated test file that exercises each layer above.

```typescript
// tests/api/openapi.spec.ts
import { execSync } from "node:child_process";
import fs from "node:fs";
import { describe, test, expect, beforeAll } from "vitest";
import Ajv from "ajv";
import addFormats from "ajv-formats";
import yaml from "js-yaml";
import { faker } from "@faker-js/faker";
import { z } from "zod";
import type { paths, components } from "../../src/generated/api";

// ---------- Setup ----------

type CreateUserBody =
  paths["/users"]["post"]["requestBody"]["content"]["application/json"];
type User = components["schemas"]["User"];

const SPEC_PATH = "openapi.yaml";
const BASE_URL = process.env.API_BASE_URL ?? "http://localhost:3000";

const spec = yaml.load(fs.readFileSync(SPEC_PATH, "utf8")) as any;

const ajv = new Ajv({ allErrors: true, strict: false });
addFormats(ajv);
for (const [name, schema] of Object.entries<any>(spec.components.schemas)) {
  ajv.addSchema(schema, `#/components/schemas/${name}`);
}

function validateAgainstSpec(schemaRef: string, payload: unknown): void {
  const validate = ajv.getSchema(schemaRef);
  if (!validate) throw new Error(`Schema ${schemaRef} not found in spec`);
  if (!validate(payload)) {
    throw new Error(
      `Schema ${schemaRef} validation failed:\n${ajv.errorsText(
        validate.errors,
        { separator: "\n" }
      )}`
    );
  }
}

// Zod schema mirroring components.schemas.User for nicer error messages
const UserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  name: z.string().min(1).max(100),
  role: z.enum(["admin", "user"]),
  createdAt: z.string().datetime(),
});

// ---------- 1. Spectral inline smoke check ----------

describe("OpenAPI spec hygiene", () => {
  test("spec passes Spectral lint", () => {
    // Ignore non-zero exit if there are only warnings; rely on dedicated CI step
    // for strict gating. This test is a fast feedback signal during local dev.
    const result = execSync(
      "npx spectral lint openapi.yaml --format json --quiet",
      { encoding: "utf8" }
    );
    const findings = JSON.parse(result);
    const errors = findings.filter((f: any) => f.severity === 0);
    expect(errors).toEqual([]);
  });
});

// ---------- 2 + 3. Generated types + request validation pre-send ----------

describe("POST /users", () => {
  test("creates a user with a spec-conformant body", async () => {
    // Generated type guarantees we are building a valid shape at compile time
    const body: CreateUserBody = {
      email: "ada@example.com",
      name: "Ada Lovelace",
    };

    // Runtime check: catches drift between spec and generated types
    validateAgainstSpec("#/components/schemas/CreateUserRequest", body);

    const res = await fetch(`${BASE_URL}/users`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });

    expect(res.status).toBe(201);

    // Response validation
    const json = (await res.json()) as User;
    const user = UserSchema.parse(json);

    expect(user.email).toBe(body.email);
    expect(user.role).toBe("user");
  });

  // ---------- 4. Faker-driven generative case ----------

  test("accepts 25 random spec-valid users", async () => {
    for (let i = 0; i < 25; i++) {
      const body: CreateUserBody = {
        email: faker.internet.email(),
        name: faker.person.fullName().slice(0, 100),
      };

      validateAgainstSpec("#/components/schemas/CreateUserRequest", body);

      const res = await fetch(`${BASE_URL}/users`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(body),
      });

      // 201 for new, 409 for duplicate email (Faker can collide)
      expect([201, 409]).toContain(res.status);

      if (res.status === 201) {
        const json = await res.json();
        UserSchema.parse(json);
      }
    }
  });

  // ---------- 5. Boundary value example ----------

  test("rejects name longer than maxLength", async () => {
    const body = {
      email: faker.internet.email(),
      name: "a".repeat(101), // maxLength is 100
    };

    // Pre-send: prove our test is genuinely sending an invalid name
    expect(() =>
      validateAgainstSpec("#/components/schemas/CreateUserRequest", body)
    ).toThrow();

    const res = await fetch(`${BASE_URL}/users`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });

    expect(res.status).toBe(400);
  });
});

// ---------- 6. oasdiff invocation as a separate step ----------

describe.skipIf(!process.env.RUN_OASDIFF)("breaking change check", () => {
  test("no breaking changes vs main", () => {
    // Requires `oasdiff` on PATH and `openapi.base.yaml` checked out from main.
    expect(() =>
      execSync(
        "oasdiff breaking openapi.base.yaml openapi.yaml --fail-on ERR",
        { stdio: "pipe" }
      )
    ).not.toThrow();
  });
});
```

Run with:

```bash
npx vitest run tests/api/openapi.spec.ts
RUN_OASDIFF=1 npx vitest run tests/api/openapi.spec.ts
```

---

## 10. Limitations

OpenAPI is a **structural** contract. It tells you the shape of requests and responses, the status codes, the authentication scheme. It does not tell you what the API **does**.

### What OpenAPI cannot express

- **Cross-field constraints**: "if `type === 'business'`, then `taxId` is required and `personalId` must be absent." OpenAPI 3.1 supports `oneOf`/`if`/`then` but the tooling is uneven; in practice, document these in `description` and validate them in handwritten tests.
- **Business rules**: "a user cannot transfer more than their balance," "a coupon expires 30 days after creation." These are not shape rules; they are behavior.
- **Asynchronous behavior**: webhooks, eventual consistency, "this returns 202 and the resource appears within 5 seconds." Some of this is captured in OpenAPI 3.1's `webhooks`, but timing and ordering are not.
- **Side effects**: "calling this endpoint twice with the same idempotency key returns the same response." Idempotency, retry semantics, and rate-limiting headers can be **described** in the spec but not **enforced** by it.
- **State transitions**: "you cannot DELETE a user who has active orders." This is a workflow rule; it lives in handwritten state-machine tests.
- **Performance contracts**: "p99 latency under 200ms." OpenAPI is silent on time. Use load tests.
- **Security beyond auth scheme**: the spec says "this needs a JWT," not "this JWT must have the `admin` role and the `users:write` scope and not be revoked."

### When you still need handwritten tests

- **Workflow tests** that span multiple endpoints (create order, ship it, cancel it).
- **Authorization matrix** tests across roles and resources.
- **Concurrency and race conditions**.
- **Time-dependent behavior** (expirations, schedules, cron-driven workflows).
- **Error path coverage** when the same error code can arise from many causes; the spec says "400" but only a behavior test distinguishes "missing field" from "invalid format" from "violates business rule".

### The healthy split

- **Spec-derived tests** cover the **structure** of every endpoint exhaustively, generated from the spec, with negligible maintenance cost.
- **Handwritten tests** cover **behavior** that the spec cannot express, focused, well-named, and small in number.

When the two are kept honest about which is which, the spec-derived suite catches drift and shape regressions automatically, and the handwritten suite stays small enough to actually maintain. That balance is the goal.

---

## See also

- `contract-testing.md` — consumer-driven contract testing with Pact.
- `schema-validation.md` — runtime schema validation patterns beyond OpenAPI.
- `rest-api-patterns.md` — design patterns the spec should encode.
