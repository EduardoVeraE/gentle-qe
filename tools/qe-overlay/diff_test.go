package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ---- unit tests puros (sin git) ----

func TestIsNoiseLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"blank", "", true},
		{"whitespace only", "   ", true},
		{"import open paren", "import (", true},
		{"import close paren", ")", true},
		{"plain import path", `	"github.com/gentleman-programming/gentle-ai/internal/branding"`, true},
		{"blank import alias", `	_ "embed"`, true},
		{"aliased import", `	branding "github.com/foo/branding"`, true},
		{"real code line", `	return branding.Product`, false},
		{"string used mid-expression is not import-shaped", `	x := "hola" + branding.Product`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isNoiseLine(c.line); got != c.want {
				t.Errorf("isNoiseLine(%q) = %v, quiero %v", c.line, got, c.want)
			}
		})
	}
}

func TestLineIsLegit(t *testing.T) {
	needles := []string{"qeSwapNativeAgentBody"}
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"mustContain match", `contentStr = qeSwapNativeAgentBody(a, b, c)`, true},
		{"marker ancla", `if x { // overlay Gentle-QE (ancla qe-overlay)`, true},
		{"marker overlay solo", `// (ancla qe-overlay)`, true},
		{"branding qualifier", `req.Header.Set("User-Agent", branding.Product+"-x")`, true},
		{"ninguno de los anteriores", `subAgentCapability := "capable"`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := lineIsLegit(c.line, needles); got != c.want {
				t.Errorf("lineIsLegit(%q) = %v, quiero %v", c.line, got, c.want)
			}
		})
	}
}

func TestHunkIsLegit(t *testing.T) {
	needles := []string{"qeSwapNativeAgentBody"}

	legit := diffHunk{added: []diffHunkLine{
		{lineNo: 1, text: `	subAgentCapability := "capable"`},
		{lineNo: 2, text: `	contentStr = qeSwapNativeAgentBody(a, b, c)`},
	}}
	if !hunkIsLegit(legit, needles) {
		t.Error("un hunk con una línea de ancla en cualquier posición debe ser legítimo")
	}

	spurious := diffHunk{added: []diffHunkLine{
		{lineNo: 1, text: `	subAgentCapability := "capable"`},
		{lineNo: 2, text: `	os.Setenv("BACKDOOR", "1")`},
	}}
	if hunkIsLegit(spurious, needles) {
		t.Error("un hunk sin ninguna línea de ancla/branding no debe ser legítimo")
	}

	importOnly := diffHunk{added: []diffHunkLine{
		{lineNo: 1, text: ""},
		{lineNo: 2, text: `	"fmt"`},
	}}
	if !hunkIsLegit(importOnly, needles) {
		t.Error("un hunk de solo imports/blancos (sin contenido sustantivo) debe ser legítimo")
	}
}

// ---- helpers para tests con git real ----

func runGitT(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// initRepoWithBaseline crea un repo git temporal con un commit baseline que
// contiene los archivos dados. Devuelve el SHA del commit baseline.
func initRepoWithBaseline(t *testing.T, files map[string]string) string {
	t.Helper()
	t.Chdir(t.TempDir())
	runGitT(t, "init", "-q")
	runGitT(t, "config", "commit.gpgsign", "false")
	for path, content := range files {
		mustWrite(t, path, content)
	}
	runGitT(t, "add", "-A")
	runGitT(t, "commit", "-q", "-m", "baseline")
	return strings.TrimSpace(runGitT(t, "rev-parse", "HEAD"))
}

func commitChange(t *testing.T, path, content string) {
	t.Helper()
	mustWrite(t, path, content)
	runGitT(t, "add", "-A")
	runGitT(t, "commit", "-q", "-m", "head change")
}

// ---- tests de integración (git real, repo temporal) ----

func TestDiffAnchorFilesAnchorOnlyPasses(t *testing.T) {
	baseline := initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n",
	})
	commitChange(t, "pkg/wire.go",
		"package pkg\n\nfunc Setup() {\n\tif qeHook() { // overlay Gentle-QE (ancla qe-overlay)\n\t\treturn\n\t}\n\tdoUpstreamThing()\n}\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	findings, skipped, err := diffAnchorFiles(m, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("un archivo con solo líneas de ancla no debe reportar hallazgos, tengo: %v", findings)
	}
	if len(skipped) != 0 {
		t.Fatalf("no debería haber archivos salteados, tengo: %v", skipped)
	}
}

func TestDiffAnchorFilesSpuriousLineFails(t *testing.T) {
	// filler es contenido IDÉNTICO en baseline y HEAD, lo bastante largo
	// (> 2*diffContextLines líneas) como para que git diff mantenga separados
	// el hunk de la línea de ancla legítima y el hunk de la línea espuria: así
	// la espuria no se "cuela" bajo el paraguas de legitimidad del hunk vecino.
	filler := strings.Repeat("func filler() {}\n\n", 15)
	baseline := initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n\n" + filler +
			"func Extra() {\n\tdoOtherUpstreamThing()\n}\n",
	})
	commitChange(t, "pkg/wire.go",
		"package pkg\n\nfunc Setup() {\n\tif qeHook() { // overlay Gentle-QE (ancla qe-overlay)\n\t\treturn\n\t}\n\tdoUpstreamThing()\n}\n\n"+filler+
			"func Extra() {\n\tos.Setenv(\"BACKDOOR\", \"1\")\n\tdoOtherUpstreamThing()\n}\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	findings, _, err := diffAnchorFiles(m, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("quiero 1 hallazgo por la línea espuria, tengo %d: %v", len(findings), findings)
	}
	if findings[0].file != "pkg/wire.go" {
		t.Errorf("archivo = %q, quiero pkg/wire.go", findings[0].file)
	}
	if !strings.Contains(findings[0].text, "BACKDOOR") {
		t.Errorf("el hallazgo debe apuntar a la línea espuria, no: %q", findings[0].text)
	}
	// La línea reportada debe ser la línea real en el archivo HEAD.
	headBody, err := readFile("pkg/wire.go")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(headBody, "\n")
	if findings[0].line < 1 || findings[0].line > len(lines) {
		t.Fatalf("línea reportada %d fuera de rango (%d líneas)", findings[0].line, len(lines))
	}
	if !strings.Contains(lines[findings[0].line-1], "BACKDOOR") {
		t.Errorf("línea:%d en HEAD = %q, esperaba que contuviera BACKDOOR", findings[0].line, lines[findings[0].line-1])
	}
}

func TestDiffAnchorFilesNetNewFileSkipped(t *testing.T) {
	baseline := initRepoWithBaseline(t, map[string]string{
		"README.md": "baseline\n",
	})
	commitChange(t, "pkg/netnew.go", "package pkg\n\nconst X = 1\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/netnew.go", MustContain: "X"}}}
	findings, skipped, err := diffAnchorFiles(m, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("un archivo net-new no debe reportar hallazgos, tengo: %v", findings)
	}
	if len(skipped) != 1 || !strings.Contains(skipped[0], "net-new") {
		t.Fatalf("quiero 1 nota de net-new, tengo: %v", skipped)
	}
}

func TestDiffAnchorFilesBrandingReferenceIsLegit(t *testing.T) {
	baseline := initRepoWithBaseline(t, map[string]string{
		"pkg/msg.go": "package pkg\n\nfunc Msg() string {\n\treturn \"gentle-ai\"\n}\n",
	})
	// Sin ancla marcada ni mustContain de ese archivo, pero referencia branding.*:
	// debe ser legítimo por la regla del calificador branding.
	commitChange(t, "pkg/msg.go",
		"package pkg\n\nfunc Msg() string {\n\treturn branding.Product + \" ok\"\n}\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/msg.go", MustContain: "algoQueNoAparece"}}}
	findings, _, err := diffAnchorFiles(m, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("una línea que referencia branding.* debe ser legítima, tengo: %v", findings)
	}
}

func TestHasUpstreamRemoteFalseWithoutRemote(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{"a.txt": "x\n"})
	if hasUpstreamRemote() {
		t.Error("sin remote 'upstream' configurado, hasUpstreamRemote debe ser false")
	}
}

func TestResolveBaselineErrorsWithoutUpstreamBranch(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{"a.txt": "x\n"})
	if _, err := resolveBaseline(); err == nil {
		t.Error("sin upstream/main resoluble, resolveBaseline debe devolver error")
	}
}

func TestRunDiffDegradesGracefullyWithoutUpstream(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{"a.txt": "x\n"})
	m := &manifest{}
	if code := runDiff(m); code != 0 {
		t.Errorf("sin remote upstream, runDiff debe degradar con exit 0, tengo %d", code)
	}
}

// El diff real contra el repo real (esta checkout) debe pasar sin hallazgos, y
// debe correr sin pánico incluso si el remote 'upstream' no está disponible en el
// entorno donde corre el test (CI sin ese remote configurado).
func TestDiffAnchorFilesRealRepo(t *testing.T) {
	t.Chdir(repoRoot(t))
	if !hasUpstreamRemote() {
		t.Skip("remote 'upstream' no disponible en este entorno; el diff real requiere `git fetch upstream`")
	}
	baseline, err := resolveBaseline()
	if err != nil {
		t.Fatalf("resolveBaseline: %v", err)
	}
	m, err := loadManifest()
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	findings, _, err := diffAnchorFiles(m, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("el repo real no debe tener ediciones espurias contra el baseline upstream, tengo: %v", findings)
	}
}
