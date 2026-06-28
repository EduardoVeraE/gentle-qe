# OWASP Mobile Top 10 — 2024 Final Release

> Reference: <https://owasp.org/www-project-mobile-top-10/>
> Status: Final 2024 release (published as the "2023" risk slugs but ratified as the 2024 list).
> Audience: QA engineers and security testers validating Android (.apk/.aab) and iOS (.ipa) applications.

This document maps each OWASP Mobile Top 10 (2024) category to a concrete QA testing playbook covering both **Android** and **iOS**. Mobile threat models differ from web/API: the binary is shipped to the attacker, the OS sandbox is the primary trust boundary, and the device may be rooted/jailbroken. Treat the client as **fully untrusted** — every control on the device can be bypassed given enough time.

## OS-specific differences (read first)

| Concern | Android | iOS |
|---------|---------|-----|
| Package format | APK / AAB (zip + DEX bytecode) | IPA (zip + Mach-O native binary) |
| Reverse engineering ease | Easier — Dalvik bytecode decompiles to readable Java/Kotlin | Harder — ARM64 native code, requires Hopper/Ghidra/IDA |
| Secure storage | Keystore (hardware-backed on most devices), EncryptedSharedPreferences | Keychain (Secure Enclave on A7+) |
| Sandbox bypass prerequisite | Root (Magisk, KernelSU) | Jailbreak (palera1n, checkra1n, Dopamine) |
| Manifest / config | `AndroidManifest.xml` (binary XML inside APK) | `Info.plist`, entitlements, `embedded.mobileprovision` |
| Network config | `network_security_config.xml`, `cleartextTrafficPermitted` | ATS in `Info.plist` (`NSAppTransportSecurity`) |
| IPC surface | Activities, Services, BroadcastReceivers, ContentProviders | URL schemes, Universal Links, App Groups, XPC |
| Default debuggability | `android:debuggable` flag | Code-sign restrictions; debug requires resigning on non-jailbroken |

> **Device authorization required.** Dynamic testing (Frida, objection, Burp interception, MobSF dynamic analyzer) MUST be performed on a device you are authorized to test. Use rooted/jailbroken **lab devices** or emulators/simulators. Never run instrumentation against production apps on user devices without written authorization. Confirm scope, NDA, and rules of engagement before starting.

## Quick reference

| ID | Category | Primary risk | Static hot-spot | Dynamic hot-spot |
|----|----------|--------------|-----------------|------------------|
| M1 | Improper Credential Usage | Hardcoded secrets / insecure credential handling | Strings, resources, BuildConfig, plist | Memory dumps, network capture |
| M2 | Inadequate Supply Chain Security | Compromised SDKs, signing, build pipeline | `build.gradle`, `Podfile`, lockfiles, signatures | Network calls from third-party SDKs |
| M3 | Insecure Authentication/Authorization | Bypassable login, weak session, IDOR | Auth client code, token storage | Token replay, role escalation |
| M4 | Insufficient Input/Output Validation | Injection, XSS in WebView, deeplink abuse | WebView usage, intent filters, URL handlers | Fuzzing inputs, deeplink payloads |
| M5 | Insecure Communication | Cleartext / weak TLS / no pinning | NSC XML, ATS config, cert pinning code | TLS interception, downgrade |
| M6 | Inadequate Privacy Controls | PII leakage in logs, clipboard, backups | Log calls, analytics SDKs, backup flags | Log capture, clipboard scrape |
| M7 | Insufficient Binary Protections | Reverse engineering, tampering | Obfuscation, integrity checks, debuggable flag | Frida hooks, repackaging |
| M8 | Security Misconfiguration | Exported components, debug mode in prod | Manifest exports, Info.plist, file perms | Drozer attack surface, debug attach |
| M9 | Insecure Data Storage | Plaintext on disk, world-readable files | SharedPreferences, SQLite, plist, NSUserDefaults | File pull from sandbox, snapshot review |
| M10 | Insufficient Cryptography | Weak algorithms, hardcoded keys, bad RNG | Crypto API calls, key generation | Runtime hooking of crypto APIs |

---

## M01 — Improper Credential Usage

### Definition
Hardcoded credentials and insecure credential handling create unauthorized paths to sensitive functionality and backend systems. Includes embedded API keys, default passwords, transmitted-cleartext secrets, and credentials stored in insecure local locations.

### Common attack vectors
- **Android:**
  - API keys in `strings.xml`, `BuildConfig`, `gradle.properties`, embedded `.properties` assets.
  - OAuth client secrets in DEX (decompile with jadx).
  - Credentials in `SharedPreferences` (XML in `/data/data/<pkg>/shared_prefs/`).
  - Credentials embedded in native libs (`.so`) via `strings`.
- **iOS:**
  - Hardcoded keys in `Info.plist`, asset catalogs, embedded `.plist` resources.
  - Secrets compiled into Mach-O binary — extract with `strings` or class-dump.
  - Credentials in `NSUserDefaults` (plaintext plist).
  - URL schemes carrying credentials in deep links.

### How to test

#### Static
- **Android:**
  ```bash
  apktool d app.apk -o app_decoded
  jadx -d app_jadx app.apk
  grep -rEi 'api[_-]?key|secret|password|token|bearer|aws_|firebase' app_jadx/sources/
  strings app_decoded/lib/arm64-v8a/*.so | grep -Ei 'key|token|secret'
  ```
- **iOS:**
  ```bash
  unzip app.ipa -d app_ipa
  cd app_ipa/Payload/MyApp.app
  strings MyApp | grep -Ei 'api[_-]?key|secret|password|token|aws_|firebase'
  plutil -convert xml1 -o - Info.plist | grep -Ei 'key|secret|token'
  class-dump MyApp -o /tmp/headers
  ```
- Run MobSF: secret scanner flags hardcoded patterns in both APK and IPA.
- Trufflehog / gitleaks against decompiled sources for entropy-based detection.

#### Dynamic
- **Android (rooted):**
  ```bash
  frida -U -n com.app.x -l dump-strings.js
  # Hook KeyStore / SharedPreferences APIs to observe credential reads
  objection -g com.app.x explore
  android keystore list
  android shell_exec "run-as com.app.x cat shared_prefs/auth.xml"
  ```
- **iOS (jailbroken):**
  ```bash
  objection -g com.app.x explore
  ios keychain dump
  ios nsuserdefaults get
  frida-trace -U -n MyApp -m "-[NSURLRequest setValue:forHTTPHeaderField:]"
  ```
- Capture HTTPS via Burp/mitmproxy with user CA installed; observe `Authorization` headers and credential payloads.

### Tools
MobSF (static + dynamic), jadx, apktool, Frida, objection, drozer (Android), class-dump, Hopper, otool, plutil, trufflehog, Burp Suite, mitmproxy.

### Example commands / scripts
```bash
# MobSF API scan (server running locally)
curl -F 'file=@app.apk' http://localhost:8000/api/v1/upload -H "Authorization: $MOBSF_KEY"
curl -X POST -d "hash=$HASH&scan_type=apk" http://localhost:8000/api/v1/scan -H "Authorization: $MOBSF_KEY"

# Frida script: log every SharedPreferences.getString call
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var SP = Java.use('android.app.SharedPreferencesImpl');
  SP.getString.overload('java.lang.String', 'java.lang.String').implementation = function (k, d) {
    var v = this.getString(k, d);
    console.log('[SP] ' + k + ' = ' + v);
    return v;
  };
});
JS

# objection one-liner: dump iOS keychain
objection -g "com.app.x" explore -s "ios keychain dump --json /tmp/kc.json"
```

### What "passing" looks like
- Zero secrets recoverable via static analysis (or only revocable, scoped tokens with short TTL).
- All credentials fetched at runtime over TLS from a trusted backend after user authentication.
- No credentials in logs, crash reports, analytics, or URL parameters.
- Tokens stored exclusively in Keystore (Android) / Keychain (iOS), never in SharedPreferences/NSUserDefaults.
- Build artifacts (release APK/IPA) free of debug `.properties`, `.env`, mapping files, or test credentials.

### Mapping
- CWE-798 Use of Hard-coded Credentials
- CWE-522 Insufficiently Protected Credentials
- Related Web/API: A07:2021 Identification and Authentication Failures, API2:2023 Broken Authentication.

---

## M02 — Inadequate Supply Chain Security

### Definition
Attackers manipulate app functionality by compromising the development, build, signing, or distribution pipeline. Includes malicious or vulnerable third-party SDKs, compromised signing keys, malicious insiders, and tampered build infrastructure.

### Common attack vectors
- **Android:**
  - Vulnerable Gradle dependencies, transitive AAR libraries with embedded malware.
  - Signing key theft (`upload-keystore.jks` exfiltration) → Play Store impersonation.
  - Compromised Gradle plugins or Maven repositories.
  - Side-loaded SDK (analytics/ads) collecting PII covertly.
- **iOS:**
  - Compromised CocoaPods/SPM/Carthage dependencies.
  - Malicious Xcode (XcodeGhost-style) or build server compromise.
  - Provisioning profile / certificate theft → enterprise distribution abuse.
  - Closed-source SDKs (Mach-O `.framework`) shipped without inspection.

### How to test

#### Static
- **Android:**
  ```bash
  ./gradlew :app:dependencies --configuration releaseRuntimeClasspath > deps.txt
  # SBOM + vuln scan
  trivy fs --scanners vuln,license .
  osv-scanner --lockfile=app/build.gradle.kts
  # Verify signing
  apksigner verify --print-certs app-release.apk
  jarsigner -verify -verbose -certs app-release.apk
  ```
- **iOS:**
  ```bash
  cat Podfile.lock; swift package show-dependencies
  trivy fs --scanners vuln Podfile.lock Package.resolved
  # Verify code signature
  codesign -dv --verbose=4 MyApp.app
  codesign --verify --deep --strict MyApp.app
  security cms -D -i embedded.mobileprovision
  ```
- Inspect each third-party SDK's network behavior (manifest declarations, `Info.plist` `NSAppTransportSecurity` exceptions, declared permissions).

#### Dynamic
- Run app behind Burp/mitmproxy with **all** outbound traffic captured; classify by SNI which SDK is calling out.
- Use Frida to hook `java.net.URL.<init>` (Android) or `NSURLSession` (iOS) and log every requested host:
  ```bash
  frida -U -n com.app.x -l hook-url.js
  ```
- Compare in-the-wild app behavior against declared SDK list — undeclared callouts indicate piggybacking.

### Tools
MobSF, OWASP Dependency-Check, Trivy, OSV-Scanner, Snyk, Frida, apksigner, jarsigner, codesign, security cms, jadx, otool.

### Example commands / scripts
```bash
# OWASP Dependency-Check on Android project
dependency-check.sh --scan ./app/build/outputs --format HTML --out reports/

# iOS: list every dynamic library loaded by the binary
otool -L MyApp.app/MyApp

# Verify supply chain: extract every embedded framework and class-dump
ls MyApp.app/Frameworks/*.framework | while read fw; do
  echo "=== $fw ==="; otool -L "$fw/$(basename $fw .framework)"
done
```

### What "passing" looks like
- SBOM produced for every release; all dependencies pinned with hash verification.
- Reproducible builds; CI signs artifacts in HSM/KMS, signing key never on developer machines.
- Vulnerability scan in CI fails build on CVSS >= 7 in production deps.
- All third-party SDKs reviewed, their privacy/network behavior documented.
- Code signature verifies; provisioning profile uses correct distribution certificate; no enterprise profiles in App Store builds.

### Mapping
- CWE-1357 Reliance on Insufficiently Trustworthy Component
- CWE-829 Inclusion of Functionality from Untrusted Control Sphere
- Related: A06:2021 Vulnerable and Outdated Components.

---

## M03 — Insecure Authentication / Authorization

### Definition
Mobile-specific weaknesses in authentication (login bypass, weak PINs, replayable tokens, reliance on biometrics as sole factor) and authorization (client-side role enforcement, IDOR, missing server checks). Mobile devices' input constraints and offline modes amplify these issues.

### Common attack vectors
- **Android:**
  - Login bypassed by directly hitting backend `/api/v1/admin/*` with a captured token.
  - Auth performed only client-side; protected screens shown after a local flag flip (Frida hook).
  - Tokens with no expiry stored in SharedPreferences.
  - Biometric prompt that does not bind to a Keystore key (`BiometricPrompt.CryptoObject` missing).
- **iOS:**
  - LocalAuthentication (`evaluatePolicy`) used without binding to a Keychain item with `kSecAccessControlBiometryAny`.
  - Touch/Face ID success treated as "logged in" with no server attestation.
  - Session token in NSUserDefaults; replayed from another device.
  - URL scheme accepting commands without verifying caller.

### How to test

#### Static
- **Android:**
  - jadx the auth package; look for `if (response.isSuccessful) { showAdminMenu(); }` patterns.
  - Look at `BiometricPrompt` usage — must pass a `CryptoObject` bound to a Keystore key with `setUserAuthenticationRequired(true)`.
  - Check `OkHttp` interceptors for static or device-derived bearer tokens.
- **iOS:**
  - class-dump and Hopper the auth controller; check for client-only checks.
  - Search for `LAContext.evaluatePolicy` without subsequent Keychain access with biometric ACL.
  - Inspect Info.plist URL scheme handlers for unauthenticated commands.

#### Dynamic
```bash
# Android: bypass biometric / root detection / SSL pinning
frida -U -n com.app.x --codeshare pcipolloni/universal-android-ssl-pinning-bypass-with-frida
objection -g com.app.x explore -s "android sslpinning disable"
objection -g com.app.x explore -s "android root disable"

# iOS: same
objection -g com.app.x explore -s "ios sslpinning disable"
objection -g com.app.x explore -s "ios jailbreak disable"
frida -U -n MyApp --codeshare dki/ios-biometric-bypass
```
- With Burp in-line, replay tokens after logout — server should reject.
- Tamper with JWT (`alg: none`, swap user ID, extend `exp`).
- IDOR: change resource IDs in API calls; observe whether server validates ownership.
- Test offline mode: airplane-mode + try privileged action → should fail closed.

### Tools
Frida, objection, Burp Suite (Repeater + Intruder), MobSF, JWT_Tool, drozer, class-dump.

### Example commands / scripts
```bash
# Frida hook: force biometric callback to success on Android
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var Cb = Java.use('androidx.biometric.BiometricPrompt$AuthenticationCallback');
  Cb.onAuthenticationFailed.implementation = function () {
    console.log('[*] Suppressed onAuthenticationFailed');
  };
});
JS

# JWT tampering in Burp Repeater (bash demo)
HEADER=$(echo -n '{"alg":"none","typ":"JWT"}' | base64 | tr -d '=' | tr '/+' '_-')
PAYLOAD=$(echo -n '{"sub":"admin","exp":9999999999}' | base64 | tr -d '=' | tr '/+' '_-')
echo "$HEADER.$PAYLOAD."
```

### What "passing" looks like
- All authentication and authorization decisions enforced server-side.
- Tokens short-lived, rotated, revocable; refresh flow uses bound device key.
- No client-side admin/role flags trusted by backend.
- Biometric unlocks a Keystore/Keychain key required for token use; bypass via Frida does NOT yield API access.
- IDOR tests fail (server returns 403/404 for resources not owned by caller).
- 4-digit PINs not allowed as primary auth; offline access requires verified local integrity check.

### Mapping
- CWE-287 Improper Authentication, CWE-285 Improper Authorization, CWE-639 IDOR, CWE-307 Improper Restriction of Excessive Auth Attempts.
- Related Web/API: A01:2021 Broken Access Control, A07:2021 Auth Failures, API1:2023 BOLA, API2:2023 Broken Auth, API5:2023 BFLA.

---

## M04 — Insufficient Input/Output Validation

### Definition
Mobile apps failing to validate or sanitize data from external sources (user input, network responses, IPC messages, deep links, WebView content) enable injection attacks, XSS in embedded WebViews, command injection, and path traversal.

### Common attack vectors
- **Android:**
  - WebView with `setJavaScriptEnabled(true)` + `addJavascriptInterface` exposing app methods to attacker-controlled HTML.
  - Deep links / intent filters accepting URLs that drive privileged actions without origin checks.
  - `ContentProvider` accepting attacker-supplied selection clauses → SQL injection.
  - SQLite raw queries built via string concatenation.
- **iOS:**
  - `WKWebView` loading attacker HTML with `WKScriptMessageHandler` exposing native APIs.
  - Universal Links / URL schemes accepting commands without state verification.
  - Predicate strings built with `NSPredicate(format:)` and untrusted input.
  - `XMLParser` accepting external entities.

### How to test

#### Static
- **Android:**
  ```bash
  jadx app.apk
  grep -rE "addJavascriptInterface|loadUrl|loadData|setJavaScriptEnabled\(true\)" out/sources/
  grep -rE "rawQuery|execSQL" out/sources/
  apktool d app.apk
  grep -A3 "intent-filter" app/AndroidManifest.xml
  ```
- **iOS:**
  - class-dump for `WKWebView` configurations; look for `WKUserContentController.add(_:name:)` exposing handlers.
  - Inspect `Info.plist` `CFBundleURLTypes` and `com.apple.developer.associated-domains`.

#### Dynamic
- Fuzz every text input, deep link, intent extra, and URL scheme handler:
  ```bash
  # Android: send crafted intent
  adb shell am start -a android.intent.action.VIEW -d "myapp://action?cmd=$(payload)" com.app.x
  adb shell am start -W -n com.app.x/.MainActivity --es payload "<script>alert(1)</script>"

  # iOS Simulator: open URL scheme
  xcrun simctl openurl booted "myapp://config?server=evil.com"
  ```
- WebView XSS: intercept HTTP, replace content with `<script>JSBridge.exec('reveal()')</script>`.
- drozer (Android): probe exported components for injection.
  ```bash
  drozer console connect
  run app.provider.query content://com.app.x.provider/users --selection "1=1)"
  run app.activity.start --component com.app.x .DebugActivity --extra string cmd "id"
  ```

### Tools
drozer, MobSF, Frida, objection, Burp Suite, mitmproxy, jadx, class-dump, Hopper, ffuf (for HTTP fuzzing of mobile API), radamsa (input mutation).

### Example commands / scripts
```bash
# Fuzz WebView with mitmproxy script
mitmdump -s inject_xss.py
# inject_xss.py:
# def response(flow):
#     if "text/html" in flow.response.headers.get("content-type",""):
#         flow.response.content = flow.response.content.replace(b"</body>",
#             b"<script>window.AndroidBridge && AndroidBridge.exec('id')</script></body>")

# Intent fuzzing
for p in "../../../../etc/passwd" "%00" "<script>" "' OR 1=1--"; do
  adb shell am start -a android.intent.action.VIEW -d "myapp://x?q=$p" com.app.x
  sleep 1
done
```

### What "passing" looks like
- All inputs validated against allow-lists; reject by default.
- WebView usage minimized; if used, JavaScript disabled by default, bridges scoped to non-sensitive APIs, content is app-bundled or strict-origin.
- Parameterized queries everywhere; no string concatenation into SQL.
- Deep links idempotent and authenticated; sensitive actions require additional confirmation.
- Server validates and re-validates everything received from client.

### Mapping
- CWE-20 Improper Input Validation, CWE-79 XSS, CWE-89 SQL Injection, CWE-78 OS Command Injection, CWE-22 Path Traversal, CWE-94 Code Injection.
- Related Web/API: A03:2021 Injection, API8:2023 Security Misconfiguration.

---

## M05 — Insecure Communication

### Definition
Sensitive data traversing communication channels (mobile-to-server, mobile-to-mobile, device-to-local) without proper protection. Includes cleartext transmission, weak/deprecated TLS, accepting invalid certificates, and missing certificate pinning.

### Common attack vectors
- **Android:**
  - `usesCleartextTraffic="true"` in manifest or per-domain in `network_security_config.xml`.
  - Custom `TrustManager` that returns void on `checkServerTrusted` (accepts any cert).
  - `WebViewClient.onReceivedSslError` calling `proceed()` regardless of error.
  - Missing certificate pinning → MITM with user-installed CA.
- **iOS:**
  - `NSAllowsArbitraryLoads = true` in `Info.plist` (ATS disabled globally).
  - Per-domain `NSExceptionAllowsInsecureHTTPLoads` overrides.
  - URLSessionDelegate's `urlSession(_:didReceive:completionHandler:)` calling `.useCredential` for any challenge.
  - No pinning → MITM with installed profile CA.

### How to test

#### Static
- **Android:**
  ```bash
  apktool d app.apk
  cat app/AndroidManifest.xml | grep -E "usesCleartextTraffic|networkSecurityConfig"
  cat app/res/xml/network_security_config.xml
  jadx app.apk; grep -rE "X509TrustManager|checkServerTrusted|onReceivedSslError|HostnameVerifier"
  ```
- **iOS:**
  ```bash
  plutil -convert xml1 -o - Info.plist | grep -A20 NSAppTransportSecurity
  class-dump MyApp -o headers/; grep -rE "didReceive|serverTrust|kSecTrust" headers/
  ```

#### Dynamic
- Configure device proxy → Burp/mitmproxy. Install Burp/mitmproxy CA as **user** CA (not system).
- If interception works without bypass → no pinning. If only system CA works → app trusts system store but not user CAs (Android 7+ default).
- If neither works → pinning is in place; bypass to confirm pinning is the ONLY protection (it should be, but pinning alone isn't sufficient if the network layer is otherwise weak).
  ```bash
  # Android pinning bypass
  objection -g com.app.x explore -s "android sslpinning disable"
  frida -U -n com.app.x --codeshare pcipolloni/universal-android-ssl-pinning-bypass-with-frida

  # iOS pinning bypass
  objection -g com.app.x explore -s "ios sslpinning disable"
  ```
- Run testssl.sh against backend hostnames discovered:
  ```bash
  testssl.sh --severity HIGH api.app.com
  ```
- Try TLS downgrade (mitmproxy `--tls-version-client`) — app must refuse TLS < 1.2.

### Tools
Burp Suite, mitmproxy, Frida, objection, MobSF, testssl.sh, sslyze, Wireshark.

### Example commands / scripts
```bash
# Run mitmproxy in transparent mode and pipe to inspector
mitmweb --listen-port 8080 --ssl-insecure

# Verify HSTS / cert chain
sslyze --regular api.app.com

# Quick capture of all hosts the app talks to during a 60s session
mitmdump -w capture.flow & ; sleep 60; mitmproxy -nr capture.flow -s 'def request(f): print(f.request.host)'
```

### What "passing" looks like
- All traffic over TLS 1.2+ (prefer TLS 1.3); cleartext disabled (`usesCleartextTraffic="false"`, ATS enabled).
- Certificate pinning implemented for production hostnames (preferably leaf or intermediate, with backup pin and rotation plan).
- App refuses MITM with user-installed CA out-of-the-box.
- No PII, credentials, or session tokens transmitted over SMS/MMS/notifications/URL params.
- Backend supports only modern cipher suites; HSTS enforced.

### Mapping
- CWE-319 Cleartext Transmission of Sensitive Information, CWE-295 Improper Certificate Validation, CWE-326 Inadequate Encryption Strength, CWE-757 Selection of Less-Secure Algorithm During Negotiation.
- Related Web/API: A02:2021 Cryptographic Failures.

---

## M06 — Inadequate Privacy Controls

### Definition
Insufficient protection of Personally Identifiable Information (PII) — names, addresses, payment data, IPs, plus sensitive categories (health, religion, sexuality, politics). Vulnerability arises from over-collection, leakage in logs/clipboard/URLs/backups, and weak storage.

### Common attack vectors
- **Android:**
  - PII written to `Logcat` via `Log.d/e`; readable on debuggable builds and via `adb logcat` on USB-attached devices.
  - PII in URL query parameters → server access logs, web analytics, browser history.
  - Auto-backup (`android:allowBackup="true"`) including PII; pulled with `adb backup`.
  - PII written to clipboard (`ClipboardManager`) — readable by all foreground apps on Android < 10.
  - PII in crash reports (Firebase Crashlytics, etc.) without scrubbing.
- **iOS:**
  - PII in `NSLog`/`os_log` (visible in Console.app over USB).
  - PII in pasteboard (`UIPasteboard.general`) — accessible to other apps; iOS 14+ shows toast.
  - Backups via iTunes/Finder containing unprotected PII.
  - Snapshot images in `Library/Caches/Snapshots/` showing PII (taken when app backgrounds).
  - Analytics SDKs (Firebase, AppsFlyer, etc.) collecting more than declared.

### How to test

#### Static
- **Android:**
  ```bash
  jadx app.apk
  grep -rE "Log\.(d|e|i|v|w)|Timber\.|System\.out\.println" out/sources/
  grep -rE "ClipboardManager|setPrimaryClip" out/sources/
  apktool d app.apk; grep -E "allowBackup|fullBackupContent|hasFragileUserData" app/AndroidManifest.xml
  ```
- **iOS:**
  - class-dump → grep `NSLog`, `os_log`, `UIPasteboard`.
  - Review `Info.plist` for declared purpose strings (`NSCameraUsageDescription`, etc.) — must match actual use.
  - Review privacy manifest (`PrivacyInfo.xcprivacy`, required since 2024).

#### Dynamic
```bash
# Android: capture logs while exercising PII flows
adb logcat -c; adb logcat | grep -iE "email|phone|ssn|card|passport|@"

# Android: snapshot the file system after each action
adb shell run-as com.app.x ls -laR /data/data/com.app.x/

# iOS: stream system logs from device
idevicesyslog | grep -iE "email|phone|card"
# Or in simulator
xcrun simctl spawn booted log stream --level debug | grep MyApp

# Pull iOS app sandbox via objection
objection -g com.app.x explore -s "env"  # find Documents path
objection -g com.app.x explore -s "ios plist cat <Library/Preferences/com.app.x.plist>"
```
- Background the app and check for plaintext snapshot:
  - **iOS:** `Library/Caches/Snapshots/` — should be blank/blurred view, not actual data.
  - **Android:** confirm `FLAG_SECURE` on activities with PII (also blocks screenshots).
- Inspect outbound analytics traffic for PII not declared in privacy policy.

### Tools
MobSF, Frida, objection, idevicesyslog, ADB, ios-deploy, mitmproxy (for analytics inspection), Pidcat (Android log filter).

### Example commands / scripts
```bash
# Frida: hook NSLog / os_log on iOS
frida -U -n MyApp -l - <<'JS'
var nslog = Module.findExportByName(null, 'NSLog');
Interceptor.attach(nslog, { onEnter: function(args){
  console.log('[NSLog] ' + ObjC.Object(args[0]).toString());
}});
JS

# Frida: hook android.util.Log on Android
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var Log = Java.use('android.util.Log');
  ['d','e','i','v','w'].forEach(function (m) {
    Log[m].overload('java.lang.String','java.lang.String').implementation = function (t, msg) {
      console.log('[Log.'+m+'] '+t+': '+msg);
      return this[m](t, msg);
    };
  });
});
JS
```

### What "passing" looks like
- Data minimization enforced: only collect PII needed for the feature.
- No PII in logs (`Log`/`NSLog`/Crashlytics) — scrubbed or replaced with redaction tokens.
- Sensitive screens marked `FLAG_SECURE` (Android) / blank snapshot view on backgrounding (iOS).
- `allowBackup="false"` (or `fullBackupContent` excluding sensitive paths); `hasFragileUserData="false"` so uninstall wipes data.
- Privacy policy and Apple privacy manifest accurately describe collection, purpose, retention, third parties.
- Compliant with GDPR/CCPA/LGPD/etc. (lawful basis, consent, deletion rights).

### Mapping
- CWE-359 Exposure of Private Information, CWE-532 Insertion of Sensitive Information into Log File, CWE-200 Exposure of Sensitive Information.
- Related: GDPR Art. 5/25/32, CCPA, A02:2021 Cryptographic Failures (data exposure).

---

## M07 — Insufficient Binary Protections

### Definition
Lack of measures to prevent reverse engineering and code tampering of the app binary. Two attack categories:
1. **Reverse engineering** — extracting secrets, algorithms, business logic, AI models.
2. **Code tampering** — modifying the binary to bypass paywalls/license checks/security checks, or redistributing trojanized versions.

### Common attack vectors
- **Android:**
  - Unminified, unobfuscated DEX → jadx produces near-original Java/Kotlin.
  - No integrity check → repackaged APK runs identically (rebuild with apktool, re-sign with own key).
  - No anti-debug → attach Frida/`gdb`/`lldb` freely; no detection of `android_dlopen_ext` hooks or `gum-js-loop` thread.
  - No root detection or trivial check (`/system/bin/su` exists?) bypassed by Magisk DenyList.
- **iOS:**
  - Mach-O in App Store is encrypted (FairPlay) on device but easily decrypted on jailbroken (`frida-ios-dump`, `bagbak`).
  - No anti-debug (no `ptrace(PT_DENY_ATTACH)` or `sysctl` checks).
  - No jailbreak detection or trivial path checks (`/Applications/Cydia.app`).
  - No anti-Frida (no detection of `gum-js-loop` thread, `frida-server` port, suspicious dyld images).

### How to test

#### Static
- **Android:**
  ```bash
  # Decompile and assess readability
  jadx -d out app.apk
  # Look for control-flow obfuscation, string encryption, name mangling
  apktool d app.apk; ls app_decoded/smali*  # multiple smali_classes2/3 = multidex but not obfuscation
  # Check ProGuard/R8 mapping
  unzip -l app.aab | grep mapping.txt
  # Check debuggable
  aapt dump badging app.apk | grep -i debug
  ```
- **iOS:**
  ```bash
  otool -hv MyApp.app/MyApp | grep -i pie     # PIE enabled?
  otool -Iv MyApp.app/MyApp | grep stack_chk  # stack canaries?
  otool -hv MyApp.app/MyApp | grep -i encrypted   # cryptid 1 = FairPlay encrypted
  jtool2 --sig MyApp                          # signature/entitlements
  ```

#### Dynamic
- **Android:**
  ```bash
  # Repackage test
  apktool d app.apk -o work && apktool b work -o tampered.apk
  apksigner sign --ks my.keystore tampered.apk
  adb install tampered.apk   # if it runs, no integrity check

  # Frida attach test
  frida-ps -U | grep com.app.x
  frida -U -n com.app.x   # if it crashes/exits, anti-Frida present
  objection -g com.app.x explore -s "android root disable"
  ```
- **iOS:**
  ```bash
  # Decrypt + dump
  frida-ios-dump -H 192.168.1.10  # or bagbak
  # Anti-debug test
  lldb -n MyApp; (lldb) attach --name MyApp
  # Jailbreak detection bypass
  objection -g MyApp explore -s "ios jailbreak disable"
  ```

### Anti-tamper, anti-debug, obfuscation specifics
- **Obfuscation (defense vs reverse engineering):**
  - Android: ProGuard/R8 (name mangling), DexGuard / commercial (control-flow flattening, string encryption), native (NDK + LLVM-Obfuscator).
  - iOS: Swift compiler does some name mangling; commercial tools (iXGuard, Promon SHIELD) for control-flow obfuscation, string encryption, native code conversion.
- **Anti-tamper (defense vs code modification):**
  - Self-signing check: app verifies its own APK/IPA signing certificate matches expected hash; runtime exits/reports if mismatch.
  - Resource integrity: hash critical assets at build, verify at runtime.
  - Server-side attestation: SafetyNet/Play Integrity API (Android), DeviceCheck/App Attest (iOS) — server validates token before granting features.
- **Anti-debug:**
  - Android: detect `TracerPid != 0` in `/proc/self/status`; check for `gum-js-loop`/`gmain` threads; detect `frida-server` listening on port 27042.
  - iOS: `ptrace(PT_DENY_ATTACH, 0, 0, 0)`; `sysctl(KERN_PROC_PID, ..., kp_proc.p_flag & P_TRACED)`; detect debugger via `getppid() != 1`.
- **Anti-hooking:** scan loaded modules (`/proc/self/maps` on Android, `_dyld_image_count` on iOS) for `frida`, `substrate`, `cynject`.
- **Defense in depth:** assume any single check is bypassable; layer multiple checks, do them in native code, randomize timing, fail late and obscurely (don't `exit()` immediately — corrupt state and report to backend).

### Tools
jadx, apktool, MobSF, Frida, objection, otool, Hopper, Ghidra, IDA Pro, class-dump, jtool2, frida-ios-dump, bagbak, RMS-Runtime-Mobile-Security.

### Example commands / scripts
```bash
# Repackage attack on Android
apktool d app.apk -o work
# edit smali to skip license check, then rebuild
apktool b work -o tampered.apk
zipalign -p 4 tampered.apk tampered-aligned.apk
apksigner sign --ks evil.keystore tampered-aligned.apk
adb install -r tampered-aligned.apk

# Frida bypass of common Android root detection
frida -U -n com.app.x --codeshare dzonerzy/fridantiroot

# iOS: dump decrypted IPA from jailbroken device
frida-ios-dump -H 192.168.1.10 -o MyApp.ipa "MyApp"
```

### What "passing" looks like
- Android release uses R8 with full mode; commercial obfuscation if app holds high-value assets.
- iOS release uses Swift, stripped symbols, no debug strings; commercial obfuscation if needed.
- Repackaged/resigned APK fails integrity check and is reported to backend.
- Frida/debugger attachment is detected; app degrades or exits with backend report.
- Server-side attestation (Play Integrity / App Attest) gates sensitive features so even a fully tampered client cannot abuse the backend.
- High-value secrets are NOT in the binary at all (fetched at runtime, scoped, revocable).

### Mapping
- CWE-693 Protection Mechanism Failure, CWE-489 Active Debug Code, CWE-656 Reliance on Security Through Obscurity (when used alone), CWE-1278 Missing Protection Against Hardware Reverse Engineering.

---

## M08 — Security Misconfiguration

### Definition
Improper setup of platform security settings, permissions, IPC exports, file permissions, debug flags, and default credentials enabling vulnerabilities and unauthorized access.

### Common attack vectors
- **Android:**
  - `android:debuggable="true"` left in release → attach `jdb`/Frida without root.
  - Exported activities, services, broadcast receivers, content providers (`exported="true"` without permission).
  - World-readable/writable files (mode `0666`); `MODE_WORLD_READABLE` SharedPreferences.
  - `allowBackup="true"` enabling `adb backup` extraction.
  - Custom permissions with `protectionLevel="normal"` instead of `signature`.
- **iOS:**
  - `get-task-allow=true` entitlement in production (allows debugger attach).
  - URL schemes claimed without verifying source app (`UIApplication.openURL`).
  - Overly permissive App Group / Keychain Access Group.
  - Missing `NSAppTransportSecurity`, `NSCameraUsageDescription`, etc.
  - Background modes not actually needed but declared (privacy red flag).

### How to test

#### Static
- **Android:**
  ```bash
  apktool d app.apk
  # Manifest review
  grep -E "exported|debuggable|allowBackup|protectionLevel|usesCleartextTraffic" app/AndroidManifest.xml
  # File mode usage in code
  jadx app.apk; grep -rE "MODE_WORLD_READABLE|MODE_WORLD_WRITEABLE|setReadable\(true, false\)" out/
  ```
- **iOS:**
  ```bash
  codesign -d --entitlements :- MyApp.app | plutil -convert xml1 -o - -
  plutil -convert xml1 -o - Info.plist
  # Look for: get-task-allow, com.apple.security.application-groups, keychain-access-groups, URL types
  ```

#### Dynamic
- **Android (drozer):**
  ```bash
  drozer console connect
  run app.package.attacksurface com.app.x          # lists exposed components
  run app.activity.info -a com.app.x               # exported activities
  run app.provider.info -a com.app.x               # content providers
  run app.provider.finduri com.app.x
  run scanner.provider.injection -a com.app.x
  run scanner.provider.traversal -a com.app.x
  run app.broadcast.send --component com.app.x .Receiver --extra string cmd "id"
  ```
- Try attaching debugger to release build:
  ```bash
  adb shell ps | grep com.app.x
  adb forward tcp:8700 jdwp:<pid>   # only succeeds if debuggable
  ```
- iOS: try installing on jailbroken device and check `entitlements`; check whether app launches with `lldb` attach.

### Tools
drozer, MobSF, apktool, Frida, objection, codesign, jtool2, ADB, RMS.

### Example commands / scripts
```bash
# Quick scan: dump all exported components
aapt dump xmltree app.apk AndroidManifest.xml | \
  awk '/E: (activity|service|receiver|provider)/,/E: \//' | \
  grep -E "name=|exported="

# iOS: list every URL scheme handled
plutil -extract CFBundleURLTypes xml1 -o - Info.plist

# Android: pull backup if allowed
adb backup -apk -noshared com.app.x -f backup.ab
dd if=backup.ab bs=1 skip=24 | openssl zlib -d | tar -xvf -
```

### What "passing" looks like
- Release builds: `debuggable=false`, `allowBackup=false`, no `MODE_WORLD_*`.
- Only intentionally public components exported, with custom permissions at `signature` level.
- Content providers protected with `android:permission` and `grantUriPermissions` granular.
- iOS: `get-task-allow=false`, minimal entitlements, App Groups scoped, accurate purpose strings.
- File permissions: app-private (0600) on sensitive files; SQLite databases inside `getFilesDir()` not external storage.
- Secure defaults applied; reviewed by checklist in CI (e.g., MobSF score gating).

### Mapping
- CWE-16 Configuration, CWE-732 Incorrect Permission Assignment, CWE-276 Incorrect Default Permissions, CWE-489 Active Debug Code, CWE-926 Improper Export of Android Application Components.
- Related Web/API: A05:2021 Security Misconfiguration, API8:2023 Security Misconfiguration.

---

## M09 — Insecure Data Storage

### Definition
Mobile apps failing to adequately protect sensitive data stored on the device — credentials, tokens, PII, payment data, encryption keys — through unencrypted storage, world-readable files, plaintext databases, unprotected caches, or insecure backups.

### Common attack vectors
- **Android storage surfaces (in priority order of risk):**
  - **SharedPreferences** (`/data/data/<pkg>/shared_prefs/*.xml`) — plaintext XML by default.
  - **SQLite** (`/data/data/<pkg>/databases/*.db`) — plaintext unless using SQLCipher.
  - **Internal files** (`/data/data/<pkg>/files/`) — sandboxed but plaintext.
  - **External storage** (`/sdcard/`) — world-readable historically; scoped storage on Android 10+.
  - **Cache** (`/data/data/<pkg>/cache/`, WebView cache) — often forgotten.
  - **Logs / crash reports** — Crashlytics, Logcat.
  - **Keystore** — secure if used correctly; misuse (extracting keys) negates benefits.
- **iOS storage surfaces:**
  - **Keychain** — correct location for secrets; verify accessibility class (`kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` for highest).
  - **NSUserDefaults** — plist in `Library/Preferences/`, plaintext.
  - **Core Data / SQLite** — plaintext unless encrypted.
  - **Documents/Library/Caches** — plaintext.
  - **iCloud / iTunes backup** — controlled via `NSURLIsExcludedFromBackupKey` and Data Protection class.
  - **Snapshots** in `Library/Caches/Snapshots/` (taken on backgrounding).

### How to test

#### Static
- **Android:**
  ```bash
  jadx app.apk
  grep -rE "getSharedPreferences|openOrCreateDatabase|MODE_PRIVATE|MODE_WORLD_READABLE" out/sources/
  grep -rE "SQLCipher|SQLiteDatabase\.openDatabase|Cipher\.getInstance" out/sources/
  grep -rE "EncryptedSharedPreferences|MasterKey" out/sources/    # Jetpack Security
  grep -rE "AndroidKeyStore|KeyGenParameterSpec" out/sources/
  ```
- **iOS:**
  ```bash
  class-dump MyApp -o headers/
  grep -rE "NSUserDefaults|UserDefaults" headers/
  grep -rE "kSecClass|SecItemAdd|kSecAttrAccessible" headers/
  grep -rE "FileProtectionType|completeUntilFirstUserAuthentication|complete" headers/
  ```

#### Dynamic
- **Android (root):**
  ```bash
  adb shell run-as com.app.x ls -laR
  adb shell run-as com.app.x cat shared_prefs/auth.xml
  adb shell run-as com.app.x sqlite3 databases/app.db ".tables" ".dump"
  objection -g com.app.x explore -s "android keystore list"
  objection -g com.app.x explore -s "android shell_exec 'cat shared_prefs/*.xml'"
  ```
- **iOS (jailbreak):**
  ```bash
  objection -g MyApp explore
  ios keychain dump --json /tmp/kc.json
  ios nsuserdefaults get
  ios plist cat Library/Preferences/com.app.x.plist
  ios cookies get
  env  # find Documents path; pull files via scp from /var/mobile/Containers/Data/Application/<UUID>/
  ```
- Background the app, then extract `Library/Caches/Snapshots/<bundle-id>/` — must not contain sensitive screen content.
- Pull a backup (`adb backup` on Android, iTunes encrypted backup on iOS) and verify sensitive data is excluded or encrypted at rest.

### Tools
objection, Frida, jadx, MobSF, ADB, sqlite3, SQLCipher tools, plutil, ios-deploy, iFunBox, RMS.

### Example commands / scripts
```bash
# Frida: hook KeyStore on Android to log every key access
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var KS = Java.use('java.security.KeyStore');
  KS.getKey.implementation = function (alias, pwd) {
    console.log('[Keystore.getKey] alias=' + alias);
    return this.getKey(alias, pwd);
  };
});
JS

# Frida: hook iOS Keychain
frida -U -n MyApp -l - <<'JS'
var SecItemAdd = Module.findExportByName('Security', 'SecItemAdd');
Interceptor.attach(SecItemAdd, { onEnter: function (args) {
  console.log('[SecItemAdd] ' + ObjC.Object(args[0]).toString());
}});
JS

# Objection one-liners
objection -g com.app.x explore -s "android hooking watch class_method 'androidx.security.crypto.EncryptedSharedPreferences\$Editor.putString'"
objection -g MyApp     explore -s "ios hooking watch method '-[NSUserDefaults setObject:forKey:]'"
```

### What "passing" looks like
- **Android:** secrets in Keystore (with `setUserAuthenticationRequired` for highest sensitivity); structured data in EncryptedSharedPreferences; databases via SQLCipher with key derived from Keystore-protected secret. No PII on external storage.
- **iOS:** secrets in Keychain with appropriate accessibility class; files protected with `NSFileProtectionComplete` or `CompleteUnlessOpen`; sensitive UI not captured in app snapshot (set blank/blurred view in `applicationDidEnterBackground`).
- Backups exclude sensitive files; iCloud sync disabled for sensitive items.
- No sensitive data in logs, crash reports, WebView cache, or HTTP cache.
- Wipe-on-uninstall behavior validated (`hasFragileUserData=false` on Android).

### Mapping
- CWE-312 Cleartext Storage of Sensitive Information, CWE-922 Insecure Storage of Sensitive Information, CWE-359 Exposure of Private Information, CWE-200 Information Exposure.
- Related Web/API: A02:2021 Cryptographic Failures.

---

## M10 — Insufficient Cryptography

### Definition
Use of weak or improperly implemented cryptography that fails to protect confidentiality and integrity. Includes deprecated algorithms (MD5, SHA-1, DES, RC4), short keys, ECB mode, hardcoded keys/IVs, predictable RNGs, missing salting, and broken protocol implementations.

### Common attack vectors
- **Android:**
  - `Cipher.getInstance("AES")` — defaults to AES/ECB/PKCS5Padding (ECB is deterministic and leaks structure).
  - Hardcoded keys/IVs in source.
  - `new Random()` for crypto (use `SecureRandom`).
  - MD5/SHA-1 for password hashing (use Argon2/PBKDF2/bcrypt/scrypt).
  - Custom "encryption" (XOR, Base64).
- **iOS:**
  - `CommonCrypto` with hardcoded key.
  - Same ECB mode pitfalls.
  - `arc4random()` is fine for non-crypto, but bespoke crypto code likely misuses it.
  - Custom rolling-your-own AES wrappers without authenticated mode (use AES-GCM or ChaCha20-Poly1305).

### How to test

#### Static
- **Android:**
  ```bash
  jadx app.apk
  grep -rE 'Cipher\.getInstance\("[^"]+"\)' out/sources/
  grep -rE '"(DES|RC4|MD5|SHA-?1|AES/ECB|RSA/ECB)' out/sources/
  grep -rE '\bnew Random\b|\.nextBytes\(' out/sources/
  grep -rE 'KeyGenerator|SecretKeySpec|IvParameterSpec' out/sources/
  ```
- **iOS:**
  ```bash
  class-dump MyApp -o headers/
  grep -rE "kCCAlgorithmAES|kCCOptionECBMode|CC_MD5|CC_SHA1" headers/
  otool -Iv MyApp | grep -E "CCCrypt|SecKeyCreate|arc4random"
  ```

#### Dynamic
```bash
# Frida: log every Cipher.getInstance to see modes/algorithms in use
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var C = Java.use('javax.crypto.Cipher');
  C.getInstance.overload('java.lang.String').implementation = function (t) {
    console.log('[Cipher] ' + t);
    return this.getInstance(t);
  };
});
JS

# iOS: hook CCCrypt
frida -U -n MyApp -l - <<'JS'
var p = Module.findExportByName('libcommonCrypto.dylib', 'CCCrypt');
Interceptor.attach(p, { onEnter: function (args) {
  console.log('[CCCrypt] op=' + args[0] + ' alg=' + args[1] + ' mode=' + args[2]);
}});
JS
```
- Tamper-test ciphertext: flip a byte in stored ciphertext, observe whether app accepts it (no MAC = no integrity).
- Capture multiple ciphertexts of identical plaintext: identical output = ECB mode.

### Tools
Frida, objection, MobSF, jadx, class-dump, Hopper, Burp Suite, custom Python (Crypto.Cipher) for ciphertext analysis.

### Example commands / scripts
```bash
# Dump key material via Frida (Android)
frida -U -n com.app.x -l - <<'JS'
Java.perform(function () {
  var SKS = Java.use('javax.crypto.spec.SecretKeySpec');
  SKS.$init.overload('[B','java.lang.String').implementation = function (b, alg) {
    var bytes = '';
    for (var i = 0; i < b.length; i++) bytes += ('0'+(b[i]&0xff).toString(16)).slice(-2);
    console.log('[SecretKeySpec] alg=' + alg + ' key=' + bytes);
    return this.$init(b, alg);
  };
});
JS

# Detect ECB by encrypting two identical blocks and comparing
echo -n "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" > /tmp/plain
# trigger app to encrypt; observe whether two 16-byte blocks of ciphertext are identical
```

### What "passing" looks like
- AES-256-GCM (or ChaCha20-Poly1305) for symmetric encryption with random 96-bit nonce per message.
- RSA-OAEP-2048+ or ECDH P-256+ / X25519 for asymmetric.
- Argon2id (preferred) or PBKDF2-SHA256 / bcrypt / scrypt for password hashing, with per-user random salts.
- Keys generated by Keystore/Keychain (hardware-backed); never `new Random()` or hardcoded.
- All cryptographic outputs include integrity (authenticated encryption or HMAC).
- TLS 1.2+ with modern cipher suites only; no fallback to deprecated protocols.
- Compliance with NIST SP 800-131A and platform recommendations (Android Crypto API guidelines, Apple CryptoKit best practices).

### Mapping
- CWE-327 Use of a Broken or Risky Cryptographic Algorithm, CWE-326 Inadequate Encryption Strength, CWE-330 Use of Insufficiently Random Values, CWE-338 Use of Cryptographically Weak PRNG, CWE-759 Use of a One-Way Hash without a Salt, CWE-916 Use of Password Hash With Insufficient Computational Effort.
- Related Web/API: A02:2021 Cryptographic Failures.

---

## Appendix — Recommended toolbelt

| Tool | Platform | Type | Use |
|------|----------|------|-----|
| MobSF | Both | Static + Dynamic | One-shot scan, API for CI |
| jadx | Android | Static | DEX → Java decompilation |
| apktool | Android | Static | APK unpack/repack, manifest decoding |
| drozer | Android | Dynamic | Attack surface enumeration, IPC fuzzing |
| Frida | Both | Dynamic | Runtime instrumentation, hooks |
| objection | Both | Dynamic | Frida-powered swiss army knife |
| class-dump | iOS | Static | Objective-C header extraction |
| Hopper / Ghidra / IDA | Both | Static | Native binary disassembly |
| otool | iOS | Static | Mach-O inspection |
| Burp Suite | Both | Dynamic | HTTPS interception, replay |
| mitmproxy | Both | Dynamic | Scriptable proxy |
| testssl.sh / sslyze | Backend | Static-ish | TLS configuration audit |
| Trivy / OSV-Scanner | Both | Static | Dependency CVE scan |
| frida-ios-dump / bagbak | iOS | Dynamic | Decrypt FairPlay-protected IPA |

## Appendix — Required environment

- Android: rooted lab device or emulator (AVD with Magisk, Genymotion, Corellium); ADB; Android Studio.
- iOS: jailbroken lab device (palera1n / checkra1n / Dopamine compatible) OR Corellium virtual device; macOS host with Xcode, libimobiledevice, ios-deploy.
- Network: mitmproxy or Burp on a host the device can reach; ability to install custom CA.
- **Authorization:** signed scope/RoE before any dynamic test. Production apps and user-owned devices are NEVER in scope without explicit written consent.

## References
- OWASP Mobile Top 10 (2024 final): <https://owasp.org/www-project-mobile-top-10/>
- OWASP MASVS / MASTG: <https://mas.owasp.org/>
- Android security best practices: <https://developer.android.com/topic/security/best-practices>
- Apple Platform Security: <https://support.apple.com/guide/security/welcome/web>
- NIST SP 800-163 (vetting mobile applications), SP 800-131A (crypto transitions).
