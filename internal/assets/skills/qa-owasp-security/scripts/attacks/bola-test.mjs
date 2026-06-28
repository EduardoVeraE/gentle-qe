#!/usr/bin/env node
// bola-test.mjs — Broken Object Level Authorization tester.
// Iterate IDs, classify responses, optional horizontal-priv-esc cross-check.
// Pure Node 18+, no external deps.
// OWASP: API1:2023 BOLA.

'use strict';

import { argv, exit, stdout, stderr } from 'node:process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const SCRIPT_NAME = 'bola-test';
const ID_MARKER = '{{ID}}';

function usage() {
  stdout.write(`Usage: bola-test.mjs --target <url-with-${ID_MARKER}> --auth-token <token> [options]

Iterate object IDs against an authenticated endpoint and classify responses.
Use --secondary-token to test horizontal privilege escalation across two
real accounts (the strongest BOLA signal).

Options:
  --target <url>             URL pattern containing the ${ID_MARKER} marker
                             e.g. https://api.example.com/users/${ID_MARKER}/profile
  --auth-token <token>       Bearer token for the primary account
  --secondary-token <token>  Bearer for a second account (horizontal priv esc)
  --id-range <a-b>           Inclusive numeric range, e.g. 1-100
  --id-list <csv>            Comma-separated explicit IDs (string or numeric)
  --rate <rps>               Max requests per second (default: 5)
  --method <verb>            HTTP method (default: GET)
  --header <n=v>             Extra header (repeatable)
  --out <dir>                Output dir (default: ./security-out/${SCRIPT_NAME}/<ts>/)
  --severity-threshold <s>   low|medium|high|critical (default: high)
  --auth-header <name>       Header to send the token in (default: Authorization)
  --auth-prefix <prefix>     Token prefix (default: "Bearer ")
  -h, --help                 Show help and exit

Example:
  bola-test.mjs --target 'https://api/users/${ID_MARKER}/profile' \\
    --auth-token \$ALICE_TOKEN --secondary-token \$BOB_TOKEN --id-range 1-50
`);
}

function parseArgs(args) {
  const opts = {
    target: '', authToken: '', secondaryToken: '', idRange: '', idList: '',
    rate: 5, method: 'GET', headers: {}, out: '', severityThreshold: 'high',
    authHeader: 'Authorization', authPrefix: 'Bearer ',
  };
  for (let i = 2; i < args.length; i++) {
    const a = args[i];
    const next = () => args[++i];
    switch (a) {
      case '-h': case '--help': usage(); exit(0); break;
      case '--target': opts.target = next(); break;
      case '--auth-token': opts.authToken = next(); break;
      case '--secondary-token': opts.secondaryToken = next(); break;
      case '--id-range': opts.idRange = next(); break;
      case '--id-list': opts.idList = next(); break;
      case '--rate': opts.rate = parseInt(next(), 10) || 5; break;
      case '--method': opts.method = next().toUpperCase(); break;
      case '--header': {
        const [k, v] = next().split('=');
        opts.headers[k] = v ?? '';
        break;
      }
      case '--out': opts.out = next(); break;
      case '--severity-threshold': opts.severityThreshold = next(); break;
      case '--auth-header': opts.authHeader = next(); break;
      case '--auth-prefix': opts.authPrefix = next(); break;
      default:
        stderr.write(`Unknown option: ${a}\n`); usage(); exit(64);
    }
  }
  return opts;
}

function expandIds(opts) {
  if (opts.idList) return opts.idList.split(',').map((s) => s.trim()).filter(Boolean);
  if (opts.idRange) {
    const [a, b] = opts.idRange.split('-').map((s) => parseInt(s, 10));
    if (Number.isNaN(a) || Number.isNaN(b) || b < a) {
      throw new Error(`Bad --id-range: ${opts.idRange}`);
    }
    const out = [];
    for (let i = a; i <= b; i++) out.push(String(i));
    return out;
  }
  throw new Error('Provide --id-range or --id-list.');
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function probe(url, init) {
  const start = Date.now();
  try {
    const res = await fetch(url, init);
    const text = await res.text();
    return {
      status: res.status, size: text.length,
      durationMs: Date.now() - start,
      bodyPreview: text.slice(0, 240), error: null,
    };
  } catch (e) {
    return { status: null, size: 0, durationMs: Date.now() - start, bodyPreview: '', error: e.message };
  }
}

function classify(probeRes) {
  if (probeRes.error) return 'ERROR';
  const s = probeRes.status;
  if (s >= 200 && s < 300) return 'ACCESSIBLE';
  if (s === 401)            return 'UNAUTHENTICATED';
  if (s === 403)            return 'DENIED';
  if (s === 404)            return 'NOT_FOUND';
  if (s >= 500)             return 'SERVER_ERROR';
  return `OTHER_${s}`;
}

function csvEscape(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
}

async function main() {
  stderr.write(`[!] AUTHORIZATION REQUIRED
[!] BOLA tests iterate object IDs against authenticated endpoints.
[!] Use only with explicit permission and accounts you control.
`);
  const opts = parseArgs(argv);
  if (!opts.target || !opts.target.includes(ID_MARKER) || !opts.authToken) {
    stderr.write(`Error: --target (with ${ID_MARKER}) and --auth-token are required.\n`);
    usage(); exit(64);
  }
  const ids = expandIds(opts);
  const ts = new Date().toISOString().replace(/[:.]/g, '-');
  const outDir = opts.out || join('./security-out', SCRIPT_NAME, ts);
  mkdirSync(outDir, { recursive: true });

  const interval = Math.max(1, Math.floor(1000 / opts.rate));
  const headersFor = (token) => ({
    ...opts.headers,
    [opts.authHeader]: `${opts.authPrefix}${token}`,
  });

  const rows = [];
  let primaryAccessible = 0;
  let crossAccountAccessible = 0;

  for (const id of ids) {
    const url = opts.target.replaceAll(ID_MARKER, encodeURIComponent(id));

    const primary = await probe(url, { method: opts.method, headers: headersFor(opts.authToken) });
    const primaryClass = classify(primary);
    if (primaryClass === 'ACCESSIBLE') primaryAccessible++;

    let secondary = null, secondaryClass = null;
    if (opts.secondaryToken) {
      await sleep(interval);
      secondary = await probe(url, { method: opts.method, headers: headersFor(opts.secondaryToken) });
      secondaryClass = classify(secondary);
      // Strongest BOLA signal: same resource accessible by both accounts.
      if (primaryClass === 'ACCESSIBLE' && secondaryClass === 'ACCESSIBLE') {
        crossAccountAccessible++;
      }
    }

    rows.push({
      id, url,
      primary: { class: primaryClass, status: primary.status, size: primary.size },
      secondary: secondary ? { class: secondaryClass, status: secondary.status, size: secondary.size } : null,
    });

    await sleep(interval);
  }

  const csvHeader = 'id,primary_class,primary_status,primary_size,secondary_class,secondary_status,secondary_size';
  const csvLines = rows.map((r) => [
    r.id, r.primary.class, r.primary.status ?? '', r.primary.size,
    r.secondary?.class ?? '', r.secondary?.status ?? '', r.secondary?.size ?? '',
  ].map(csvEscape).join(','));
  writeFileSync(join(outDir, 'matrix.csv'), [csvHeader, ...csvLines].join('\n'));

  const report = {
    target: opts.target, ids: ids.length, rate: opts.rate, timestamp: ts,
    summary: {
      primaryAccessible, crossAccountAccessible,
      hasSecondary: !!opts.secondaryToken,
    },
    rows,
  };
  writeFileSync(join(outDir, 'report.json'), JSON.stringify(report, null, 2));

  stderr.write(`---\nIDs probed: ${ids.length}\n` +
    `Primary ACCESSIBLE: ${primaryAccessible}\n` +
    `Cross-account ACCESSIBLE (BOLA): ${crossAccountAccessible}\n` +
    `Output: ${outDir}\n`);

  if (!['low', 'medium', 'high', 'critical'].includes(opts.severityThreshold)) {
    stderr.write(`Invalid --severity-threshold: ${opts.severityThreshold}\n`);
    exit(64);
  }
  // Cross-account accessible == confirmed BOLA. Primary-only accessible may
  // be legitimate (the user's own resources), so it is reported but not gating.
  if (crossAccountAccessible > 0) exit(1);
  exit(0);
}

main().catch((e) => { stderr.write(`Fatal: ${e.stack || e.message}\n`); exit(3); });
