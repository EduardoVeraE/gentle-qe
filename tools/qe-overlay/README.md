# qe-overlay — mantenimiento del overlay SDET (Gentle-QE)

Gentle-QE es un **overlay aditivo** sobre el upstream `gentleman-programming/gentle-ai`.
Su valor es el comportamiento (persona SDET, ISTQB, shift-left, skills de testing), no la
marca. Esta carpeta contiene la herramienta y el manifiesto que mantienen ese overlay
barato de re-aplicar en cada actualización del upstream.

## Principio

Cada línea editada en un archivo del upstream = conflicto recurrente en cada sync.
Cada archivo nuevo del fork = cero conflicto. Por eso el overlay se concentra en
archivos `_qe.go` / net-new y reduce las ediciones inline a unos pocos puntos de
anclaje (marcados en el código con `// ... (ancla qe-overlay)`).

## Capas del overlay

1. **Assets net-new** — directorios bajo `internal/assets/` (7 skills QA, persona SDET).
   Se auto-embeben vía el glob `//go:embed all:skills …`; agregar uno no toca código.
2. **Wiring Go aislado** — archivos `_qe.go` (`types_qe.go`, `catalog/skills_qe.go`,
   `presets_qe.go`, `persona/inject_qe.go`, `tui/screens/*_qe.go`, `cli/qe_defaults.go`)
   más ~8 líneas de delegación inline. Incluye `qeNeutralizeRegionalVoice`
   (inject_qe.go): neutraliza en runtime la directiva rioplatense/voseo de los
   assets gentleman del upstream (espeja el registro del usuario, fallback a
   español neutro LatAm) sin editar los `.md`; la guarda de drift vive en
   `persona/regional_voice_qe_test.go`.
3. **Branding mínimo** — `internal/branding`, referenciado solo en los sitios funcionales
   del self-update / version. El **module path Go y el state dir `.gentle-ai` NO se
   rebrandean** (alta dispersión, romperían el build / generarían fricción).
4. **Override** — `README.md`, `.goreleaser.yaml`, `scripts/install.sh` resuelven a favor
   del fork vía `.gitattributes merge=ours`.

## Uso de la herramienta

```bash
go run ./tools/qe-overlay verify         # ¿overlay intacto? ¿drift del upstream?
go run ./tools/qe-overlay verify --diff  # verify + diff real (ver abajo)
go run ./tools/qe-overlay diff           # solo el diff real contra el baseline upstream
go run ./tools/qe-overlay accept         # absorbe skills nuevos del upstream al manifiesto
go run ./tools/qe-overlay apply          # guía para re-aplicar anclas perdidas (ver abajo)
go run ./tools/qe-overlay apply --write  # + reinserción automática en el caso inambiguo
```

`verify` (exit ≠ 0 si hay problemas) chequea: directorios net-new presentes,
**net-new instalables** (sin assets embebidos de 0 bytes y con `SKILL.md` en la raíz —
un solo archivo vacío aborta toda la inyección de skills al instalar, y `//go:embed all:`
arrastra hasta dotfiles vacíos como `.gitkeep`), archivos `_qe.go` presentes, anclas de
branding intactas, delegaciones inline intactas, y skills upstream nuevos sin clasificar.
Corre en CI (job *Overlay Guard*).

`verify` valida las anclas por **contenido** (`strings.Contains`): confirma que la
línea de anclaje siga presente, pero NO detecta si además se coló una línea de lógica
espuria (no-ancla) en un archivo upstream. Para eso existe `diff` (implementación en
`diff.go`):

- Calcula el baseline como `git merge-base HEAD upstream/main` (no `upstream/main`
  directo, para no mezclar commits del upstream posteriores a la base del fork).
- Para cada archivo con anclas (`inlineAnchors` + `brandingAnchors`) corre
  `git diff -U10 <baseline> HEAD -- <archivo>` y agrupa el diff en hunks.
- Un **hunk** es legítimo si alguna de sus líneas agregadas contiene un
  `mustContain` registrado para ese archivo, un marcador (`overlay Gentle-QE` /
  `ancla qe-overlay`), o una referencia a `branding.*` (branding minimalista sin
  ancla explícita en la línea puntual). Un hunk sin ninguna línea sustantiva (solo
  imports/blancos) también es legítimo — un import agregado que no se usa ya rompe
  el build, y si se usa, la línea de uso pasa por el chequeo normal.
- Cualquier otro hunk se reporta línea por línea (`archivo:línea`) y el exit code es 1.
- Archivos net-new (sin contraparte en el baseline, p. ej.
  `internal/branding/branding.go`) se saltean: el concepto de "edición espuria de
  upstream" no aplica — ya lo cubre la sección de presencia de `verify`.
- Si el remote `upstream` no está configurado o `upstream/main` no es resoluble
  localmente (falta `git fetch upstream`), `diff` degrada con un warning y exit 0
  — no bloquea CI en checkouts sin ese remote.

**Límite conocido**: la legitimidad se evalúa a nivel de *hunk* (bloque de diff
contiguo, contexto `-U10`), no línea por línea. Esto es deliberado: un cambio
legítimo de anclaje suele tocar varias líneas contiguas (una guarda + su cuerpo, un
comentario explicativo + la línea marcada) y alcanza con que UNA lleve el marcador.
La contrapartida es que una línea espuria insertada a menos de ~20 líneas de una
línea de ancla legítima, dentro del mismo hunk, no se detectaría. Ver
`tools/qe-overlay/diff_test.go` para los casos cubiertos.

El manifiesto `overlay.json` es la **fuente de verdad** del overlay: edítalo cuando
agregues un skill QA, un nuevo archivo `_qe.go` o un nuevo punto de anclaje.

### `apply` — ayuda a re-aplicar anclas perdidas (implementación: `apply.go`)

Durante un sync, un conflicto de merge en un archivo upstream editado por el overlay
puede resolverse a favor del upstream y **perder una línea de ancla** (inline o de
branding). `verify` DETECTA esto (`mustContain` ausente) pero no ayuda a restaurarlo.
`apply` sí:

- Para cada ancla perdida, busca en el historial de git del fork (`git log -S<mustContain>`)
  el último commit donde esa línea todavía existía, y muestra esa línea + su contexto
  inmediato (3 líneas antes/después) para que un humano la ubique y reinserte en segundos.
- Con `--write`, además intenta una **reinserción automática conservadora**: solo cuando
  la línea inmediatamente anterior (en el historial) es sustantiva y aparece **exacta y
  únicamente una vez** en el archivo actual, inserta la línea de ancla justo después. Si el
  punto de inserción es ambiguo (0 o 2+ matches) o la línea de contexto es trivial (`}`,
  `)`, blancos), no escribe nada y cae al reporte guiado para esa ancla — nunca adivina.

**Por qué no hay reinserción 100% automática**: el manifiesto solo guarda
`{file, mustContain}` — el símbolo que debe existir, no la línea completa ni su punto de
inserción exacto. Reconstruir ambos de forma confiable en el caso general (múltiples
apariciones de la línea de contexto, refactors del upstream que mueven la función, etc.)
no es factible con esa metadata sin arriesgar una reinserción en el lugar equivocado —
peor que no reinsertar nada. `apply --write` cubre el caso común y seguro (contexto único);
todo lo demás queda en el reporte guiado, que sigue siendo mucho más rápido que buscar a
mano en `git log`.

Exit code: `0` si no hay nada que restaurar, o si `--write` restauró todo lo que había.
`1` si queda al menos un ancla sin restaurar (reporte guiado impreso, requiere acción
manual). Igual que `diff`, requiere el historial de git local (no un remote específico:
busca en cualquier commit alcanzable desde donde se corre, típicamente `HEAD`).

```bash
go run ./tools/qe-overlay apply
go run ./tools/qe-overlay apply --write
go build ./... && go run ./tools/qe-overlay verify   # confirmar tras --write
```

## Runbook de sync (cada release menor del upstream)

> Runbook completo y canónico (cadencia, gotchas, manejo de bugs upstream,
> checklist): [`docs/qe-upstream-sync.md`](../../docs/qe-upstream-sync.md).
> Resumen operativo:

```bash
# Una sola vez por clon: habilitar el driver merge=ours del .gitattributes
git config merge.ours.driver true

# --no-tags es CRÍTICO: el fork limpió los ~236 tags heredados del upstream;
# sin --no-tags se re-traen. Solo necesitamos la rama main.
git fetch --no-tags upstream main
git checkout -b sync/vX.Y.Z main
git merge upstream/main            # README/.goreleaser/install.sh => merge=ours auto

go run ./tools/qe-overlay verify   # reporta overlay roto / branding drift / skills nuevos
#   - delegación inline perdida  -> `go run ./tools/qe-overlay apply` (guía o --write)
#   - branding perdido           -> apuntar el sitio a branding.* (`apply` también ayuda)
#   - skill nuevo del upstream   -> decidir si va a un preset QE, luego `accept`

go build ./...
go test ./... -run Golden -update  # regenerar goldens si cambió testdata upstream
go test ./...                       # verde
go run ./tools/qe-overlay accept   # actualizar el baseline de upstream conocido

# commit + merge a main + (si toca release) scripts/release.sh
```

> Nota: el flag `-update` para goldens del TUI vive solo en el paquete `internal/tui`
> (`go test ./internal/tui -update`), no en `internal/tui/screens`.
