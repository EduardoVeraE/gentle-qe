package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanContentForBrand(t *testing.T) {
	forbidden := []string{"gentle-qa", "gentle_qa"}

	cases := []struct {
		name    string
		content string
		want    int
	}{
		{"marca vieja en frontmatter", "author: x (adapted for gentle-qa)", 1},
		{"gentle-ai es intencional, no es fuga", "module github.com/gentleman-programming/gentle-ai", 0},
		{"marca actual no es fuga", "Gentle-QE rocks, gentle-qe everywhere", 0},
		{"id de bead se ignora", "ver gentle-qa-i9p para el contexto", 0},
		{"case-insensitive", "GENTLE-QA en mayúsculas", 1},
		{"identificador con guion bajo", "export GENTLE_QA=1", 1},
		{"dos fugas en una línea", "gentle-qa y de nuevo gentle-qa", 2},
		{"id de bead junto a una fuga real", "gentle-qa-i9p pero también gentle-qa suelto", 1},
		{"contenido limpio", "nada que ver aquí", 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := scanContentForBrand(c.content, forbidden)
			if len(got) != c.want {
				t.Fatalf("scanContentForBrand(%q) = %d hits, quiero %d (%+v)", c.content, len(got), c.want, got)
			}
		})
	}
}

func TestScanContentForBrandReportsLineAndText(t *testing.T) {
	content := "línea uno\nauthor: adapted for gentle-qa\nlínea tres"
	got := scanContentForBrand(content, []string{"gentle-qa"})
	if len(got) != 1 {
		t.Fatalf("quiero 1 hit, tengo %d (%+v)", len(got), got)
	}
	if got[0].line != 2 {
		t.Errorf("línea = %d, quiero 2", got[0].line)
	}
	if got[0].text != "gentle-qa" {
		t.Errorf("text = %q, quiero %q", got[0].text, "gentle-qa")
	}
}

func TestBrandScanTargets(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "skills/qa/SKILL.md", "x")
	mustWrite(t, "skills/qa/sub/ref.md", "x")
	mustWrite(t, "generic/persona-sdet.md", "x")
	mustWrite(t, "generic/notmd.go", "x") // overlayFile no-markdown: se ignora

	m := &manifest{NetNewDirs: []string{"skills/qa", "skills/falta"}}
	m.OverlayFiles = []string{"generic/persona-sdet.md", "generic/notmd.go"}

	got, err := brandScanTargets(m)
	if err != nil {
		t.Fatal(err)
	}
	// Espera el contenido del dir net-new (recursivo) + solo el .md del overlay;
	// el dir faltante no rompe, el .go se excluye.
	if len(got) != 3 {
		t.Fatalf("quiero 3 targets, tengo %d: %v", len(got), got)
	}
}

func TestDetectBrandLeaks(t *testing.T) {
	t.Chdir(t.TempDir())
	mustWrite(t, "skills/qa/SKILL.md", "author: adapted for gentle-qa\n")
	mustWrite(t, "skills/qa/clean.md", "todo gentle-qe aquí\n")
	mustWrite(t, "generic/persona-sdet.md", "persona ok, ver bead gentle-qa-i9p\n")

	m := &manifest{NetNewDirs: []string{"skills/qa"}}
	m.OverlayFiles = []string{"generic/persona-sdet.md"}
	m.BrandLeak.Forbidden = []string{"gentle-qa", "gentle_qa"}

	leaks, err := detectBrandLeaks(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(leaks) != 1 {
		t.Fatalf("quiero 1 fuga, tengo %d: %v", len(leaks), leaks)
	}
	if !strings.Contains(leaks[0], "SKILL.md") {
		t.Errorf("la fuga debería apuntar a SKILL.md, no: %q", leaks[0])
	}
}

func TestDetectBrandLeaksSinForbidden(t *testing.T) {
	m := &manifest{NetNewDirs: []string{"."}}
	leaks, err := detectBrandLeaks(m)
	if err != nil {
		t.Fatal(err)
	}
	if leaks != nil {
		t.Fatalf("sin forbidden no debe escanear nada, tengo: %v", leaks)
	}
}
