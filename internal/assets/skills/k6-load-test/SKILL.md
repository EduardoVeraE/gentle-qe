---
name: k6-load-test
description: "Trigger: k6, load test, stress, spike, soak, SLO thresholds, performance. Write k6 performance test scenarios for QE."
license: Apache-2.0
metadata:
  author: dengineproblem (adapted for gentle-qa)
  version: "1.1"
  source: https://skills.sh/dengineproblem/agents-monorepo/k6-load-test
---

## ISTQB Mapping

| Aspect | Value |
|--------|-------|
| Test level | System Testing, Integration Testing |
| Test type | Non-functional — Performance, Reliability, Scalability |
| Techniques | Boundary Value Analysis (load thresholds), Risk-based testing |
| Test oracle | SLO/SLA definitions = expected behavior under load |

**Core principle**: Performance tests validate Non-Functional Requirements (NFRs). Every test MUST trace back to a documented NFR. If there's no NFR, define it before writing a single line of k6 code.

---

## When to Use

| Test type | Goal | When to run |
|-----------|------|-------------|
| Load | Validate normal traffic | Before every release |
| Stress | Find breaking point | Monthly / major releases |
| Spike | Validate auto-scaling | When infra changes |
| Soak | Detect memory leaks | Weekly / pre-release |
| Breakpoint | Find absolute limit | Capacity planning only |

**Not here**: functional correctness, API contract, UI behavior — use the right layer.

---

## Critical Patterns

### Pattern 1: NFR Definition (do this FIRST)

Define NFRs before writing any k6 script. This is your **test oracle**.

```markdown
# NFR: Checkout API
- p95 response time < 500ms under normal load (100 concurrent users)
- p99 response time < 1000ms under normal load
- Error rate < 1% under normal load
- System must handle 5x normal load before degradation (stress threshold)
- Recovery time < 2 minutes after spike subsides
```

```javascript
// thresholds ARE the NFR in executable form
export const options = {
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],  // NFR: latency
    http_req_failed:   ['rate<0.01'],                  // NFR: error rate
    'http_req_duration{endpoint:checkout}': ['p(95)<800'], // per-endpoint NFR
  },
};
```

### Pattern 2: Load Test (baseline validation)

```javascript
// tests/performance/load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const checkoutDuration = new Trend('checkout_duration');

export const options = {
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed:   ['rate<0.01'],
    error_rate:        ['rate<0.05'],
  },
  stages: [
    { duration: '2m', target: 50 },   // Ramp-up
    { duration: '5m', target: 50 },   // Steady state (normal load)
    { duration: '2m', target: 0 },    // Ramp-down
  ],
};

export default function () {
  const res = http.get(`${__ENV.BASE_URL}/api/products`, {
    tags: { endpoint: 'products' },
  });

  const ok = check(res, {
    'status 200':        (r) => r.status === 200,
    'latency < 500ms':   (r) => r.timings.duration < 500,
    'body not empty':    (r) => r.body.length > 0,
  });

  errorRate.add(!ok);
  sleep(1 + Math.random()); // Realistic think time: 1-2s
}
```

### Pattern 3: Stress Test (breaking point)

```javascript
export const options = {
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed:   ['rate<0.10'],
  },
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },   // Normal load baseline
    { duration: '2m', target: 200 },   // 2x — monitor degradation
    { duration: '5m', target: 200 },
    { duration: '2m', target: 400 },   // 4x — approaching limit
    { duration: '5m', target: 400 },
    { duration: '3m', target: 0 },     // Recovery
  ],
};
```

### Pattern 4: Spike Test (auto-scaling validation)

```javascript
export const options = {
  thresholds: {
    http_req_duration: ['p(95)<3000'],  // More lenient — spike scenario
    http_req_failed:   ['rate<0.15'],
  },
  stages: [
    { duration: '1m',  target: 50 },    // Normal baseline
    { duration: '10s', target: 1000 },  // Spike — 20x surge
    { duration: '3m',  target: 1000 },  // Hold — auto-scaling must kick in
    { duration: '10s', target: 50 },    // Back to normal
    { duration: '3m',  target: 50 },    // Recovery validation
    { duration: '30s', target: 0 },
  ],
};
```

### Pattern 5: Data-Driven with SharedArray

```javascript
import { SharedArray } from 'k6/data';

// SharedArray: loaded ONCE, shared across all VUs — use always for test data
const users = new SharedArray('users', () => JSON.parse(open('./data/users.json')));

export default function () {
  const user = users[__VU % users.length]; // Distribute users across VUs

  const loginRes = http.post(
    `${__ENV.BASE_URL}/api/auth/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(loginRes, { 'login ok': (r) => r.status === 200 });
  const token = loginRes.json('token');

  const res = http.get(`${__ENV.BASE_URL}/api/orders`, {
    headers: { Authorization: `Bearer ${token}` },
    tags: { endpoint: 'orders' },
  });

  check(res, { 'orders ok': (r) => r.status === 200 });
  sleep(2);
}
```

### Pattern 6: User Journey with Groups

```javascript
import { group } from 'k6';

export const options = {
  thresholds: {
    // Per-group thresholds trace back to per-step NFRs
    'http_req_duration{group:::Browse}':    ['p(95)<400'],
    'http_req_duration{group:::Cart}':      ['p(95)<300'],
    'http_req_duration{group:::Checkout}':  ['p(95)<800'],
  },
  stages: [
    { duration: '2m', target: 50 },
    { duration: '5m', target: 50 },
    { duration: '2m', target: 0 },
  ],
};

export default function () {
  group('Browse', () => {
    const r = http.get(`${__ENV.BASE_URL}/api/products`);
    check(r, { 'ok': (r) => r.status === 200 });
    sleep(2);
  });

  group('Cart', () => {
    const r = http.post(`${__ENV.BASE_URL}/api/cart`,
      JSON.stringify({ productId: 1, qty: 1 }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    check(r, { 'added': (r) => r.status === 201 });
    sleep(1);
  });

  group('Checkout', () => {
    const r = http.post(`${__ENV.BASE_URL}/api/orders`,
      JSON.stringify({ paymentMethod: 'card' }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    check(r, { 'ordered': (r) => r.status === 201 });
  });

  sleep(1);
}
```

### Pattern 7: CI/CD Quality Gate

```yaml
# .github/workflows/performance.yml
name: Performance Gate

on:
  push:
    branches: [main]

jobs:
  k6:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run load test
        uses: grafana/k6-action@v0.3.1
        with:
          filename: tests/performance/load-test.js
          flags: --out json=results.json
        env:
          BASE_URL: ${{ secrets.STAGING_URL }}

      - name: Upload results artifact
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: k6-results-${{ github.run_id }}
          path: results.json
```

---

## Anti-patterns — Never Do This

| Anti-pattern | Why it fails | Fix |
|---|---|---|
| No thresholds defined | No oracle = no test | Define NFRs first, then encode as thresholds |
| `sleep(0)` or no sleep | Hammers server unrealistically | Use `sleep(1 + Math.random())` |
| `open()` inside default fn | Reads file per iteration = disk thrash | Use `SharedArray` |
| Testing in production | Affects real users | Use staging with prod-like data |
| Threshold too tight for CI | Flaky gate, noise | Set thresholds to p95 NFR, not best case |
| Ignoring error rate | Tests pass while 20% fail | Always add `http_req_failed` threshold |
| No tags on requests | Can't slice metrics by endpoint | Add `tags: { endpoint: 'name' }` |

---

## Test Oracle Checklist

Before calling a performance test complete:
- [ ] Every threshold maps to a documented NFR
- [ ] Error rate threshold is defined (`http_req_failed`)
- [ ] Per-endpoint thresholds for critical paths (checkout, login, search)
- [ ] Baseline run stored — next run compares against it
- [ ] Recovery behavior validated (after stress/spike, does latency return to baseline?)

---

## Decision Tree

```
What type of test?
├── Validate normal traffic → Load test
├── Find where system breaks → Stress test
├── Validate auto-scaling → Spike test
├── Find memory leaks → Soak test (8h minimum)
└── Capacity planning → Breakpoint test

Threshold failed in CI?
├── p95 latency too high → Profile app (N+1 queries? Missing cache?)
├── Error rate too high → Check 5xx logs, circuit breakers
├── Only on spike → Check auto-scaling config, warm-up time
└── Flaky threshold → Check test data isolation, env stability

New NFR to add?
1. Document NFR in prose (p95 < Xms under Y users)
2. Encode as k6 threshold
3. Run baseline to confirm current state
4. Add to CI gate
```

---

## Commands

```bash
k6 run tests/performance/load-test.js
k6 run -e BASE_URL=https://staging.example.com tests/performance/load-test.js
k6 run --out json=results.json tests/performance/load-test.js
k6 run --vus 1 --iterations 1 tests/performance/load-test.js  # Smoke check
```

---

## Resources

- [k6 docs](https://grafana.com/docs/k6/latest/)
- [k6 thresholds](https://grafana.com/docs/k6/latest/using-k6/thresholds/)
- [ISTQB — Non-functional testing](https://www.istqb.org)
- [skills.sh source](https://skills.sh/dengineproblem/agents-monorepo/k6-load-test)
