# Anclas inline del overlay QE (Gentle-QE)

> Generado a partir de `tools/qe-overlay/overlay.json` el 2026-07-11, verificado contra el código de la rama `feat/qe-simplified-installer`. Documento vivo: regenerar/actualizar cada vez que `overlay.json` cambie.

## Qué es una "ancla inline"

El fork Gentle-QE mantiene su comportamiento sobre `gentleman-programming/gentle-ai` por dos vías:

1. **Archivos `_qe.go` / net-new** (capa 1-2 del overlay, ver `tools/qe-overlay/README.md`): cero conflicto en merges, porque el upstream no los toca.
2. **Anclas inline**: líneas de ~1 renglón insertadas en archivos **propios del upstream** que delegan a una función del overlay. Cada una es un punto de conflicto potencial en cada `git merge upstream/main`, porque el upstream puede editar esa misma línea o su entorno inmediato.

`tools/qe-overlay/overlay.json` registra estas anclas en dos listas:

- **`inlineAnchors`** (21 entradas): delegación de comportamiento — el fork inyecta lógica (persona SDET, presets QE, defaults, supresión de features community/OpenCode, contenido SDD de testing).
- **`brandingAnchors`** (11 entradas): identidad de producto — el fork apunta usos de literales (`"gentle-ai"`, nombre de owner, etc.) a las constantes de `internal/branding`. Nótese que dos de las 11 (`internal/branding/branding.go` × 2) no son un punto de merge contra el upstream — son un auto-chequeo del propio archivo fuente de verdad, no una delegación en código ajeno.

`tools/qe-overlay verify` (corre en CI, job *Overlay Guard*) falla si algún `mustContain` desaparece de su archivo — es la única fuente de verdad automatizada; este documento es la capa humana de contexto sobre *por qué* cada línea existe y *cuánto duele* perderla en un merge.

También existe un tercer mecanismo, `brandLeak` (no tabulado aquí porque no es una lista de anclas puntuales): un scan de contenido que busca literales `gentle-qa`/`gentle_qa` (marca vieja del fork, ya renombrada) en todo el contenido net-new. No es un punto de conflicto de merge — es una guarda de higiene de marca.

## Verificación realizada

Se hizo `grep` de cada `mustContain` contra su `file` referenciado. **Resultado: 32/32 anclas (21 inline + 11 branding) están presentes hoy en el código.** No se encontró ninguna desalineación `mustContain` ↔ código.

Hallazgo secundario (no es una desalineación, es un gap de convención): 4 de las 21 anclas inline no llevan el comentario estándar `// overlay Gentle-QE (ancla qe-overlay)` que sí llevan las otras 17 — usan en su lugar un bloque de comentario más largo tipo `// QE override: ...`. Ver sección de oportunidades, punto 5.

## Tabla de anclas inline, por área

### Skills (2 anclas)

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/components/skills/presets.go:41` | `qeSkillsForPreset` | Guard al inicio de `SkillsForPreset()`: si el preset es uno de los QE, devuelve la lista QE antes de caer al `switch` de presets upstream. | Medio — función corta y estable, pero el `switch` de abajo puede ganar casos nuevos. |
| `internal/components/skills/presets.go:70` | `qeAllSkills` | Append al final de `AllSkillIDs()`: agrega los IDs de skills QE a la lista completa de skills conocidos. | Bajo — línea de `append` aislada al final de la función. |

### Persona (4 anclas)

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/components/persona/inject.go:79` | `qeNeutralizeRegionalVoice(personaContent(` | Envuelve la llamada central `personaContent(adapter.Agent(), persona)` para neutralizar voseo/rioplatense antes de escribir el archivo de persona. | Alto — línea núcleo de la función principal de inyección; cualquier refactor del flujo de inyección la toca. |
| `internal/components/persona/inject.go:299` | `qeNeutralizeRegionalVoice(assets.MustRead("kimi/output-style-gentleman.md"))` | Neutraliza el output-style "gentleman" de Kimi antes de escribirlo (Módulo 2 de la inyección). | Medio — anidada en un `switch` de selección de output-style que sí cambia cuando se agregan personas. |
| `internal/components/persona/inject.go:350` | `qeNeutralizeRegionalVoice(assets.MustRead("claude/output-style-gentleman.md"))` | Mismo patrón que arriba pero para Claude Code (Módulo 3, output-style + merge de settings). | Medio — mismo motivo. |
| `internal/components/persona/inject.go:497` | `model.PersonaSDET` | Caso del `switch` en `personaContent()`: cuando la persona es SDET, delega a `qePersonaContent(agent, persona)`. | Medio — el `switch` de personas es el punto donde el upstream agrega personas nuevas. |

### TUI / instalador (10 anclas)

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/tui/screens/preset.go:17` | `qePresetOptionsForBuild` | Toda la función delega: devuelve las opciones de preset construidas por el overlay en vez de las del build. | Bajo-medio — archivo de 53 líneas, función wrapper completa; solo choca si upstream cambia la firma. |
| `internal/tui/screens/persona.go:13` | `qePersonaOptionsForBuild` | Mismo patrón, para las opciones de persona del screen de selección. | Bajo-medio — mismo motivo (45 líneas). |
| `internal/tui/screens/welcome.go:55` | `qeWelcomeOptions` | Mismo patrón, para las opciones del welcome screen (incluye lógica de perfiles/engines). | Medio — el screen de welcome (244 líneas) es más activo que preset/persona. |
| `internal/tui/model.go:589` | `model.PresetQESDET` | En `NewModel()`: fuerza `componentsForPreset(model.PresetQESDET, model.PersonaSDET)` como default en vez del preset upstream. | Alto — `NewModel()` es la función de inicialización central del modelo TUI, muy tocada por el upstream. |
| `internal/tui/model.go:607` | `QEDefaultSDDMode` | Misma función `NewModel()`: fuerza `SDDMode=single` (seam ON) vía `model.QEDefaultSDDMode(selection.SDDMode)`. | Alto — mismo motivo que la anterior; ambas viven en la misma función crítica. |
| `internal/tui/model.go:1488` | `qeWelcomeCanonicalCursor` | Dentro de `confirmSelection()`, caso `ScreenWelcome`: reescribe el cursor antes de interpretar la selección del usuario. | Alto — `confirmSelection()` es el switch de navegación principal, área de alto churn. |
| `internal/tui/model.go:3739` | `qeSuppressOpenCodePlugins` | Guard al inicio de `shouldShowOpenCodePluginsScreen()`: si está activo, oculta el screen de plugins de OpenCode. | Medio — guard aislado al tope de una función corta. |
| `internal/tui/model.go:3758` | `qeSuppressCommunityTools` | Guard al inicio de `shouldShowCommunityToolsScreen()`: oculta el screen de herramientas community. | Medio — mismo patrón que la anterior. |
| `internal/tui/model.go:3987` | `qeFilterPickerFlow` | Al final de la función que arma la secuencia de screens del picker flow: filtra/ajusta la lista antes de devolverla. | Alto — el upstream agrega screens nuevos a este flujo con frecuencia; la línea de retorno es el punto de fricción típico. |
| `internal/components/uninstall/service.go:535` | `_qe-sdd` | Excluye el directorio `_qe-sdd` de un uninstall genérico de skills (junto con `_shared` y prefijo `sdd-`). | Bajo-medio — lista de exclusión estática, la lógica de uninstall cambia poco. |

### CLI (2 anclas)

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/cli/validate.go:60` | `QEDefaultSDDMode` | En la validación de flags no interactivos: fuerza el mismo default de `SDDMode` que `model.go:607`, para el flujo CLI. | Medio — misma constante que en TUI, pero función de validación más chica y estable. |
| `internal/cli/validate.go:93` | `qeDefaultPreset` | `return qeDefaultPreset, nil` como fallback de preset cuando no se especifica uno. | Medio — línea de retorno aislada dentro de un `switch`/resolución de preset. |

### SDD (3 anclas)

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/components/skills/inject.go:90` | `QESDDTestingContent` | Al materializar skills embebidos: si hay contenido QE para ese `(id, relPath)`, lo sustituye por el dev asset upstream (fail-open si no hay match). | Alto — dentro del loop de materialización general de skills; upstream toca esto cuando cambia cómo se instalan skills. |
| `internal/components/sdd/inject.go:729` | `qeSwapNativeAgentBody` | Al inyectar sub-agentes SDD: reemplaza el body del wrapper nativo por contenido QE, preservando la sección de capability y ANTES de inyectar guía CodeGraph. | Muy alto — vive en el archivo más grande y complejo del overlay (2443 líneas), en la lógica de sub-agentes SDD, el área de mayor actividad upstream. |
| `internal/components/sdd/prompts.go:83` | `QESDDTestingContent` | Mismo patrón que `skills/inject.go`, pero para el pipeline de prompts SDD (no de skills instalados): sustituye contenido antes de inyectar guía CodeGraph. | Alto — mismo mecanismo, tercera copia del mismo patrón (ver oportunidades). |

## Tabla de anclas de branding

| Archivo upstream | Símbolo `mustContain` | Qué hace la delegación | Riesgo de conflicto |
|---|---|---|---|
| `internal/branding/branding.go` | `EduardoVeraE` | Constante `Owner` — no es un punto de merge contra upstream, es auto-chequeo del archivo fuente de verdad (propio del fork). | Ninguno |
| `internal/branding/branding.go` | `gentle-qe` | Constantes `Product`/`Repo` — mismo motivo. | Ninguno |
| `internal/update/registry.go:21` | `branding.Owner` | Campo `Owner:` de un struct literal que describe la herramienta actual para el self-update. | Bajo |
| `internal/update/check.go:158` | `branding.Repo` | Condición que identifica si un `tool` detectado es el propio binario (`Name`, `Owner`, `Repo`). | Bajo-medio |
| `internal/update/advisory.go:30` | `branding.Owner` | `var` a nivel de paquete: construye la URL del advisory de seguridad con owner/repo del fork. | Bajo |
| `internal/update/instructions.go` | `branding.Product` | 4 usos: caso de `switch`, mensajes de instrucciones de upgrade específicos por gestor de paquetes. | Medio — el `switch` gana casos cuando se agregan gestores de paquetes. |
| `internal/update/upgrade/strategy.go` | `branding.Product` | 2 usos en archivo de 782 líneas: comparación de identidad de la tool y mensaje de salida antes del reemplazo de binario. | Medio-alto — archivo de estrategia de upgrade, área activa. |
| `internal/app/app.go` | `branding.Product` | 3 usos en archivo central de 949 líneas: impresión de versión, mensajes de error de comando desconocido, uso de `skill-registry`. | Medio-alto — `app.go` es el entrypoint de comandos, tocado con frecuencia. |
| `internal/tui/styles/styles.go:31` | `branding.Display` | Construye el string del banner de versión (`"Gentle-QE vX — Unified AI Ecosystem..."`). | Bajo |
| `internal/app/help.go:11` | `branding.Product` | Variable local usada para armar el texto de ayuda. | Bajo |
| `internal/cli/doctor.go` | `branding.Product` | 6 usos en archivo de 417 líneas: lista de tools conocidas y remedios sugeridos en el health check. | Medio — múltiples puntos de uso suben la probabilidad de colisión, aunque `doctor.go` cambia con poca frecuencia. |

## Oportunidades de minimización (recomendaciones — no aplicadas)

1. **`internal/tui/model.go` concentra 6 de las 21 anclas inline** (líneas 589, 607, 1488, 3739, 3758, 3987) en el archivo más grande y más activo del repo (4605 líneas). Es el punto de mayor densidad de riesgo. Sugerencias puntuales:
   - `589` (`componentsForPreset(model.PresetQESDET, ...)`) y `607` (`QEDefaultSDDMode`) viven en la misma función `NewModel()` y son ambas "aplicar defaults QE a la selección inicial". Podrían colapsarse en una sola llamada `qeApplyModelDefaults(&selection)` inmediatamente después de construir `selection`, reduciendo 2 anclas a 1 en esa función.
   - `3739` (`qeSuppressOpenCodePlugins`) y `3758` (`qeSuppressCommunityTools`) son guards estructuralmente idénticos en dos funciones hermanas (`shouldShowOpenCodePluginsScreen` / `shouldShowCommunityToolsScreen`). No son fácilmente fusionables sin tocar ambas funciones, pero si el upstream llega a fusionar esos dos screens algún día, el overlay debería seguirle el paso con una sola ancla.

2. **`internal/components/persona/inject.go` concentra 4 anclas**, tres de ellas (`79`, `299`, `350`) son el mismo patrón `qeNeutralizeRegionalVoice(assets.MustRead(...))` aplicado a tres call-sites distintos. Consolidar de verdad requeriría interceptar `assets.MustRead` en el punto de lectura (un `qeMustRead()` que neutralice internamente), pero eso ampliaría la superficie de la ancla (pasaría de 3 líneas triviales a una función que intercepta lecturas de assets) — no está claro que sea una mejora neta. Vale la pena evaluarlo solo si el upstream agrega un cuarto asset "gentleman" con voseo.

3. **El patrón `QESDDTestingContent` / `qeSwapNativeAgentBody` está triplicado** en tres pipelines de materialización distintos (`skills/inject.go:90`, `sdd/inject.go:729`, `sdd/prompts.go:83`). Esto refleja que el upstream mismo tiene tres pipelines separados (instalación de skills, sub-agentes SDD, prompts SDD) — no es redundancia del overlay, es la superficie real de integración. **No es consolidable sin tocar la arquitectura del upstream.** Es, sin embargo, la zona de mayor riesgo acumulado: `sdd/inject.go` (2443 líneas) es el archivo más complejo y más propenso a refactors del repo. Recomendación: priorizar tests de regresión de esta ancla (ya existe `internal/components/sdd/qe_sdd_override_qe_test.go` en `overlayFiles` — verificar que cubra específicamente la línea 729, no solo el contenido QE en general).

4. **Las 3 anclas en `internal/tui/screens/{preset,persona,welcome}.go` ya son wrappers completos** (toda la función delega). Son, en la práctica, casi net-new — el único acoplamiento real es la firma de la función. No requieren consolidación; podría evaluarse convertirlas formalmente en archivos `_qe.go` si el upstream llegara a eliminar por completo esas funciones wrapper (poco probable dado su rol de entrypoint de screen).

5. **Las 11 anclas de branding ya son mínimas** (un token de constante por línea) y de bajísimo costo de reaplicación individual. No se recomienda ninguna acción — es el patrón ejemplar del overlay ("apuntar a `branding.*`"). Única observación: 2 de las 11 (`branding.go` mismo) no son puntos de merge reales, son auto-chequeo; podrían moverse fuera de `brandingAnchors` a una validación separada tipo "constants present" si se quiere que el conteo de `brandingAnchors` refleje solo puntos de conflicto de merge.

6. **Inconsistencia de convención de comentario**: 17 de las 21 anclas inline llevan `// overlay Gentle-QE (ancla qe-overlay)` al final de la línea, pero 4 no lo llevan (`tui/model.go:607`, `components/skills/inject.go:90`, `components/sdd/inject.go:729`, `components/sdd/prompts.go:83`) — usan en su lugar un bloque `// QE override: ...` más largo arriba de la línea. Esto no rompe `qe-overlay verify` (que hace match por `mustContain`, no por el comentario), pero sí rompe el `grep "ancla qe-overlay"` que el README del overlay documenta como forma humana de ubicar todos los puntos de anclaje. Recomendación: estandarizar — agregar el sufijo corto también en estas 4 líneas (puede convivir con el bloque explicativo largo) para que un solo `grep` encuentre las 21.

## Resumen numérico

- **Anclas inline**: 21 (skills 2, persona 4, tui/instalador 10, cli 2, sdd 3).
- **Anclas de branding**: 11 (2 de ellas son auto-chequeo de `branding.go`, no puntos de merge).
- **Total en `overlay.json`**: 32.
- **Desalineaciones `mustContain` ↔ código encontradas**: 0.
- **Archivo de mayor concentración de riesgo**: `internal/tui/model.go` (6 anclas) y `internal/components/sdd/inject.go` (1 ancla, pero en el archivo más complejo del repo).
