// apply.go implementa el modo `qe-overlay apply`: ayuda a RE-APLICAR anclas
// (inline y de branding) perdidas tras un merge del upstream.
//
// Motivación: `verify` solo DETECTA que un mustContain ya no está en el archivo
// (overlay roto); no ayuda a restaurarlo. El manifiesto solo guarda
// {file, mustContain} — el símbolo que debe existir, no la línea completa ni su
// punto de inserción — así que un re-apply 100% automático y confiable no es
// factible con esa metadata sola. `apply` compensa buscando en el HISTORIAL DE
// GIT del fork la última versión conocida-buena de esa línea (vía `git log -S`)
// y mostrándola con contexto, para que un humano la reinserte en segundos. Con
// --write, además intenta una reinserción automática, pero SOLO en el caso
// inambiguo (línea de contexto previa única en el archivo actual) — nunca
// adivina un punto de inserción.
package main

import (
	"fmt"
	"os"
	"strings"
)

// anchorContextLines es cuántas líneas de contexto (antes/después) se muestran
// alrededor de la línea de ancla recuperada del historial, en el reporte guiado
// de `apply`. Menor que diffContextLines (10): acá el contexto es para que un
// humano ubique el punto de inserción a simple vista, no para agrupar hunks de
// diff.
const anchorContextLines = 3

// runApply es el entrypoint del modo `apply`. Devuelve el exit code del proceso:
// 0 si no había nada que restaurar o si --write restauró todo; 1 si queda al
// menos un ancla sin restaurar (requiere intervención manual con la guía
// impresa).
func runApply(m *manifest, write bool) int {
	missing := missingAnchors(m)
	if len(missing) == 0 {
		fmt.Println("✓ qe-overlay apply: no hay anclas perdidas, nada que restaurar.")
		return 0
	}

	unresolved := 0
	for _, a := range missing {
		guide, err := buildAnchorGuide(a)
		if err != nil {
			fmt.Printf("✗ %s: no se pudo reconstruir la guía de restauración de %q: %v\n", a.File, a.MustContain, err)
			unresolved++
			continue
		}

		if write {
			applied, reason := tryAutoRestore(a, guide)
			if applied {
				fmt.Printf("✓ %s: ancla %q reinsertada automáticamente (última vez vista en %s)\n", a.File, a.MustContain, shortSHA(guide.sha))
				continue
			}
			fmt.Printf("• %s: reinserción automática no aplicó (%s); guía manual abajo:\n", a.File, reason)
		}

		fmt.Print(guide.String())
		unresolved++
	}

	if unresolved == 0 {
		fmt.Println("✓ qe-overlay apply: todas las anclas perdidas fueron reinsertadas. Corré `go build ./... && go run ./tools/qe-overlay verify` para confirmar.")
		return 0
	}
	return 1
}

// missingAnchors devuelve las anclas (inline + branding con mustContain) que
// `verify` reportaría como perdidas: archivo legible pero sin el mustContain.
// Un archivo directamente ilegible no se incluye acá — no hay dónde reinsertar
// nada; ya lo reporta `verify` como overlay roto (archivo faltante).
func missingAnchors(m *manifest) []anchor {
	var out []anchor
	check := func(a anchor) {
		if a.MustContain == "" {
			return
		}
		body, err := readFile(a.File)
		if err != nil {
			return
		}
		if !strings.Contains(body, a.MustContain) {
			out = append(out, a)
		}
	}
	for _, a := range m.InlineAnchors {
		check(a)
	}
	for _, a := range m.BrandingAnchors {
		check(a)
	}
	return out
}

// anchorGuide es la guía de restauración reconstruida del historial de git para
// una sola ancla perdida.
type anchorGuide struct {
	file        string
	mustContain string
	sha         string
	lineNo      int // línea (1-based) en el archivo tal como estaba en sha
	before      []string
	anchorLine  string
	after       []string
}

// String formatea la guía para un humano: dónde estaba, en qué commit, y la
// línea + contexto para ubicar el punto de inserción equivalente en el archivo
// actual.
func (g anchorGuide) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "✗ ancla perdida en %s: falta %q\n", g.file, g.mustContain)
	fmt.Fprintf(&b, "  última vez vista en commit %s (git show %s:%s)\n", shortSHA(g.sha), shortSHA(g.sha), g.file)
	fmt.Fprintf(&b, "  línea original y contexto (línea %d del archivo en ese commit):\n", g.lineNo)
	start := g.lineNo - len(g.before)
	for i, l := range g.before {
		fmt.Fprintf(&b, "      %d\t%s\n", start+i, l)
	}
	fmt.Fprintf(&b, "    > %d\t%s\n", g.lineNo, g.anchorLine)
	for i, l := range g.after {
		fmt.Fprintf(&b, "      %d\t%s\n", g.lineNo+1+i, l)
	}
	fmt.Fprintln(&b, "  Ubicá ese contexto en el archivo actual y reinsertá la línea marcada con '>' en el punto equivalente.")
	return b.String()
}

// buildAnchorGuide busca en el historial de git del fork la última versión del
// archivo que contenía mustContain y extrae esa línea + su contexto inmediato.
//
// Recorre los commits que tocaron ese string en ese archivo (`git log -S`,
// newest-first) y devuelve el primero cuyo contenido en ESE commit todavía
// contiene mustContain — es decir, la última versión conocida-buena antes de
// que se perdiera (el propio merge de sync suele ser el commit siguiente, que
// ya no lo tiene y por eso no aparece acá).
func buildAnchorGuide(a anchor) (anchorGuide, error) {
	out, err := gitOutput("log", "--pretty=format:%H", "-S"+a.MustContain, "--", a.File)
	if err != nil {
		return anchorGuide{}, fmt.Errorf("no se pudo buscar en el historial: %w", err)
	}
	shas := nonEmptyLines(out)
	if len(shas) == 0 {
		return anchorGuide{}, fmt.Errorf("el string no aparece en ningún commit del historial de %s (¿se renombró el archivo o el ancla?)", a.File)
	}

	for _, sha := range shas {
		content, err := gitOutput("show", sha+":"+a.File)
		if err != nil {
			continue // el archivo no existía con ese path en ese commit (rename, etc.)
		}
		lines := strings.Split(content, "\n")
		for i, l := range lines {
			if !strings.Contains(l, a.MustContain) {
				continue
			}
			beforeStart := i - anchorContextLines
			if beforeStart < 0 {
				beforeStart = 0
			}
			afterEnd := i + 1 + anchorContextLines
			if afterEnd > len(lines) {
				afterEnd = len(lines)
			}
			return anchorGuide{
				file:        a.File,
				mustContain: a.MustContain,
				sha:         sha,
				lineNo:      i + 1,
				before:      append([]string(nil), lines[beforeStart:i]...),
				anchorLine:  l,
				after:       append([]string(nil), lines[i+1:afterEnd]...),
			}, nil
		}
	}
	return anchorGuide{}, fmt.Errorf("ningún commit en el historial de %s conserva una versión con %q", a.File, a.MustContain)
}

// tryAutoRestore intenta una reinserción conservadora de la línea de ancla en el
// archivo actual: solo si la línea inmediatamente anterior (en el historial) es
// sustantiva (no trivial) y aparece EXACTA y ÚNICAMENTE en el archivo actual.
// Devuelve (false, motivo) cuando no puede garantizar el punto de inserción sin
// ambigüedad — preferimos no escribir antes que reinsertar en el lugar
// equivocado.
func tryAutoRestore(a anchor, g anchorGuide) (bool, string) {
	if len(g.before) == 0 {
		return false, "sin línea de contexto previa en el historial"
	}
	prev := g.before[len(g.before)-1]
	if !isSubstantiveContextLine(prev) {
		return false, "la línea de contexto previa es demasiado trivial para ubicarla sin ambigüedad"
	}

	body, err := readFile(a.File)
	if err != nil {
		return false, "no se pudo leer el archivo actual: " + err.Error()
	}
	lines := strings.Split(body, "\n")

	matches := 0
	matchIdx := -1
	for i, l := range lines {
		if l == prev {
			matches++
			matchIdx = i
		}
	}
	if matches != 1 {
		return false, fmt.Sprintf("la línea de contexto previa aparece %d veces en el archivo actual (necesito exactamente 1)", matches)
	}

	// Si la línea de ancla ya está presente justo ahí (carrera con otra
	// restauración), no dupliques.
	if matchIdx+1 < len(lines) && lines[matchIdx+1] == g.anchorLine {
		return false, "la línea de ancla ya está presente en el punto de inserción"
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:matchIdx+1]...)
	newLines = append(newLines, g.anchorLine)
	newLines = append(newLines, lines[matchIdx+1:]...)

	if err := os.WriteFile(a.File, []byte(strings.Join(newLines, "\n")), 0o644); err != nil {
		return false, "no se pudo escribir el archivo: " + err.Error()
	}
	return true, ""
}

// isSubstantiveContextLine descarta líneas de contexto demasiado comunes para
// ubicar un punto de inserción sin ambigüedad (blancos, cierres de bloque
// sueltos como "}" o ")").
func isSubstantiveContextLine(l string) bool {
	t := strings.TrimSpace(l)
	if len(t) < 4 {
		return false
	}
	switch t {
	case "}", ")", "{", "(", "]", "[", "},", "),":
		return false
	}
	return true
}

// nonEmptyLines separa s por líneas y descarta las vacías/blancas.
func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
