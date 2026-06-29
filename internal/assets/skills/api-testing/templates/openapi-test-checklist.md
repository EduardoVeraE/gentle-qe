<!-- Skill: api-testing · Template: openapi-test-checklist -->
<!-- Placeholders: {{api_name}}, {{spec_version}}, {{review_date}}, {{owner}}, {{owner_users}}, {{owner_orders}} -->

# OpenAPI operation gate — {{api_name}}

Spec version: `{{spec_version}}` · Reviewed: {{review_date}} · Owner: {{owner}}

One row per operation in the OpenAPI document. An operation is NOT release-ready until every column is green.

## Legend

- **Spectral lint pass**: `spectral lint openapi.yaml` returns 0 errors for this operation.
- **Schema valid (request)**: A real request matches `requestBody` / `parameters` schema.
- **Schema valid (response 2xx)**: A real success response matches the documented schema.
- **Schema valid (response 4xx/5xx)**: At least one error response per documented status validates.
- **Examples valid**: All `examples` / `example` blocks parse against their schemas.
- **Security defined**: `security` is set (operation-level or root-level inheritance is explicit).
- **Breaking-change diff**: `oasdiff` vs. previous release shows no breaking change (or change is accepted).
- **Tested at level**: C = Component, I = Integration, S = System. Multiple allowed.
- **Owner**: GitHub handle accountable for this operation.
- **Last reviewed**: ISO date of last sign-off.

## Checklist

| Operation | Spectral lint pass | Schema valid (request) | Schema valid (response 2xx) | Schema valid (response 4xx/5xx) | Examples valid | Security defined | Breaking-change diff | Tested at level (C/I/S) | Owner | Last reviewed |
| --------- | ------------------ | ---------------------- | --------------------------- | ------------------------------- | -------------- | ---------------- | -------------------- | ----------------------- | ----- | ------------- |
| `GET /users` | [ ] | N/A | [ ] | [ ] | [ ] | [ ] | [ ] | C, I, S | {{owner_users}} | {{review_date}} |
| `POST /users` | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | C, I, S | {{owner_users}} | {{review_date}} |
| `GET /users/{id}` | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | I, S | {{owner_users}} | {{review_date}} |
| `PATCH /users/{id}` | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | C, I, S | {{owner_users}} | {{review_date}} |
| `DELETE /users/{id}` | [ ] | N/A | [ ] | [ ] | [ ] | [ ] | [ ] | I, S | {{owner_users}} | {{review_date}} |
| `GET /users/{id}/orders` | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | I, S | {{owner_orders}} | {{review_date}} |

## How to fill a row

1. Run `spectral lint` — fix all errors before checking the box.
2. Capture a real request/response from an integration test (NOT a mock).
3. Validate against the schema with the same library the runtime uses (Ajv, openapi-core, Bean Validation).
4. Confirm `oasdiff base.yaml head.yaml` is clean — or attach the change-acceptance link.
5. Confirm at least one automated test exists at each marked level.
6. Update `Last reviewed` and your handle.

## Anti-patterns

- "It works locally" — without an automated assertion the row stays unchecked.
- Marking 2xx valid without a 4xx case — error contracts are part of the API.
- Skipping `Security defined` because the gateway "handles it" — the spec must still declare it.
- Reviewing in bulk after a release — review per PR, per operation.

## Per-release ritual

- Generate the diff: `oasdiff diff -base prev.yaml -revision head.yaml -fail-on ERR`.
- Re-run the full checklist for any operation touched in the diff.
- Archive a snapshot of this file alongside the release tag.

## Status meanings

- `[ ]` — not yet verified; treat as red.
- `[x]` — verified by an automated test in the current commit.
- `N/A` — column does not apply (e.g., no request body for `GET`, no 4xx documented for `/health`).
- `WAIVED` — temporarily accepted with a linked ticket and an expiry date. Waivers older than 30 days fail CI.

## Required artifacts per row

For every checked operation the repo MUST contain:

1. A test file referencing the operation (`tests/api/<resource>/<operation>.spec.ts` or equivalent).
2. A schema-validation assertion using the same library as the runtime (Ajv, openapi-core, Bean Validation).
3. At least one negative case (4xx) — a green 2xx alone is not enough.
4. A trace ID captured in the test output for debuggability.

## How operations enter and leave this list

- A new operation in the OpenAPI spec MUST appear here within the same PR that adds it.
- A removed operation MUST be deleted from this checklist in the same PR (and added to the deprecation log).
- Renamed operations: delete the old row, add the new one — do NOT edit in place; the audit trail matters.

## Failure modes to watch

- `2xx` schema valid but `examples` outdated — examples are part of the contract; lint catches drift only if you keep them under Spectral rules.
- `Security defined` checked because root-level `security` is set, but the operation overrides with `security: []` (public). Make the override explicit and tested.
- `Breaking-change diff` checked despite a removed enum value — `oasdiff` flags this as ERR; never override without a written acceptance.
- `Tested at level` lists `S` only — system tests are slow and noisy; if the logic can be exercised at `C` or `I`, push it down.
