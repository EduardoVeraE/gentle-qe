# Contrato de comportamiento — overlay Gentle-QE

Este documento enumera todas las divergencias de comportamiento conocidas del
fork **Gentle-QE** respecto al upstream `gentleman-programming/gentle-ai`. Es
el "contrato" de qué hace distinto el fork: cada entrada indica QUÉ cambia,
DÓNDE vive el mecanismo, y CÓMO se verifica con tests.

Fuente de verdad estructural: `tools/qe-overlay/overlay.json` (lista todos los
`overlayFiles`, `netNewDirs`, `brandingAnchors` e `inlineAnchors` que
`qe-overlay verify` vigila). Este doc explica el *comportamiento* que esos
anclajes producen; `overlay.json` es la lista mecánica de *dónde* viven.

Filosofía compartida por todas las divergencias: **cero edición de contenido
upstream**. El fork solo agrega archivos net-new (`*_qe.go`, assets nuevos) y
anclajes de una línea (`inlineAnchors`) en archivos upstream, para que un
merge del upstream nunca genere conflictos y el overlay sea re-aplicable.

---

## 1. Branding

**Qué**: nombre de producto `gentle-qe`, marca visible `Gentle-QE`, owner de
GitHub `EduardoVeraE` (no el owner upstream), repo `gentle-qe`, directorio de
estado `~/.gentle-qe`, User-Agent `gentle-qe-update-check` para el chequeo de
actualizaciones. El **module path Go no se rebrandea**
(`github.com/gentleman-programming/gentle-ai` se mantiene a propósito — tocarlo
rompería el ecosistema de imports).

**Dónde**:
- `internal/branding/branding.go` — archivo overlay completo, constantes
  `Product`, `Display`, `Owner`, `Repo`, `StateDir`, `UserAgent`.
- 11 `brandingAnchors` en `overlay.json` fuerzan a que estos archivos upstream
  usen `branding.*` en vez de literales: `internal/update/registry.go`,
  `internal/update/check.go`, `internal/update/advisory.go`,
  `internal/update/instructions.go`, `internal/update/upgrade/strategy.go`,
  `internal/app/app.go`, `internal/app/help.go`,
  `internal/tui/styles/styles.go`, `internal/cli/doctor.go`.

**Cómo se verifica**: `tools/qe-overlay verify` — `mustContain` sobre cada
anchor de marca, más la sección `brandLeak.forbidden: ["gentle-qa",
"gentle_qa"]` que persigue la marca vieja del fork (renombrado histórico),
no "gentle-ai".

---

## 2. Persona SDET por defecto + neutralización de voseo

**Qué**: el fork instala una única persona, `PersonaSDET` — perfil "Senior
SDET / QE, ISTQB certified" (filosofía ISTQB, pesticide paradox, test pyramid,
defect clustering, etc.), compuesto de `persona-sdet.md` más un slot opcional
`lineamientos-personales.md` (vacío por defecto, nunca tocado por upstream,
editable libremente por el maintainer sin conflictos de sync).

Además, cualquier asset **gentleman** del upstream que lleve voz rioplatense
pasa por `qeNeutralizeRegionalVoice` **en tiempo de instalación** (el .md
upstream nunca se edita en disco):

| Texto upstream (rioplatense) | Reemplazo del fork |
|---|---|
| `"use warm natural Rioplatense Spanish (voseo) without overloading..."` | `"mirror the user's own Spanish register and regional voice; if there is no clear signal, default to neutral Latin American Spanish (tuteo: \"tú\", \"puedes\") without regional slang."` |
| `"Never inject Rioplatense slang, voseo,"` | `"Never inject regional slang, voseo,"` |

Se aplica a 8 assets upstream: `claude/persona-gentleman.md`,
`claude/output-style-gentleman.md`, `generic/persona-gentleman.md`,
`hermes/persona-gentleman.md`, `kimi/persona-gentleman.md`,
`kimi/output-style-gentleman.md`, `kiro/persona-gentleman.md`,
`opencode/persona-gentleman.md`.

**Dónde**:
- `internal/components/persona/inject_qe.go` — `qePersonaContent`,
  `qeNeutralizeRegionalVoice`, mapa `qeRegionalVoiceReplacements`.
- Anchors inline en `internal/components/persona/inject.go` (líneas ~79, 299,
  350, 495-497).

**Cómo se verifica**: `internal/components/persona/regional_voice_qe_test.go`
- `TestQENeutralizeRegionalVoiceCoversUpstreamAssets` — drift-guard: si el
  upstream reformula el texto rioplatense, el reemplazo deja de matchear y el
  test falla señalando el asset exacto a actualizar.
- `TestInjectGentlemanInstallsNeutralLatamPolicy` — end-to-end: instala la
  persona gentleman para claude-code/opencode/kimi en un HOME temporal y
  verifica ausencia de "Rioplatense" + presencia de la política del fork en
  los archivos ya instalados.

---

## 3. Presets QE

**Qué**: 4 presets net-new que **reemplazan** los presets dev del upstream en
el build de producción:

| Preset | Skills |
|---|---|
| `qe-sdet` | las 7 skills QA completas (stack SDET completo) |
| `qe-front` | `playwright-e2e-testing` + `a11y-playwright-testing` |
| `qe-api` | `api-testing` |
| `qe-perf` | `k6-load-test` |

**Dónde**:
- `internal/model/types_qe.go` — IDs (`PresetQESDET`, `PresetQEFront`,
  `PresetQEAPI`, `PresetQEPerf`).
- `internal/components/skills/presets_qe.go` — `qePresetSkills`,
  `qeSkillsForPreset`, `qeAllSkills`.
- Anchors en `internal/components/skills/presets.go:40-41,66,70` —
  `SkillsForPreset`/`AllSkillIDs` delegan primero al overlay.

**Cómo se verifica**: `internal/tui/screens/persona_preset_qe_test.go` →
`TestPresetOptions_QEBuildContainsOnlyQEPresets`, más la cobertura de
instalación real de skills (ver §4).

---

## 4. Skills QA net-new

**Qué**: 7 skills nuevas registradas al catálogo: `qa-manual-istqb`,
`qa-owasp-security`, `api-testing`, `playwright-e2e-testing`,
`a11y-playwright-testing`, `k6-load-test`, `selenium-e2e-java`. Se agregan vía
`append` en `init()` — **nunca editan** el slice `mvpSkills` del upstream.

**Dónde**:
- `internal/catalog/skills_qe.go` — registro (`qaSkills` + `init()`).
- Directorios en `netNewDirs` de `overlay.json`:
  `internal/assets/skills/{qa-manual-istqb,qa-owasp-security,api-testing,
  playwright-e2e-testing,a11y-playwright-testing,k6-load-test,
  selenium-e2e-java}`.

**Cómo se verifica**:
`internal/components/skills/qa_injection_qe_test.go` →
`TestInjectQASkillsToAgentsRealFilesystem` — test de regresión real (bug
histórico: assets embebidos vacíos en `qa-owasp-security` abortaban toda la
instalación de skills para "qwen-code", sin que el suite E2E upstream lo
detectara). Instala las 7 skills en HOME temporal para
qwen-code/opencode/claude-code y verifica: ningún `SKILL.md` vacío, ningún
archivo instalado de 0 bytes, y que el subárbol
`references/scripts/templates` de `qa-owasp-security` instala con contenido
real.

---

## 5. SDD-testing override (`gentle-qe-589`)

**Qué**: el ciclo SDD (`explore → propose → spec → design → tasks → apply →
verify`) emite artefactos de **test-design** en vez de artefactos de
desarrollo de producto:

| Fase | Reencuadre QE |
|---|---|
| `explore` | riesgo (likelihood × impact), testabilidad, defect clustering (20/80), inventario de oráculos |
| `propose` | capability → test-requirement / oracle |
| `spec` | oracle-first, escenarios GIVEN/WHEN/THEN, particiones límite/negativas |
| `design` | la "Testing Strategy" pasa a ser el documento entero: niveles de test, técnicas ISTQB nombradas, pirámide de test, priorización por riesgo |
| `tasks` | escenarios/casos automatizables etiquetados por capa de pirámide + técnica ISTQB |
| `apply` | escribe CÓDIGO DE TEST (specs, fixtures, Page Objects) usando las skills QA del fork |
| `verify` | ejecución + coverage por riesgo (no 100% vanidad) + flakiness |
| `archive`/`onboard`/`init` | quedan neutrales — **fail-open** al contenido dev upstream (no existe asset QE para estas fases) |

**Mecanismo**: 9 assets net-new en `internal/assets/skills/_qe-sdd/` (7
overrides de fase + 2 siblings strict-TDD: `apply.strict-tdd.md`,
`verify.strict-tdd-verify.md`) + un único helper
`QESDDTestingContent(skillID, fileName)`
(`internal/components/skills/inject_qe.go`), gateado por `IsSDDSkill`,
fail-open (`ok=false`) si no hay asset QE para esa fase/archivo.

**Los 4 injector paths** (verificados en código, no solo en el design doc):

1. **Path 1** — `internal/components/skills/inject.go:90`
   (`InjectWithCapability`, dentro del `WalkDir`): swap total de cada `.md`
   walked (SKILL.md + siblings strict-tdd) para todo adapter
   `SupportsSkills()` (Claude, Cursor, Kiro, Kimi, y cualquier futuro adapter).
2. **Path 2** — `internal/components/sdd/inject.go:722` (step 3c, wrappers de
   sub-agente nativo, ej. `claude/agents/sdd-apply.md`): **body-swap** vía
   `qeSwapNativeAgentBody` (`internal/components/sdd/inject_qe.go`) —
   preserva `name`/`model`/`effort`/`tools` del frontmatter byte-idénticos,
   reescribe solo la línea `description:` (soporta scalar YAML folded/literal
   multilinea), reemplaza el body entero. Corre antes de
   `injectCodeGraphGuidanceIntoPrompt` para que la guía CodeGraph aterrice en
   el nuevo body QE. No-op defensivo si el wrapper no tiene fence de
   frontmatter.
3. **Path 3** — `internal/components/sdd/prompts.go:83`
   (`WriteSharedPromptFiles`, prompts compartidos OpenCode/Kilocode): swap de
   `SKILL.md` antes de la inyección de guía CodeGraph.
4. **Path 4** — `internal/components/uninstall/service.go:535` (robustez, no
   inyección): agrega `_qe-sdd` al skip-set del desinstalador de skills, para
   que un uninstall de `ComponentSkills` no borre los overrides QE.

**Strict-TDD reencuadrado**: `apply.strict-tdd.md` y
`verify.strict-tdd-verify.md` redefinen el ciclo RED → GREEN → REFACTOR como
automatización de tests, no código de producción:
- **RED** = escribir el assert/oráculo que falla primero.
- **GREEN** = el Page Object/fixture mínimo que lo hace pasar.
- **REFACTOR** = limpiar el test sin romperlo.
Nunca "escribir código de producción" — la verificación strict-mode confirma
que cada oráculo RED falló genuinamente antes de GREEN y que ningún test pasa
vacuamente.

**Cómo se verifica**:
`internal/components/sdd/qe_sdd_override_qe_test.go` (batería principal):
- `TestQEStructuralOracleAllSevenPhases` — headings requeridos por fase (ej.
  `design` debe contener `## Test levels`, `Test pyramid`, `risk-based`).
- `TestQEStructuralOracleRejectsKeywordStuffing`.
- `TestQENegativeMarkersAreDevVerified` — ausencia de strings dev reales
  (confirmados presentes en el upstream real que guardan, ej. "Implement
  code changes"/"writes code" en el wrapper de `sdd-apply`, "production
  code" en los bodies SKILL).
- `TestQEOverride_Path1_SkillsInjectionSwapsAllPhasesAndSiblings`,
  `TestQEOverride_Path2_NativeSubAgentBodySwapPreservesFrontmatter`,
  `TestQEOverride_Path2_NoDuplicateFrontmatterOrSectionMarkers`,
  `TestQEOverride_Path3_SharedPromptFilesSwapsBeforeCodeGraphGuidance` — uno
  por path.
- `TestQEOverride_GateFailOpenForNonQEPhases` — `ok=false` para
  `sdd-archive`/`sdd-onboard`/`sdd-init` y para IDs no-SDD.
- `TestQEOverride_StrictTDDOracle`.
- 6 tests unitarios del parser `qeSwapNativeAgentBody`
  (`TestQeSwapNativeAgentBody_*`): fence ausente = no-op defensivo,
  description folded/single-line, blank-line continuation, strip de
  frontmatter propio del asset QE, campos preservados byte-idénticos.

Path 4 se verifica en
`internal/components/uninstall/qe_sdd_override_uninstall_test.go` →
`TestComponentOperationsSkills_PreservesQESDDOverlayDirectory`,
`TestComponentOperationsSDD_PreservesQESDDOverlayDirectory`.

---

## 6. Instalador TUI simplificado (`gentle-qe-dl9`)

**Qué**: flujo **unconditional** en el build de producción (ningún flag, env
var, ni estado persistido puede restaurarlo al flujo dev completo):

- **Oculta 6 pantallas**: Claude Model Picker, Kiro Model Picker, Codex Model
  Picker, SDDMode (vía filtro de `pickerFlowSlice`), CommunityTools y
  OpenCodePlugins (vía guards `shouldShowCommunityToolsScreen`/
  `shouldShowOpenCodePluginsScreen` que retornan `false` temprano).
  **StrictTDD queda visible** — se considera disciplina SDET valiosa, y
  ocultar una pantalla nunca implica remover el componente subyacente
  (`ComponentSDD` se mantiene).
- **Filtra listas Persona/Preset**: `PersonaOptions()` = solo
  `[PersonaSDET]`; `PresetOptions()` = solo los 4 presets QE. En producción
  es **reemplazo total**, nunca `append`.
- **Colapsa el Welcome menu** a 7 entradas esenciales: Start installation,
  Sync configs, Upgrade + Sync, Manage backups, Managed uninstall, Quit, más
  cualquier entrada "Upgrade tools...". El colapso matchea por **label**, no
  por posición, porque "OpenCode SDD Profiles" se inserta condicionalmente
  upstream (rompería índices fijos).
- **Remapeo de cursor de dos gaps** (`qeWelcomeCanonicalCursor`): el índice
  colapsado (0-6) se traduce al índice canónico que espera el `switch`
  upstream **intacto**, saltando el "leader gap" (Configure models/Create
  Agent/OpenCode Plugins/[OpenCode SDD Profiles condicional]) y el "tail gap"
  (CommunityTools, que se ubica entre Managed uninstall y Quit). Es el punto
  más frágil de la feature — cualquier reordenamiento del menú upstream lo
  rompe silenciosamente si no se actualiza la tabla de remapeo.
- **SDDMode default = `SDDModeSingle`**, seteado **explícitamente** (no el
  zero-value `""`) tanto en el TUI (`NewModel`,
  `internal/tui/model.go:589-607` vía `model.QEDefaultSDDMode`) como en el
  path CLI (`internal/cli/validate.go:60`). Es explícito porque `""` se
  auto-promueve silenciosamente a `SDDModeMulti` cuando hay perfiles OpenCode
  presentes o detectados (`internal/app/app.go:664-666`,
  `internal/cli/sync.go:681-682`) — fijar `single` cierra ese path de
  auto-promoción de forma determinística. `single` no reduce capacidad SDD:
  el ciclo multi-agente de `gentle-qe-589` corre sobre inyección de
  skills/prompts (`ComponentSDD`), intacto bajo `single`; el knob `SDDMode`
  solo gobierna asignación de modelo por fase en OpenCode.

**Testing seam** (`internal/model/seam_qe.go`): variable compartida
`model.QEInstallerFlow` (`const qeFlowDefault = true`), vive en el paquete
`model` para evitar dependencias cruzadas tui→cli/screens. En el binario de
producción nunca se compilan archivos `_test.go`, así que la variable queda
`true` siempre → el flujo QE es incondicional para usuarios reales. Cada
paquete afectado (`tui`, `screens`, `cli`) tiene un `TestMain` net-new
(`qe_seam_test.go`) que la apaga (`false`) antes de correr sus tests, para
que los ~93 tests de flujo dev del upstream sigan pasando sin edición; los
tests QE la reactivan localmente con `defer` restore y **nunca** llaman
`t.Parallel()` (mutan una var de paquete compartida — requisito documentado
inline en `seam_qe.go`). `QEDefaultSDDMode(cur)` solo fuerza `single` cuando
el seam está ON y `cur == ""`; con el seam OFF devuelve el valor upstream sin
tocar.

**Dónde**:
- `internal/tui/model_qe.go` — `qeFilterPickerFlow`,
  `qeSuppressCommunityTools`, `qeSuppressOpenCodePlugins`,
  `qeWelcomeCanonicalCursor`.
- `internal/tui/screens/welcome_qe.go` — `qeWelcomeOptions`.
- `internal/tui/screens/persona_qe.go` / `preset_qe.go` — filtros puros
  (`qeFilter*Options`, siempre QE-only) + wrappers seam-aware
  (`qe*OptionsForBuild`: seam OFF reproduce el comportamiento shippeado hoy —
  dev + `append` QE —, seam ON reemplaza totalmente).
- `internal/model/seam_qe.go` / `types_qe.go` — la seam compartida y
  `QEDefaultSDDMode`.
- 8 anchors de una línea en archivos upstream: `internal/tui/model.go`
  (líneas ~589, 607, 1488, 3739, 3758, 3987), `internal/tui/screens/
  welcome.go:55`, `internal/tui/screens/persona.go:13`,
  `internal/tui/screens/preset.go:17`.

**Cómo se verifica**:
- `internal/tui/model_qe_test.go` — `TestQEWelcomeCanonicalCursor_RemapTable`
  (7×2 casos, con/sin OpenCode detectado), `TestQEWelcomeCanonicalCursor_
  QuitBoundary`, `TestQEFilterPickerFlow_ExcludesDevOnlyScreens`,
  `TestQEModel_PickerFlowSliceNeverExposesDevOnlyScreens`,
  `TestQESuppressCommunityTools_AlwaysTrue`,
  `TestQESuppressOpenCodePlugins_AlwaysTrue`.
- `internal/tui/installer_flow_qe_test.go` — integración Bubbletea:
  `TestQEInstallerFlow_HiddenScreensNeverActivated`,
  `TestQEInstallerFlow_SDDModeDefaultsToSingle`,
  `TestQEInstallerFlow_StrictTDDScreenIsReachable`,
  `TestQEInstallerFlow_WelcomeCollapsedQuitDispatchesQuit`,
  `TestQEInstallerFlow_ReverseNavigationMirrorsForwardAndAvoidsHiddenScreens`
  (navegación reversa = espejo exacto de la forward).
- `internal/tui/screens/welcome_qe_test.go` —
  `TestQEWelcomeOptions_CollapsesToSevenEssentials`,
  `TestQEWelcomeOptions_ExcludesDevOnlyEntries`,
  `TestQEWelcomeOptions_StructuralFailFast` (invariante `len==7`, para que un
  cambio futuro al keep-set sin actualizar el remap falle ruidosamente).
- `internal/tui/screens/persona_preset_qe_test.go` —
  `TestPersonaOptions_QEBuildContainsOnlySDET`,
  `TestPresetOptions_QEBuildContainsOnlyQEPresets`,
  `TestQEFilterPersonaOptions_IgnoresDevInput`,
  `TestQEFilterPresetOptions_IgnoresDevInput`.
- `internal/tui/qe_seam_test.go`, `internal/tui/screens/qe_seam_test.go`,
  `internal/cli/qe_seam_test.go` — `TestMain` que apaga el seam por paquete.
- `internal/model/seam_qe_test.go` → `TestQEFlowDefaultIsOnInProduction` —
  prueba que la const `qeFlowDefault == true` (inmutable, no afectada por el
  flip de `TestMain`), garantizando que el binario shippeado corre con el
  filtro ON.

---

## 7. Logo / wordmark (menor, cosmético)

**Qué**: reemplaza el logo neón "rose" del upstream por el wordmark ANSI
Shadow "GENTLE · QE — Quality, Engineered", reasignando la var `logoLines`
del paquete `styles` vía `init()` (el archivo upstream `logo.go` nunca se
edita).

**Dónde**: `internal/tui/styles/logo_qe.go`.

**Cómo se verifica**: no se encontró test dedicado para este archivo en el
overlay — riesgo bajo por ser puramente cosmético, pero es un gap de
cobertura si se busca 100% de trazabilidad comportamiento→test.

---

## Gaps / puntos abiertos detectados durante esta investigación

- `tools/qe-overlay verify` solo hace *presence check* (`strings.Contains`
  vía `mustContain`), no un diff línea a línea contra upstream. Está
  documentado como limitación aceptada en `openspec/changes/gentle-qe-dl9/
  design.md` (Req 6) y como mejora cross-cutting pendiente en un bead
  separado — no cubre ediciones espurias fuera de los anchors registrados.
- El Open Question de `gentle-qe-589/design.md` ("confirmar el set exacto de
  negative markers con el maintainer") sigue sin marcar `[x]` resuelto en el
  documento, aunque el test `TestQENegativeMarkersAreDevVerified` ya existe,
  pasa, y sus markers están verificados contra el texto upstream real.
- El logo (`logo_qe.go`, §7) no tiene test dedicado.
