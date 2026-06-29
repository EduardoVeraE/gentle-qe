#!/usr/bin/env node
// api_artifacts.mjs — CLI to scaffold API testing artifacts from templates.
//
// Usage:
//   node api_artifacts.mjs list
//   node api_artifacts.mjs help [<template>]
//   node api_artifacts.mjs create <template> --out <dir> [--<placeholder> <value> ...] [--strip-hints]
//   node api_artifacts.mjs --self-test
//
// Make executable:
//   chmod +x api_artifacts.mjs
//
// Pure Node.js 18+. No external dependencies. ESM only.

import { readdirSync, readFileSync, writeFileSync, existsSync, mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { dirname, join, resolve, basename, isAbsolute } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";
import { tmpdir } from "node:os";
import process from "node:process";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const TEMPLATES_DIR = resolve(__dirname, "..", "templates");

// ---------- Template descriptions (short, human-friendly) ----------
const DESCRIPTIONS = {
  "api-test-plan": "API testing plan: scope, environments, tooling, schedule, exit criteria.",
  "openapi-test-checklist": "OpenAPI/Swagger contract coverage checklist per endpoint and operation.",
  "mandatory-headers-checklist": "Required request/response headers gate (auth, tracing, security, caching).",
  "contract-test-charter": "Consumer-driven contract test charter with provider states and interactions.",
};

// ---------- Minimal argument parser ----------
// Returns { positional: [...], flags: { name: value | true } }
function parseArgs(argv) {
  const positional = [];
  const flags = {};
  let i = 0;
  while (i < argv.length) {
    const a = argv[i];
    if (a.startsWith("--")) {
      const name = a.slice(2);
      const next = argv[i + 1];
      if (next === undefined || next.startsWith("--")) {
        flags[name] = true;
        i += 1;
      } else {
        flags[name] = next;
        i += 2;
      }
    } else {
      positional.push(a);
      i += 1;
    }
  }
  return { positional, flags };
}

// ---------- Template I/O helpers ----------
function listTemplates() {
  if (!existsSync(TEMPLATES_DIR)) return [];
  return readdirSync(TEMPLATES_DIR)
    .filter((f) => f.endsWith(".md"))
    .map((f) => f.replace(/\.md$/, ""))
    .sort();
}

function templatePath(name) {
  return join(TEMPLATES_DIR, `${name}.md`);
}

function readTemplate(name) {
  const p = templatePath(name);
  if (!existsSync(p)) {
    const available = listTemplates().join(", ") || "(none)";
    console.error(`[ERROR] Unknown template: "${name}". Available: ${available}`);
    process.exit(1);
  }
  return readFileSync(p, "utf8");
}

function lineCount(content) {
  return content.split("\n").length;
}

// Parse the `<!-- Placeholders: {{a}}, {{b}}, ... -->` manifest comment.
function parsePlaceholdersFromManifest(content) {
  const match = content.match(/<!--\s*Placeholders:\s*([^>]+?)\s*-->/);
  if (!match) return [];
  const raw = match[1];
  const tokens = raw.match(/\{\{[a-z_0-9]+\}\}/g) || [];
  // dedupe in order
  const seen = new Set();
  const out = [];
  for (const t of tokens) {
    const name = t.slice(2, -2);
    if (!seen.has(name)) {
      seen.add(name);
      out.push(name);
    }
  }
  return out;
}

// Extract a `<!-- e.g., ... -->` hint that appears immediately after a placeholder
// occurrence in the template. Used by `help` to show example values.
function extractHints(content, placeholders) {
  const hints = {};
  for (const p of placeholders) {
    const re = new RegExp(`\\{\\{${p}\\}\\}\\s*<!--\\s*([^>]+?)\\s*-->`);
    const m = content.match(re);
    if (m) hints[p] = m[1].trim();
  }
  return hints;
}

// ---------- Defaults ----------
function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

function gitUserName() {
  try {
    const name = execSync("git config user.name", { stdio: ["ignore", "pipe", "ignore"] })
      .toString()
      .trim();
    return name || null;
  } catch {
    return null;
  }
}

// ---------- Slug ----------
function slugify(s) {
  return String(s)
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60) || "untitled";
}

// ---------- Path safety ----------
function safeResolveOut(outArg) {
  // Reject any segment containing ".." to avoid traversal surprises.
  const parts = outArg.split(/[\\/]+/);
  if (parts.some((p) => p === "..")) {
    console.error(`[ERROR] --out path must not contain ".." segments: ${outArg}`);
    process.exit(1);
  }
  return isAbsolute(outArg) ? outArg : resolve(process.cwd(), outArg);
}

// ---------- Commands ----------
function cmdList() {
  const templates = listTemplates();
  if (templates.length === 0) {
    console.log("(no templates found)");
    return;
  }
  for (const t of templates) {
    const content = readFileSync(templatePath(t), "utf8");
    const lc = lineCount(content);
    const desc = DESCRIPTIONS[t] || "";
    console.log(`${t} (${lc} lines) — ${desc}`);
  }
}

function generalHelp() {
  console.log(`api_artifacts — scaffold API testing artifacts from templates.

Commands:
  list                                  List available templates
  help [<template>]                     Show placeholders for a template
  create <template> --out <dir> [...]   Create a filled artifact
  --self-test                           Run internal smoke tests

Create flags:
  --out <dir>             Output directory (default: ./out)
  --<placeholder> <val>   Replace {{placeholder}} globally in the template
  --strip-hints           Remove "<!-- e.g., ... -->" hint comments from the output
  --title <text>          Used to slug the output filename

Auto-filled when not provided:
  --date     today's ISO date (YYYY-MM-DD)
  --author   git config user.name (if available)

Available templates:`);
  for (const t of listTemplates()) console.log(`  - ${t}`);
}

function cmdHelp(template) {
  if (!template) {
    generalHelp();
    return;
  }
  const content = readTemplate(template);
  const placeholders = parsePlaceholdersFromManifest(content);
  const hints = extractHints(content, placeholders);
  console.log(`Template: ${template}`);
  console.log(`File:     ${templatePath(template)}`);
  console.log(`Lines:    ${lineCount(content)}`);
  console.log(`Description: ${DESCRIPTIONS[template] || "(no description)"}`);
  console.log("");
  console.log(`Placeholders (${placeholders.length}):`);
  for (const p of placeholders) {
    const hint = hints[p] ? `  # ${hints[p]}` : "";
    console.log(`  --${p}${hint}`);
  }
  console.log("");
  console.log(`Usage:`);
  console.log(`  node api_artifacts.mjs create ${template} --out ./out \\`);
  const sample = placeholders.slice(0, 3).map((p) => `--${p} "..."`).join(" ");
  if (sample) console.log(`    ${sample} ...`);
}

function stripManifests(content) {
  // Remove the leading Skill + Placeholders comments. They are metadata for the
  // CLI, not for the rendered artifact.
  return content
    .replace(/^<!--\s*Skill:[^>]*-->\s*\n?/m, "")
    .replace(/^<!--\s*Placeholders:[^>]*-->\s*\n?/m, "");
}

function stripHints(content) {
  // Remove "<!-- e.g., ... -->" inline hints. Keep ordinary HTML comments alone.
  return content.replace(/\s*<!--\s*e\.g\.,[^>]*-->/g, "");
}

function applyReplacements(content, values) {
  let out = content;
  for (const [name, value] of Object.entries(values)) {
    const re = new RegExp(`\\{\\{${name}\\}\\}`, "g");
    out = out.replace(re, String(value));
  }
  return out;
}

function findRemainingPlaceholders(content) {
  const tokens = content.match(/\{\{[a-z_0-9]+\}\}/g) || [];
  return Array.from(new Set(tokens));
}

function cmdCreate(args) {
  const { positional, flags } = args;
  const template = positional[1];
  if (!template) {
    console.error(`[ERROR] create requires a <template> argument.`);
    console.error(`        Available: ${listTemplates().join(", ") || "(none)"}`);
    process.exit(1);
  }

  const content = readTemplate(template);
  const placeholders = parsePlaceholdersFromManifest(content);

  // Defaults
  const provided = { ...flags };
  if (!provided.date) provided.date = todayISO();
  if (!provided.author) {
    const name = gitUserName();
    if (name) provided.author = name;
  }

  // Out dir
  const outArg = typeof provided.out === "string" ? provided.out : "./out";
  const outDir = safeResolveOut(outArg);
  if (!existsSync(outDir)) mkdirSync(outDir, { recursive: true });

  // Strip hints option
  const stripHintsFlag = provided["strip-hints"] === true;

  // Build replacement map (only known placeholders + provided values; ignore CLI-only flags)
  const replacements = {};
  for (const p of placeholders) {
    if (provided[p] !== undefined && provided[p] !== true) {
      replacements[p] = provided[p];
    }
  }

  // Render
  let rendered = applyReplacements(content, replacements);
  rendered = stripManifests(rendered);
  if (stripHintsFlag) rendered = stripHints(rendered);

  // Warn about missing placeholders
  const missing = placeholders.filter((p) => !(p in replacements));
  if (missing.length > 0) {
    console.error(`[WARN] missing: ${missing.map((m) => `{{${m}}}`).join(", ")}`);
  }

  // Filename
  const titleForSlug = (provided.title && provided.title !== true) ? provided.title : "untitled";
  const filename = `${template}-${todayISO()}-${slugify(titleForSlug)}.md`;
  const outPath = join(outDir, filename);

  writeFileSync(outPath, rendered, "utf8");
  console.log(`Created: ${outPath}`);
  return outPath;
}

// ---------- Self-test ----------
function selfTest() {
  let failed = 0;
  const log = (ok, msg) => {
    if (!ok) failed += 1;
    console.log(`${ok ? "PASS" : "FAIL"}  ${msg}`);
  };

  const templates = listTemplates();
  log(templates.length === 4, `list templates → found ${templates.length} (expected 4)`);

  const tmp = mkdtempSync(join(tmpdir(), "api-artifacts-"));
  try {
    for (const t of templates) {
      const content = readTemplate(t);
      const placeholders = parsePlaceholdersFromManifest(content);
      // Build a sample value for every placeholder.
      const flags = { out: tmp, title: `selftest ${t}`, "strip-hints": true };
      for (const p of placeholders) {
        if (!(p in flags)) flags[p] = `SAMPLE_${p.toUpperCase()}`;
      }
      const argv = ["create", t];
      for (const [k, v] of Object.entries(flags)) {
        argv.push(`--${k}`);
        if (v !== true) argv.push(String(v));
      }
      const args = parseArgs(argv);
      const outPath = cmdCreate(args);
      const written = readFileSync(outPath, "utf8");
      const remaining = findRemainingPlaceholders(written);
      log(
        remaining.length === 0,
        `render ${t} → ${remaining.length === 0 ? "no placeholders left" : `LEFTOVER ${remaining.join(", ")}`}`,
      );
      // Manifests must be stripped from output
      log(
        !/^<!--\s*Skill:/m.test(written) && !/^<!--\s*Placeholders:/m.test(written),
        `manifests stripped from ${basename(outPath)}`,
      );
    }
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }

  if (failed === 0) {
    console.log("\nself-test: OK");
    process.exit(0);
  } else {
    console.log(`\nself-test: FAILED (${failed} check(s))`);
    process.exit(1);
  }
}

// ---------- Entrypoint ----------
function main() {
  const raw = process.argv.slice(2);
  if (raw.length === 0) {
    generalHelp();
    process.exit(0);
  }

  // Top-level flag: --self-test
  if (raw.includes("--self-test")) {
    selfTest();
    return;
  }

  const args = parseArgs(raw);
  const cmd = args.positional[0];

  switch (cmd) {
    case "list":
      cmdList();
      break;
    case "help":
      cmdHelp(args.positional[1]);
      break;
    case "create":
      cmdCreate(args);
      break;
    default:
      console.error(`[ERROR] Unknown command: ${cmd}`);
      generalHelp();
      process.exit(1);
  }
}

main();
