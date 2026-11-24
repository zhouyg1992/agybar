package schema

import "testing"

func TestParseRealAgyPayload(t *testing.T) {
	raw := `{
		"cwd": "/Users/presence/code/github/agybar",
		"session_id": "",
		"conversation_id": "",
		"transcript_path": "/Users/presence/.gemini/antigravity/brain/.system_generated/logs/transcript.jsonl",
		"model": {"id": "Orchestrator (Neon)", "display_name": "Orchestrator (Neon)"},
		"workspace": {"current_dir": "/Users/presence/code/github/agybar", "project_dir": "/Users/presence/code/github/agybar"},
		"version": "1.5.16",
		"context_window": {
			"total_input_tokens": 0,
			"total_output_tokens": 0,
			"context_window_size": 1048576,
			"used_percentage": 0,
			"remaining_percentage": 100,
			"current_usage": null
		},
		"exceeds_200k_tokens": null,
		"product": "antigravity",
		"quota": {
			"3p-5h": {"remaining_fraction": 1, "reset_time": "2026-11-24T12:46:48Z", "reset_in_seconds": 17109},
			"3p-weekly": {"remaining_fraction": 0.98408, "reset_time": "2026-12-01T02:46:48Z", "reset_in_seconds": 585909},
			"gemini-5h": {"remaining_fraction": 0.97935826, "reset_time": "2026-11-24T12:25:01Z", "reset_in_seconds": 15802},
			"gemini-weekly": {"remaining_fraction": 0.94954187, "reset_time": "2026-12-01T02:25:01Z", "reset_in_seconds": 584602}
		},
		"agent_state": "idle",
		"vcs": {"type": "git"},
		"cycle_mode": "auto",
		"sandbox": {"enabled": false},
		"plan_tier": "Google AI Pro",
		"email": "user@example.com",
		"terminal_width": 158
	}`

	p, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p.Version != "1.5.16" {
		t.Errorf("Version = %q, want 1.5.16", p.Version)
	}
	if p.Product != "antigravity" {
		t.Errorf("Product = %q, want antigravity", p.Product)
	}
	if p.CycleMode != "auto" {
		t.Errorf("CycleMode = %q, want auto", p.CycleMode)
	}
	if p.VCS == nil || p.VCS.Type != "git" || p.VCS.Branch != "" {
		t.Errorf("unexpected VCS: %+v", p.VCS)
	}
	if len(p.Quota) != 4 {
		t.Errorf("len(Quota) = %d, want 4", len(p.Quota))
	}
	if p.TerminalWidth != 158 {
		t.Errorf("TerminalWidth = %d, want 158", p.TerminalWidth)
	}
	if got := p.EffectiveDir(); got != "/Users/presence/code/github/agybar" {
		t.Errorf("EffectiveDir() = %q", got)
	}
}

func TestEffectiveUsedTokens(t *testing.T) {
	cwNull := &ContextWindow{
		TotalInputTokens:  50400,
		TotalOutputTokens: 8100,
		CurrentUsage:      []byte("null"),
	}
	if got := cwNull.EffectiveUsedTokens(); got != 58500 {
		t.Errorf("with null current_usage: got %d, want 58500", got)
	}

	cwCurrent := &ContextWindow{
		TotalInputTokens:  300000,
		TotalOutputTokens: 50000,
		CurrentUsage:      []byte(`{"input_tokens":12000,"cache_creation_input_tokens":3000,"cache_read_input_tokens":45000,"output_tokens":2000}`),
	}
	if got := cwCurrent.EffectiveUsedTokens(); got != 62000 {
		t.Errorf("with populated current_usage: got %d, want 62000", got)
	}
}
