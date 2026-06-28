# Security Tooling Catalog

A practical, opinionated catalog of security testing tools used during application
security QA. Each entry follows the same structure so a tester can quickly evaluate
whether the tool fits the testing stage, target, and budget.

> **Authorization first.** Many tools in this catalog actively probe systems
> (DAST, fuzzing, network recon, mobile instrumentation). NEVER run them against
> systems you do not own or have written permission to test. Tools flagged with
> **REQUIRES AUTHORIZATION** below must only run against assets explicitly in
> scope of a contract, bug bounty program, or internal pentest engagement.

## Quick reference

| Category               | Recommended OSS                        | Recommended commercial                |
| ---------------------- | -------------------------------------- | ------------------------------------- |
| DAST                   | OWASP ZAP                              | Burp Suite Pro                        |
| SAST                   | semgrep                                | SonarQube (Enterprise) / CodeQL (GHE) |
| SCA                    | Trivy / OWASP Dependency-Check         | Snyk                                  |
| Secrets detection      | gitleaks                               | GitHub Advanced Security              |
| Container scanning     | Trivy                                  | Snyk Container / Prisma Cloud         |
| IaC scanning           | Checkov                                | Snyk IaC / Prisma Cloud               |
| Auth / JWT testing     | jwt_tool                               | Burp Suite Pro (JWT extension)        |
| Mobile (Android/iOS)   | MobSF + Frida + objection              | NowSecure / Corellium                 |
| Network / Recon        | nmap + Amass + Subfinder               | Tenable Nessus                        |
| Fuzzing                | ffuf / wfuzz / AFL++                   | Mayhem / Burp Pro Intruder            |
| CI/CD orchestration    | semgrep CI + Trivy + gitleaks actions  | Snyk + GitHub Advanced Security       |

## Per-tool entry structure

Every entry below follows this format:

```
### <tool name>
- **Purpose**: 1 sentence
- **Install**: brew/npm/pip/docker command
- **Basic command**: realistic single-line invocation
- **When to use**: 1-2 sentences
- **OWASP categories addressed**: A01..A10 references
- **Output**: format (JSON, SARIF, HTML)
- **License / cost**: OSS / commercial / freemium
```

---

## 1. DAST (Dynamic Application Security Testing)

DAST tools probe a running application from the outside, like an attacker would.
They are essential for catching runtime issues that SAST cannot see (auth flows,
session handling, server misconfigurations).

> **REQUIRES AUTHORIZATION.** All DAST tools below send live traffic to the
> target. Only run against staging/test environments or assets explicitly in scope.

### OWASP ZAP

- **Purpose**: Full-featured DAST proxy and active scanner from the OWASP Foundation.
- **Install**: `brew install --cask zap` (macOS) or `docker pull zaproxy/zap-stable`
- **Basic command**: `docker run -t zaproxy/zap-stable zap-baseline.py -t https://staging.example.com -r zap-report.html`
- **When to use**: Default OSS DAST. Use the baseline scan in CI, the full scan
  during release hardening, and the desktop UI for manual exploration with the proxy.
- **OWASP categories addressed**: A01, A02, A03, A05, A07, A09
- **Output**: HTML, JSON, XML, SARIF (via `-f sarif`)
- **License / cost**: Apache 2.0, free

Common entry points (`zap-baseline.py --help` essentials):

```
zap-baseline.py -t <URL>                # passive scan, no attacks
zap-baseline.py -t <URL> -j             # AJAX spider for SPAs
zap-baseline.py -t <URL> -r report.html # write HTML report
zap-baseline.py -t <URL> -J report.json # write JSON
zap-full-scan.py  -t <URL> -r full.html # active scan; LOUD, authorized only
```

### Burp Suite Community

- **Purpose**: Intercepting proxy with manual testing tools (Repeater, Decoder, Comparer).
- **Install**: download from `https://portswigger.net/burp/communitydownload`
- **Basic command**: launch the GUI, configure browser proxy at `127.0.0.1:8080`
- **When to use**: Manual exploration of auth flows, parameter tampering,
  request replay during exploratory testing.
- **OWASP categories addressed**: A01, A03, A07
- **Output**: project file (binary), manual export of requests/responses
- **License / cost**: free (no active scanner, no Intruder rate limiting)

### Burp Suite Professional

- **Purpose**: Burp Community plus active scanner, full-speed Intruder, BApp
  store extensions, and CI integration via the Enterprise edition.
- **Install**: licensed installer from PortSwigger.
- **Basic command**: use the GUI scanner; CLI scans require Burp Suite Enterprise.
- **When to use**: Engagements that need a deep active scanner, JWT and SAML
  extensions, or reproducible commercial reports for clients.
- **OWASP categories addressed**: A01, A02, A03, A05, A07, A08, A09
- **Output**: HTML, XML, project file
- **License / cost**: commercial; per-user annual license. **REQUIRES PAID LICENSE.**

### Nuclei

- **Purpose**: Template-driven scanner for known CVEs, exposures, and misconfigurations.
- **Install**: `brew install nuclei` or `go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest`
- **Basic command**: `nuclei -u https://staging.example.com -severity high,critical -o nuclei.txt`
- **When to use**: Fast, signature-based pass against a target to catch known
  issues; great in CI for newly deployed environments.
- **OWASP categories addressed**: A05, A06, A09
- **Output**: text, JSON (`-jsonl`), SARIF (`-sarif-export`)
- **License / cost**: MIT, free

### Wapiti

- **Purpose**: Lightweight DAST scanner with modules for XSS, SQLi, command
  injection, file inclusion, and more.
- **Install**: `pip install wapiti3`
- **Basic command**: `wapiti -u https://staging.example.com -f html -o wapiti-report`
- **When to use**: Quick second opinion alongside ZAP, or when a smaller,
  scriptable scanner is preferred.
- **OWASP categories addressed**: A03, A05, A07
- **Output**: HTML, JSON, XML, TXT
- **License / cost**: GPL, free

---

## 2. SAST (Static Application Security Testing)

SAST tools analyze source code or compiled artifacts without running them. They
catch insecure patterns early (in IDE, on commit, in PR) and are the cheapest
place to fix vulnerabilities.

### semgrep

- **Purpose**: Fast, pattern-based static analysis with a large rule registry.
- **Install**: `brew install semgrep` or `pip install semgrep`
- **Basic command**: `semgrep --config p/owasp-top-ten --error --sarif --output semgrep.sarif .`
- **When to use**: Default SAST for any polyglot codebase. Run locally, in
  pre-commit, and in CI. Author custom rules for project-specific patterns.
- **OWASP categories addressed**: A01, A02, A03, A04, A07, A08
- **Output**: text, JSON, SARIF, JUnit XML
- **License / cost**: LGPL (CLI free); semgrep Cloud is commercial freemium

Common semgrep invocations:

```
semgrep --config auto .                       # auto-detect language and apply registry
semgrep --config p/owasp-top-ten .            # OWASP Top 10 ruleset
semgrep --config p/security-audit .           # broader security audit
semgrep --config ./rules/ .                   # custom rules directory
semgrep --sarif --output semgrep.sarif .      # SARIF for GitHub code scanning
semgrep --severity ERROR --error .            # fail CI only on errors
```

### CodeQL

- **Purpose**: Semantic code analysis engine from GitHub; queries the code as a
  database to find variant patterns of known vulnerabilities.
- **Install**: download CLI from `https://github.com/github/codeql-cli-binaries`
- **Basic command**: `codeql database create db --language=javascript --source-root . && codeql database analyze db --format=sarif-latest --output=codeql.sarif javascript-security-and-quality.qls`
- **When to use**: GitHub-hosted projects (free for public repos via Code Scanning)
  and engagements where deep dataflow analysis is required.
- **OWASP categories addressed**: A01, A03, A07, A08
- **Output**: SARIF, CSV
- **License / cost**: free for OSS on github.com; commercial for private repos
  (bundled with GitHub Advanced Security).

### SonarQube

- **Purpose**: Code quality and security platform with rules per language.
- **Install**: `docker run -d -p 9000:9000 sonarqube:lts-community`
- **Basic command**: `sonar-scanner -Dsonar.projectKey=myapp -Dsonar.sources=.`
- **When to use**: When the team already uses Sonar for quality and wants
  security rules in the same dashboard. The Community Edition has limited
  security rules; Developer/Enterprise add taint analysis.
- **OWASP categories addressed**: A01, A03, A05, A07
- **Output**: web dashboard, JSON via API
- **License / cost**: Community is free (LGPL); Developer/Enterprise are commercial.

### Bandit (Python)

- **Purpose**: Python-specific SAST focused on common security pitfalls.
- **Install**: `pip install bandit`
- **Basic command**: `bandit -r src/ -f json -o bandit.json`
- **When to use**: Any Python project, especially when semgrep rules are too
  generic. Pairs well with `pip-audit` for dependencies.
- **OWASP categories addressed**: A02, A03, A08
- **Output**: text, JSON, XML, HTML, SARIF
- **License / cost**: Apache 2.0, free

### gosec (Go)

- **Purpose**: Go-specific SAST scanner.
- **Install**: `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- **Basic command**: `gosec -fmt sarif -out gosec.sarif ./...`
- **When to use**: Any Go project; catches common issues like hardcoded creds,
  weak crypto, and unhandled errors with security implications.
- **OWASP categories addressed**: A02, A03, A05, A08
- **Output**: text, JSON, SARIF, JUnit XML
- **License / cost**: Apache 2.0, free

### ESLint security plugins

- **Purpose**: Catch insecure JavaScript/TypeScript patterns at lint time.
- **Install**: `npm install --save-dev eslint-plugin-security eslint-plugin-no-unsanitized`
- **Basic command**: `npx eslint --ext .js,.ts,.tsx --plugin security src/`
- **When to use**: All JS/TS projects, in IDE and pre-commit. Complement with
  semgrep for cross-language coverage.
- **OWASP categories addressed**: A03, A07
- **Output**: text, JSON, SARIF (via `--format @microsoft/eslint-formatter-sarif`)
- **License / cost**: MIT, free

---

## 3. SCA / Dependency scanning

Software Composition Analysis (SCA) finds known vulnerabilities (CVEs) in your
direct and transitive dependencies. Run on every PR and on a schedule to catch
newly disclosed CVEs.

### Snyk

- **Purpose**: Commercial SCA + SAST + IaC + Container platform.
- **Install**: `npm install -g snyk` then `snyk auth`
- **Basic command**: `snyk test --severity-threshold=high --sarif-file-output=snyk.sarif`
- **When to use**: When you want a single vendor across SCA/IaC/Container with
  a managed vulnerability database, and your org can pay per-seat.
- **OWASP categories addressed**: A06, A08, A09
- **Output**: text, JSON, SARIF, HTML (with `snyk-to-html`)
- **License / cost**: freemium; free tier limited tests/month, paid plans for teams.

### Trivy

- **Purpose**: Multi-purpose scanner: SCA, container, filesystem, IaC, secrets.
- **Install**: `brew install trivy` or `docker pull aquasec/trivy`
- **Basic command**: `trivy fs --severity HIGH,CRITICAL --format sarif -o trivy.sarif .`
- **When to use**: Default OSS SCA + container scanner; runs offline with cached
  vulnerability DB; trivial to drop into CI.
- **OWASP categories addressed**: A05, A06, A08, A09
- **Output**: table, JSON, SARIF, CycloneDX, SPDX
- **License / cost**: Apache 2.0, free

Common Trivy invocations:

```
trivy fs .                                    # scan filesystem (deps + secrets + IaC)
trivy image myorg/myapp:1.2.3                 # scan a container image
trivy config infra/                           # scan IaC (Terraform, K8s, etc.)
trivy fs --severity HIGH,CRITICAL .           # only high+ severity
trivy fs --format sarif -o trivy.sarif .      # SARIF for GitHub code scanning
trivy fs --ignore-unfixed .                   # skip vulns without fixes
trivy image --download-db-only                # pre-warm the DB in CI cache
```

### npm audit

- **Purpose**: Built-in Node.js dependency auditor.
- **Install**: ships with npm
- **Basic command**: `npm audit --omit=dev --audit-level=high --json > npm-audit.json`
- **When to use**: Any Node.js project. Pair with `npm audit fix` for automated
  upgrades, but verify with tests; transitive bumps can break.
- **OWASP categories addressed**: A06
- **Output**: text, JSON
- **License / cost**: free

### pip-audit

- **Purpose**: Python dependency auditor (PyPI Advisory Database + OSV).
- **Install**: `pip install pip-audit`
- **Basic command**: `pip-audit -r requirements.txt --format json --output pip-audit.json`
- **When to use**: Any Python project; supports requirements.txt, Pipfile.lock,
  poetry.lock, and PEP 621 pyproject.toml.
- **OWASP categories addressed**: A06
- **Output**: text, JSON, CycloneDX, markdown
- **License / cost**: Apache 2.0, free

### OWASP Dependency-Check

- **Purpose**: Long-standing SCA tool, broad ecosystem support (Java, .NET,
  Node, Python, Ruby).
- **Install**: `brew install dependency-check`
- **Basic command**: `dependency-check --scan ./ --format SARIF --out dependency-check`
- **When to use**: Polyglot or Java/.NET-heavy projects; when you need NVD-only
  data without a SaaS dependency.
- **OWASP categories addressed**: A06
- **Output**: HTML, XML, JSON, SARIF, CSV
- **License / cost**: Apache 2.0, free

### Grype

- **Purpose**: Container and filesystem SCA from Anchore (companion to Syft).
- **Install**: `brew install grype`
- **Basic command**: `grype dir:. -o sarif > grype.sarif`
- **When to use**: When you already use Syft for SBOMs, or want a focused
  vulnerability scanner that consumes CycloneDX/SPDX SBOMs.
- **OWASP categories addressed**: A06, A08
- **Output**: table, JSON, CycloneDX, SARIF
- **License / cost**: Apache 2.0, free

### Dependabot

- **Purpose**: GitHub-native automated dependency updates and security alerts.
- **Install**: enable in repo Settings; commit `.github/dependabot.yml`.
- **Basic command**: declarative config, no CLI.
- **When to use**: Any GitHub repo; lowest-friction way to keep dependencies
  patched. Combine with branch protection requiring SCA checks to pass.
- **OWASP categories addressed**: A06
- **Output**: PRs, security alerts in repo Security tab
- **License / cost**: free (public + private repos on github.com)

---

## 4. Secrets detection

Find credentials, API keys, and tokens accidentally committed to source control
or present in build artifacts.

### gitleaks

- **Purpose**: Fast secrets scanner for git repos and filesystem.
- **Install**: `brew install gitleaks`
- **Basic command**: `gitleaks detect --source . --report-format sarif --report-path gitleaks.sarif`
- **When to use**: Pre-commit hook (`gitleaks protect`) and CI on every PR; also
  for auditing the full git history of a repo before open-sourcing.
- **OWASP categories addressed**: A02, A07
- **Output**: JSON, SARIF, CSV
- **License / cost**: MIT, free

Common gitleaks invocations:

```
gitleaks detect --source .                    # scan committed history
gitleaks detect --source . --no-git           # scan filesystem (no git)
gitleaks protect --staged                     # pre-commit; only staged files
gitleaks detect --report-format sarif --report-path gl.sarif .
gitleaks detect --config .gitleaks.toml .     # use custom rules
gitleaks detect --redact .                    # mask secrets in output
```

### trufflehog

- **Purpose**: Secrets scanner with credential verification (calls APIs to
  confirm a leaked key is live).
- **Install**: `brew install trufflehog`
- **Basic command**: `trufflehog filesystem . --json > trufflehog.json`
- **When to use**: When verification matters (incident response, scope of
  exposure). Use carefully; verification means outbound API calls.
- **OWASP categories addressed**: A02, A07
- **Output**: JSON, text
- **License / cost**: AGPL, free

### detect-secrets

- **Purpose**: Yelp's secrets scanner; baseline-driven so the CI gate only
  fails on new secrets, not the entire history.
- **Install**: `pip install detect-secrets`
- **Basic command**: `detect-secrets scan > .secrets.baseline`
- **When to use**: Brownfield projects with legacy secrets you cannot remove
  immediately; the baseline lets you track and gate new ones.
- **OWASP categories addressed**: A02, A07
- **Output**: JSON baseline, text on audit
- **License / cost**: Apache 2.0, free

### GitHub secret scanning

- **Purpose**: Native GitHub feature; scans pushes against partner provider
  patterns and (with push protection) blocks commits at push time.
- **Install**: enable in repo Settings -> Code security.
- **Basic command**: declarative; no CLI.
- **When to use**: All GitHub repos, full stop. Free for public repos; part of
  GitHub Advanced Security on private repos.
- **OWASP categories addressed**: A02, A07
- **Output**: alerts in repo Security tab
- **License / cost**: free for public; commercial via GHAS for private.

---

## 5. Container / IaC scanning

Find vulnerabilities in container images, Dockerfiles, Kubernetes manifests,
Terraform, CloudFormation, and other infrastructure code.

### Trivy (containers)

- **Purpose**: Container image scanning (also covered above for SCA).
- **Install**: see SCA section.
- **Basic command**: `trivy image --severity HIGH,CRITICAL --format sarif -o trivy-image.sarif myorg/myapp:1.2.3`
- **When to use**: Every image build; gate the registry push on critical CVEs.
- **OWASP categories addressed**: A05, A06, A08
- **Output**: table, JSON, SARIF, CycloneDX
- **License / cost**: Apache 2.0, free

### Checkov

- **Purpose**: Policy-as-code scanner for Terraform, CloudFormation, Kubernetes,
  Helm, ARM, Dockerfile, and more.
- **Install**: `pip install checkov`
- **Basic command**: `checkov -d . --output sarif --output-file checkov.sarif`
- **When to use**: All IaC repos; complements Trivy with deeper policy rules
  (CIS benchmarks, AWS/GCP/Azure best practices).
- **OWASP categories addressed**: A05, A08
- **Output**: text, JSON, SARIF, JUnit XML, GitHub annotations
- **License / cost**: Apache 2.0, free (Bridgecrew/Prisma Cloud commercial).

### tfsec

- **Purpose**: Terraform-focused static analyzer (now part of Trivy).
- **Install**: `brew install tfsec` (still available standalone)
- **Basic command**: `tfsec . --format sarif --out tfsec.sarif`
- **When to use**: Terraform-only repos that want a focused, fast tool.
  Aquasec recommends migrating to `trivy config` for new projects.
- **OWASP categories addressed**: A05
- **Output**: text, JSON, SARIF, JUnit, CSV
- **License / cost**: MIT, free

### kube-bench

- **Purpose**: Runs the CIS Kubernetes Benchmark against a live cluster.
- **Install**: `brew install kube-bench` or run as a Job in-cluster.
- **Basic command**: `kube-bench run --json > kube-bench.json`
- **When to use**: After cluster provisioning and on a schedule; produces an
  auditable benchmark report for compliance.
- **OWASP categories addressed**: A05
- **Output**: text, JSON, JUnit
- **License / cost**: Apache 2.0, free

### kube-hunter

- **Purpose**: Active scanner for Kubernetes attack surface.
- **Install**: `pip install kube-hunter` (legacy; project is archived)
- **Basic command**: `kube-hunter --remote <cluster-ip>`
- **When to use**: Black-box assessment of a cluster from outside or inside.
  Project is archived; treat results as advisory and prefer kube-bench plus
  modern tools (Trivy, Falco) for ongoing monitoring.
  **REQUIRES AUTHORIZATION** — actively probes the cluster.
- **OWASP categories addressed**: A05, A08
- **Output**: text, JSON
- **License / cost**: Apache 2.0, free

---

## 6. Auth / JWT testing

Authentication and authorization bugs are some of the highest-impact findings.
These tools focus on JWT handling, OAuth flows, and authz logic.

### jwt_tool

- **Purpose**: Toolkit for testing, tampering, and brute-forcing JWTs.
- **Install**: `git clone https://github.com/ticarpi/jwt_tool && pip install -r jwt_tool/requirements.txt`
- **Basic command**: `python3 jwt_tool.py <token> -T`
- **When to use**: Manual JWT analysis: alg=none, weak HS256 secrets, kid
  injection, jwk header injection, RS->HS confusion.
- **OWASP categories addressed**: A01, A02, A07
- **Output**: text, file outputs
- **License / cost**: GPL, free

### jwt.io debugger

- **Purpose**: Browser-based JWT decoder and signature verifier.
- **Install**: web only — `https://jwt.io`
- **Basic command**: paste a token into the browser UI.
- **When to use**: Quick decode and signature check during exploratory testing.
  Do NOT paste production tokens into a third-party site.
- **OWASP categories addressed**: A02, A07
- **Output**: web UI
- **License / cost**: free hosted; the underlying libraries are OSS.

### Authz0

- **Purpose**: Authorization testing tool that builds a per-role request matrix
  and replays it to find authz bypasses.
- **Install**: `go install github.com/hahwul/authz0@latest`
- **Basic command**: `authz0 -u urls.yaml`
- **When to use**: Multi-role apps; complement manual Burp Repeater testing
  with automated request replay across roles.
- **OWASP categories addressed**: A01
- **Output**: text, JSON
- **License / cost**: MIT, free

### Postman / Newman scripts

- **Purpose**: Scriptable API request collections; easy to build per-role
  smoke suites and authz checks.
- **Install**: `npm install -g newman` (CLI runner); Postman desktop app for authoring.
- **Basic command**: `newman run authz-suite.postman_collection.json -e staging.postman_environment.json --reporters cli,json`
- **When to use**: Lightweight authz regression suites that QA can maintain;
  combine with environment files per role and a smoke run in CI.
- **OWASP categories addressed**: A01, A07
- **Output**: text, JSON, HTML, JUnit
- **License / cost**: Postman freemium; Newman is OSS (Apache 2.0).

---

## 7. Mobile

Mobile testing covers static analysis of binaries, runtime instrumentation, and
device-level inspection. **Only run instrumentation tools against devices and
apps you own or have explicit permission to test.**

### MobSF (Mobile Security Framework)

- **Purpose**: Automated SAST + DAST for Android (APK/AAB) and iOS (IPA).
- **Install**: `docker pull opensecurity/mobile-security-framework-mobsf`
- **Basic command**: `docker run -it --rm -p 8000:8000 opensecurity/mobile-security-framework-mobsf:latest`
- **When to use**: First pass on any mobile build. The dynamic analyzer
  **REQUIRES AUTHORIZATION** and a rooted/jailbroken device or emulator that you own.
- **OWASP categories addressed**: A02, A04, A05, A07, A08
- **Output**: HTML, JSON, PDF
- **License / cost**: GPL, free

Common MobSF invocations:

```
# Start the server
docker run -it --rm -p 8000:8000 opensecurity/mobile-security-framework-mobsf:latest

# Upload via API (requires API key from the UI)
curl -F "file=@app.apk" http://localhost:8000/api/v1/upload \
  -H "Authorization: <api-key>"

# Trigger static scan
curl -X POST http://localhost:8000/api/v1/scan \
  -H "Authorization: <api-key>" \
  -d "hash=<hash-from-upload>"

# Download PDF report
curl http://localhost:8000/api/v1/download_pdf \
  -H "Authorization: <api-key>" \
  -d "hash=<hash>" -o report.pdf
```

### Frida

- **Purpose**: Dynamic instrumentation toolkit; inject JS into running apps.
- **Install**: `pip install frida-tools` and push `frida-server` to the device.
- **Basic command**: `frida -U -f com.example.app -l hook.js --no-pause`
- **When to use**: Bypass certificate pinning, hook crypto calls, dump
  in-memory secrets, or trace API usage during dynamic testing.
  **REQUIRES AUTHORIZATION** and rooted/jailbroken target device.
- **OWASP categories addressed**: A02, A07, A08
- **Output**: stdout, custom JS exports
- **License / cost**: wxWindows / Apache 2.0, free

### objection

- **Purpose**: Frida-powered runtime mobile exploration with pre-built commands
  (SSL bypass, keychain dump, file system explorer).
- **Install**: `pip install objection`
- **Basic command**: `objection -g com.example.app explore`
- **When to use**: Quick triage on a rooted device without writing custom
  Frida scripts. **REQUIRES AUTHORIZATION.**
- **OWASP categories addressed**: A02, A04, A07
- **Output**: interactive REPL, file dumps
- **License / cost**: GPL, free

### jadx

- **Purpose**: Decompile Android DEX/APK to readable Java.
- **Install**: `brew install jadx`
- **Basic command**: `jadx -d out app.apk`
- **When to use**: Static review of Android apps; combine with grep for
  hardcoded URLs, secrets, and crypto misuse.
- **OWASP categories addressed**: A02, A04, A07, A08
- **Output**: Java source tree, GUI
- **License / cost**: Apache 2.0, free

### apktool

- **Purpose**: Decode and rebuild Android APKs (smali level), inspect resources
  and AndroidManifest.
- **Install**: `brew install apktool`
- **Basic command**: `apktool d app.apk -o app-decoded`
- **When to use**: Manifest review, resource inspection, repackaging for
  debugging (always against apps you own).
- **OWASP categories addressed**: A04, A05, A08
- **Output**: smali, decoded resources
- **License / cost**: Apache 2.0, free

### drozer

- **Purpose**: Android security assessment framework focused on IPC, content
  providers, and exported components.
- **Install**: see `https://github.com/WithSecureLabs/drozer`
- **Basic command**: `drozer console connect`
- **When to use**: Deep Android attack-surface analysis; **REQUIRES AUTHORIZATION**.
- **OWASP categories addressed**: A01, A04, A08
- **Output**: console output
- **License / cost**: BSD, free

### class-dump (iOS)

- **Purpose**: Extract Objective-C class declarations from a Mach-O binary.
- **Install**: `brew install class-dump`
- **Basic command**: `class-dump -H -o headers/ AppBinary`
- **When to use**: iOS reverse engineering; quick way to enumerate classes,
  selectors, and protocols. Modern Swift binaries need additional tooling
  (e.g. `Hopper`, `Ghidra`).
- **OWASP categories addressed**: A04, A08
- **Output**: header files
- **License / cost**: BSD, free

---

## 8. Network / Recon

Reconnaissance maps the attack surface. **REQUIRES AUTHORIZATION** for any
active probing of assets you do not own.

### nmap

- **Purpose**: De facto network scanner; port discovery, service detection,
  scriptable via NSE.
- **Install**: `brew install nmap`
- **Basic command**: `nmap -sV -sC -oX nmap.xml staging.example.com`
- **When to use**: Pre-engagement reconnaissance, validating that internal
  services are not exposed, and confirming firewall rules.
- **OWASP categories addressed**: A05
- **Output**: text, XML, JSON (via `--script-args` + parsers), grepable
- **License / cost**: NPSL (free for most uses).

### masscan

- **Purpose**: Internet-scale port scanner; orders of magnitude faster than
  nmap for raw port discovery.
- **Install**: `brew install masscan`
- **Basic command**: `sudo masscan -p1-65535 10.0.0.0/24 --rate=1000 -oJ masscan.json`
- **When to use**: Large IP ranges; pair with nmap for service detection on
  the discovered ports. **REQUIRES AUTHORIZATION** at scale.
- **OWASP categories addressed**: A05
- **Output**: text, JSON, XML, list
- **License / cost**: AGPL, free

### Amass

- **Purpose**: Subdomain enumeration via passive sources and active resolution.
- **Install**: `brew install amass`
- **Basic command**: `amass enum -d example.com -o amass.txt`
- **When to use**: Bug bounty recon, asset inventory for a domain, supply-chain
  surface mapping (passive mode is safe; active mode probes DNS).
- **OWASP categories addressed**: A05
- **Output**: text, JSON, GraphML
- **License / cost**: Apache 2.0, free

### Subfinder

- **Purpose**: Fast passive subdomain discovery.
- **Install**: `brew install subfinder`
- **Basic command**: `subfinder -d example.com -o subs.txt`
- **When to use**: First step of recon; pipe into `httpx` and `nuclei` for
  a baseline pass.
- **OWASP categories addressed**: A05
- **Output**: text, JSON
- **License / cost**: MIT, free

---

## 9. Fuzzing

Fuzzing sends malformed or unexpected input to find crashes, memory bugs, or
unhandled cases.

### wfuzz

- **Purpose**: Web application fuzzer for parameters, paths, and headers.
- **Install**: `pip install wfuzz`
- **Basic command**: `wfuzz -c -z file,wordlist.txt --hc 404 https://staging.example.com/FUZZ`
- **When to use**: Directory brute-forcing, parameter discovery, header
  fuzzing during manual testing. **REQUIRES AUTHORIZATION.**
- **OWASP categories addressed**: A01, A05
- **Output**: text, JSON, HTML, magictree
- **License / cost**: GPL, free

### ffuf

- **Purpose**: Modern, very fast HTTP fuzzer in Go.
- **Install**: `brew install ffuf`
- **Basic command**: `ffuf -u https://staging.example.com/FUZZ -w wordlist.txt -mc 200,301,302 -o ffuf.json -of json`
- **When to use**: Default OSS web fuzzer; replaces wfuzz for most modern use cases.
  **REQUIRES AUTHORIZATION.**
- **OWASP categories addressed**: A01, A05
- **Output**: text, JSON, CSV, HTML
- **License / cost**: MIT, free

### AFL++

- **Purpose**: Coverage-guided fuzzer for native binaries.
- **Install**: `brew install afl++`
- **Basic command**: `afl-fuzz -i corpus/ -o findings/ -- ./target @@`
- **When to use**: Fuzzing C/C++/Rust binaries (parsers, codecs, native libs).
  Long-running by nature; runs continuously, not in CI.
- **OWASP categories addressed**: A03, A08
- **Output**: crash and hang directories on disk
- **License / cost**: Apache 2.0, free

### restler

- **Purpose**: Stateful REST API fuzzer driven by an OpenAPI/Swagger spec.
- **Install**: `docker pull mcr.microsoft.com/restlerfuzzer/restler`
- **Basic command**: see project docs; runs a `compile` then `fuzz-lean` then `fuzz`.
- **When to use**: APIs with a maintained OpenAPI spec; finds dependency-order
  bugs that random fuzzers miss. **REQUIRES AUTHORIZATION.**
- **OWASP categories addressed**: A01, A03, A05
- **Output**: HTML, JSON, logs
- **License / cost**: MIT, free

---

## 10. CI/CD integration

The goal: scanners run on every PR, results land in a single dashboard
(GitHub Code Scanning, GitLab Security Dashboard), and high-severity findings
block merge. Below are minimal integrations for the most common stacks.

### GitHub Actions — minimal security CI

```yaml
name: security

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read
  security-events: write   # required to upload SARIF

jobs:
  semgrep:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: returntocorp/semgrep-action@v1
        with:
          config: p/owasp-top-ten
          generateSarif: "1"
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: semgrep.sarif

  trivy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: fs
          format: sarif
          output: trivy.sarif
          severity: HIGH,CRITICAL
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy.sarif

  gitleaks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  zap-baseline:
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    steps:
      - uses: zaproxy/action-baseline@v0.10.0
        with:
          target: https://staging.example.com
```

### GitLab CI — minimal security pipeline

```yaml
include:
  - template: Security/SAST.gitlab-ci.yml
  - template: Security/Secret-Detection.gitlab-ci.yml
  - template: Security/Dependency-Scanning.gitlab-ci.yml
  - template: Security/Container-Scanning.gitlab-ci.yml

trivy_fs:
  stage: test
  image:
    name: aquasec/trivy:latest
    entrypoint: [""]
  script:
    - trivy fs --severity HIGH,CRITICAL --exit-code 1 --format json -o gl-trivy.json .
  artifacts:
    when: always
    paths: [gl-trivy.json]
```

### Gating rules (pragmatic defaults)

- **Block merge** on:
  - Any new HIGH or CRITICAL SCA finding (Trivy / Snyk / npm audit) with a fix available.
  - Any HIGH+ semgrep / CodeQL / SonarQube finding introduced in the diff.
  - Any new gitleaks / GitHub secret-scanning hit (with push protection on, this never reaches CI).
  - DAST critical findings during the release pipeline (not on every PR — too slow).
- **Warn but do not block** on:
  - MEDIUM or informational findings.
  - HIGH+ findings without an available fix (`--ignore-unfixed`); track via issue tracker.
- **Schedule, do not gate**:
  - Full ZAP active scan, Nuclei full template run, full Amass enumeration:
    nightly or weekly against staging.
- **Always upload SARIF** to GitHub Code Scanning or GitLab Security Dashboard
  so triage happens in one place, not in raw CI logs.

### Suppressing false positives

- Prefer baselines (`detect-secrets`, semgrep `--baseline-ref`) over inline
  `# nosec` / `// nosemgrep` annotations; baselines force a deliberate review.
- Document every suppression in code review with a reason and an expiry.
- Re-evaluate suppressions during quarterly security reviews.

---

## Tools that REQUIRE explicit authorization

This is a non-exhaustive recap. Before running any of these against a target,
confirm written authorization is in place (engagement letter, bug bounty scope,
internal pentest agreement, or "I own this device" for mobile):

- **OWASP ZAP** active/full scans (passive baseline is generally safe).
- **Burp Suite Professional** active scanner and Intruder at speed.
- **Nuclei** with active/intrusive templates.
- **Wapiti** active scans.
- **kube-hunter** active mode.
- **Frida**, **objection**, **drozer** — only on devices and apps you own.
- **MobSF** dynamic analyzer — only on owned devices/emulators.
- **masscan**, **nmap** at scale, **Amass** active mode.
- **wfuzz**, **ffuf**, **restler** — any active fuzzing.
- **Burp Suite Professional** itself — REQUIRES PAID LICENSE per user.

When in doubt: do not run it. Confirm scope first, in writing.
