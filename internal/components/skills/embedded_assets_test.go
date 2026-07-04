package skills

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
)

// TestEmbeddedSkillAssetsAreNonEmpty guards against zero-byte files (e.g. a
// stray .gitkeep, .gitignore or .DS_Store) sneaking into the embedded skills
// tree via `//go:embed all:skills`. Inject aborts the entire skills component
// on the first empty asset (see inject.go: "embedded asset %q is empty"), so a
// single empty file breaks installation for every agent. Walking the real
// embedded FS here catches that at CI time instead of at install time.
func TestEmbeddedSkillAssetsAreNonEmpty(t *testing.T) {
	var empty []string
	err := fs.WalkDir(assets.FS, "skills", func(assetPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		content, readErr := assets.Read(assetPath)
		if readErr != nil {
			t.Fatalf("read embedded asset %q: %v", assetPath, readErr)
		}
		if len(content) == 0 {
			empty = append(empty, assetPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded skills: %v", err)
	}
	if len(empty) > 0 {
		t.Fatalf("embedded skills contain %d empty asset(s), which abort skill injection at install time:\n  %s",
			len(empty), strings.Join(empty, "\n  "))
	}
}

// TestEmbeddedSkillAssetsInjectableSanity walks every embedded skill directory
// the way Inject does and asserts each one is copyable end-to-end. It mirrors
// the failure the installer hit for "qa-owasp-security" so a regression in any
// skill (not just that one) is caught here.
func TestEmbeddedSkillAssetsInjectableSanity(t *testing.T) {
	entries, err := fs.ReadDir(assets.FS, "skills")
	if err != nil {
		t.Fatalf("read embedded skills dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := path.Join("skills", e.Name())
		err := fs.WalkDir(assets.FS, skillDir, func(assetPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			content, readErr := assets.Read(assetPath)
			if readErr != nil {
				return readErr
			}
			if len(content) == 0 {
				t.Errorf("skill %q: embedded asset %q is empty", e.Name(), assetPath)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("skill %q: walk embedded assets: %v", e.Name(), err)
		}
	}
}
