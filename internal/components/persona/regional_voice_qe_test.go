package persona

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// regional_voice_qe_test.go — guarda del overlay Gentle-QE para la política de
// idioma del fork: la persona gentleman instalada espeja el registro español
// del usuario con fallback a español latinoamericano neutro (tuteo), nunca
// impone voseo rioplatense. Ver qeNeutralizeRegionalVoice en inject_qe.go.
//
// Archivo net-new del fork (el upstream nunca lo toca). Registrado en
// tools/qe-overlay/overlay.json.

const qeMirrorRegisterLine = "mirror the user's own Spanish register"

// gentlemanVoiceAssets son los assets upstream de gentleman que llevan la voz
// regional y pasan por qeNeutralizeRegionalVoice antes de instalarse.
//
// claude/persona-gentleman.md y kimi/persona-gentleman.md quedan afuera a
// propósito: desde el refactor "residual persona" del upstream (design.md
// Decision 1), Claude Code y Kimi mueven la directiva de idioma/tono al
// output-style twin (claude/output-style-gentleman.md,
// kimi/output-style-gentleman.md — ambos SÍ están en esta lista) y dejan la
// persona.md residual sin esa directiva. La neutralización se sigue aplicando
// sobre el contenido de persona.md en runtime (es incondicional en el call
// site), pero no hay "Rioplatense" que reemplazar ahí — cubrir el twin de
// output-style es suficiente.
var gentlemanVoiceAssets = []string{
	"claude/output-style-gentleman.md",
	"generic/persona-gentleman.md",
	"hermes/persona-gentleman.md",
	"kimi/output-style-gentleman.md",
	"kiro/persona-gentleman.md",
	"opencode/persona-gentleman.md",
}

// TestQENeutralizeRegionalVoiceCoversUpstreamAssets es la guarda de drift del
// sync: si el upstream reformula la directiva rioplatense, el mapa de
// reemplazos deja de matchear y este test falla señalando el asset exacto.
func TestQENeutralizeRegionalVoiceCoversUpstreamAssets(t *testing.T) {
	for _, path := range gentlemanVoiceAssets {
		t.Run(path, func(t *testing.T) {
			raw := assets.MustRead(path)
			if !strings.Contains(raw, "Rioplatense") {
				// El upstream quitó la voz regional de este asset: la
				// neutralización quedó sin trabajo aquí, pero conviene
				// revisar si el fork aún necesita este reemplazo.
				t.Fatalf("%s ya no contiene 'Rioplatense': el upstream cambió la voz regional; revisa qeRegionalVoiceReplacements y esta lista", path)
			}

			normalized := qeNeutralizeRegionalVoice(raw)
			if strings.Contains(normalized, "Rioplatense") {
				t.Fatalf("%s sigue conteniendo 'Rioplatense' tras neutralizar: el upstream reformuló la línea; actualiza qeRegionalVoiceReplacements en inject_qe.go", path)
			}
			if strings.Contains(raw, "use warm natural Rioplatense Spanish") && !strings.Contains(normalized, qeMirrorRegisterLine) {
				t.Fatalf("%s no recibió la política del fork (%q) tras neutralizar", path, qeMirrorRegisterLine)
			}
		})
	}
}

// TestInjectGentlemanInstallsNeutralLatamPolicy verifica de punta a punta que
// los archivos instalados con la persona gentleman llevan la política del fork
// y no la directiva rioplatense del upstream.
//
// Claude Code y Kimi tienen "residual channel" (design.md Decision 1): su
// persona.md/CLAUDE.md del system prompt ya viene residual/slim del upstream
// (sin la directiva de idioma) y la directiva vive solo en el output-style
// twin. Para esos dos, el system prompt solo se verifica con
// assertNoRioplatense (nunca debería tener la directiva vieja, pero tampoco
// se le exige la política nueva); el output-style twin sí lleva la
// verificación completa. OpenCode no tiene ese twin — su persona.md
// (AGENTS.md) sigue llevando la directiva completa, así que se verifica ahí.
func TestInjectGentlemanInstallsNeutralLatamPolicy(t *testing.T) {
	t.Run("claude-code", func(t *testing.T) {
		home := t.TempDir()
		if _, err := Inject(home, claudeAdapter(), model.PersonaGentleman); err != nil {
			t.Fatalf("Inject(claude, gentleman) error = %v", err)
		}
		assertNoRioplatense(t, filepath.Join(home, ".claude", "CLAUDE.md"))
		assertNeutralLatamPolicy(t, filepath.Join(home, ".claude", "output-styles", "gentleman.md"))
	})

	t.Run("opencode", func(t *testing.T) {
		home := t.TempDir()
		if _, err := Inject(home, opencodeAdapter(), model.PersonaGentleman); err != nil {
			t.Fatalf("Inject(opencode, gentleman) error = %v", err)
		}
		assertNeutralLatamPolicy(t, filepath.Join(home, ".config", "opencode", "AGENTS.md"))
	})

	t.Run("kimi", func(t *testing.T) {
		home := t.TempDir()
		if _, err := Inject(home, kimiAdapter(), model.PersonaGentleman); err != nil {
			t.Fatalf("Inject(kimi, gentleman) error = %v", err)
		}
		assertNoRioplatense(t, filepath.Join(home, ".kimi", "persona.md"))
		assertNeutralLatamPolicy(t, filepath.Join(home, ".kimi", "output-style.md"))
	})
}

// assertNoRioplatense verifies the installed file never carries the upstream
// Rioplatense directive, without requiring the fork's replacement policy line
// to be present. Used for residual system-prompt files (see comment on
// TestInjectGentlemanInstallsNeutralLatamPolicy) where the directive text
// lives elsewhere (the output-style twin), not in this file at all.
func assertNoRioplatense(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if strings.Contains(string(content), "Rioplatense") {
		t.Fatalf("%s instalado contiene 'Rioplatense'; la neutralización del overlay no se aplicó", path)
	}
}

func assertNeutralLatamPolicy(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	text := string(content)
	if strings.Contains(text, "Rioplatense") {
		t.Fatalf("%s instalado contiene 'Rioplatense'; la neutralización del overlay no se aplicó", path)
	}
	if !strings.Contains(text, qeMirrorRegisterLine) {
		t.Fatalf("%s instalado no contiene la política del fork %q", path, qeMirrorRegisterLine)
	}
	if !strings.Contains(text, "neutral Latin American Spanish") {
		t.Fatalf("%s instalado no contiene el fallback a español neutro LatAm", path)
	}
}
