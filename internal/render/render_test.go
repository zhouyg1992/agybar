package render

import (
	"strings"
	"testing"

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
