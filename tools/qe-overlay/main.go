// Command qe-overlay verifica y mantiene el overlay SDET/QE (Gentle-QE) sobre el
// upstream gentleman-programming/gentle-ai.
//
// Modos:
//
//	go run ./tools/qe-overlay verify   Chequea que el overlay siga intacto tras un
//	                                   merge del upstream y reporta drift (skills
//	                                   nuevos del upstream, anclas reescritas).
//	go run ./tools/qe-overlay accept   Absorbe el drift detectado actualizando el
//	                                   manifiesto (known upstream skills).
//
// No tiene dependencias externas: lee tools/qe-overlay/overlay.json con la stdlib.
// Está pensado para correrse desde la raíz del repo (donde está go.mod).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const manifestPath = "tools/qe-overlay/overlay.json"

type anchor struct {
	File           string `json:"file"`
	MustContain    string `json:"mustContain"`
	MustNotContain string `json:"mustNotContain,omitempty"`
}

type manifest struct {
	Comment         string   `json:"_comment,omitempty"`
	NetNewDirs      []string `json:"netNewDirs"`
	OverlayFiles    []string `json:"overlayFiles"`
	BrandingAnchors []anchor `json:"brandingAnchors"`
	InlineAnchors   []anchor `json:"inlineAnchors"`
	UpstreamSkills  struct {
		Dir   string   `json:"dir"`
		Known []string `json:"known"`
	} `json:"upstreamSkills"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: qe-overlay <verify|accept>")
		os.Exit(2)
	}

	m, err := loadManifest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo el manifiesto: %v\n", err)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "verify":
		os.Exit(runVerify(m))
	case "accept":
		os.Exit(runAccept(m))
	default:
		fmt.Fprintf(os.Stderr, "modo desconocido %q (usa verify|accept)\n", os.Args[1])
		os.Exit(2)
	}
}

func loadManifest() (*manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", manifestPath, err)
	}
	return &m, nil
}

// runVerify devuelve 0 si el overlay está intacto, 1 si hay hallazgos.
func runVerify(m *manifest) int {
	var problems []string
	var notes []string

	// 1. Directorios net-new presentes.
	for _, d := range m.NetNewDirs {
		if !isDir(d) {
			problems = append(problems, "falta directorio net-new (overlay roto): "+d)
		}
	}

	// 2. Archivos del overlay presentes.
	for _, f := range m.OverlayFiles {
		if !isFile(f) {
			problems = append(problems, "falta archivo del overlay (overlay roto): "+f)
		}
	}

	// 3. Anclas de branding.
	for _, a := range m.BrandingAnchors {
		body, err := readFile(a.File)
		if err != nil {
			problems = append(problems, "ancla de branding ilegible: "+a.File+" ("+err.Error()+")")
			continue
		}
		if a.MustContain != "" && !strings.Contains(body, a.MustContain) {
			problems = append(problems, fmt.Sprintf("branding perdido en %s: falta %q (¿un merge re-introdujo gentle-ai? apuntar a branding.*)", a.File, a.MustContain))
		}
		if a.MustNotContain != "" && strings.Contains(body, a.MustNotContain) {
			problems = append(problems, fmt.Sprintf("branding revertido en %s: reapareció %q", a.File, a.MustNotContain))
		}
	}

	// 4. Anclas inline (líneas de delegación que upstream podría reescribir).
	for _, a := range m.InlineAnchors {
		body, err := readFile(a.File)
		if err != nil {
			problems = append(problems, "ancla inline ilegible: "+a.File+" ("+err.Error()+")")
			continue
		}
		if !strings.Contains(body, a.MustContain) {
			problems = append(problems, fmt.Sprintf("delegación inline perdida en %s: falta %q (re-aplicar la línea de delegación del overlay)", a.File, a.MustContain))
		}
	}

	// 5. Drift del upstream: skills nuevos no clasificados.
	newSkills, err := detectNewUpstreamSkills(m)
	if err != nil {
		problems = append(problems, "no se pudo escanear skills: "+err.Error())
	}
	for _, s := range newSkills {
		notes = append(notes, fmt.Sprintf("skill nuevo del upstream sin clasificar: %q — decidí si va a algún preset QE, luego corré `qe-overlay accept`", s))
	}

	// Reporte.
	if len(problems) == 0 && len(notes) == 0 {
		fmt.Println("✓ qe-overlay: overlay intacto, sin drift.")
		return 0
	}
	for _, p := range problems {
		fmt.Println("✗ " + p)
	}
	for _, n := range notes {
		fmt.Println("• " + n)
	}
	if len(problems) > 0 {
		return 1
	}
	// Solo notas (drift no bloqueante): salida 0 pero visible.
	return 0
}

// runAccept absorbe los skills nuevos del upstream en el manifiesto (known).
func runAccept(m *manifest) int {
	newSkills, err := detectNewUpstreamSkills(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo escanear skills: %v\n", err)
		return 2
	}
	if len(newSkills) == 0 {
		fmt.Println("✓ qe-overlay: nada que aceptar (sin skills nuevos del upstream).")
		return 0
	}

	m.UpstreamSkills.Known = append(m.UpstreamSkills.Known, newSkills...)
	sort.Strings(m.UpstreamSkills.Known)

	if err := saveManifest(m); err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo guardar el manifiesto: %v\n", err)
		return 2
	}
	fmt.Printf("✓ qe-overlay: %d skill(s) upstream agregados a known: %s\n", len(newSkills), strings.Join(newSkills, ", "))
	return 0
}

// detectNewUpstreamSkills devuelve los directorios bajo upstreamSkills.Dir que no
// son ni QE (netNewDirs) ni upstream conocidos (known).
func detectNewUpstreamSkills(m *manifest) ([]string, error) {
	entries, err := os.ReadDir(m.UpstreamSkills.Dir)
	if err != nil {
		return nil, err
	}

	known := map[string]struct{}{}
	for _, k := range m.UpstreamSkills.Known {
		known[k] = struct{}{}
	}
	qe := map[string]struct{}{}
	for _, d := range m.NetNewDirs {
		qe[filepath.Base(d)] = struct{}{}
	}

	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if _, ok := known[name]; ok {
			continue
		}
		if _, ok := qe[name]; ok {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func saveManifest(m *manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(manifestPath, data, 0o644)
}

func isDir(p string) bool  { fi, err := os.Stat(p); return err == nil && fi.IsDir() }
func isFile(p string) bool { fi, err := os.Stat(p); return err == nil && !fi.IsDir() }

func readFile(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
