package main

import (
	"strings"
	"testing"
)

// ---- unit tests puros (sin git) ----

func TestMissingAnchorsAnchorPresentIsNotMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tqeHook()\n}\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	got := missingAnchors(m)
	if len(got) != 0 {
		t.Fatalf("ancla presente no debe reportarse como perdida, tengo: %v", got)
	}
}

func TestMissingAnchorsDetectsAbsenceInlineAndBranding(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n")
	mustWrite(t, "pkg/brand.go", "package pkg\n\nconst X = 1\n")

	m := &manifest{
		InlineAnchors:   []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}},
		BrandingAnchors: []anchor{{File: "pkg/brand.go", MustContain: "branding.Product"}},
	}
	got := missingAnchors(m)
	if len(got) != 2 {
		t.Fatalf("quiero 2 anclas perdidas (inline + branding), tengo %d: %v", len(got), got)
	}
}

func TestMissingAnchorsIgnoresUnreadableFile(t *testing.T) {
	t.Chdir(t.TempDir())
	m := &manifest{InlineAnchors: []anchor{{File: "no/existe.go", MustContain: "qeHook"}}}
	got := missingAnchors(m)
	if len(got) != 0 {
		t.Fatalf("un archivo ilegible ya lo reporta verify (overlay roto), apply no debe duplicarlo: %v", got)
	}
}

func TestMissingAnchorsSkipsEmptyMustContain(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "pkg/brand.go", "package pkg\n")
	m := &manifest{BrandingAnchors: []anchor{{File: "pkg/brand.go", MustContain: ""}}}
	if got := missingAnchors(m); len(got) != 0 {
		t.Fatalf("un ancla sin mustContain (solo mustNotContain) no aplica a apply: %v", got)
	}
}

func TestIsSubstantiveContextLine(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"}", false},
		{")", false},
		{"   ", false},
		{"", false},
		{"x", false},
		{"\tif cfg.Preset == model.PresetQESDET {", true},
		{"return qeHook()", true},
	}
	for _, c := range cases {
		if got := isSubstantiveContextLine(c.line); got != c.want {
			t.Errorf("isSubstantiveContextLine(%q) = %v, quiero %v", c.line, got, c.want)
		}
	}
}

// ---- tests de integración (git real, repo temporal) ----

// TestBuildAnchorGuideFindsHistoricalLine simula el escenario del bead: un
// merge (aquí, un commit cualquiera) pierde la línea de ancla. buildAnchorGuide
// debe encontrar el último commit donde SÍ estaba, junto con su contexto.
func TestBuildAnchorGuideFindsHistoricalLine(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {\n\tif qeHook() { // ancla qe-overlay\n\t\treturn\n\t}\n\tdoUpstreamThing()\n}\n",
	})
	goodSHA := strings.TrimSpace(runGitT(t, "rev-parse", "HEAD"))
	// El merge "pierde" la línea de ancla (simula el conflicto real).
	commitChange(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n")

	a := anchor{File: "pkg/wire.go", MustContain: "qeHook"}
	guide, err := buildAnchorGuide(a)
	if err != nil {
		t.Fatalf("buildAnchorGuide: %v", err)
	}
	if guide.sha != goodSHA {
		t.Errorf("sha = %s, quiero el commit bueno %s", guide.sha, goodSHA)
	}
	if !strings.Contains(guide.anchorLine, "qeHook") {
		t.Errorf("anchorLine = %q, esperaba que contuviera qeHook", guide.anchorLine)
	}
	if len(guide.before) == 0 {
		t.Fatal("esperaba al menos una línea de contexto previa")
	}
	if !strings.Contains(guide.before[len(guide.before)-1], "func Setup") {
		t.Errorf("línea previa = %q, esperaba la firma de Setup", guide.before[len(guide.before)-1])
	}
	// El reporte formateado debe nombrar el archivo, el mustContain y el sha corto.
	report := guide.String()
	for _, want := range []string{a.File, a.MustContain, shortSHA(goodSHA)} {
		if !strings.Contains(report, want) {
			t.Errorf("el reporte debería mencionar %q, no lo hace:\n%s", want, report)
		}
	}
}

func TestBuildAnchorGuideNoHistoryErrors(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {}\n",
	})
	a := anchor{File: "pkg/wire.go", MustContain: "neverExisted"}
	if _, err := buildAnchorGuide(a); err == nil {
		t.Fatal("un mustContain que nunca existió en el historial debe devolver error")
	}
}

func TestMissingAnchorsRealRepoHasNothingToRestore(t *testing.T) {
	t.Chdir(repoRoot(t))
	m, err := loadManifest()
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if got := missingAnchors(m); len(got) != 0 {
		t.Fatalf("el repo real no debe tener anclas perdidas hoy, tengo: %v", got)
	}
}

func TestTryAutoRestoreUniqueContextInsertsAnchorLine(t *testing.T) {
	t.Chdir(t.TempDir())
	// tryAutoRestore no invoca git: opera directo sobre el archivo actual (ya
	// sin la línea de ancla, como quedaría tras un merge que la perdió) y la
	// guía ya construida (simulando lo que buildAnchorGuide devolvería).
	mustWrite(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n")

	a := anchor{File: "pkg/wire.go", MustContain: "qeHook"}
	g := anchorGuide{
		file:        a.File,
		mustContain: a.MustContain,
		sha:         "deadbeef",
		lineNo:      4,
		before:      []string{"func Setup() {"},
		anchorLine:  "\tif qeHook() { // ancla qe-overlay",
		after:       []string{"\t\treturn", "\t}"},
	}

	applied, reason := tryAutoRestore(a, g)
	if !applied {
		t.Fatalf("esperaba reinserción exitosa (contexto único), motivo: %s", reason)
	}

	got, err := readFile(a.File)
	if err != nil {
		t.Fatal(err)
	}
	want := "package pkg\n\nfunc Setup() {\n\tif qeHook() { // ancla qe-overlay\n\tdoUpstreamThing()\n}\n"
	if got != want {
		t.Errorf("contenido tras restore =\n%s\nquiero:\n%s", got, want)
	}
}

func TestTryAutoRestoreAmbiguousContextDoesNotWrite(t *testing.T) {
	t.Chdir(t.TempDir())
	original := "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n\nfunc Setup() {\n\tdoOtherThing()\n}\n"
	mustWrite(t, "pkg/wire.go", original)

	a := anchor{File: "pkg/wire.go", MustContain: "qeHook"}
	g := anchorGuide{
		before:     []string{"func Setup() {"}, // aparece 2 veces en el archivo actual
		anchorLine: "\tif qeHook() { // ancla qe-overlay",
	}

	applied, reason := tryAutoRestore(a, g)
	if applied {
		t.Fatal("un contexto ambiguo (2 matches) no debe reinsertar nada")
	}
	if !strings.Contains(reason, "2 veces") {
		t.Errorf("motivo = %q, esperaba que mencionara la ambigüedad", reason)
	}

	got, err := readFile(a.File)
	if err != nil {
		t.Fatal(err)
	}
	if got != original {
		t.Error("el archivo no debe modificarse cuando la reinserción es ambigua")
	}
}

func TestTryAutoRestoreTrivialContextDoesNotWrite(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n}\n")

	a := anchor{File: "pkg/wire.go", MustContain: "qeHook"}
	g := anchorGuide{
		before:     []string{"}"}, // línea trivial: no debe intentar ubicarla
		anchorLine: "\tqeHook()",
	}
	applied, reason := tryAutoRestore(a, g)
	if applied {
		t.Fatal("una línea de contexto trivial no debe usarse para reinsertar")
	}
	if reason == "" {
		t.Error("esperaba un motivo no vacío")
	}
}

func TestTryAutoRestoreAlreadyPresentDoesNotDuplicate(t *testing.T) {
	t.Chdir(t.TempDir())
	original := "package pkg\n\nfunc Setup() {\n\tqeHook()\n\tdoUpstreamThing()\n}\n"
	mustWrite(t, "pkg/wire.go", original)

	a := anchor{File: "pkg/wire.go", MustContain: "qeHook"}
	g := anchorGuide{
		before:     []string{"func Setup() {"},
		anchorLine: "\tqeHook()",
	}
	applied, reason := tryAutoRestore(a, g)
	if applied {
		t.Fatal("si la línea de ancla ya está en el punto de inserción, no debe duplicarse")
	}
	if !strings.Contains(reason, "ya está presente") {
		t.Errorf("motivo = %q", reason)
	}
	got, _ := readFile(a.File)
	if got != original {
		t.Error("el archivo no debe modificarse")
	}
}

// ---- runApply end-to-end (git real) ----

func TestRunApplyNothingMissingReturnsZero(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tqeHook()\n}\n")
	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	if code := runApply(m, false); code != 0 {
		t.Errorf("sin nada que restaurar, runApply debe devolver 0, tengo %d", code)
	}
}

func TestRunApplyDryRunReportsAndDoesNotWrite(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {\n\tif qeHook() { // ancla qe-overlay\n\t\treturn\n\t}\n\tdoUpstreamThing()\n}\n",
	})
	commitChange(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n")
	lostBody, err := readFile("pkg/wire.go")
	if err != nil {
		t.Fatal(err)
	}

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	code := runApply(m, false)
	if code != 1 {
		t.Errorf("con un ancla perdida y sin --write, runApply debe devolver 1, tengo %d", code)
	}
	got, err := readFile("pkg/wire.go")
	if err != nil {
		t.Fatal(err)
	}
	if got != lostBody {
		t.Error("sin --write, apply no debe escribir el archivo (solo reportar)")
	}
}

func TestRunApplyWriteRestoresUnambiguousAnchor(t *testing.T) {
	initRepoWithBaseline(t, map[string]string{
		"pkg/wire.go": "package pkg\n\nfunc Setup() {\n\tif qeHook() { // ancla qe-overlay\n\t\treturn\n\t}\n\tdoUpstreamThing()\n}\n",
	})
	commitChange(t, "pkg/wire.go", "package pkg\n\nfunc Setup() {\n\tdoUpstreamThing()\n}\n")

	m := &manifest{InlineAnchors: []anchor{{File: "pkg/wire.go", MustContain: "qeHook"}}}
	code := runApply(m, true)
	if code != 0 {
		t.Errorf("una reinserción no ambigua exitosa debe devolver 0, tengo %d", code)
	}
	got, err := readFile("pkg/wire.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "qeHook") {
		t.Errorf("el archivo debería tener la línea de ancla reinsertada, tengo:\n%s", got)
	}
	if missing := missingAnchors(m); len(missing) != 0 {
		t.Errorf("tras --write, el ancla no debería seguir reportándose como perdida: %v", missing)
	}
}

func TestRunApplyRealRepoNothingMissing(t *testing.T) {
	t.Chdir(repoRoot(t))
	m, err := loadManifest()
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if code := runApply(m, false); code != 0 {
		t.Errorf("el repo real no debe tener anclas perdidas hoy, runApply debería devolver 0, tengo %d", code)
	}
}
