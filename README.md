<div align="center">

<img alt="Gentle-QE banner" src="docs/assets/brand/gentle-qe-banner.png" width="820" />

<h1>Gentle-QE</h1>

<p><strong>Gentle-QE — Unified AI Ecosystem for Testing and Reliability.</strong></p>

<p>Configures your AI coding assistant (Claude Code, Cursor, OpenCode, and more) with a Senior QE/SDET persona — ISTQB, shift-left, risk-based — plus a full set of ready-to-use testing skills.</p>

<p>
<a href="https://github.com/EduardoVeraE/gentle-qe/releases"><img src="https://img.shields.io/github/v/release/EduardoVeraE/gentle-qe" alt="Release"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
<img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey" alt="Platform">
</p>

</div>

---

## What It Does

**Gentle-QE turns your AI coding agent into a Senior QA/SDET.** Instead of an assistant that only writes code, you get one that also thinks about how that code breaks: ISTQB test techniques, shift-left, risk-based prioritization, security checks, and testing workflows baked in from the start.

**Before**: "My agent writes code, but nobody's thinking about how it breaks."

**After**: Your agent thinks like a Senior QA/SDET — testing skills, security checks, and reliability workflows, ready to use.

Supports 15+ AI coding agents, including Claude Code, Cursor, OpenCode, Gemini CLI, VS Code Copilot, Codex, and Windsurf.

---

## Install

```bash
# macOS / Linux
brew tap EduardoVeraE/homebrew-tap
brew install gentle-qe

# Windows
scoop bucket add eduardoverae https://github.com/EduardoVeraE/scoop-bucket
scoop install gentle-qe
```

> **Short aliases.** On macOS/Linux (Homebrew), `gentle-qe` is also installed as `evqe` and `qe`, so every command below works under any of the three names. On Windows (Scoop) only `gentle-qe` is created; to get a short alias, add one to your PowerShell profile, e.g. `Set-Alias evqe gentle-qe`.

---

## Basic Usage

Run the installer and follow the interactive prompts to pick your AI agents and a QE preset:

```bash
gentle-qe install
```

By default it configures your agents globally. To scope the install to a single project instead:

```bash
gentle-qe install --scope=workspace
```

Once installed, restart your AI agent — the QE persona and skills are active immediately.

Check the health of your install anytime with:

```bash
gentle-qe doctor
```

---

## What You Get

- **SDET persona** — ISTQB, shift-left, risk-based thinking, instead of a generic coding assistant.
- **QE skills**, ready to use in your agent:

  | Skill | Focus |
  | --- | --- |
  | Manual testing (ISTQB) | Test cases, techniques, and templates |
  | Security / OWASP | Pentest scripts, threat models, vulnerability reports |
  | API & contract testing | REST/GraphQL functional and contract tests |
  | Playwright E2E | End-to-end testing (BDD) |
  | Accessibility | Playwright + axe-core, WCAG checks |
  | k6 load testing | Load, stress, spike, and soak scenarios |
  | Selenium (Java) | End-to-end testing with WebDriver |

- **Presets** that bundle these skills into ready-made stacks: `qe-sdet` (full stack, default), `qe-front` (E2E + accessibility), `qe-api` (API/contract testing), `qe-perf` (performance).

---

## More Info

Full usage guide, supported agents, and platform notes: see the [`docs/`](docs) folder, starting with [Usage](docs/usage.md).

---

## Credits

Gentle-QE is a QE/SDET-focused fork of **[Gentleman-Programming/gentle-ai](https://github.com/Gentleman-Programming/gentle-ai)**, maintained by [@EduardoVeraE](https://github.com/EduardoVeraE).

---

<div align="center">
<img alt="Gentle-QE" src="docs/assets/brand/gentle-qe-logo.png" width="180" />
<br/>
<a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
</div>
