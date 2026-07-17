package styles

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// RenderLogoFit picks the full brand scene only when the terminal fits it
// unwrapped; anything smaller (or unknown, i.e. zero) must fall back to the
// plain wordmark so Welcome never overflows short/narrow terminals.
func TestRenderLogoFit_FallsBackBelowThresholds(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		wantScene     bool
	}{
		{"unknown size", 0, 0, false},
		{"narrow terminal", qeBrandSceneMinWidth - 1, qeBrandSceneMinHeight, false},
		{"short terminal", qeBrandSceneMinWidth, qeBrandSceneMinHeight - 1, false},
		{"exact minimum", qeBrandSceneMinWidth, qeBrandSceneMinHeight, true},
		{"large terminal", 120, 60, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := RenderLogoFit(tc.width, tc.height)
			wantLines := len(qeLogoLines)
			if tc.wantScene {
				wantLines += len(qeBrandSceneLines)
			}
			if got := lipgloss.Height(out); got != wantLines {
				t.Fatalf("RenderLogoFit(%d, %d) height = %d, want %d", tc.width, tc.height, got, wantLines)
			}
			if got := strings.Contains(out, "Q U A L I T Y"); !got {
				t.Fatalf("RenderLogoFit(%d, %d) lost the wordmark tagline", tc.width, tc.height)
			}
		})
	}
}

// The scene must stay narrower than the Welcome frame budget that
// qeBrandSceneMinWidth promises (welcomeContentWidth subtracts 10 for frame
// chrome), otherwise lipgloss wraps it and the height math above lies.
func TestRenderLogoFit_SceneFitsPromisedWidth(t *testing.T) {
	const frameChrome = 10
	budget := qeBrandSceneMinWidth - frameChrome
	for i, line := range qeBrandSceneLines {
		if w := lipgloss.Width(line); w > budget {
			t.Fatalf("scene line %d width = %d, want <= %d", i, w, budget)
		}
	}
	for i, line := range qeLogoLines {
		if w := lipgloss.Width(line); w > budget {
			t.Fatalf("wordmark line %d width = %d, want <= %d", i, w, budget)
		}
	}
}

// Every foreground color emitted by the logo — wordmark fallback and full
// scene alike — must be the Gentle-QE brand color #0A91B2 (RGB 10;145;178).
// TrueColor is forced because tests run without a TTY, where lipgloss would
// otherwise strip colors and hide a wrong palette.
func TestLogo_RendersInBrandColorOnly(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	const brandRGB = "38;2;10;145;178" // #0A91B2
	fgSeq := regexp.MustCompile(`38;2;\d+;\d+;\d+`)

	for name, out := range map[string]string{
		"wordmark fallback": RenderLogo(),
		"full scene":        RenderLogoFit(120, 60),
	} {
		seqs := fgSeq.FindAllString(out, -1)
		if len(seqs) == 0 {
			t.Fatalf("%s: no truecolor foreground sequences emitted", name)
		}
		for _, seq := range seqs {
			if seq != brandRGB {
				t.Fatalf("%s: found foreground %q, want only %q", name, seq, brandRGB)
			}
		}
		t.Logf("%s: %d styled segments, all %s", name, len(seqs), brandRGB)
	}
}
