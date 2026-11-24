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
		for _, st := range []string{"minimal", "pill"} {
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
