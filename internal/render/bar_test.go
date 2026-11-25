package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zhouyg1992/agybar/internal/theme"
)

func TestBarStyleNamesAndResolve(t *testing.T) {
	want := []string{"blocks", "smooth", "braille", "slider", "dots", "micro"}
	got := BarStyleNames()
	if len(got) != len(want) {
		t.Fatalf("BarStyleNames() length = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("BarStyleNames()[%d] = %q, want %q", i, got[i], want[i])
		}
		if resolved, ok := ResolveBarStyle(want[i]); !ok || resolved != want[i] {
			t.Errorf("ResolveBarStyle(%q) = (%q, %v), want (%q, true)", want[i], resolved, ok, want[i])
		}
	}

	for _, tc := range []struct {
		in   string
		want string
	}{
		{in: " Slider ", want: "slider"},
		{in: "DOTS", want: "dots"},
		{in: "MICRO", want: "micro"},
		{in: " Blocks ", want: "blocks"},
	} {
		resolved, ok := ResolveBarStyle(tc.in)
		if !ok || resolved != tc.want {
			t.Errorf("ResolveBarStyle(%q) = (%q, %v), want (%q, true)", tc.in, resolved, ok, tc.want)
		}
	}

	for _, invalid := range []string{"", "   ", "unknown", "bar"} {
		if resolved, ok := ResolveBarStyle(invalid); ok {
			t.Errorf("ResolveBarStyle(%q) = (%q, true), want (_, false)", invalid, resolved)
		}
	}
}

func TestProgressBarRule1Styles(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("AGYBAR_COLOR", "")

	pal := theme.Get("catppuccin")
	mutedRGB := fmt.Sprintf("38;2;%d;%d;%dm", pal.Muted.R, pal.Muted.G, pal.Muted.B)
	goodRGB := fmt.Sprintf("38;2;%d;%d;%dm", pal.Good.R, pal.Good.G, pal.Good.B)

	styles := []struct {
		name  string
		glyph string
	}{
		{name: "blocks", glyph: "▰"},
		{name: "dots", glyph: "•"},
		{name: "micro", glyph: "▪"},
	}

	const width = 10
	for _, st := range styles {
		for _, pct := range []float64{0, 4, 52, 100} {
			raw := ProgressBar(pct, width, st.name, pal)
			plain := stripANSI(raw)
			if visibleLen(raw) != width {
				t.Errorf("style=%s pct=%.0f visibleLen=%d, want %d (plain=%q)", st.name, pct, visibleLen(raw), width, plain)
			}
			if strings.Contains(plain, "▱") {
				t.Errorf("style=%s pct=%.0f should not contain ▱: %q", st.name, pct, plain)
			}
			if gotPlain := strings.Repeat(st.glyph, width); plain != gotPlain {
				t.Errorf("style=%s pct=%.0f plain=%q, want %q", st.name, pct, plain, gotPlain)
			}

			switch pct {
			case 0:
				if count := strings.Count(raw, mutedRGB+st.glyph); count != width {
					t.Errorf("style=%s pct=0 muted cells=%d, want %d (raw=%q)", st.name, count, width, raw)
				}
			case 4:
				if !strings.Contains(raw, goodRGB+st.glyph) {
					t.Errorf("style=%s pct=4 missing filled cell in Good color: %q", st.name, raw)
				}
				if count := strings.Count(raw, mutedRGB+st.glyph); count != width-1 {
					t.Errorf("style=%s pct=4 muted cells=%d, want %d (raw=%q)", st.name, count, width-1, raw)
				}
			case 52:
				if count := strings.Count(raw, mutedRGB+st.glyph); count != 5 {
					t.Errorf("style=%s pct=52 muted cells=%d, want 5 (raw=%q)", st.name, count, raw)
				}
			case 100:
				if strings.Contains(raw, mutedRGB) {
					t.Errorf("style=%s pct=100 should have no muted cells: %q", st.name, raw)
				}
			}
		}
	}
}

func TestProgressBarSlider(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("AGYBAR_COLOR", "")

	pal := theme.Get("catppuccin")
	mutedRGB := fmt.Sprintf("38;2;%d;%d;%dm", pal.Muted.R, pal.Muted.G, pal.Muted.B)

	cases := []struct {
		pct       float64
		thumbIdx  int
		boldThumb bool
	}{
		{pct: 0, thumbIdx: 0, boldThumb: false},
		{pct: 4, thumbIdx: 0, boldThumb: true},
		{pct: 52, thumbIdx: 5, boldThumb: true},
		{pct: 93, thumbIdx: 8, boldThumb: true},
		{pct: 100, thumbIdx: 9, boldThumb: true},
	}

	const width = 10
	for _, tc := range cases {
		raw := ProgressBar(tc.pct, width, "slider", pal)
		plain := stripANSI(raw)
		if visibleLen(raw) != width {
			t.Errorf("pct=%.0f visibleLen=%d, want %d", tc.pct, visibleLen(raw), width)
		}
		if dots := strings.Count(plain, "•"); dots != 1 {
			t.Errorf("pct=%.0f expected 1 thumb glyph •, got %d in %q", tc.pct, dots, plain)
		}
		if dashes := strings.Count(plain, "─"); dashes != width-1 {
			t.Errorf("pct=%.0f expected %d track glyphs ─, got %d in %q", tc.pct, width-1, dashes, plain)
		}
		runes := []rune(plain)
		if runes[tc.thumbIdx] != '•' {
			t.Errorf("pct=%.0f thumb at rune index %d is %q, want • (plain=%q)", tc.pct, tc.thumbIdx, runes[tc.thumbIdx], plain)
		}

		if tc.boldThumb {
			if !strings.Contains(raw, "1;38;2;") || !strings.Contains(raw, "m•") {
				t.Errorf("pct=%.0f expected bold active thumb, got %q", tc.pct, raw)
			}
		} else {
			if !strings.Contains(raw, mutedRGB+"•") {
				t.Errorf("pct=%.0f expected muted thumb, got %q", tc.pct, raw)
			}
			if strings.Contains(raw, "1;38;2;") {
				t.Errorf("pct=%.0f should not use bold thumb when pct==0: %q", tc.pct, raw)
			}
		}
	}

	if got := ProgressBar(50, 0, "slider", pal); got != "" {
		t.Errorf("width=0 expected empty string, got %q", got)
	}
	if got := stripANSI(ProgressBar(50, 1, "slider", pal)); got != "•" {
		t.Errorf("width=1 expected %q, got %q", "•", got)
	}
	if got := stripANSI(ProgressBar(-10, 10, "slider", pal)); got != "•─────────" {
		t.Errorf("negative pct expected clamped 0%% slider, got %q", got)
	}
	if got := stripANSI(ProgressBar(150, 10, "slider", pal)); got != "─────────•" {
		t.Errorf(">100 pct expected clamped 100%% slider, got %q", got)
	}
}

func TestProgressBarAllStyles256Color(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("AGYBAR_COLOR", "")

	pal := theme.Get("catppuccin")
	for _, bs := range BarStyleNames() {
		for _, pct := range []float64{0, 4, 52, 93} {
			raw := ProgressBar(pct, 10, bs, pal)
			if strings.Contains(raw, "38;2;") || strings.Contains(raw, "48;2;") {
				t.Errorf("style=%s pct=%.0f must not emit 24-bit SGR on Apple_Terminal: %q", bs, pct, raw)
			}
			if !strings.Contains(raw, "38;5;") {
				t.Errorf("style=%s pct=%.0f missing 256-color fg SGR: %q", bs, pct, raw)
			}
		}
	}
}
