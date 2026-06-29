# API Testing Levels — ISTQB CTFL Applied

ISTQB Foundation Level (CTFL) defines four test levels: **Component**, **Integration**, **System**, and **Acceptance**. Each level has a different scope, oracle, mocking strategy, and bug profile. Applied to APIs, the model maps cleanly to:

- **Component test** — one endpoint in isolation, downstream collaborators mocked
- **Integration test** — the endpoint plus real adjacent services (DB, queue, sibling APIs)
- **System test** — the full API surface running against a real environment, real auth, real data
- **Acceptance test** — business stakeholder validation; usually a thin wrapper over system tests

This document focuses on the first three (the engineering-owned levels). Acceptance is product-driven and reuses system-test infrastructure.

## Why Levels Matter for API Testing

Levels are not a bureaucratic ceremony. Each level answers a different question, has a different oracle, and catches a different class of bug. Mixing them produces tests that are slow, flaky, and unclear about what they prove.

| Concern             | Component                  | Integration                  | System                        |
| ------------------- | -------------------------- | ---------------------------- | ----------------------------- |
| **Oracle**          | Code + schema              | Provider response + DB state | BDD scenario from product     |
| **Downstream deps** | Mocked (MSW, WireMock)     | Real (Testcontainers)        | Real (staging / ephemeral)    |
| **Auth**            | Stubbed token              | Real local IdP or stub       | Real IdP, real tokens         |
| **Data**            | In-memory fixtures         | Seeded DB, cleaned per test  | Test tenant in real env       |
| **Speed**           | < 100 ms                   | 1-5 s                        | 5-60 s                        |
| **Stability**       | Deterministic              | Mostly deterministic         | Subject to env drift          |
| **Bug profile**     | Schema, validation, mapping | Wiring, retries, timeouts    | Cross-feature, business rules |

If you write everything as a system test, you get slow flaky suites that fail for reasons unrelated to the change. If you write everything as a component test, you ship code where the wiring between services is broken in production. The pyramid is not optional — it is what makes the suite trustworthy AND fast.

A note up front: **not every endpoint needs all three levels**. A trivial `GET /health` endpoint needs one component test and nothing else. A payment endpoint needs all three. Decide based on risk and complexity, not dogma.

---

## 1. Component Test (Single Endpoint)

### Scope

One endpoint, one HTTP method. The unit under test is the request handler plus its immediate validation, serialization, and response shaping. **Everything downstream is mocked**: database, queues, sibling services, external APIs, clocks, ID generators.

### Oracle

The oracle is the code itself plus the response schema. You assert what the handler should produce given a controlled input.

### What to Assert

- HTTP status code (happy path AND error paths)
- Response schema (Zod, JSON Schema)
- Response headers in scope: `Content-Type`, `Location`, `ETag`, error contract
- Request validation: rejects malformed input with `400` or `422`
- Auth gate: missing/invalid token returns `401`
- Authorization gate: wrong role returns `403`

### What NOT to Assert at This Level

- Cross-service flows (that is integration)
- Real DB behavior, transactions, constraints (integration)
- Race conditions and concurrent writes (integration or system)
- End-to-end business rules that span multiple endpoints (system)
- Performance characteristics (load testing, separate concern)

### Isolation Pattern

The handler talks to a `Repository` or `Client` abstraction. In the component test, that abstraction is replaced with a mock that returns canned responses or throws canned errors. The test never opens a TCP socket to a real downstream service.

### Tools

| Stack       | HTTP driver                       | Mocking                                 | Schema                       |
| ----------- | --------------------------------- | --------------------------------------- | ---------------------------- |
| TypeScript  | Playwright `request` / Supertest  | MSW (network), `vi.mock` / Jest mocks   | Zod                          |
| Java        | REST Assured / MockMvc            | WireMock, Mockito                       | `json-schema-validator`      |

### TypeScript Example — `POST /orders`

Happy path. Note: this uses `@playwright/test`'s `request` fixture against a server started in-process with mocked dependencies. The order service exposes a factory that accepts a mocked repository and notifier.

```typescript
import { test, expect, request as pwRequest } from "@playwright/test";
import { z } from "zod";
import { startTestServer } from "../helpers/test-server";

const OrderResponse = z.object({
  id: z.string().uuid(),
  status: z.literal("PENDING"),
  total: z.number().positive(),
  createdAt: z.string().datetime(),
});

const ErrorResponse = z.object({
  code: z.string(),
  message: z.string(),
  details: z.array(z.object({ field: z.string(), reason: z.string() })).optional(),
});

test.describe("POST /orders — component", () => {
  let baseURL: string;
  let stop: () => Promise<void>;
  const repo = {
    save: async (order: unknown) => ({ ...(order as object), id: "11111111-1111-1111-1111-111111111111" }),
  };
  const notifier = { publish: async () => undefined };

  test.beforeAll(async () => {
    const server = await startTestServer({ repo, notifier });
    baseURL = server.url;
    stop = server.stop;
  });

  test.afterAll(async () => {
    await stop();
  });

  test("happy path: returns 201 with valid order", async () => {
    const ctx = await pwRequest.newContext({ baseURL });
    const res = await ctx.post("/orders", {
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
        "Idempotency-Key": "key-123",
      },
      data: { items: [{ sku: "ABC", qty: 2, unitPrice: 10 }] },
    });

    expect(res.status()).toBe(201);
    expect(res.headers()["location"]).toMatch(/\/orders\/[0-9a-f-]+/);
    expect(res.headers()["content-type"]).toContain("application/json");

    const body = await res.json();
    expect(() => OrderResponse.parse(body)).not.toThrow();
    expect(body.total).toBe(20);
  });
});
```

Error path. Same setup, asserts the error contract.

```typescript
test("error path: 422 when items is empty", async () => {
  const ctx = await pwRequest.newContext({ baseURL });
  const res = await ctx.post("/orders", {
    headers: {
      Authorization: "Bearer test-token",
      "Content-Type": "application/json",
    },
    data: { items: [] },
  });

  expect(res.status()).toBe(422);
  const body = await res.json();
  expect(() => ErrorResponse.parse(body)).not.toThrow();
  expect(body.code).toBe("VALIDATION_ERROR");
  expect(body.details).toContainEqual({ field: "items", reason: "MUST_NOT_BE_EMPTY" });
});

test("error path: 401 when Authorization header missing", async () => {
  const ctx = await pwRequest.newContext({ baseURL });
  const res = await ctx.post("/orders", {
    headers: { "Content-Type": "application/json" },
    data: { items: [{ sku: "ABC", qty: 1, unitPrice: 10 }] },
  });

  expect(res.status()).toBe(401);
  expect(res.headers()["www-authenticate"]).toBeDefined();
});
```

Why this is a component test: the repo is a stub returning a fixed UUID, the notifier is a no-op, and there is no DB, no queue, no real IdP. The assertions cover **only** what the handler controls: status, schema, headers, error shape.

### Java Example — `POST /orders`

Happy path with REST Assured + MockMvc + Mockito. Downstream `OrderRepository` and `Notifier` are mocked.

```java
import static io.restassured.RestAssured.given;
import static io.restassured.module.mockmvc.RestAssuredMockMvc.*;
import static org.hamcrest.Matchers.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import io.restassured.module.mockmvc.RestAssuredMockMvc;
import io.restassured.module.jsv.JsonSchemaValidator;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.test.web.servlet.MockMvc;

@SpringBootTest
class OrderControllerComponentTest {

    @Autowired private MockMvc mockMvc;
    @MockBean private OrderRepository repo;
    @MockBean private Notifier notifier;

    @BeforeEach
    void setup() {
        RestAssuredMockMvc.mockMvc(mockMvc);
        when(repo.save(any())).thenAnswer(inv -> {
            Order o = inv.getArgument(0);
            o.setId("11111111-1111-1111-1111-111111111111");
            return o;
        });
    }

    @Test
    @DisplayName("happy path: returns 201 with valid order")
    void createOrderHappyPath() {
        given()
            .header("Authorization", "Bearer test-token")
            .header("Idempotency-Key", "key-123")
            .contentType("application/json")
            .body("{\"items\":[{\"sku\":\"ABC\",\"qty\":2,\"unitPrice\":10}]}")
        .when()
            .post("/orders")
        .then()
            .statusCode(201)
            .header("Location", matchesRegex(".*/orders/[0-9a-f-]+"))
            .contentType("application/json")
            .body("status", equalTo("PENDING"))
            .body("total", equalTo(20))
            .body(JsonSchemaValidator.matchesJsonSchemaInClasspath("schemas/order.json"));
    }
}
```

Error path.

```java
@Test
@DisplayName("error path: 422 when items is empty")
void createOrderEmptyItems() {
    given()
        .header("Authorization", "Bearer test-token")
        .contentType("application/json")
        .body("{\"items\":[]}")
    .when()
        .post("/orders")
    .then()
        .statusCode(422)
        .body("code", equalTo("VALIDATION_ERROR"))
        .body("details.find { it.field == 'items' }.reason", equalTo("MUST_NOT_BE_EMPTY"));
}

@Test
@DisplayName("error path: 401 when Authorization header missing")
void createOrderNoAuth() {
    given()
        .contentType("application/json")
        .body("{\"items\":[{\"sku\":\"ABC\",\"qty\":1,\"unitPrice\":10}]}")
    .when()
        .post("/orders")
    .then()
        .statusCode(401)
        .header("WWW-Authenticate", notNullValue());
}
```

---

## 2. Integration Test (Multi-Service)

### Scope

The endpoint plus its **real** downstream dependencies. Real database (Postgres in a container), real message broker (Kafka, RabbitMQ in a container), real adjacent service or a contract-tested fake. The test verifies that the endpoint correctly **wires** all of these together.

### Oracle

Two oracles, both must hold:

1. **Provider response** — what the endpoint returns to the caller
2. **Side-effects in collaborators** — what got written to the DB, what message was published, what the downstream service was called with

A test that only checks the response misses half the integration story.

### What to Assert

- Contract adherence: request/response shape against OpenAPI or Pact contract
- Propagation: trace headers, correlation IDs flow to downstream calls
- Retries: transient failures in downstream services are retried per policy
- Timeouts: slow downstream returns `504` or fallback within the SLO
- Idempotency under partial failure: duplicate request with same `Idempotency-Key` produces same outcome even if first attempt crashed mid-flight
- Transactional integrity: failure mid-flow does not leave orphaned rows or unconsumed messages
- DB state: the row was written, with the expected columns, in the expected status

### Tools

| Stack       | DB / Queue / Service                  | Contract                       |
| ----------- | ------------------------------------- | ------------------------------ |
| TypeScript  | Testcontainers (Node), Docker Compose | Pact, OpenAPI validator        |
| Java        | Testcontainers (Java), JUnit 5        | Pact JVM, OpenAPI validator    |

### TypeScript Example — `POST /orders` Integration

Real Postgres via Testcontainers, real `notification-service` (or a Pact-verified fake). Asserts response **and** DB row **and** outbound notification call.

```typescript
import { test, expect, request as pwRequest } from "@playwright/test";
import { PostgreSqlContainer, StartedPostgreSqlContainer } from "@testcontainers/postgresql";
import { GenericContainer, StartedTestContainer } from "testcontainers";
import { Pool } from "pg";
import { startApp } from "../../src/app";

let pg: StartedPostgreSqlContainer;
let notif: StartedTestContainer;
let db: Pool;
let baseURL: string;
let stopApp: () => Promise<void>;

test.beforeAll(async () => {
  pg = await new PostgreSqlContainer("postgres:16").start();
  notif = await new GenericContainer("notification-service:test").withExposedPorts(8080).start();

  db = new Pool({ connectionString: pg.getConnectionUri() });
  await db.query("CREATE TABLE orders (id uuid PRIMARY KEY, status text, total numeric, created_at timestamptz)");

  const app = await startApp({
    dbUrl: pg.getConnectionUri(),
    notifierUrl: `http://${notif.getHost()}:${notif.getMappedPort(8080)}`,
  });
  baseURL = app.url;
  stopApp = app.stop;
});

test.afterAll(async () => {
  await stopApp();
  await db.end();
  await pg.stop();
  await notif.stop();
});

test("POST /orders persists row and notifies", async () => {
  const ctx = await pwRequest.newContext({ baseURL });
  const res = await ctx.post("/orders", {
    headers: {
      Authorization: "Bearer integration-token",
      "Idempotency-Key": "int-key-1",
      traceparent: "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01",
    },
    data: { items: [{ sku: "ABC", qty: 2, unitPrice: 10 }] },
  });

  expect(res.status()).toBe(201);
  const body = await res.json();

  // DB side-effect
  const { rows } = await db.query("SELECT id, status, total FROM orders WHERE id = $1", [body.id]);
  expect(rows).toHaveLength(1);
  expect(rows[0].status).toBe("PENDING");
  expect(Number(rows[0].total)).toBe(20);

  // Notification side-effect — query the notif service's recording endpoint
  const notifCtx = await pwRequest.newContext({
    baseURL: `http://${notif.getHost()}:${notif.getMappedPort(8080)}`,
  });
  const recorded = await notifCtx.get(`/_test/recorded?orderId=${body.id}`);
  expect(recorded.status()).toBe(200);
  const events = await recorded.json();
  expect(events).toHaveLength(1);
  expect(events[0].traceparent).toMatch(/^00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-/);
});

test("POST /orders is idempotent under same Idempotency-Key", async () => {
  const ctx = await pwRequest.newContext({ baseURL });
  const payload = {
    headers: { Authorization: "Bearer integration-token", "Idempotency-Key": "int-key-2" },
    data: { items: [{ sku: "ABC", qty: 1, unitPrice: 10 }] },
  };
  const a = await ctx.post("/orders", payload);
  const b = await ctx.post("/orders", payload);

  expect(a.status()).toBe(201);
  expect(b.status()).toBe(201);
  expect((await a.json()).id).toBe((await b.json()).id);

  const { rows } = await db.query("SELECT count(*)::int AS n FROM orders WHERE id = $1", [(await a.json()).id]);
  expect(rows[0].n).toBe(1);
});
```

### Java Example — `POST /orders` Integration

```java
import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.*;

import org.junit.jupiter.api.*;
import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.sql.*;

@Testcontainers
class OrderIntegrationTest {

    @Container static PostgreSQLContainer<?> pg = new PostgreSQLContainer<>("postgres:16");
    @Container static GenericContainer<?> notif = new GenericContainer<>("notification-service:test").withExposedPorts(8080);

    static String baseUrl;

    @BeforeAll
    static void boot() throws Exception {
        try (Connection c = DriverManager.getConnection(pg.getJdbcUrl(), pg.getUsername(), pg.getPassword());
             Statement s = c.createStatement()) {
            s.execute("CREATE TABLE orders (id uuid PRIMARY KEY, status text, total numeric, created_at timestamptz)");
        }
        baseUrl = AppRunner.start(pg.getJdbcUrl(), pg.getUsername(), pg.getPassword(),
            "http://" + notif.getHost() + ":" + notif.getMappedPort(8080));
    }

    @Test
    @DisplayName("persists row and notifies downstream")
    void createOrderPersistsAndNotifies() throws Exception {
        String id = given()
            .baseUri(baseUrl)
            .header("Authorization", "Bearer integration-token")
            .header("Idempotency-Key", "int-key-1")
            .header("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
            .contentType("application/json")
            .body("{\"items\":[{\"sku\":\"ABC\",\"qty\":2,\"unitPrice\":10}]}")
        .when()
            .post("/orders")
        .then()
            .statusCode(201)
            .extract().path("id");

        try (Connection c = DriverManager.getConnection(pg.getJdbcUrl(), pg.getUsername(), pg.getPassword());
             PreparedStatement ps = c.prepareStatement("SELECT status, total FROM orders WHERE id = ?::uuid")) {
            ps.setString(1, id);
            try (ResultSet rs = ps.executeQuery()) {
                Assertions.assertTrue(rs.next());
                Assertions.assertEquals("PENDING", rs.getString("status"));
                Assertions.assertEquals(20.0, rs.getDouble("total"), 0.001);
            }
        }

        given()
            .baseUri("http://" + notif.getHost() + ":" + notif.getMappedPort(8080))
        .when()
            .get("/_test/recorded?orderId=" + id)
        .then()
            .statusCode(200)
            .body("size()", equalTo(1))
            .body("[0].traceparent", startsWith("00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-"));
    }
}
```

Why this is integration: real DB, real downstream service, real wire format. The test does NOT rebuild the user journey or test business outcomes — it tests **the wiring** between this endpoint and the systems it touches.

---

## 3. System Test (Full E2E API)

### Scope

The entire API surface running in a real environment: real auth provider (OAuth2/OIDC), real DB with realistic data, real adjacent services, real network. The test exercises a **business journey** that crosses multiple endpoints.

### Oracle

The oracle is a **BDD scenario from product**: given a user with X, when they do Y, then the system reaches state Z. The oracle is not the code, not the schema — it is the business rule.

### What to Assert

- The user journey reaches the expected business outcome
- Cross-feature interactions hold (creating an order updates inventory, triggers notification, generates an invoice)
- Business rules enforced end-to-end (e.g. order over $10K requires manager approval)
- Real auth, real tokens, real session management
- Realistic data volumes do not break pagination or filtering

### Tools

| Stack       | Driver                         | Environment                       |
| ----------- | ------------------------------ | --------------------------------- |
| TypeScript  | Playwright `request` fixture   | Ephemeral env or staging          |
| Java        | REST Assured                   | Ephemeral env or staging          |

### TypeScript Example — Full Journey

Login → create order → confirm payment → fetch order. Real staging environment. Uses Playwright `request` fixture (NOT browser context — this is API testing).

```typescript
import { test, expect, request as pwRequest, APIRequestContext } from "@playwright/test";

const STAGING = process.env.STAGING_URL ?? "https://api.staging.example.com";

async function login(ctx: APIRequestContext, user: string, pass: string): Promise<string> {
  const res = await ctx.post("/auth/token", {
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    form: { grant_type: "password", username: user, password: pass, scope: "orders" },
  });
  expect(res.status()).toBe(200);
  return (await res.json()).access_token as string;
}

test.describe("Order journey — system", () => {
  test("user creates, pays, and fetches an order", async () => {
    const ctx = await pwRequest.newContext({ baseURL: STAGING });
    const token = await login(ctx, process.env.E2E_USER!, process.env.E2E_PASS!);
    const auth = { Authorization: `Bearer ${token}` };

    const create = await ctx.post("/orders", {
      headers: { ...auth, "Idempotency-Key": `sys-${Date.now()}` },
      data: { items: [{ sku: "ABC", qty: 2, unitPrice: 10 }] },
    });
    expect(create.status()).toBe(201);
    const order = await create.json();
    expect(order.status).toBe("PENDING");

    const pay = await ctx.post(`/orders/${order.id}/payments`, {
      headers: auth,
      data: { method: "card", token: "tok_test_visa" },
    });
    expect(pay.status()).toBe(200);

    // Eventual consistency — poll for CONFIRMED
    let fetched: { status: string } = { status: "PENDING" };
    for (let i = 0; i < 10 && fetched.status !== "CONFIRMED"; i++) {
      await new Promise((r) => setTimeout(r, 500));
      const get = await ctx.get(`/orders/${order.id}`, { headers: auth });
      expect(get.status()).toBe(200);
      fetched = await get.json();
    }
    expect(fetched.status).toBe("CONFIRMED");
  });
});
```

### Java Example — Full Journey

```java
import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.*;
import io.restassured.RestAssured;
import org.junit.jupiter.api.*;

class OrderSystemTest {

    static String token;
    static final String BASE = System.getenv().getOrDefault("STAGING_URL", "https://api.staging.example.com");

    @BeforeAll
    static void login() {
        RestAssured.baseURI = BASE;
        token = given()
            .contentType("application/x-www-form-urlencoded")
            .formParam("grant_type", "password")
            .formParam("username", System.getenv("E2E_USER"))
            .formParam("password", System.getenv("E2E_PASS"))
            .formParam("scope", "orders")
        .when()
            .post("/auth/token")
        .then()
            .statusCode(200)
            .extract().path("access_token");
    }

    @Test
    @DisplayName("user creates, pays, and fetches an order")
    void orderJourney() throws InterruptedException {
        String id = given()
            .header("Authorization", "Bearer " + token)
            .header("Idempotency-Key", "sys-" + System.currentTimeMillis())
            .contentType("application/json")
            .body("{\"items\":[{\"sku\":\"ABC\",\"qty\":2,\"unitPrice\":10}]}")
        .when()
            .post("/orders")
        .then()
            .statusCode(201)
            .body("status", equalTo("PENDING"))
            .extract().path("id");

        given()
            .header("Authorization", "Bearer " + token)
            .contentType("application/json")
            .body("{\"method\":\"card\",\"token\":\"tok_test_visa\"}")
        .when()
            .post("/orders/" + id + "/payments")
        .then()
            .statusCode(200);

        String status = "PENDING";
        for (int i = 0; i < 10 && !"CONFIRMED".equals(status); i++) {
            Thread.sleep(500);
            status = given()
                .header("Authorization", "Bearer " + token)
            .when()
                .get("/orders/" + id)
            .then()
                .statusCode(200)
                .extract().path("status");
        }
        Assertions.assertEquals("CONFIRMED", status);
    }
}
```

Why this is a system test: real auth, real env, multi-endpoint flow, business outcome (`CONFIRMED`) is the assertion.

---

## 4. Decision Matrix — Which Level for Which Concern

Pick the LOWEST level that can give you the answer. Higher levels are slower, flakier, and more expensive to maintain.

| Concern                                  | Component | Integration | System | Notes                                                                       |
| ---------------------------------------- | :-------: | :---------: | :----: | --------------------------------------------------------------------------- |
| Request/response schema                  |     X     |             |        | Component is enough — schema is owned by the handler.                       |
| Validation rules (required, format)      |     X     |             |        | Component. Pure input rejection.                                            |
| HTTP status codes (happy + error)        |     X     |             |        | Component, with mocked downstream errors.                                   |
| Auth (401)                               |     X     |             |        | Component; auth gate is in the handler.                                     |
| Authorization (403, role/tenant)         |     X     |     X       |        | Component for simple roles; integration if policy lives in another service. |
| Headers (Content-Type, Location, ETag)   |     X     |             |        | Component.                                                                  |
| Error contract shape                     |     X     |             |        | Component.                                                                  |
| DB persistence                           |           |     X       |        | Integration. Real DB only.                                                  |
| Transactions, rollback on failure        |           |     X       |        | Integration with induced failure (kill DB mid-write, etc).                  |
| Retries, timeouts, circuit breaker       |           |     X       |        | Integration with controllable downstream (toxiproxy, WireMock delays).     |
| Idempotency under partial failure        |           |     X       |        | Integration. Component idempotency is too shallow.                          |
| Trace propagation, correlation IDs       |           |     X       |        | Integration — assert on downstream collaborator.                            |
| Outbound message published               |           |     X       |        | Integration with real broker.                                               |
| Race conditions, concurrent writes       |           |     X       |   X    | Integration if scoped to one endpoint; system if cross-endpoint.            |
| Business rule spanning ≥ 2 endpoints     |           |             |   X    | System. Use a BDD scenario.                                                 |
| Cross-feature interaction                |           |             |   X    | System.                                                                     |
| Real OAuth2/OIDC flow                    |           |             |   X    | System.                                                                     |
| User journey (login → action → outcome)  |           |             |   X    | System.                                                                     |
| Performance under load                   |           |             |        | None of these — use a load tester (k6, Gatling).                            |
| Security (XSS, SQLi, BOLA, JWT attacks)  |           |             |        | None of these — use a dedicated security skill.                             |

Rule of thumb: if you can answer the question with a mock, do it at component level. If you need real wiring but not real users, integration. If the question is "does the business outcome happen?", system.

---

## 5. Anti-Patterns

### Testing Business Rules at Component Level

Symptom: a component test for `POST /orders` that asserts "if total > $10K and user is not manager, reject". The handler returns the rejection, sure — but the rule lives in a policy service, the user role comes from an IdP, and the threshold is a config value. A component test mocks all of those, so what you are testing is **your mock**, not the rule.

Fix: assert the rule at integration (real policy service) or system (real env). At component level, only assert that the handler **propagates** a rejection from the policy service correctly.

### Mocking Everything in an Integration Test

Symptom: the test sets up Testcontainers Postgres, then mocks the repository that talks to Postgres. Now the DB is running but unused. You have a slow component test that pretends to be an integration test.

Fix: at integration level, **use the real collaborator**. If you mock it, you are at component level — drop the container and call it what it is.

### Running E2E for Trivial Validations

Symptom: a system test that logs in, hits `POST /orders` with `items: []`, and asserts `422`. This took 30 seconds, used a real IdP, and proved nothing that a 50ms component test would not have proven.

Fix: validation belongs at component level. Reserve system tests for **journeys** and **cross-feature outcomes**.

### One Giant System Test for Everything

Symptom: a single 400-line system test that creates a user, logs in, creates an order, pays, refunds, cancels, deletes the user. When it fails, you have no idea where.

Fix: one journey per system test. Split by business outcome. Independent setup/teardown.

### Component Tests Without Schema Validation

Symptom: tests assert `status === 201` and stop there. A field rename ships and silently breaks consumers.

Fix: **every** component test asserts the response against a schema. Zod or JSON Schema, no exceptions. See `schema-validation.md`.

### Integration Tests That Only Check the Response

Symptom: integration test posts an order, asserts `201`, ends. Did the row get written? Did the notification fire? Did the trace propagate? You have no idea.

Fix: at integration level, assert on **side-effects in collaborators**, not just the response.

### Flaky System Tests Treated as "Just Retry"

Symptom: system suite has a 30% retry rate. The team adds `retries: 3` in the config and moves on.

Fix: flakiness is a signal. Either the env is unstable (fix the env), the test has a real race (fix the test), or it relies on data that drifts (use a tenant per run, clean up). Retries paper over real bugs.

---

## 6. Coverage Goal Per Level — The API Test Pyramid

Rough targets for a typical service. Adjust based on risk profile.

| Level       | % of suite | Speed target  | Stability target   |
| ----------- | :--------: | ------------- | ------------------ |
| Component   |   ~70%     | < 100 ms each | > 99.5% pass rate  |
| Integration |   ~25%     | 1-5 s each    | > 99% pass rate    |
| System      |   ~5%      | 5-60 s each   | > 95% pass rate    |

### Why this shape

- **Component tests** are cheap, fast, deterministic. They catch the most bugs (validation, schema, mapping, error contract) for the least cost. Write many.
- **Integration tests** are where real wiring lives. Every external collaborator deserves at least one integration test per endpoint that uses it. Don't write 50 — write the 5 that prove the wiring.
- **System tests** are expensive and brittle. They prove **business outcomes** end-to-end. One per critical journey, no more. If you find yourself adding a 20th system test, you are probably testing something that belongs at integration or component.

### Per-Endpoint Sketch

For a non-trivial endpoint like `POST /orders`:

- **Component**: ~10-15 tests. Happy path, each validation rule, each auth/authz path, each error mapping from downstream, schema, headers.
- **Integration**: ~3-5 tests. DB persistence, idempotency under partial failure, trace propagation, outbound notification, retry on transient downstream failure.
- **System**: ~1 test. The order journey (login → create → pay → confirm → fetch).

For a trivial endpoint like `GET /health`:

- **Component**: 1 test. Returns `200` with `{ status: "ok" }`.
- **Integration**: 0.
- **System**: covered implicitly by every other system test.

### Honesty Check

If your service is a thin CRUD layer over one DB, you probably do not need system tests at all — integration plus a smoke test in CI is enough. If your service orchestrates 5 downstream calls and has complex business rules, you need all three levels and you need to invest in integration the most.

The pyramid is a guideline, not a prescription. The principle is: **push every test to the lowest level that can prove the thing**. That is what makes a suite fast, stable, and worth running.

---

## See Also

- [REST API Patterns](./rest-api-patterns.md) — CRUD, pagination, error patterns
- [Playwright API Testing](./playwright-api-testing.md) — request fixture, Supertest patterns
- [REST Assured Testing](./rest-assured-testing.md) — Java patterns, AssertJ, JSON Schema
- [Schema Validation](./schema-validation.md) — Zod, JSON Schema, strict vs loose
- [Contract Testing](./contract-testing.md) — request/response contracts, idempotency
- [OpenAPI-Driven Testing](./openapi-driven-testing.md) — spec-as-source-of-truth workflow
