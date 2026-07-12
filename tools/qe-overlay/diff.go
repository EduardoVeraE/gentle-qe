// diff.go implementa el modo `qe-overlay diff` (alias: `qe-overlay verify --diff`):
// una verificación REAL de "zero upstream content edits" comparando el árbol actual
// contra el baseline del upstream (merge-base), en vez de solo comprobar que las
// anclas EXISTAN en el archivo (lo que hace `verify` vía strings.Contains).
//
// Motivación: `verify` no distingue "el ancla está" de "el ancla está Y no se coló
// una línea de lógica espuria (no-ancla) en un archivo upstream". Este modo sí lo
// hace, a costa de requerir el remote `upstream` fetcheado localmente.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// overlayMarkers son los comentarios que marcan una línea de anclaje legítima en
// código upstream editado por el overlay (ver README de tools/qe-overlay).
var overlayMarkers = []string{"overlay Gentle-QE", "ancla qe-overlay"}

// brandingQualifier es el paquete propio del fork: cualquier línea que lo referencia
// es contenido de branding minimalista legítimo, aunque no calce con el mustContain
// puntual registrado para ese archivo (branding.Product/.Owner/.Repo/.Display/
// .StateDir/.UserAgent aparecen sueltos en varios sitios funcionales sin ancla
// explícita — ver internal/branding).
const brandingQualifier = "branding."

// importLineRE reconoce una línea de import de Go aislada: "path", _ "path" o
// alias "path". Estas líneas nunca son "lógica espuria" por sí solas — si el import
// agregado no se usa en ningún lado, el build ya rompe; si se usa, esa línea de uso
// sí pasa por el chequeo de legitimidad normal.
var importLineRE = regexp.MustCompile(`^(_ |[A-Za-z_][A-Za-z0-9_]* )?"[^"]*"$`)

// diffContextLines es el contexto de `git diff -U<n>` usado para agrupar hunks.
// 3 (el default de git) fragmenta cambios legítimos de una sola función en varios
// hunks cuando las líneas de anclaje quedan a más de ~6 líneas de otras líneas
// del mismo cambio (visto en internal/components/sdd/inject.go). 10 fusiona esos
// casos reales sin fusionar cambios genuinamente distantes dentro del mismo
// archivo (ver tools/qe-overlay/diff_test.go y el README).
const diffContextLines = 10

type diffFinding struct {
	file string
	line int
	text string
}

func (f diffFinding) String() string {
	return fmt.Sprintf("%s:%d: %s", f.file, f.line, strings.TrimSpace(f.text))
}

// runDiff es el entrypoint del modo `diff`. Devuelve el exit code del proceso.
func runDiff(m *manifest) int {
	if !hasUpstreamRemote() {
		fmt.Println("⚠ qe-overlay diff: el remote 'upstream' no está configurado o no es fetcheable.")
		fmt.Println("  Este chequeo requiere `git remote add upstream <url upstream> && git fetch upstream`.")
		fmt.Println("  Se omite el diff real; el overlay solo queda validado por `verify` (anclas por contenido).")
		return 0
	}

	baseline, err := resolveBaseline()
	if err != nil {
		fmt.Printf("⚠ qe-overlay diff: no se pudo resolver el merge-base con upstream/main: %v\n", err)
		fmt.Println("  Se omite el diff real (¿corriste `git fetch upstream`?).")
		return 0
	}

	findings, skipped, err := diffAnchorFiles(m, baseline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error corriendo el diff real: %v\n", err)
		return 2
	}

	if len(findings) == 0 {
		fmt.Printf("✓ qe-overlay diff: sin ediciones espurias contra el baseline upstream (%s).\n", shortSHA(baseline))
		for _, s := range skipped {
			fmt.Println("• " + s)
		}
		return 0
	}

	fmt.Printf("✗ qe-overlay diff: %d línea(s) espuria(s) detectadas contra el baseline upstream (%s):\n", len(findings), shortSHA(baseline))
	for _, f := range findings {
		fmt.Println("✗ " + f.String())
	}
	for _, s := range skipped {
		fmt.Println("• " + s)
	}
	return 1
}

// hasUpstreamRemote comprueba que el remote 'upstream' existe y que su rama main
// es resoluble localmente (requiere haber corrido `git fetch upstream` alguna vez).
func hasUpstreamRemote() bool {
	if err := runGit("remote", "get-url", "upstream"); err != nil {
		return false
	}
	return runGit("rev-parse", "--verify", "-q", "upstream/main") == nil
}

// resolveBaseline calcula el punto común entre HEAD y upstream/main. Usamos el
// merge-base (no upstream/main directo) porque el fork vive detrás de upstream/main:
// diffear contra upstream/main directo mezclaría commits del upstream posteriores
// a la base del fork con el diff real del overlay.
func resolveBaseline() (string, error) {
	out, err := gitOutput("merge-base", "HEAD", "upstream/main")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// diffAnchorFiles corre el diff real por cada archivo anclado (inlineAnchors +
// brandingAnchors) contra baseline y devuelve los hallazgos de líneas espurias.
// Los archivos net-new (sin contraparte en baseline, p.ej. internal/branding/branding.go)
// se saltean: el concepto de "edición espuria de upstream" no aplica a un archivo
// 100% propio del fork — ya lo cubre la sección de presencia de `verify`.
func diffAnchorFiles(m *manifest, baseline string) ([]diffFinding, []string, error) {
	mustContain := map[string][]string{}
	addAnchor := func(a anchor) {
		mustContain[a.File] = append(mustContain[a.File], a.MustContain)
	}
	var files []string
	seen := map[string]bool{}
	addFile := func(f string) {
		if !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	for _, a := range m.InlineAnchors {
		addAnchor(a)
		addFile(a.File)
	}
	for _, a := range m.BrandingAnchors {
		if a.MustContain == "" {
			continue
		}
		addAnchor(a)
		addFile(a.File)
	}
	sort.Strings(files)

	var findings []diffFinding
	var skipped []string
	for _, file := range files {
		if !fileExistsAtRef(baseline, file) {
			skipped = append(skipped, fmt.Sprintf("%s: net-new (sin contraparte en el baseline upstream), diff real no aplica", file))
			continue
		}
		hunks, err := diffHunks(baseline, file)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", file, err)
		}
		needles := mustContain[file]
		for _, h := range hunks {
			if hunkIsLegit(h, needles) {
				continue
			}
			for _, l := range h.added {
				if isNoiseLine(l.text) {
					continue
				}
				findings = append(findings, diffFinding{file: file, line: l.lineNo, text: l.text})
			}
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].file != findings[j].file {
			return findings[i].file < findings[j].file
		}
		return findings[i].line < findings[j].line
	})
	sort.Strings(skipped)
	return findings, skipped, nil
}

// hunkIsLegit decide si un hunk completo es legítimo: legítimo si CUALQUIERA de
// sus líneas agregadas contiene un mustContain registrado para el archivo, un
// marcador de ancla, o una referencia al paquete branding; o si el hunk entero es
// "ruido" (imports/líneas en blanco) sin contenido sustantivo.
//
// Agrupar a nivel de hunk (no de línea suelta) es una decisión deliberada: un
// cambio legítimo de anclaje suele tocar varias líneas contiguas (una guarda +
// su cuerpo, un comentario explicativo + la línea marcada), y solo UNA de ellas
// necesita llevar el marcador o el mustContain. Ver el README para el límite
// conocido de esta heurística (no es equivalente a revisar línea por línea).
func hunkIsLegit(h diffHunk, needles []string) bool {
	hasSubstance := false
	for _, l := range h.added {
		if isNoiseLine(l.text) {
			continue
		}
		hasSubstance = true
		if lineIsLegit(l.text, needles) {
			return true
		}
	}
	// Hunk sin ninguna línea sustantiva (solo imports/blancos): no hay nada que
	// pueda ser una edición espuria de lógica.
	return !hasSubstance
}

func lineIsLegit(text string, needles []string) bool {
	for _, n := range needles {
		if n != "" && strings.Contains(text, n) {
			return true
		}
	}
	for _, marker := range overlayMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return strings.Contains(text, brandingQualifier)
}

// isNoiseLine identifica líneas que nunca son "lógica espuria" por sí solas:
// líneas en blanco y líneas de import aisladas (si el import no se usa en
// ningún lado, el build revienta; si se usa, la línea de uso pasa por el
// chequeo de legitimidad normal).
func isNoiseLine(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" || t == "import (" || t == ")" {
		return true
	}
	return importLineRE.MatchString(t)
}

type diffHunkLine struct {
	lineNo int
	text   string
}

type diffHunk struct {
	added []diffHunkLine
}

var hunkHeaderRE = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// diffHunks corre `git diff -U<diffContextLines> baseline HEAD -- file` y parsea
// los hunks en líneas agregadas con su número de línea en el archivo nuevo (HEAD).
func diffHunks(baseline, file string) ([]diffHunk, error) {
	out, err := gitOutput("diff", fmt.Sprintf("-U%d", diffContextLines), baseline, "HEAD", "--", file)
	if err != nil {
		return nil, err
	}
	var hunks []diffHunk
	var cur *diffHunk
	newLine := 0
	for _, raw := range strings.Split(out, "\n") {
		if m := hunkHeaderRE.FindStringSubmatch(raw); m != nil {
			if cur != nil {
				hunks = append(hunks, *cur)
			}
			cur = &diffHunk{}
			fmt.Sscanf(m[1], "%d", &newLine)
			continue
		}
		if cur == nil {
			continue // preámbulo (diff --git, ---, +++)
		}
		if raw == "" {
			continue
		}
		switch raw[0] {
		case '+':
			if strings.HasPrefix(raw, "+++") {
				continue
			}
			cur.added = append(cur.added, diffHunkLine{lineNo: newLine, text: raw[1:]})
			newLine++
		case ' ':
			newLine++
		case '-':
			// línea vieja: no avanza el contador del archivo nuevo.
		case '\\':
			// "\ No newline at end of file": ignorar.
		}
	}
	if cur != nil {
		hunks = append(hunks, *cur)
	}
	return hunks, nil
}

// fileExistsAtRef comprueba si file existe en el árbol de ref.
func fileExistsAtRef(ref, file string) bool {
	return runGit("cat-file", "-e", ref+":"+file) == nil
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func runGit(args ...string) error {
	_, err := gitOutput(args...)
	return err
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
