package skills

import "testing"

// sdd_gating_qe_test.go — gate del override SDD test-design del overlay
// Gentle-QE. Net-new. Verifica que SetQETestDesignSDD apaga el override en el
// único choke point (QESDDTestingContent), haciendo que los 3 paths caigan
// fail-open al SDD de desarrollo upstream. Sin t.Parallel: el flag es global.

// TestQESDDTestingContent_GatedOffFallsOpenToDev: con el override ON (default)
// una skill SDD devuelve contenido QE; con el override OFF devuelve ok=false
// para que el caller sirva el asset dev upstream sin cambios.
func TestQESDDTestingContent_GatedOffFallsOpenToDev(t *testing.T) {
	t.Cleanup(func() { SetQETestDesignSDD(true) }) // restaura el default del paquete

	SetQETestDesignSDD(true)
	if _, ok := QESDDTestingContent("sdd-apply", "SKILL.md"); !ok {
		t.Fatalf("override ON: QESDDTestingContent(sdd-apply) ok=false, want true (QE content)")
	}

	SetQETestDesignSDD(false)
	if _, ok := QESDDTestingContent("sdd-apply", "SKILL.md"); ok {
		t.Fatalf("override OFF: QESDDTestingContent(sdd-apply) ok=true, want false (fail-open to upstream dev SDD)")
	}
}

// TestQESDDTestingContent_DefaultIsOn: el default del paquete preserva el
// comportamiento histórico (override incondicional) para los callers directos
// que no optan por apagarlo (tests directos + install SDET).
func TestQESDDTestingContent_DefaultIsOn(t *testing.T) {
	if !qeTestDesignSDDEnabled {
		t.Fatalf("qeTestDesignSDDEnabled default = false, want true")
	}
}
