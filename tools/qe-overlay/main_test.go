package main

import "testing"

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
