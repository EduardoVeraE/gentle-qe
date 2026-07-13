package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCodeGraphInitValidatesCanonicalProjectAndPropagatesInitFailure(t *testing.T) {
	// Canonicalize workspace immediately: on macOS, t.TempDir() returns a path
	// under /var/folders, a symlink to /private/var/folders. Production code
	// (canonicalCodeGraphProjectRoot) resolves that symlink via
	// filepath.EvalSymlinks before comparing/passing paths around, so this
	// test's own expectations must be built from the same canonical form or
	// they mismatch on any machine where TMPDIR is symlinked — unrelated to
	// the QE overlay, this is a portability gap in the plain string path.
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks(workspace) error = %v", err)
	}
	root := filepath.Join(workspace, "project")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}

	originalRoot := codeGraphGitTopLevel
	originalInit := codeGraphInit
	originalHome := codeGraphUserHomeDir
	originalTemp := codeGraphTempDir
	t.Cleanup(func() {
		codeGraphGitTopLevel = originalRoot
		codeGraphInit = originalInit
		codeGraphUserHomeDir = originalHome
		codeGraphTempDir = originalTemp
	})
	codeGraphGitTopLevel = func(path string) (string, error) {
		if path != root {
			t.Fatalf("git root path = %q, want %q", path, root)
		}
		return root, nil
	}
	codeGraphUserHomeDir = func() (string, error) { return filepath.Join(workspace, "home"), nil }
	codeGraphTempDir = func() string { return filepath.Join(workspace, "temporary") }

	var output bytes.Buffer
	var called []string
	codeGraphInit = func(name string, args ...string) error {
		called = append([]string{name}, args...)
		return nil
	}
	if err := RunCodeGraph([]string{"init", "--cwd", root}, &output); err != nil {
		t.Fatalf("RunCodeGraph() error = %v", err)
	}
	if got, want := strings.Join(called, " "), "codegraph init "+root; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
	if !strings.Contains(output.String(), root) {
		t.Fatalf("output = %q, want canonical root", output.String())
	}

	codeGraphInit = func(string, ...string) error { return errors.New("init failed") }
	if err := RunCodeGraph([]string{"init", "--cwd", root}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "init failed") {
		t.Fatalf("subprocess error = %v, want propagated init failure", err)
	}
}

func TestRunCodeGraphInitRejectsUnsafeOrUnrecognizedRoots(t *testing.T) {
	workspace := t.TempDir()
	home := filepath.Join(workspace, "home")
	temp := filepath.Join(workspace, "temporary")
	for _, path := range []string{home, temp} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	outside := filepath.Join(temp, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(workspace, "escape")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Fatal(err)
	}

	originalRoot := codeGraphGitTopLevel
	originalInit := codeGraphInit
	originalHome := codeGraphUserHomeDir
	originalTemp := codeGraphTempDir
	t.Cleanup(func() {
		codeGraphGitTopLevel = originalRoot
		codeGraphInit = originalInit
		codeGraphUserHomeDir = originalHome
		codeGraphTempDir = originalTemp
	})
	codeGraphGitTopLevel = func(path string) (string, error) {
		if path == filepath.Join(workspace, "not-a-project") {
			return "", errors.New("not a git repository")
		}
		return path, nil
	}
	codeGraphUserHomeDir = func() (string, error) { return home, nil }
	codeGraphTempDir = func() string { return temp }
	codeGraphInit = func(string, ...string) error { t.Fatal("codegraph init must not run for rejected roots"); return nil }

	for _, path := range []string{"", string(filepath.Separator), home, temp, outside, symlink, filepath.Join(workspace, "not-a-project")} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if err := RunCodeGraph([]string{"init", "--cwd", path}, &bytes.Buffer{}); err == nil {
				t.Fatalf("RunCodeGraph(%q) error = nil, want rejection", path)
			}
		})
	}
}

func TestRunCodeGraphInitAcceptsProjectBelowHome(t *testing.T) {
	// See the canonicalization comment in
	// TestRunCodeGraphInitValidatesCanonicalProjectAndPropagatesInitFailure:
	// production canonicalizes both root and home via filepath.EvalSymlinks
	// before the "is root below home" check, so this test's own home/root
	// values must be derived from an already-canonical workspace.
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks(workspace) error = %v", err)
	}
	home := filepath.Join(workspace, "home")
	root := filepath.Join(home, "work", "project-feature")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	originalRoot := codeGraphGitTopLevel
	originalInit := codeGraphInit
	originalHome := codeGraphUserHomeDir
	originalTemp := codeGraphTempDir
	t.Cleanup(func() {
		codeGraphGitTopLevel = originalRoot
		codeGraphInit = originalInit
		codeGraphUserHomeDir = originalHome
		codeGraphTempDir = originalTemp
	})
	codeGraphGitTopLevel = func(path string) (string, error) { return path, nil }
	codeGraphUserHomeDir = func() (string, error) { return home, nil }
	codeGraphTempDir = func() string { return filepath.Join(workspace, "temporary") }

	called := false
	codeGraphInit = func(name string, args ...string) error {
		called = name == "codegraph" && len(args) == 2 && args[0] == "init" && args[1] == root
		return nil
	}
	if err := RunCodeGraph([]string{"init", "--cwd", root}, &bytes.Buffer{}); err != nil {
		t.Fatalf("RunCodeGraph() error = %v", err)
	}
	if !called {
		t.Fatal("codegraph init was not called for a project below HOME")
	}
}
