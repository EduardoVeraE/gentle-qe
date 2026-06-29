# STRIDE Threat Modeling for Security Testing

A self-contained reference for QA engineers, security testers, and developers who
need to systematically discover threats in a system **before** they become
vulnerabilities. STRIDE is a structured way to ask "what can go wrong?" against
a model of the system, not the system itself.

---

## 1. Introduction

### What STRIDE is

STRIDE is a threat classification framework created at Microsoft (Loren
Kohnfelder and Praerit Garg, 1999) to help engineers enumerate threats during
design. It is a **mnemonic** for six threat categories:

| Letter | Threat                  | Violates                |
| ------ | ----------------------- | ----------------------- |
| S      | Spoofing                | Authenticity            |
| T      | Tampering               | Integrity               |
| R      | Repudiation             | Non-repudiation         |
| I      | Information disclosure  | Confidentiality         |
| D      | Denial of service       | Availability            |
| E      | Elevation of privilege  | Authorization           |

STRIDE is **not** a vulnerability scanner, a pentest, or a checklist of fixes.
It is a structured brainstorming method that turns a Data Flow Diagram (DFD)
into a list of threats, which then drive mitigations and security tests.

### When to do it

- **Early design** — before code exists, when changes are cheapest.
- **Major architecture change** — new auth system, new data store, new trust
  boundary, new third-party integration.
- **Pre-release security review** — final sweep before a feature ships.
- **Post-incident** — after a breach, re-model the affected component.
- **Compliance** — many frameworks (PCI-DSS, ISO 27001, SOC 2, HIPAA) require
  documented threat analysis.

It is **not** a one-time exercise. Threat models live alongside the system and
get updated whenever the architecture changes. Treat them like specs.

### Who participates

A good threat modeling session has three roles:

- **Architect / tech lead** — knows the design, draws the DFD.
- **Security engineer / QA** — drives the STRIDE walkthrough, asks the
  uncomfortable "what if" questions.
- **Product owner / business analyst** — knows what the assets are worth and
  which threats matter most.

Optional but valuable: SRE/ops (knows the deployment surface), a developer who
built the component (knows the implementation gotchas).

Keep the group small (3–6 people) and time-boxed (60–90 minutes per session).

---

## 2. The 6 STRIDE elements

### S — Spoofing

**Definition.** An attacker impersonates a legitimate principal — user, service,
device, or process — to gain trust they should not have.

**Security property violated.** Authenticity (proving who you are).

**Example threats.**

- Stolen credentials let an attacker log in as a user.
- A forged JWT (weak signing key, `alg:none`) is accepted by the API.
- DNS poisoning routes traffic to a malicious server impersonating the real one.
- A microservice trusts an `X-User-Id` header set by a compromised gateway.
- Session fixation: attacker forces a victim to use a session ID the attacker
  knows.

**OWASP overlap.** Web Top 10 2025 A07 (Identification and Authentication
Failures), API Top 10 2023 API2 (Broken Authentication).

### T — Tampering

**Definition.** An attacker modifies data — in transit, at rest, in memory, or
in code — to alter system behavior or corrupt records.

**Security property violated.** Integrity.

**Example threats.**

- A man-in-the-middle modifies an HTTP request (no TLS or stripped TLS).
- An attacker edits a hidden form field (`price=0.01`) before submission.
- SQL injection rewrites the query: `'; UPDATE accounts SET balance=...`.
- A malicious npm package modifies build output (supply chain).
- An attacker writes to a shared S3 bucket without integrity checks.
- Mobile app binary is patched to bypass payment validation.

**OWASP overlap.** Web Top 10 2025 A03 (Injection), A08 (Software and Data
Integrity Failures), API Top 10 2023 API3 (Broken Object Property Level
Authorization), Mobile Top 10 2024 M7 (Insufficient Binary Protections).

### R — Repudiation

**Definition.** A user (legitimate or hostile) performs an action and later
denies doing it, and the system cannot prove otherwise.

**Security property violated.** Non-repudiation (auditability).

**Example threats.**

- No audit log of who deleted a record.
- Logs exist but are mutable by the same user whose actions they record.
- Admin actions logged without timestamps or actor identity.
- Shared admin account — "admin did it" but which human?
- Logs lack signing or hashing, so they can be edited after the fact.

**OWASP overlap.** Web Top 10 2025 A09 (Security Logging and Monitoring
Failures), API Top 10 2023 API9 (Improper Inventory Management — partial),
Mobile Top 10 2024 M8 (Security Misconfiguration).

### I — Information disclosure

**Definition.** Sensitive data is exposed to a party that should not see it.

**Security property violated.** Confidentiality.

**Example threats.**

- Stack traces with DB credentials returned in 500 responses.
- API returns full user object including `password_hash` or `ssn`.
- S3 bucket misconfigured as public.
- JWT contains PII in the payload (it is base64, not encrypted).
- TLS not enforced; tokens sent over HTTP.
- Verbose error messages leak whether a username exists ("user not found" vs
  "wrong password").
- Mobile app caches auth tokens in unencrypted SharedPreferences/UserDefaults.

**OWASP overlap.** Web Top 10 2025 A02 (Cryptographic Failures), A05 (Security
Misconfiguration), API Top 10 2023 API3 (Broken Object Property Level
Authorization), Mobile Top 10 2024 M2 (Inadequate Supply Chain Security), M9
(Insecure Data Storage).

### D — Denial of service

**Definition.** An attacker makes the system unavailable to legitimate users.

**Security property violated.** Availability.

**Example threats.**

- Volumetric DDoS saturates bandwidth.
- An unauthenticated endpoint runs an O(n^2) DB query — one curl can pin CPU.
- Regex with catastrophic backtracking (ReDoS) on user input.
- File upload accepts 10 GB files and writes them to disk synchronously.
- Login endpoint has no rate limit — attacker locks out every account by
  triggering lockout.
- "Billing DoS" — attacker triggers expensive paid API calls (SMS, email,
  cloud functions) until budget is exhausted.

**OWASP overlap.** API Top 10 2023 API4 (Unrestricted Resource Consumption),
Web Top 10 2025 A05 (Security Misconfiguration — partial).

### E — Elevation of privilege

**Definition.** An attacker gains capabilities they should not have — either as
a higher-tier user (vertical) or as a different same-tier user (horizontal).

**Security property violated.** Authorization.

**Example threats.**

- IDOR: `GET /api/orders/123` returns another user's order.
- Mass assignment: `PATCH /users/me { "role": "admin" }` and the server
  accepts the field.
- A "user" JWT has its `role` claim flipped to `admin` (weak verification).
- Container escape via misconfigured Docker socket mount.
- Path traversal: `?file=../../etc/passwd`.
- A second-factor bypass — "enroll new device" endpoint works without the
  current factor.

**OWASP overlap.** Web Top 10 2025 A01 (Broken Access Control), API Top 10
2023 API1 (Broken Object Level Authorization), API5 (Broken Function Level
Authorization).

---

## 3. STRIDE workflow

A STRIDE session follows five steps. Skipping steps is the most common reason
threat models produce noise instead of signal.

### Step 1 — Define system boundaries

Decide **what is in scope** and write it down.

- Which services, endpoints, users, data stores?
- Which integrations (auth providers, payment, email)?
- Which environments (prod, staging)?
- What is **explicitly out of scope** (the OS kernel, the cloud provider's
  hypervisor, the user's device)?

Output: a one-paragraph "scope statement" pinned to the top of the document.

### Step 2 — Build a Data Flow Diagram (DFD)

A DFD has **four** elements:

| Element              | Symbol  | Examples                                     |
| -------------------- | ------- | -------------------------------------------- |
| External entity      | square  | End user, third-party API, mobile app        |
| Process              | circle  | Web server, Lambda, microservice             |
| Data store           | parallel lines | Postgres, Redis, S3 bucket, log file  |
| Data flow            | arrow   | HTTP request, DB query, message queue event  |
| Trust boundary       | dashed line | Internet ↔ DMZ, DMZ ↔ internal, user ↔ kernel |

Keep the DFD at the right zoom level. A single diagram with 50 boxes is useless.
Draw a Level-0 (context) diagram, then drill into the components that matter.

### Step 3 — Walk STRIDE on each element

For each element of the DFD, ask all six STRIDE questions. Not every threat
applies to every element — use this matrix:

```
                  S    T    R    I    D    E
External entity   X         X
Process           X    X    X    X    X    X
Data store             X    X*   X    X
Data flow              X         X    X
```

(* Repudiation on data stores when logs/audit trails live there.)

For each applicable cell, ask: **"How could an attacker do this here?"**
Capture every plausible threat — quantity over quality at this stage. You will
prune later.

### Step 4 — Rate each threat

You need a way to prioritize. Two common approaches:

**DREAD (legacy).** Score 1–10 on each axis, average them:

- **D**amage potential — how bad if it happens?
- **R**eproducibility — how easy to reproduce?
- **E**xploitability — how much skill/access required?
- **A**ffected users — how many?
- **D**iscoverability — how easy to find the vuln?

DREAD is criticized for subjectivity. Many teams use a simpler **risk matrix**:

```
              Likelihood
          Low    Medium    High
High      Med    High      Critical
Med       Low    Med       High      Impact
Low       Info   Low       Med
```

Or align with CVSS 3.1/4.0 if you already use it for vulnerability tracking.

### Step 5 — Document mitigations and tests

For every threat above your risk threshold, record:

- **Mitigation** — what control reduces the risk (input validation, MFA, rate
  limit, audit log, RBAC check, etc.)?
- **Owner** — which team/person?
- **Test** — how does QA verify the mitigation works? Unit test, integration
  test, fuzz test, manual pentest case, monitoring alert?
- **Residual risk** — what risk remains after the mitigation, and is it
  acceptable?

The output of the threat model is **not** the diagram. It is the threat list
with mitigations and test cases. That list goes into the backlog.

---

## 4. DFD notation (ASCII)

### External entity (square)

```
+----------+
|  User    |
+----------+
```

### Process (circle, drawn as parens)

```
  ( Web API )
```

### Data store (open rectangle / parallel lines)

```
=================
|  users table  |
=================
```

### Data flow (arrow with label)

```
User ---- POST /login ---->  ( Web API )
```

### Trust boundary (dashed line)

```
        Internet         |          DMZ
                         |
+----------+   HTTPS     |    ( Web API )
|  User    | ----------> |
+----------+             |
                         |
                  trust boundary
```

### Putting it together — small example

```
    Internet                |              DMZ                      |        Internal
                            |                                       |
+----------+   HTTPS    +-------------+   gRPC    +-------------+   |   ===================
|  User    | ---------> | API Gateway | --------> | Order Svc   | -+-> |  orders table   |
+----------+            +-------------+           +-------------+   |   ===================
                            |                          |            |
                            |                          v            |   ===================
                            |                    +-----------+      |   |   audit log    |
                            |                    | Logger    | -----+-> ===================
                            |                    +-----------+      |
                            |                                       |
                       trust boundary                          trust boundary
```

---

## 5. STRIDE → OWASP mapping

This table lets you cross-reference STRIDE threats against the catalogues your
organization probably already tracks. It is a guide, not a strict mapping —
real threats often span multiple categories.

```
+-------+----------------------------+----------------------------+----------------------------+
| STRIDE| Web Top 10 2025            | API Top 10 2023            | Mobile Top 10 2024         |
+-------+----------------------------+----------------------------+----------------------------+
|   S   | A07 AuthN Failures         | API2 Broken AuthN          | M3 Insecure AuthN/AuthZ    |
|       | A08 Software/Data Integrity| API8 Sec Misconfiguration  | M4 Insufficient Input/     |
|       |                            |                            |     Output Validation      |
+-------+----------------------------+----------------------------+----------------------------+
|   T   | A03 Injection              | API3 Broken Object         | M4 Insufficient Input/     |
|       | A08 Software/Data Integrity|     Property Level AuthZ   |     Output Validation      |
|       |                            | API8 Sec Misconfiguration  | M7 Insufficient Binary     |
|       |                            |                            |     Protections            |
+-------+----------------------------+----------------------------+----------------------------+
|   R   | A09 Logging & Monitoring   | API9 Improper Inventory    | M8 Security Misconfig      |
|       |    Failures                |     Management             |                            |
+-------+----------------------------+----------------------------+----------------------------+
|   I   | A02 Cryptographic Failures | API3 Broken Object         | M2 Inadequate Supply Chain |
|       | A05 Security Misconfig     |     Property Level AuthZ   | M9 Insecure Data Storage   |
|       |                            | API8 Sec Misconfiguration  | M10 Insufficient Crypto    |
+-------+----------------------------+----------------------------+----------------------------+
|   D   | A05 Security Misconfig     | API4 Unrestricted Resource | M8 Security Misconfig      |
|       |                            |     Consumption            |                            |
+-------+----------------------------+----------------------------+----------------------------+
|   E   | A01 Broken Access Control  | API1 Broken Object Level   | M1 Improper Credential     |
|       | A04 Insecure Design        |     AuthZ                  |     Usage                  |
|       |                            | API5 Broken Function       | M3 Insecure AuthN/AuthZ    |
|       |                            |     Level AuthZ            |                            |
+-------+----------------------------+----------------------------+----------------------------+
```

---

## 6. Attack trees

An attack tree is a hierarchical decomposition of a single attacker goal into
sub-goals. The root is the goal; children are alternative ways to achieve it
(OR), or required steps to combine (AND). Attack trees complement STRIDE: STRIDE
is breadth-first ("what can go wrong everywhere?"), attack trees are
depth-first ("how could THIS specific bad thing happen?").

### Example 1 — "Steal a user session"

```
Goal: Steal a user session
|
+-- (OR) Get the session token
|     |
|     +-- (OR) Read it from victim's machine
|     |     +-- Local malware reads cookies
|     |     +-- Browser extension exfiltrates cookies
|     |
|     +-- (OR) Sniff it from the network
|     |     +-- HTTP (no TLS) on public Wi-Fi
|     |     +-- Strip TLS (sslstrip)
|     |     +-- Compromise CA / forge cert
|     |
|     +-- (OR) Exploit the application
|     |     +-- XSS that exfiltrates document.cookie
|     |     +-- Token leaked in URL, captured in Referer header
|     |     +-- Token logged server-side and log access leaked
|     |
|     +-- (OR) Predict / forge it
|           +-- Weak random number generator
|           +-- JWT signed with weak key (HS256 + brute force)
|           +-- alg:none accepted
|
+-- (OR) Hijack without the token
      +-- Session fixation (force known token on victim)
      +-- CSRF + CSRF protection broken
```

Each leaf becomes a test case or a mitigation requirement.

### Example 2 — "Make a paid SMS API exhaust the company budget"

```
Goal: Exhaust SMS budget (financial DoS)
|
+-- (AND) Find an endpoint that triggers SMS
|     +-- Signup with phone verification
|     +-- "Forgot password" with SMS
|     +-- Two-factor enrollment
|
+-- (AND) Bypass rate limiting
      +-- (OR) No rate limit at all
      +-- (OR) Rate limit per session — rotate sessions
      +-- (OR) Rate limit per IP — use proxies
      +-- (OR) Rate limit per phone — iterate phone numbers
```

---

## 7. Tools

| Tool                              | Type           | Notes                                                         |
| --------------------------------- | -------------- | ------------------------------------------------------------- |
| Microsoft Threat Modeling Tool    | Desktop (Win)  | Free, official; STRIDE built-in; aging UI; Windows-only       |
| OWASP Threat Dragon               | Web / Electron | Free, open source; STRIDE + LINDDUN; integrates with GitHub   |
| IriusRisk                         | SaaS           | Commercial; auto-generates threat models from architecture    |
| pytm                              | Code (Python)  | Threat model as code; diff in PRs; renders DFDs via Graphviz  |
| ThreatSpec                        | Code (DSL)     | Embed threat assertions next to code as comments              |
| draw.io / diagrams.net            | Diagram only   | No threat enumeration; pair with a spreadsheet                |
| Excel / Google Sheets templates   | Manual         | Surprisingly common; works for small teams; STRIDE per row    |
| OWASP pytm GitHub                 | Library        | `tm.add(Threat(...))`; renders Markdown reports               |

For most teams new to threat modeling: start with **draw.io + a spreadsheet**
or **Threat Dragon**. Code-based tools (pytm, ThreatSpec) shine once the
practice is mature and threat models live in version control alongside specs.

---

## 8. Worked example — JWT-protected REST API

A small but realistic example. The system: a REST API for a notes app.

### 8.1 Scope

- In scope: the API gateway, the auth service, the notes service, the
  Postgres database, the audit log.
- Out of scope: the user's browser, the cloud provider's infrastructure, the
  TLS termination at the load balancer (assume TLS is correctly configured
  end-to-end).

### 8.2 DFD

```
                Internet                  |              App VPC                       |       Data tier
                                          |                                            |
+--------+   POST /login            +-------------+   gRPC GetUser   +-------------+   |   ===================
|  User  | -----------------------> | API Gateway | ----------------> | Auth Svc   | --+-> |   users table   |
| (SPA)  | <-- 200 { jwt }   --,    +-------------+ <----jwt sign---- +-------------+   |   ===================
+--------+                    |          |
    |   GET /notes (Bearer)   |          |   gRPC ListNotes
    +-------------------------+          +-------------------------+
                                                                  |
                                                                  v
                                                          +-------------+
                                                          | Notes Svc   | ----+-> ===================
                                                          +-------------+     |   |   notes table   |
                                                                  |           |   ===================
                                                                  +---------- + -----+
                                                                              |      |
                                                                              v      v
                                                                      ===================
                                                                      |   audit log    |
                                                                      ===================
                            trust boundary                       trust boundary
```

DFD elements:

- **External entity**: User (SPA in browser).
- **Processes**: API Gateway, Auth Svc, Notes Svc.
- **Data stores**: users table, notes table, audit log.
- **Data flows**: HTTPS from user, gRPC between services, SQL to DB.
- **Trust boundaries**: Internet ↔ App VPC, App VPC ↔ Data tier.

### 8.3 STRIDE walkthrough (selected, not exhaustive)

| # | Element             | STRIDE | Threat                                                     | Likelihood | Impact | Risk     | Mitigation                                                                                                | Test                                                       |
| - | ------------------- | ------ | ---------------------------------------------------------- | ---------- | ------ | -------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| 1 | API Gateway → User  | S      | Attacker forges JWT with `alg:none` or weak HS256 secret   | Medium     | High   | High     | Reject `alg:none`; use RS256/ES256 with rotated keys ≥ 2048-bit; verify `iss`, `aud`, `exp`               | Unit test on JWT verifier with crafted tokens; fuzz `alg`  |
| 2 | Notes Svc → notes   | E      | IDOR: `GET /notes/{id}` returns notes owned by other users | High       | High   | Critical | Server-side ownership check `WHERE user_id = $jwt.sub`; never trust client-supplied `user_id`             | Integration test: user A requests user B's note → 404      |
| 3 | API Gateway         | D      | Login endpoint has no rate limit — credential stuffing     | High       | Medium | High     | Per-IP and per-account rate limit; account lockout with backoff; CAPTCHA after N failures                 | Load test: 1000 logins/min should be throttled             |
| 4 | Auth Svc → users    | I      | Stack trace from DB error returns `password_hash` to user  | Medium     | High   | High     | Generic 500 message in prod; structured logging server-side; never serialize sensitive fields             | Force a DB error in staging; assert response body redacted |
| 5 | Notes Svc           | T      | Mass assignment: `PATCH /notes/{id} { "owner_id": "..." }` | High       | High   | Critical | Allow-list of editable fields in the serializer; reject unknown fields with 400                           | Integration test: attempt to PATCH `owner_id` → 400        |
| 6 | audit log           | R      | Admin actions logged with no actor → cannot prove who did it | Low      | High   | Medium   | Log `actor_id` (from JWT `sub`), `action`, `target`, `timestamp`, `request_id`; append-only; signed       | Manually trigger admin action; inspect log row             |
| 7 | Notes Svc → DB      | I      | JWT carries PII (email, name) in payload — leaked via logs | Medium     | Medium | Medium   | Keep JWT payload to opaque IDs; fetch PII from DB on demand; never log full JWTs (mask middle segment)    | Grep request logs for `eyJ`; assert masked                 |

### 8.4 Notes on the worked example

- We did not enumerate every cell of the STRIDE matrix — in a real session you
  would, but here we picked the highest-risk threats per element.
- Each row maps to at least one OWASP category: #1 → A07/API2, #2 → A01/API1,
  #3 → API4, #4 → A02/A05, #5 → API3, #6 → A09, #7 → A02.
- The "Test" column is what makes this exercise valuable for QA: every threat
  becomes either an automated test or a documented manual check.
- Residual risk discussion (not shown) would cover threats like physical access
  to the database server (out of scope) or 0-day in the JWT library
  (mitigated by dependency scanning + rapid patch process).

---

## Further reading

- Adam Shostack, *Threat Modeling: Designing for Security* (Wiley, 2014) —
  the canonical book.
- OWASP Threat Modeling Cheat Sheet —
  https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html
- Microsoft SDL Threat Modeling guidance.
- "Threat Modeling Manifesto" (2020) — values and principles that apply
  regardless of which framework you pick.

STRIDE is the most widely taught framework, but it is not the only one.
Alternatives worth knowing:

- **PASTA** — risk-centric, 7 stages, business-impact driven.
- **LINDDUN** — privacy-focused (linkability, identifiability, non-repudiation,
  detectability, disclosure of information, unawareness, non-compliance).
- **OCTAVE** — organizational risk; heavier weight, enterprise scale.
- **Trike** — defensive/auditing focus.

For a typical web/API/mobile system, STRIDE + an attack tree on top-priority
threats covers most of what QA needs. Pick the framework that matches the
system; do not pick a framework and force the system into it.
