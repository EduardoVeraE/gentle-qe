<!-- Skill: api-testing · Template: contract-test-charter -->
<!-- Placeholders: {{consumer_name}}, {{consumer_repo}}, {{provider_name}}, {{provider_repo}}, {{contract_approach}}, {{pact_broker_url}}, {{openapi_spec_url}}, {{verification_cadence}}, {{can_i_deploy_policy}}, {{owner_consumer}}, {{owner_provider}}, {{last_verified_date}}, {{known_drift}} -->

# Contract test charter — {{consumer_name}} ↔ {{provider_name}}

One charter per consumer-provider pair. If a consumer talks to N providers, you have N charters.

## Identity

| Field | Value |
| ----- | ----- |
| Consumer | {{consumer_name}} (`{{consumer_repo}}`) |
| Provider | {{provider_name}} (`{{provider_repo}}`) |
| Contract approach | {{contract_approach}} <!-- e.g., Pact (consumer-driven) + OpenAPI (provider spec) — both required to pass --> |
| Pact broker URL | {{pact_broker_url}} <!-- omit if approach is OpenAPI-only --> |
| OpenAPI spec URL | {{openapi_spec_url}} <!-- omit if approach is Pact-only --> |
| Owner (consumer side) | {{owner_consumer}} |
| Owner (provider side) | {{owner_provider}} |
| Last verified | {{last_verified_date}} |

## Approach selection — why this combo

- **Pact only**: closed ecosystem, both sides under our control, fast feedback per consumer.
- **OpenAPI only**: external/public API, multiple unknown consumers, spec is the law.
- **Both**: internal API with external consumers OR a migration phase. Pact catches consumer assumptions; OpenAPI catches provider drift. Cost: two pipelines to keep green.

State the chosen reason explicitly: {{contract_approach}}.

## Interactions

Each row is one request/response pair the consumer depends on. Every row MUST have a verifiable test.

| # | Description | Method | Path | Request shape | Response status | Response shape | Provider state required |
| - | ----------- | ------ | ---- | ------------- | --------------- | -------------- | ----------------------- |
| 1 | Fetch order by id | GET | `/orders/{id}` | `id: uuid` in path, `Authorization` header | 200 | `Order` | `order {id} exists and belongs to caller` |
| 2 | Fetch order — not found | GET | `/orders/{id}` | unknown id | 404 | `Problem` | `no order with id {id}` |
| 3 | Create order | POST | `/orders` | `CreateOrder` body, `Idempotency-Key` | 201 | `Order` + `Location` header | `caller has active cart` |
| 4 | Create order — duplicate key | POST | `/orders` | same body + same `Idempotency-Key` within TTL | 200 | original `Order` | `previous order exists for that key` |
| 5 | Cancel order | POST | `/orders/{id}/cancel` | empty body | 204 | empty | `order {id} is in state PLACED` |

Add or remove rows to reflect the real consumer surface — do NOT pad.

## States required (provider side)

The provider MUST implement state handlers for every `Provider state required` value above. State handlers seed the data needed for verification and clean it up afterwards. Examples:

- `order {id} exists and belongs to caller`
- `no order with id {id}`
- `caller has active cart`
- `previous order exists for that key`
- `order {id} is in state PLACED`

State handlers live in the provider repo and are versioned with it.

## Verification cadence

{{verification_cadence}} <!-- e.g., consumer publishes contract on every PR; provider verifies on every PR and nightly against the latest tagged consumer contracts; pre-deploy gate runs can-i-deploy -->

| Trigger | What runs | Where |
| ------- | --------- | ----- |
| Consumer PR | Generate + publish contract to broker; tag with branch | Consumer CI |
| Provider PR | Verify all `main`-tagged consumer contracts | Provider CI |
| Nightly | Verify all environment-tagged contracts (`prod`, `stage`) | Provider CI |
| Pre-deploy | `can-i-deploy --pacticipant {{provider_name}} --version $SHA --to-environment prod` | Deploy pipeline |

## can-i-deploy policy

{{can_i_deploy_policy}} <!-- e.g., A version may deploy to <env> only if every contract tagged with <env> has been verified successfully against that version. Unverified == blocked. No overrides without two-owner sign-off recorded in the broker. -->

Hard rules:
- A red contract blocks deploy. No "we'll fix it after" — fix or roll back the consumer.
- Overrides require BOTH owners to sign off and leave a comment in the broker.
- Stale contracts (no run in N days) are treated as red.

## Known acceptable drift

Some breaks are intentional and explicitly allowlisted. Each entry MUST have an expiry date.

| Field | Drift | Reason | Expires | Approved by |
| ----- | ----- | ------ | ------- | ----------- |
| {{known_drift}} <!-- e.g., Order.legacyTotalCents --> | Field removed from response | Consumer migrated to `Order.total` in v3.2 | 2026-09-01 | {{owner_consumer}} + {{owner_provider}} |

If the table is empty, write "None" — do NOT delete the section.

## Review

- Owners review this charter on every contract change and at minimum quarterly.
- Update `Last verified` whenever the verification pipeline goes green end-to-end.
- Archive the charter alongside the consumer's release tag.
