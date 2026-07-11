package uninstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// --- Path 4: generic skills uninstall must never remove the _qe-sdd overlay ---
//
// The _qe-sdd overlay directory is a build-time-only asset (never installed
// on disk as its own skill directory), but the generic ComponentSkills
// uninstall path walks every directory name under the embedded "skills/" tree
// and must defensively skip "_qe-sdd" exactly like it skips "_shared" and any
// "sdd-*" directory. Before this test, that protection was only checked by
// tools/qe-overlay's static "mustContain" grep on service.go — a real bug
// that removed the skip condition (or renamed it) would still pass that
// check. This test exercises the real op.apply(...) filesystem operations.

func TestComponentOperationsSkills_PreservesQESDDOverlayDirectory(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()

	svc, err := NewService(homeDir, workspaceDir, "dev")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter, ok := svc.registry.Get(model.AgentClaudeCode)
	if !ok {
		t.Fatal("claude adapter not found in registry")
	}

	skillsDir := adapter.SkillsDir(homeDir)

	// A real, installable skill directory that the generic uninstall SHOULD remove.
	goTestingDir := filepath.Join(skillsDir, "go-testing")
	if err := os.MkdirAll(goTestingDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(go-testing) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(goTestingDir, "SKILL.md"), []byte("go testing skill"), 0o644); err != nil {
		t.Fatalf("WriteFile(go-testing/SKILL.md) error = %v", err)
	}

	// The QE overlay directory, simulated on disk. The generic skills
	// uninstall must never remove this — it must be skipped just like it
	// skips "_shared" and "sdd-*" directories.
	qeOverlayDir := filepath.Join(skillsDir, "_qe-sdd")
	if err := os.MkdirAll(qeOverlayDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(_qe-sdd) error = %v", err)
	}
	qeOverlayFile := filepath.Join(qeOverlayDir, "apply.md")
	if err := os.WriteFile(qeOverlayFile, []byte("qe apply override"), 0o644); err != nil {
		t.Fatalf("WriteFile(_qe-sdd/apply.md) error = %v", err)
	}

	ops, _, err := svc.componentOperations(adapter, model.ComponentSkills)
	if err != nil {
		t.Fatalf("componentOperations(ComponentSkills) error = %v", err)
	}
	if len(ops) == 0 {
		t.Fatal("componentOperations(ComponentSkills) returned no operations — test setup invalid")
	}

	for _, op := range ops {
		if _, _, err := op.apply(op.path); err != nil {
			t.Fatalf("op.apply(%q) error = %v", op.path, err)
		}
	}

	if _, err := os.Stat(goTestingDir); !os.IsNotExist(err) {
		t.Fatalf("go-testing skill dir should have been removed by the real uninstall op, stat err = %v", err)
	}
	if _, err := os.Stat(qeOverlayFile); err != nil {
		t.Fatalf("_qe-sdd overlay file must survive a real skills uninstall, stat err = %v", err)
	}
	if _, err := os.Stat(qeOverlayDir); err != nil {
		t.Fatalf("_qe-sdd overlay directory must survive a real skills uninstall, stat err = %v", err)
	}
}

// TestComponentOperationsSDD_PreservesQESDDOverlayDirectory covers the
// SDD-component uninstall path (managedSDDSkillIDs — sdd-* phase dirs plus
// judgment-day) with the same real-filesystem assertion: a manually placed
// _qe-sdd directory alongside the managed SDD skill directories must survive
// a full SDD component uninstall, since _qe-sdd is never one of the managed
// SDD skill IDs.
func TestComponentOperationsSDD_PreservesQESDDOverlayDirectory(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()

	svc, err := NewService(homeDir, workspaceDir, "dev")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter, ok := svc.registry.Get(model.AgentClaudeCode)
	if !ok {
		t.Fatal("claude adapter not found in registry")
	}

	skillsDir := adapter.SkillsDir(homeDir)

	sddApplyDir := filepath.Join(skillsDir, "sdd-apply")
	if err := os.MkdirAll(sddApplyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(sdd-apply) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sddApplyDir, "SKILL.md"), []byte("sdd apply skill"), 0o644); err != nil {
		t.Fatalf("WriteFile(sdd-apply/SKILL.md) error = %v", err)
	}

	qeOverlayDir := filepath.Join(skillsDir, "_qe-sdd")
	if err := os.MkdirAll(qeOverlayDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(_qe-sdd) error = %v", err)
	}
	qeOverlayFile := filepath.Join(qeOverlayDir, "apply.md")
	if err := os.WriteFile(qeOverlayFile, []byte("qe apply override"), 0o644); err != nil {
		t.Fatalf("WriteFile(_qe-sdd/apply.md) error = %v", err)
	}

	ops, _, err := svc.componentOperations(adapter, model.ComponentSDD)
	if err != nil {
		t.Fatalf("componentOperations(ComponentSDD) error = %v", err)
	}

	for _, op := range ops {
		if _, _, err := op.apply(op.path); err != nil {
			t.Fatalf("op.apply(%q) error = %v", op.path, err)
		}
	}

	if _, err := os.Stat(sddApplyDir); !os.IsNotExist(err) {
		t.Fatalf("sdd-apply skill dir should have been removed by the real SDD uninstall op, stat err = %v", err)
	}
	if _, err := os.Stat(qeOverlayFile); err != nil {
		t.Fatalf("_qe-sdd overlay file must survive a real SDD component uninstall, stat err = %v", err)
	}
	if _, err := os.Stat(qeOverlayDir); err != nil {
		t.Fatalf("_qe-sdd overlay directory must survive a real SDD component uninstall, stat err = %v", err)
	}
}
