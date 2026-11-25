package render

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zhouyg1992/agybar/internal/config"
	"github.com/zhouyg1992/agybar/internal/schema"
	"github.com/zhouyg1992/agybar/internal/theme"
)

func ptrF(v float64) *float64 { return &v }
func ptrI(v int64) *int64     { return &v }

func samplePayload(usedPct float64, width int) *schema.Payload {
	return &schema.Payload{
		CWD:        "/Users/presence/code/github/agybar",
		AgentState: "thinking",
		Model: &schema.ModelInfo{
			ID:          "gemini-4-neon",
			DisplayName: "Gemini 4 Neon",
			Effort:      "medium",
		},
		ContextWindow: &schema.ContextWindow{
			TotalInputTokens:  92400,
			TotalOutputTokens: 11600,
			ContextWindowSize: 200000,
			UsedPercentage:    ptrF(usedPct),
		},
		Quota: map[string]schema.QuotaBucket{
			"5h": {
				RemainingFraction: ptrF(0.82),
				ResetInSeconds:    ptrI(8100),
			},
			"weekly": {
				RemainingFraction: ptrF(0.64),
				ResetInSeconds:    ptrI(259200),
			},
		},
		VCS: &schema.VCSInfo{
			Branch: "main",
			Dirty:  true,
		},
		Subagents: []schema.SubagentInfo{
			{Name: "research", Status: "running"},
		},
		TaskCount:     1,
		TerminalWidth: width,
	}
}

func TestRenderWideAndCompact(t *testing.T) {
	cfg := config.Default()

	wide := Render(samplePayload(52, 120), cfg)
	if !strings.Contains(wide, "\n") {
		t.Fatalf("expected 2-line output on wide terminal, got: %q", stripANSI(wide))
	}
	plainWide := stripANSI(wide)
	for _, want := range []string{"Gemini 4 Neon", "52%", "5h:82%", "main*", "thinking"} {
		if !strings.Contains(plainWide, want) {
			t.Errorf("wide output missing %q: %s", want, plainWide)
		}
	}

	narrow := Render(samplePayload(52, 72), cfg)
	if strings.Contains(narrow, "\n") {
		t.Fatalf("expected 1-line compact output on narrow terminal, got: %q", stripANSI(narrow))
	}
}

func TestRenderHighUsageCompactHint(t *testing.T) {
	cfg := config.Default()
	out := stripANSI(Render(samplePayload(93, 120), cfg))
	if !strings.Contains(out, "/compact") {
		t.Errorf("expected /compact hint at 93%% usage, got: %s", out)
	}
}

func TestRenderAllThemesAndStyles(t *testing.T) {
	p := samplePayload(68, 120)
	for _, th := range theme.Names() {
		for _, st := range []string{"minimal", "badge", "pill"} {
			cfg := config.Default()
			cfg.Theme = th
			cfg.Style = st
			out := Render(p, cfg)
			if visibleLen(out) == 0 {
				t.Errorf("empty output for theme=%s style=%s", th, st)
			}
		}
	}
}

func TestRenderEmptyPayload(t *testing.T) {
	out := Render(&schema.Payload{}, config.Default())
	if out == "" {
		t.Error("expected non-nil safe render on empty payload")
	}
}

func TestModelEffortDedup(t *testing.T) {
	p := samplePayload(40, 120)
	p.Model.DisplayName = "Gemini 4 Neon (Medium)"
	p.Model.Effort = "medium"

	minimalOut := stripANSI(Render(p, config.Default()))
	if strings.Contains(minimalOut, "[med]") {
		t.Errorf("expected duplicate [med] to be omitted, got: %s", minimalOut)
	}

	pillCfg := config.Default()
	pillCfg.Style = "pill"
	pillOut := stripANSI(Render(p, pillCfg))
	if strings.Contains(pillOut, "· med") {
		t.Errorf("expected duplicate pill effort to be omitted, got: %s", pillOut)
	}

	p.Model.DisplayName = "Gemini 4 Neon"
	withEffort := stripANSI(Render(p, config.Default()))
	if !strings.Contains(withEffort, "[med]") {
		t.Errorf("expected [med] when not in display name, got: %s", withEffort)
	}
}

func TestIdleStateUsesSubtextColor(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("AGYBAR_COLOR", "")

	p := samplePayload(20, 120)
	p.AgentState = "idle"
	pal := theme.Get("catppuccin")
	out := Render(p, config.Default())

	wantRGB := fmt.Sprintf("38;2;%d;%d;%dm", pal.Subtext.R, pal.Subtext.G, pal.Subtext.B)
	mutedRGB := fmt.Sprintf("38;2;%d;%d;%dm○ idle", pal.Muted.R, pal.Muted.G, pal.Muted.B)
	if !strings.Contains(out, wantRGB+"○ idle") {
		t.Errorf("expected idle state to use Subtext color (%s), got: %q", wantRGB, out)
	}
	if strings.Contains(out, mutedRGB) {
		t.Errorf("idle state should not use Muted color: %q", out)
	}
}

func TestPillRendersArtifactsAndSandbox(t *testing.T) {
	p := samplePayload(45, 120)
	p.ArtifactCount = 3
	p.Sandbox = &schema.SandboxInfo{Enabled: true}

	cfg := config.Default()
	cfg.Style = "pill"
	out := stripANSI(Render(p, cfg))
	for _, want := range []string{"◈ 3", "⊘ sandbox"} {
		if !strings.Contains(out, want) {
			t.Errorf("pill output missing %q: %s", want, out)
		}
	}
}

func TestCompactOmitsEffortInMinimalAndPill(t *testing.T) {
	p := samplePayload(45, 72)
	minOut := stripANSI(Render(p, config.Default()))
	if strings.Contains(minOut, "[med]") {
		t.Errorf("minimal compact should omit [med], got: %s", minOut)
	}

	pillCfg := config.Default()
	pillCfg.Style = "pill"
	pillOut := stripANSI(Render(p, pillCfg))
	if strings.Contains(pillOut, "· med") {
		t.Errorf("pill compact should omit effort, got: %s", pillOut)
	}
}

func TestQuotaModelMatchingAndResetTimeFallback(t *testing.T) {
	future := time.Now().Add(2*time.Hour + 30*time.Minute).UTC().Format(time.RFC3339)
	p := &schema.Payload{
		Model: &schema.ModelInfo{ID: "Orchestrator (Neon)", DisplayName: "Orchestrator (Neon)"},
		Quota: map[string]schema.QuotaBucket{
			"3p-5h":         {RemainingFraction: ptrF(0.10), ResetInSeconds: ptrI(1800)},
			"3p-weekly":     {RemainingFraction: ptrF(0.20), ResetInSeconds: ptrI(36000)},
			"gemini-5h":     {RemainingFraction: ptrF(0.95), ResetTime: future},
			"gemini-weekly": {RemainingFraction: ptrF(0.90), ResetInSeconds: ptrI(86400)},
		},
		TerminalWidth: 120,
	}

	out := stripANSI(Render(p, config.Default()))
	if !strings.Contains(out, "gemini-5h:95%(2h") || !strings.Contains(out, "gemini-7d:90%(1d)") {
		t.Errorf("expected gemini buckets prioritized with reset_time fallback for Orchestrator (Neon), got: %s", out)
	}
	if strings.Contains(out, "3p-5h") {
		t.Errorf("did not expect 3p-5h when 2 gemini buckets match: %s", out)
	}

	p.Model = &schema.ModelInfo{ID: "claude-sonnet", DisplayName: "Claude Sonnet"}
	outClaude := stripANSI(Render(p, config.Default()))
	if !strings.Contains(outClaude, "3p-5h:10%") || !strings.Contains(outClaude, "3p-7d:20%") {
		t.Errorf("expected 3p buckets prioritized for Claude Sonnet, got: %s", outClaude)
	}
}

func TestTerminalAppFallbackTo256Color(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("AGYBAR_COLOR", "")

	p := samplePayload(68, 120)
	for _, th := range theme.Names() {
		for _, st := range []string{"minimal", "badge", "pill"} {
			cfg := config.Default()
			cfg.Theme = th
			cfg.Style = st
			out := Render(p, cfg)
			if strings.Contains(out, "38;2;") || strings.Contains(out, "48;2;") {
				t.Errorf("Apple_Terminal output must not emit 24-bit SGR (theme=%s style=%s): %q", th, st, out)
			}
			if !strings.Contains(out, "38;5;") {
				t.Errorf("Apple_Terminal output missing 256-color fg SGR (theme=%s style=%s): %q", th, st, out)
			}
			if (st == "badge" || st == "pill") && !strings.Contains(out, "48;5;") {
				t.Errorf("Apple_Terminal %s output missing 256-color bg SGR (theme=%s): %q", st, th, out)
			}
		}
	}
}

func TestBadgeAndPillCaps(t *testing.T) {
	p := samplePayload(52, 120)
	cfg := config.Default()
	cfg.Style = "badge"

	badgeOut := Render(p, cfg)
	if !strings.Contains(badgeOut, "48;") {
		t.Errorf("badge output should contain background SGR: %q", badgeOut)
	}
	if strings.Contains(badgeOut, "\ue0b6") || strings.Contains(badgeOut, "\ue0b4") {
		t.Errorf("badge output should not contain Nerd Font PUA caps: %q", badgeOut)
	}

	cfg.Style = "pill"
	pillOut := Render(p, cfg)
	if !strings.Contains(pillOut, "\ue0b6") || !strings.Contains(pillOut, "\ue0b4") {
		t.Errorf("pill output should contain Nerd Font PUA caps: %q", pillOut)
	}
}

func TestRGBToANSI256(t *testing.T) {
	cases := []struct {
		c    theme.RGB
		want uint8
	}{
		{theme.RGB{R: 49, G: 50, B: 68}, 237},
		{theme.RGB{R: 255, G: 255, B: 255}, 231},
		{theme.RGB{R: 0, G: 0, B: 0}, 16},
		{theme.RGB{R: 158, G: 206, B: 106}, 149},
	}
	for _, tc := range cases {
		if got := rgbToANSI256(tc.c); got != tc.want {
			t.Errorf("rgbToANSI256(%+v) = %d, want %d", tc.c, got, tc.want)
		}
	}
}

func TestMinimalSeparators(t *testing.T) {
	p := samplePayload(52, 120)

	presets := []struct {
		spec string
		want string
	}{
		{spec: "pipe", want: " │ "},
		{spec: "dashed", want: " ┊ "},
		{spec: "dot", want: " • "},
		{spec: "square", want: " ▪ "},
		{spec: "DOT", want: " • "},
		{spec: "", want: " │ "},
		{spec: "›", want: " › "},
		{spec: "//", want: " // "},
		{spec: "X", want: " X "},
		{spec: "a\nb", want: " ab "},
		{spec: "\n\t", want: " │ "},
	}

	for _, tc := range presets {
		cfg := config.Default()
		cfg.Separator = tc.spec
		out := stripANSI(Render(p, cfg))
		lines := strings.Split(out, "\n")
		if len(lines) != 2 {
			t.Fatalf("sep=%q expected 2 lines, got %d: %q", tc.spec, len(lines), out)
		}
		if !strings.Contains(lines[0], tc.want) {
			t.Errorf("sep=%q line 1 missing %q: %q", tc.spec, tc.want, lines[0])
		}
		if !strings.Contains(lines[1], tc.want) {
			t.Errorf("sep=%q line 2 missing %q: %q", tc.spec, tc.want, lines[1])
		}

		pNarrow := samplePayload(52, 72)
		compactOut := stripANSI(Render(pNarrow, cfg))
		if !strings.Contains(compactOut, tc.want) {
			t.Errorf("sep=%q compact line missing %q: %q", tc.spec, tc.want, compactOut)
		}
	}

	// Test "none" preset in wide ("   ") and compact ("  ") without pipe or empty ANSI resets.
	for _, spec := range []string{"none", "None"} {
		cfg := config.Default()
		cfg.Separator = spec

		wideRaw := Render(p, cfg)
		widePlain := stripANSI(wideRaw)
		wideLines := strings.Split(widePlain, "\n")
		if len(wideLines) != 2 {
			t.Fatalf("sep=%q expected 2 lines, got %q", spec, widePlain)
		}
		if strings.Contains(widePlain, "│") {
			t.Errorf("sep=%q wide output should not contain │: %q", spec, widePlain)
		}
		if !strings.Contains(wideLines[0], "   ") || !strings.Contains(wideLines[1], "   ") {
			t.Errorf("sep=%q wide lines should use 3 spaces: %q", spec, widePlain)
		}
		if strings.Contains(wideRaw, "m\x1b[0m") {
			t.Errorf("sep=%q should not emit empty ANSI wrapper around spaces: %q", spec, wideRaw)
		}

		pNarrow := samplePayload(52, 72)
		compactPlain := stripANSI(Render(pNarrow, cfg))
		if strings.Contains(compactPlain, "│") {
			t.Errorf("sep=%q compact output should not contain │: %q", spec, compactPlain)
		}
		if !strings.Contains(compactPlain, "  ") {
			t.Errorf("sep=%q compact output should use 2 spaces: %q", spec, compactPlain)
		}
	}

	// Badge and pill styles should ignore Separator and keep space-separated capsules.
	for _, st := range []string{"badge", "pill"} {
		cfg := config.Default()
		cfg.Style = st
		cfg.Separator = "dot"
		out := stripANSI(Render(p, cfg))
		if strings.Contains(out, " • ") {
			t.Errorf("style=%s should ignore minimal separator 'dot', got: %q", st, out)
		}
	}
}

func TestRenderCycleModeAtEndOfLine2(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("AGYBAR_COLOR", "")

	pal := theme.Get("catppuccin")
	cases := []struct {
		mode     string
		wantText string
		wantCol  theme.RGB
	}{
		{mode: "plan", wantText: "❖ plan", wantCol: pal.Cyan},
		{mode: "auto", wantText: "» auto", wantCol: pal.Good},
		{mode: "accept-edits", wantText: "✎ accept-edits", wantCol: pal.Warn},
		{mode: "accept_edits", wantText: "✎ accept-edits", wantCol: pal.Warn},
		{mode: "default", wantText: "◇ default", wantCol: pal.Subtext},
		{mode: "always-proceed", wantText: "⚠ always-proceed", wantCol: pal.High},
		{mode: "custom", wantText: "◇ custom", wantCol: pal.Subtext},
	}

	for _, tc := range cases {
		p := samplePayload(52, 120)
		p.ArtifactCount = 2
		p.Sandbox = &schema.SandboxInfo{Enabled: true}
		p.CycleMode = tc.mode

		wantRGB := fmt.Sprintf("38;2;%d;%d;%dm", tc.wantCol.R, tc.wantCol.G, tc.wantCol.B)
		for _, st := range []string{"minimal", "badge", "pill"} {
			cfg := config.Default()
			cfg.Style = st

			raw := Render(p, cfg)
			if !strings.Contains(raw, wantRGB) {
				t.Errorf("mode=%q style=%s missing RGB %s in %q", tc.mode, st, wantRGB, raw)
			}

			lines := strings.Split(stripANSI(raw), "\n")
			if len(lines) != 2 {
				t.Fatalf("mode=%q style=%s expected 2 lines, got %d: %q", tc.mode, st, len(lines), stripANSI(raw))
			}

			line2 := lines[1]
			sandboxIdx := strings.LastIndex(line2, "⊘ sandbox")
			modeIdx := strings.LastIndex(line2, tc.wantText)
			if sandboxIdx < 0 || modeIdx <= sandboxIdx {
				t.Errorf("mode=%q style=%s expected %q after sandbox on line 2, got: %q", tc.mode, st, tc.wantText, line2)
			}
			trimmed := strings.TrimRight(line2, " \ue0b4")
			if !strings.HasSuffix(trimmed, tc.wantText) {
				t.Errorf("mode=%q style=%s expected line 2 to end with %q, got: %q", tc.mode, st, tc.wantText, line2)
			}
		}
	}
}

func TestRenderCycleModeEmptyAndCompact(t *testing.T) {
	p := samplePayload(52, 120)
	p.Sandbox = &schema.SandboxInfo{Enabled: true}
	p.CycleMode = ""

	for _, st := range []string{"minimal", "badge", "pill"} {
		cfg := config.Default()
		cfg.Style = st
		lines := strings.Split(stripANSI(Render(p, cfg)), "\n")
		if len(lines) != 2 {
			t.Fatalf("style=%s expected 2 lines, got %d", st, len(lines))
		}
		trimmed := strings.TrimRight(lines[1], " \ue0b4")
		if !strings.HasSuffix(trimmed, "⊘ sandbox") {
			t.Errorf("style=%s empty cycle_mode should end with sandbox, got: %q", st, lines[1])
		}
	}

	pNarrow := samplePayload(52, 72)
	pNarrow.CycleMode = "plan"
	for _, st := range []string{"minimal", "badge", "pill"} {
		cfg := config.Default()
		cfg.Style = st
		out := stripANSI(Render(pNarrow, cfg))
		if strings.Contains(out, "\n") {
			t.Errorf("style=%s compact mode should stay 1 line, got: %q", st, out)
		}
		if strings.Contains(out, "❖ plan") {
			t.Errorf("style=%s compact mode should omit cycle_mode, got: %q", st, out)
		}
	}

	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("AGYBAR_COLOR", "")
	pApple := samplePayload(52, 120)
	pApple.CycleMode = "plan"
	rawApple := Render(pApple, config.Default())
	if strings.Contains(rawApple, "38;2;") || strings.Contains(rawApple, "48;2;") {
		t.Errorf("Apple_Terminal output must not emit 24-bit SGR: %q", rawApple)
	}
	if !strings.Contains(stripANSI(rawApple), "❖ plan") {
		t.Errorf("Apple_Terminal output missing mode label: %q", stripANSI(rawApple))
	}
}
