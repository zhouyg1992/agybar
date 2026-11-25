package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zhouyg1992/agybar/internal/config"
	"github.com/zhouyg1992/agybar/internal/schema"
	"github.com/zhouyg1992/agybar/internal/theme"
)

// Render turns an agy statusLine payload into a styled multi-line or single-line string.
func Render(p *schema.Payload, cfg config.Config) string {
	if p == nil {
		p = &schema.Payload{}
	}
	pal := theme.Get(cfg.Theme)
	compact := !cfg.ShowSecond || (p.TerminalWidth > 0 && p.TerminalWidth < cfg.CompactAt)

	switch strings.ToLower(strings.TrimSpace(cfg.Style)) {
	case "badge":
		return renderCapsule(p, cfg, pal, compact, false)
	case "pill":
		return renderCapsule(p, cfg, pal, compact, true)
	default:
		return renderMinimal(p, cfg, pal, compact)
	}
}

func renderMinimal(p *schema.Payload, cfg config.Config, pal theme.Palette, compact bool) string {
	sep := formatSeparator(cfg.Separator, pal, compact)

	// --- Line 1: Model + Context Bar + Tokens + Quota ---
	var line1 []string

	if m := formatModel(p, pal, compact); m != "" {
		line1 = append(line1, m)
	}

	barWidth := 10
	if compact {
		barWidth = 6
	}
	if ctxSeg := formatContext(p, cfg, pal, barWidth, compact); ctxSeg != "" {
		line1 = append(line1, ctxSeg)
	}

	if cfg.ShowTokens && !compact {
		if tok := formatTokens(p, pal); tok != "" {
			line1 = append(line1, tok)
		}
	}

	if cfg.ShowQuota {
		maxBuckets := 2
		if compact {
			maxBuckets = 1
		}
		if q := formatQuota(p, pal, maxBuckets, compact); q != "" {
			line1 = append(line1, q)
		}
	}

	if compact {
		if st := formatState(p.AgentState, pal, true); st != "" {
			line1 = append(line1, st)
		}
		return " " + strings.Join(line1, sep)
	}

	// --- Line 2: State + VCS + Workspace + Subagents/Tasks/Sandbox + Mode ---
	var line2 []string

	if st := formatState(p.AgentState, pal, false); st != "" {
		line2 = append(line2, st)
	}
	if vcs := formatVCS(p, pal); vcs != "" {
		line2 = append(line2, vcs)
	}
	if dir := formatDir(p, pal); dir != "" {
		line2 = append(line2, dir)
	}
	if sub := formatSubagents(p, pal); sub != "" {
		line2 = append(line2, sub)
	}
	if p.TaskCount > 0 {
		line2 = append(line2, fg(pal.Cyan, fmt.Sprintf("⏳ %d", p.TaskCount)))
	}
	if p.ArtifactCount > 0 {
		line2 = append(line2, fg(pal.Subtext, fmt.Sprintf("◈ %d", p.ArtifactCount)))
	}
	if p.Sandbox != nil && p.Sandbox.Enabled {
		line2 = append(line2, fg(pal.Subtext, "⊘ sandbox"))
	}
	if mode := formatCycleMode(p.CycleMode, pal); mode != "" {
		line2 = append(line2, mode)
	}

	out1 := " " + strings.Join(line1, sep)
	if len(line2) == 0 {
		return out1
	}
	out2 := " " + strings.Join(line2, sep)
	return out1 + "\n" + out2
}

func formatSeparator(spec string, pal theme.Palette, compact bool) string {
	trimmed := strings.TrimSpace(spec)
	switch strings.ToLower(trimmed) {
	case "", "pipe":
		return " " + fg(pal.Muted, "│") + " "
	case "dashed":
		return " " + fg(pal.Muted, "┊") + " "
	case "dot":
		return " " + fg(pal.Muted, "•") + " "
	case "square":
		return " " + fg(pal.Muted, "▪") + " "
	case "none":
		if compact {
			return "  "
		}
		return "   "
	default:
		sanitized := strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return -1
			}
			return r
		}, trimmed)
		sanitized = strings.TrimSpace(sanitized)
		if sanitized == "" {
			sanitized = "│"
		}
		return " " + fg(pal.Muted, sanitized) + " "
	}
}

func renderCapsule(p *schema.Payload, cfg config.Config, pal theme.Palette, compact bool, rounded bool) string {
	// --- Line 1 pills ---
	var line1 []string

	if name := p.DisplayModelName(); name != "" {
		label := "✦ " + name
		if !compact && p.Model != nil && p.Model.Effort != "" && !modelIncludesEffort(name, p.Model.Effort) {
			label += " · " + shortEffort(p.Model.Effort)
		}
		line1 = append(line1, pill(pal.SurfaceHi, pal.Accent, label, rounded))
	}

	if pct, ok := p.ContextWindow.EffectiveUsedPct(); ok {
		barWidth := 8
		if compact {
			barWidth = 5
		}
		bar := ProgressBar(pct, barWidth, cfg.BarStyle, pal)
		pctCol := pal.UsageColor(pct)
		txt := fmt.Sprintf("%s %s", bar, fgBold(pctCol, fmt.Sprintf("%.0f%%", pct)))
		if !compact && p.ContextWindow != nil && p.ContextWindow.ContextWindowSize > 0 {
			used := p.ContextWindow.EffectiveUsedTokens()
			if used > 0 {
				txt += " " + fg(pal.Subtext, fmt.Sprintf("(%s/%s)", formatCount(used), formatCount(p.ContextWindow.ContextWindowSize)))
			}
		}
		if pct >= 90 {
			txt += " " + fgBold(pal.Crit, "→ /compact")
		}
		line1 = append(line1, txt)
	}

	if cfg.ShowTokens && !compact {
		if tok := formatTokens(p, pal); tok != "" {
			line1 = append(line1, pill(pal.Surface, pal.Subtext, stripANSI(tok), rounded))
		}
	}

	if cfg.ShowQuota {
		maxBuckets := 2
		if compact {
			maxBuckets = 1
		}
		if q := formatQuota(p, pal, maxBuckets, compact); q != "" {
			line1 = append(line1, q)
		}
	}

	if compact {
		if st := formatState(p.AgentState, pal, true); st != "" {
			line1 = append(line1, st)
		}
		return " " + strings.Join(line1, " ")
	}

	// --- Line 2 pills ---
	var line2 []string
	if st := formatState(p.AgentState, pal, false); st != "" {
		line2 = append(line2, st)
	}
	if p.VCS != nil && p.VCS.Branch != "" {
		br := "⎇ " + p.VCS.Branch
		if p.VCS.Dirty {
			br += "*"
		}
		line2 = append(line2, pill(pal.Surface, pal.Accent2, br, rounded))
	}
	if dir := formatDir(p, pal); dir != "" {
		line2 = append(line2, dir)
	}
	if sub := formatSubagents(p, pal); sub != "" {
		line2 = append(line2, pill(pal.Surface, pal.Accent, stripANSI(sub), rounded))
	}
	if p.TaskCount > 0 {
		line2 = append(line2, pill(pal.Surface, pal.Cyan, fmt.Sprintf("⏳ %d", p.TaskCount), rounded))
	}
	if p.ArtifactCount > 0 {
		line2 = append(line2, pill(pal.Surface, pal.Subtext, fmt.Sprintf("◈ %d", p.ArtifactCount), rounded))
	}
	if p.Sandbox != nil && p.Sandbox.Enabled {
		line2 = append(line2, pill(pal.Surface, pal.Subtext, "⊘ sandbox", rounded))
	}
	if icon, label, col, ok := cycleModeSpec(p.CycleMode, pal); ok {
		line2 = append(line2, pill(pal.Surface, col, icon+" "+label, rounded))
	}

	out1 := " " + strings.Join(line1, " ")
	if len(line2) == 0 {
		return out1
	}
	out2 := " " + strings.Join(line2, " ")
	return out1 + "\n" + out2
}

func formatModel(p *schema.Payload, pal theme.Palette, compact bool) string {
	name := p.DisplayModelName()
	if name == "" {
		return ""
	}
	out := fgBold(pal.Accent, "✦ "+name)
	if !compact && p.Model != nil && p.Model.Effort != "" && !modelIncludesEffort(name, p.Model.Effort) {
		out += " " + fg(pal.Muted, "["+shortEffort(p.Model.Effort)+"]")
	}
	return out
}

func modelIncludesEffort(name, effort string) bool {
	nl := strings.ToLower(strings.TrimSpace(name))
	el := strings.ToLower(strings.TrimSpace(effort))
	if nl == "" || el == "" {
		return false
	}
	sh := shortEffort(el)
	for _, token := range []string{el, sh} {
		if strings.Contains(nl, "("+token+")") ||
			strings.Contains(nl, "["+token+"]") ||
			strings.HasSuffix(nl, " "+token) ||
			strings.Contains(nl, " "+token+" ") {
			return true
		}
	}
	return false
}

func formatContext(p *schema.Payload, cfg config.Config, pal theme.Palette, barWidth int, compact bool) string {
	pct, ok := p.ContextWindow.EffectiveUsedPct()
	if !ok {
		return ""
	}
	bar := ProgressBar(pct, barWidth, cfg.BarStyle, pal)
	pctCol := pal.UsageColor(pct)
	out := fmt.Sprintf("%s %s", bar, fgBold(pctCol, fmt.Sprintf("%.0f%%", pct)))

	if !compact && p.ContextWindow != nil && p.ContextWindow.ContextWindowSize > 0 {
		used := p.ContextWindow.EffectiveUsedTokens()
		if used > 0 {
			out += " " + fg(pal.Subtext, fmt.Sprintf("(%s/%s)", formatCount(used), formatCount(p.ContextWindow.ContextWindowSize)))
		}
	}
	if pct >= 90 {
		out += " " + fgBold(pal.Crit, "→ /compact")
	}
	return out
}

func formatTokens(p *schema.Payload, pal theme.Palette) string {
	if p.ContextWindow == nil {
		return ""
	}
	inTok := p.ContextWindow.TotalInputTokens
	outTok := p.ContextWindow.TotalOutputTokens
	if inTok == 0 && outTok == 0 {
		return ""
	}
	return fg(pal.Subtext, "↑"+formatCount(inTok)) + " " + fg(pal.Subtext, "↓"+formatCount(outTok))
}

type quotaEntry struct {
	name   string
	remPct float64
	reset  string
	match  bool
}

func formatQuota(p *schema.Payload, pal theme.Palette, maxBuckets int, compact bool) string {
	if len(p.Quota) == 0 && p.AICredits == nil {
		return ""
	}

	modelHint := strings.ToLower(p.DisplayModelName())
	if p.Model != nil && p.Model.ID != "" {
		modelHint += " " + strings.ToLower(p.Model.ID)
	}

	var entries []quotaEntry
	for k, b := range p.Quota {
		if b.Disabled || b.RemainingFraction == nil {
			continue
		}
		remPct := *b.RemainingFraction * 100
		if remPct < 0 {
			remPct = 0
		}
		if remPct > 100 {
			remPct = 100
		}
		reset := ""
		if b.ResetInSeconds != nil && *b.ResetInSeconds > 0 {
			reset = formatDuration(*b.ResetInSeconds)
		} else if b.ResetTime != "" {
			if t, err := time.Parse(time.RFC3339, b.ResetTime); err == nil {
				if secs := int64(time.Until(t).Seconds()); secs > 0 {
					reset = formatDuration(secs)
				}
			}
		}
		kl := strings.ToLower(k)
		matched := bucketMatchesModel(kl, modelHint)
		entries = append(entries, quotaEntry{
			name:   shortBucketName(k),
			remPct: remPct,
			reset:  reset,
			match:  matched,
		})
	}

	// Sort: matched buckets first, then lowest remaining percentage, then name.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].match != entries[j].match {
			return entries[i].match
		}
		if entries[i].remPct != entries[j].remPct {
			return entries[i].remPct < entries[j].remPct
		}
		return entries[i].name < entries[j].name
	})

	if len(entries) > maxBuckets {
		entries = entries[:maxBuckets]
	}

	var parts []string
	for _, e := range entries {
		col := pal.QuotaRemainingColor(e.remPct)
		seg := fg(pal.Subtext, e.name+":") + fgBold(col, fmt.Sprintf("%.0f%%", e.remPct))
		if !compact && e.reset != "" {
			seg += fg(pal.Muted, "("+e.reset+")")
		}
		parts = append(parts, seg)
	}

	if p.AICredits != nil && len(parts) < maxBuckets {
		parts = append(parts, fg(pal.Subtext, "cr:")+fg(pal.Cyan, fmt.Sprintf("%.0f", *p.AICredits)))
	}

	if len(parts) == 0 {
		return ""
	}
	return fg(pal.Warn, "⚡") + " " + strings.Join(parts, " "+fg(pal.Muted, "·")+" ")
}

func formatState(state string, pal theme.Palette, compact bool) string {
	s := strings.ToLower(strings.TrimSpace(state))
	if s == "" {
		return ""
	}
	var col theme.RGB
	var dot string
	switch s {
	case "thinking":
		col = pal.Accent
		dot = "◉"
	case "working", "running":
		col = pal.Accent2
		dot = "●"
	case "tool_use":
		col = pal.Cyan
		dot = "⚙"
	case "error":
		col = pal.Crit
		dot = "✖"
	default: // idle
		col = pal.Subtext
		dot = "○"
	}
	if compact {
		return fg(col, dot)
	}
	return fg(col, dot+" "+s)
}

func formatVCS(p *schema.Payload, pal theme.Palette) string {
	if p.VCS == nil || p.VCS.Branch == "" {
		return ""
	}
	out := fg(pal.Accent2, "⎇ "+p.VCS.Branch)
	if p.VCS.Dirty {
		out += fgBold(pal.Warn, "*")
	}
	return out
}

func formatDir(p *schema.Payload, pal theme.Palette) string {
	dir := p.EffectiveDir()
	if dir == "" {
		return ""
	}
	return fg(pal.Subtext, shortenPath(dir))
}

func bucketMatchesModel(bucketKeyLower, modelHintLower string) bool {
	if modelHintLower == "" {
		return false
	}
	for _, token := range []string{"gemini", "claude", "gpt", "pro", "flash", "neon", "opus", "sonnet", "haiku"} {
		if strings.Contains(bucketKeyLower, token) && strings.Contains(modelHintLower, token) {
			return true
		}
	}
	for _, token := range []string{"claude", "sonnet", "opus", "haiku", "gpt", "openai", "codex"} {
		if strings.Contains(modelHintLower, token) {
			return strings.Contains(bucketKeyLower, "3p") ||
				strings.Contains(bucketKeyLower, "claude") ||
				strings.Contains(bucketKeyLower, "gpt")
		}
	}
	for _, token := range []string{"gemini", "neon", "flash", "orchestrator"} {
		if strings.Contains(modelHintLower, token) {
			return strings.Contains(bucketKeyLower, "gemini")
		}
	}
	return false
}

func formatSubagents(p *schema.Payload, pal theme.Palette) string {
	n := p.ActiveSubagentCount()
	if n == 0 {
		return ""
	}
	label := "agent"
	if n > 1 {
		label = "agents"
	}
	return fg(pal.Accent, fmt.Sprintf("⬡ %d %s", n, label))
}

func cycleModeSpec(mode string, pal theme.Palette) (icon, label string, col theme.RGB, ok bool) {
	raw := strings.ToLower(strings.TrimSpace(mode))
	if raw == "" {
		return "", "", theme.RGB{}, false
	}
	norm := strings.NewReplacer("_", "-", " ", "-").Replace(raw)
	switch norm {
	case "plan":
		return "❖", "plan", pal.Cyan, true
	case "auto":
		return "»", "auto", pal.Good, true
	case "accept-edits":
		return "✎", "accept-edits", pal.Warn, true
	case "default":
		return "◇", "default", pal.Subtext, true
	default:
		if strings.Contains(norm, "bypass") || strings.Contains(norm, "yolo") || strings.Contains(norm, "always") {
			return "⚠", norm, pal.High, true
		}
		return "◇", norm, pal.Subtext, true
	}
}

func formatCycleMode(mode string, pal theme.Palette) string {
	icon, label, col, ok := cycleModeSpec(mode, pal)
	if !ok {
		return ""
	}
	return fg(col, icon+" "+label)
}

func shortEffort(e string) string {
	switch strings.ToLower(strings.TrimSpace(e)) {
	case "medium":
		return "med"
	case "minimal":
		return "min"
	default:
		return strings.ToLower(strings.TrimSpace(e))
	}
}

func shortBucketName(k string) string {
	s := strings.ToLower(strings.TrimSpace(k))
	s = strings.TrimPrefix(s, "quota_")
	s = strings.TrimPrefix(s, "bucket_")
	replacer := strings.NewReplacer(
		"gemini_pro", "g-pro",
		"gemini-pro", "g-pro",
		"gemini_flash", "g-flash",
		"gemini-flash", "g-flash",
		"five_hour", "5h",
		"5_hour", "5h",
		"weekly", "7d",
	)
	s = replacer.Replace(s)
	if len(s) > 12 {
		s = s[:12]
	}
	return s
}

func formatCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatDuration(sec int64) string {
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	mins := sec / 60
	if mins < 60 {
		return fmt.Sprintf("%dm", mins)
	}
	hours := mins / 60
	remMins := mins % 60
	if hours >= 24 {
		days := hours / 24
		remHours := hours % 24
		if remHours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd%dh", days, remHours)
	}
	if remMins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remMins)
}

func shortenPath(p string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(p, home) {
		p = "~" + strings.TrimPrefix(p, home)
	}
	parts := strings.Split(filepath.Clean(p), string(filepath.Separator))
	if len(parts) <= 3 {
		return p
	}
	return filepath.Join(parts[0], "…", parts[len(parts)-2], parts[len(parts)-1])
}
