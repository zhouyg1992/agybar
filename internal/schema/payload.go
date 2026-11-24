package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Payload represents the JSON telemetry piped to statusLine commands via stdin by agy CLI.
type Payload struct {
	CWD               string                 `json:"cwd"`
	SessionID         string                 `json:"session_id"`
	ConversationID    string                 `json:"conversation_id"`
	ConversationTitle string                 `json:"conversation_title"`
	TranscriptPath    string                 `json:"transcript_path"`
	Version           string                 `json:"version"`
	Product           string                 `json:"product"`
	CycleMode         string                 `json:"cycle_mode"`
	Exceeds200kTokens *bool                  `json:"exceeds_200k_tokens"`
	AgentState        string                 `json:"agent_state"`
	Model             *ModelInfo             `json:"model"`
	ContextWindow     *ContextWindow         `json:"context_window"`
	Quota             map[string]QuotaBucket `json:"quota"`
	Workspace         *WorkspaceInfo         `json:"workspace"`
	VCS               *VCSInfo               `json:"vcs"`
	Sandbox           *SandboxInfo           `json:"sandbox"`
	Subagents         []SubagentInfo         `json:"subagents"`
	ArtifactCount     int                    `json:"artifact_count"`
	TaskCount         int                    `json:"task_count"`
	TerminalWidth     int                    `json:"terminal_width"`
	PlanTier          string                 `json:"plan_tier"`
	Email             string                 `json:"email"`
	AICredits         *float64               `json:"ai_credits"`
}

type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Effort      string `json:"effort"`
}

type ContextWindow struct {
	TotalInputTokens    int64           `json:"total_input_tokens"`
	TotalOutputTokens   int64           `json:"total_output_tokens"`
	ContextWindowSize   int64           `json:"context_window_size"`
	UsedPercentage      *float64        `json:"used_percentage"`
	RemainingPercentage *float64        `json:"remaining_percentage"`
	CurrentUsage        json.RawMessage `json:"current_usage"`
}

type QuotaBucket struct {
	RemainingFraction *float64 `json:"remaining_fraction"`
	RemainingAmount   *float64 `json:"remaining_amount"`
	ResetTime         string   `json:"reset_time"`
	ResetInSeconds    *int64   `json:"reset_in_seconds"`
	Disabled          bool     `json:"disabled"`
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type VCSInfo struct {
	Type   string `json:"type"`
	Branch string `json:"branch"`
	Dirty  bool   `json:"dirty"`
}

type SandboxInfo struct {
	Enabled bool `json:"enabled"`
}

type SubagentInfo struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Parse decodes stdin JSON gracefully. Empty input returns an empty Payload without error.
func Parse(data []byte) (*Payload, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return &Payload{}, nil
	}
	var p Payload
	if err := json.Unmarshal([]byte(trimmed), &p); err != nil {
		return nil, fmt.Errorf("parse statusline payload: %w", err)
	}
	return &p, nil
}

// EffectiveUsedPct returns the context window used percentage in [0, 100].
func (c *ContextWindow) EffectiveUsedPct() (float64, bool) {
	if c == nil {
		return 0, false
	}
	if c.UsedPercentage != nil {
		v := *c.UsedPercentage
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		return v, true
	}
	if c.RemainingPercentage != nil {
		v := 100 - *c.RemainingPercentage
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		return v, true
	}
	if c.ContextWindowSize > 0 {
		total := c.TotalInputTokens + c.TotalOutputTokens
		v := float64(total) / float64(c.ContextWindowSize) * 100
		if v > 100 {
			v = 100
		}
		return v, true
	}
	return 0, false
}

// EffectiveUsedTokens returns the current context token count when available,
// falling back to total input + output tokens.
func (c *ContextWindow) EffectiveUsedTokens() int64 {
	if c == nil {
		return 0
	}
	raw := strings.TrimSpace(string(c.CurrentUsage))
	if raw != "" && raw != "null" {
		var cu struct {
			InputTokens              int64 `json:"input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
		}
		if err := json.Unmarshal([]byte(raw), &cu); err == nil {
			sum := cu.InputTokens + cu.CacheCreationInputTokens + cu.CacheReadInputTokens + cu.OutputTokens
			if sum > 0 {
				return sum
			}
		}
	}
	return c.TotalInputTokens + c.TotalOutputTokens
}

// EffectiveDir returns the preferred working directory from workspace or cwd.
func (p *Payload) EffectiveDir() string {
	if p == nil {
		return ""
	}
	if p.Workspace != nil {
		if p.Workspace.CurrentDir != "" {
			return p.Workspace.CurrentDir
		}
		if p.Workspace.ProjectDir != "" {
			return p.Workspace.ProjectDir
		}
	}
	return p.CWD
}

// DisplayModelName returns a concise human-friendly model name.
func (p *Payload) DisplayModelName() string {
	if p == nil || p.Model == nil {
		return ""
	}
	if p.Model.DisplayName != "" {
		return p.Model.DisplayName
	}
	return p.Model.ID
}

// ActiveSubagentCount returns the count of running/working subagents (or total if statuses are omitted).
func (p *Payload) ActiveSubagentCount() int {
	if p == nil || len(p.Subagents) == 0 {
		return 0
	}
	active := 0
	for _, s := range p.Subagents {
		st := strings.ToLower(s.Status)
		if st == "" || st == "running" || st == "working" || st == "active" || st == "thinking" {
			active++
		}
	}
	return active
}
