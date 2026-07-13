# Runbook de sincronización con upstream (Gentle-QE)

Procedimiento operativo para traer cambios de `gentleman-programming/gentle-ai`
al fork sin romper el overlay QE. Complementa (no duplica):

- `docs/qe-overlay-anchors.md` — mapa de las 32 anclas (21 inline + 11 branding) y su riesgo de conflicto por archivo.
- `docs/qe-behavior-contract.md` — qué comportamiento diverge del upstream y cómo se verifica cada divergencia.
- `tools/qe-overlay/README.md` — mecánica de `verify` / `diff` / `accept` y el runbook corto original.

Este doc responde "¿con qué cadencia y con qué checklist se corre el runbook?", basado en el sync real de upstream v2.x (171 commits, fusionado en `feat/qe-sync-tooling` → PR #15).

---

## 1. Cadencia recomendada

**Sync periódico, no reactivo.** Objetivo: mensual, o antes si upstream corta un minor/major release.

Por qué: el sync v2.x acumuló 171 commits porque se dejó crecer el gap sin control. Un gap grande implica:

- Más superficie de conflicto por sync (cada ancla inline es un punto de fricción; con 171 commits de por medio, varias anclas cayeron en zonas con múltiples reescrituras acumuladas en vez de una sola).
- Menos trazabilidad: es más difícil atribuir un conflicto puntual a un commit upstream concreto cuando hay 171 candidatos.
- Mayor riesgo de que un bug introducido upstream (como el data race de `detect.go`, ver §3) quede varias semanas sin que el fork lo detecte, porque `go test -race` en CI solo corre sobre el código que el fork ya tiene mergeado.

Sync frecuente = diffs chicos = revisión de anclas rápida = bajo riesgo. Es la misma lógica de "integrar seguido" que ya aplicamos a features propias.

**Cómo medir el gap actual** (correrlo antes de decidir si toca sync):

```bash
git fetch --no-tags upstream main
git rev-list --count $(git merge-base HEAD upstream/main)..upstream/main
```

Un número de dos dígitos es razonable. Tres dígitos → el sync ya se atrasó, priorizarlo.

---

## 2. Runbook paso a paso

### 2.1 Fetch (cuidado con los tags)

```bash
git fetch --no-tags upstream main
```

**`--no-tags` es obligatorio, no opcional.** El fork ya limpió los ~236 tags heredados del upstream (`git tag` en el fork hoy solo lista los 4 propios: `v0.1.0`–`v0.1.3`). Un `git fetch upstream main` sin `--no-tags` los vuelve a traer todos y contamina `git tag`/`git describe` del fork. El job *Overlay Guard* en `.github/workflows/ci.yml` ya usa `--no-tags` en CI — replicar el mismo flag en local.

### 2.2 Rama de sync

```bash
git checkout -b sync/upstream-vN main
git merge --no-ff upstream/main
```

`--no-ff` para que el merge quede como un commit propio y trazable en `git log`, igual que el sync v2.x (`46372dc chore: sync upstream v2.x (preserve QE overlay)`).

### 2.3 Resolver conflictos preservando las anclas

Cada conflicto en un archivo listado en `docs/qe-overlay-anchors.md` se resuelve así:

- **Integrar el cambio de upstream** tal cual (nueva lógica, refactor, lo que sea).
- **Mantener la línea de ancla** (el símbolo `mustContain` de esa fila en la tabla de `qe-overlay-anchors.md`) dentro del código resultante, en el mismo punto de delegación semántico aunque la línea exacta upstream haya cambiado alrededor.
- Si el archivo es de los del `.gitattributes` con `merge=ours` (`README.md`, `.goreleaser.yaml`, `scripts/install.sh`, `.github/workflows/pr-check.yml`, `internal/components/opencodeplugin/plugin.go`) — **no tocar, ya resuelve solo** a favor del fork (requiere `git config merge.ours.driver true` una vez por clon). Igual revisar manualmente si upstream trajo algo relevante en esos archivos que valga la pena portar a mano.
- Golden tests del TUI: regenerar con `go test ./internal/tui -update` (el flag `-update` vive solo en ese paquete, no en `internal/tui/screens`).

Prioridad de revisión por riesgo (usar la columna "Riesgo de conflicto" de `qe-overlay-anchors.md`): las anclas en `internal/tui/model.go` (6 anclas, archivo de 4605 líneas, alto churn upstream) y `internal/components/sdd/inject.go:729` (`qeSwapNativeAgentBody`, archivo más complejo del repo) son las de mayor probabilidad de conflicto real — revisarlas primero aunque `git merge` no las marque como conflicto.

### 2.4 Validar (red de seguridad, en este orden)

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
go run ./tools/qe-overlay verify --diff
```

- `go build` / `go vet` — primero, rápido, descarta breakage obvio antes de correr el suite completo.
- `go test ./...` — suite funcional completa (unit + los `TestMain` que apagan el seam QE por paquete).
- `go test -race ./...` — mismo suite con detector de race. Es el que cazó el bug de `detect.go` en el sync v2.x (ver §3); correrlo siempre, no solo cuando "algo huele raro".
- `go run ./tools/qe-overlay verify --diff` — el único chequeo que valida el overlay en sí: `verify` confirma que las 32 anclas + los `overlayFiles`/`netNewDirs` siguen presentes; `--diff` además compara contra `git merge-base HEAD upstream/main` y falla si algún hunk en un archivo con anclas trae líneas sin marcador (`overlay Gentle-QE` / `ancla qe-overlay` / `branding.*`) — es decir, caza ediciones espurias que `verify` solo no detecta. Exit 0 en los dos modos = overlay intacto y sin ediciones fuera de las anclas registradas.

Si `verify` reporta algo, las categorías posibles y su remedio están en `tools/qe-overlay/README.md` (§"Runbook de sync"): delegación inline perdida → reaplicar la línea marcada; branding perdido → apuntar al `branding.*` correspondiente; skill nuevo del upstream sin clasificar → decidir preset QE y correr `qe-overlay accept`.

### 2.5 PR y merge

- PR a `main` con labels `type:chore` + `size:exception` (el diff de un sync legítimamente supera el budget normal de PR).
- **No crear tag.** El fork mantiene su propio versionado (`v0.1.x`) independiente del ciclo de release upstream — un sync no es un release. Cortar tag solo cuando se decida deliberadamente un release del fork (`scripts/release.sh`, fuera del alcance de este runbook).
- Después del merge: `go run ./tools/qe-overlay accept` para actualizar el baseline de upstream conocido (si no se corrió ya en la rama de sync).

---

## 3. Manejo de bugs de upstream descubiertos durante el sync

Un sync puede destapar un bug que ya existía en el código upstream pero que el fork no había mergeado todavía — como pasó con el data race en `detectInstalledVersion` (`internal/update/detect.go`), cazado por `go test -race` en el sync v2.x (commit `5714247`): el código nuevo de detección de ownership de Homebrew leía `cmd.Process` en una goroutine mientras `cmd.Output()` lo escribía internamente.

Procedimiento:

1. **Fix mínimo**, suficiente para pasar `go test -race` — no un rediseño.
2. **Marcarlo con el comentario `overlay Gentle-QE`** inline, igual que cualquier otra ancla — así `go run ./tools/qe-overlay diff` lo acepta como hunk legítimo en vez de reportarlo como edición espuria.
3. **Dejar explícito en el commit/comentario que es un bug upstream, no comportamiento QE** — con nota de que se reporta upstream y se revierte al integrarse (ver el mensaje de `5714247` como plantilla).
4. **Crear un bead** para trackear el reporte upstream.
5. **Revertir el fix local** en el próximo sync donde el fix upstream ya esté integrado — no dejarlo indefinidamente como parche paralelo.

---

## 4. Gotchas conocidos

- **`--no-tags` en el fetch, siempre.** Ver §2.1 — sin el flag se re-traen los ~236 tags upstream que el fork ya limpió.
- **`qeWelcomeCanonicalCursor` es el punto más frágil del overlay TUI.** Si upstream reordena el menú Welcome (agrega/quita/reordena entradas), el remapeo de cursor de dos gaps (`internal/tui/model.go:1488`, ver `docs/qe-behavior-contract.md` §6) se rompe **silenciosamente** — no falla el build, cambia el destino de una tecla. Validar siempre con `TestQEWelcomeCanonicalCursor_RemapTable` y el resto de `internal/tui/model_qe_test.go` / `internal/tui/installer_flow_qe_test.go` después de cualquier sync que toque `welcome.go` o el `switch` de `confirmSelection()`.
- **Brand leaks en features nuevas del upstream.** Código net-new que upstream agrega (sin ancla, porque el fork no lo tocó) puede traer literales de marca (`gentle-ai`, owner upstream, etc.) que deberían enrutarse a `branding.*` si tocan una superficie visible (self-update, mensajes de versión, User-Agent). `go run ./tools/qe-overlay diff` los caza vía la heurística `branding.` (hunks que referencian el paquete branding sin ancla explícita se aceptan; los que no referencian nada se reportan) — pero la heurística es a nivel de hunk (`-U10`), no línea por línea (ver "Límite conocido" en `tools/qe-overlay/README.md`), así que una revisión manual de código nuevo con literales de marca sigue siendo necesaria, `diff` no es infalible.
- **Golden tests del TUI**: el flag `-update` vive solo en `internal/tui`, no en `internal/tui/screens` — correr el comando en el paquete equivocado no regenera nada y falla en silencio (test sigue rojo).
- **`merge.ours.driver`** debe habilitarse una vez por clon (`git config merge.ours.driver true`) o los archivos con `merge=ours` en `.gitattributes` no resuelven automáticamente y generan conflictos evitables.

---

## 5. Checklist final de pre-merge

- [ ] `git fetch --no-tags upstream main` — confirmar `git tag | wc -l` no creció.
- [ ] Rama `sync/upstream-vN` mergeada con `--no-ff` desde `main`.
- [ ] Todos los conflictos resueltos preservando anclas (`docs/qe-overlay-anchors.md` como checklist de referencia, especialmente las de riesgo Alto/Muy alto).
- [ ] Goldens del TUI regenerados si cambió testdata upstream (`go test ./internal/tui -update`).
- [ ] `go build ./...` verde.
- [ ] `go vet ./...` verde.
- [ ] `go test ./...` verde.
- [ ] `go test -race ./...` verde (sin excepciones — ver §3 si destapa algo).
- [ ] `go run ./tools/qe-overlay verify --diff` exit 0.
- [ ] Si se aplicó un fix de bug upstream: comentado con `overlay Gentle-QE`, bead creado para reportarlo, nota de revertir en próximo sync.
- [ ] `go run ./tools/qe-overlay accept` corrido (baseline actualizado).
- [ ] PR a `main` con labels `type:chore` + `size:exception`.
- [ ] **Ningún tag creado.**
