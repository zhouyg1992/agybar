package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zhouyg1992/agybar/internal/config"
	"github.com/zhouyg1992/agybar/internal/render"
	"github.com/zhouyg1992/agybar/internal/schema"
	"github.com/zhouyg1992/agybar/internal/theme"
	"github.com/zhouyg1992/agybar/internal/vcs"
)

var version = "0.2.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "preview":
			runPreview()
			return
		case "theme":
			runSetTheme(os.Args[2:])
			return
		case "style":
			runSetStyle(os.Args[2:])
			return
		case "bar":
			runSetBar(os.Args[2:])
			return
		case "sep", "separator":
			runSetSeparator(os.Args[2:])
			return
		case "config":
			runShowConfig()
			return
		case "version", "-v", "--version":
			fmt.Printf("agybar %s\n", version)
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	runStatusLine()
}

func runStatusLine() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	if os.Getenv("AGYBAR_DEBUG") == "1" {
		_ = os.WriteFile("/tmp/agybar-last-payload.json", data, 0o644)
	}

	payload, err := schema.Parse(data)
	if err != nil {
		// Never break the CLI UI if stdin contains unexpected data.
		fmt.Printf(" ✦ agybar (parse error)\n")
		return
	}

	payload.VCS = vcs.Resolve(payload.VCS, payload.EffectiveDir())

	cfg := config.Load()
	fmt.Println(render.Render(payload, cfg))
}

func runPreview() {
	cfg := config.Load()

	ptrF := func(v float64) *float64 { return &v }
	ptrI := func(v int64) *int64 { return &v }

	demoNormal := &schema.Payload{
		CWD:        "/Users/presence/code/github/agybar",
		AgentState: "thinking",
		CycleMode:  "auto",
		Model: &schema.ModelInfo{
			ID:          "gemini-4-neon",
			DisplayName: "Gemini 4 Neon",
			Effort:      "medium",
		},
		ContextWindow: &schema.ContextWindow{
			TotalInputTokens:  92400,
			TotalOutputTokens: 11600,
			ContextWindowSize: 200000,
			UsedPercentage:    ptrF(52),
		},
		Quota: map[string]schema.QuotaBucket{
			"5h":     {RemainingFraction: ptrF(0.82), ResetInSeconds: ptrI(8100)},
			"weekly": {RemainingFraction: ptrF(0.64), ResetInSeconds: ptrI(259200)},
		},
		VCS:           &schema.VCSInfo{Branch: "main", Dirty: true},
		Subagents:     []schema.SubagentInfo{{Name: "research", Status: "running"}, {Name: "review", Status: "running"}},
		TaskCount:     1,
		TerminalWidth: 120,
	}

	demoHigh := &schema.Payload{
		CWD:        "/Users/presence/code/github/agybar",
		AgentState: "tool_use",
		CycleMode:  "plan",
		Model: &schema.ModelInfo{
			ID:          "claude-sonnet",
			DisplayName: "Claude Sonnet",
			Effort:      "high",
		},
		ContextWindow: &schema.ContextWindow{
			TotalInputTokens:  168000,
			TotalOutputTokens: 18500,
			ContextWindowSize: 200000,
			UsedPercentage:    ptrF(93),
		},
		Quota: map[string]schema.QuotaBucket{
			"5h":     {RemainingFraction: ptrF(0.14), ResetInSeconds: ptrI(2700)},
			"weekly": {RemainingFraction: ptrF(0.38), ResetInSeconds: ptrI(172800)},
		},
		VCS:           &schema.VCSInfo{Branch: "feat/hud", Dirty: false},
		Sandbox:       &schema.SandboxInfo{Enabled: true},
		TerminalWidth: 120,
	}

	demoNarrow := *demoNormal
	demoNarrow.TerminalWidth = 72

	fmt.Println("\x1b[1m=== agybar preview (current config: theme=" + cfg.Theme + ", style=" + cfg.Style + ", bar=" + cfg.BarStyle + ", sep=" + cfg.Separator + ") ===\x1b[0m")
	fmt.Println()

	for _, th := range theme.Names() {
		c := cfg
		c.Theme = th
		fmt.Printf("\x1b[2m── theme: %s (minimal / wide) ──\x1b[0m\n", th)
		c.Style = "minimal"
		fmt.Println(render.Render(demoNormal, c))
		fmt.Println()
	}

	fmt.Printf("\x1b[2m── high usage warning (>90%%) · theme: %s ──\x1b[0m\n", cfg.Theme)
	fmt.Println(render.Render(demoHigh, cfg))
	fmt.Println()

	fmt.Printf("\x1b[2m── compact mode (narrow terminal <88 cols) · theme: %s ──\x1b[0m\n", cfg.Theme)
	fmt.Println(render.Render(&demoNarrow, cfg))
	fmt.Println()

	fmt.Printf("\x1b[2m── badge style (universal) · theme: %s ──\x1b[0m\n", cfg.Theme)
	badgeCfg := cfg
	badgeCfg.Style = "badge"
	fmt.Println(render.Render(demoNormal, badgeCfg))
	fmt.Println()

	fmt.Printf("\x1b[2m── pill style (Nerd Font / iTerm2 Powerline) · theme: %s ──\x1b[0m\n", cfg.Theme)
	pillCfg := cfg
	pillCfg.Style = "pill"
	fmt.Println(render.Render(demoNormal, pillCfg))
	fmt.Println()

	fmt.Printf("\x1b[2m── minimal separators · theme: %s ──\x1b[0m\n", cfg.Theme)
	for _, sep := range []string{"pipe", "dashed", "dot", "square", "none"} {
		sepCfg := cfg
		sepCfg.Style = "minimal"
		sepCfg.Separator = sep
		fmt.Printf("\x1b[2m[%s]\x1b[0m\n", sep)
		fmt.Println(render.Render(demoNormal, sepCfg))
	}
	fmt.Println()

	fmt.Printf("\x1b[2m── progress bar styles · theme: %s ──\x1b[0m\n", cfg.Theme)
	pal := theme.Get(cfg.Theme)
	for _, bs := range render.BarStyleNames() {
		fmt.Printf("  %-8s %s 4%%   %s 52%%   %s 93%%\n",
			bs,
			render.ProgressBar(4, 10, bs, pal),
			render.ProgressBar(52, 10, bs, pal),
			render.ProgressBar(93, 10, bs, pal),
		)
	}
	fmt.Println()
}

func runSetTheme(args []string) {
	if len(args) == 0 {
		fmt.Printf("Available themes: %s\n", strings.Join(theme.Names(), ", "))
		return
	}
	canonical, ok := theme.ResolveName(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown theme %q. Available: %s\n", strings.TrimSpace(args[0]), strings.Join(theme.Names(), ", "))
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Theme = canonical
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Theme set to %s (%s)\n", canonical, config.Path())
}

func runSetStyle(args []string) {
	if len(args) == 0 {
		fmt.Println("Available styles: minimal, badge, pill")
		return
	}
	st := strings.ToLower(args[0])
	if st != "minimal" && st != "badge" && st != "pill" {
		fmt.Fprintf(os.Stderr, "Unknown style %q. Available: minimal, badge, pill\n", st)
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Style = st
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Style set to %s (%s)\n", st, config.Path())
}

func runSetBar(args []string) {
	if len(args) == 0 {
		fmt.Printf("Available bar styles: %s\n", strings.Join(render.BarStyleNames(), ", "))
		return
	}
	bs, ok := render.ResolveBarStyle(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown bar style %q. Available: %s\n", strings.TrimSpace(args[0]), strings.Join(render.BarStyleNames(), ", "))
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.BarStyle = bs
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Bar style set to %s (%s)\n", bs, config.Path())
}

func normalizeSeparatorArg(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("separator cannot be empty (use \"none\" for blank spacing, or pass a preset/custom character)")
	}
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("separator cannot contain control characters")
		}
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "pipe", "dashed", "dot", "square", "none":
		return lower, nil
	default:
		return trimmed, nil
	}
}

func runSetSeparator(args []string) {
	if len(args) == 0 {
		fmt.Println(`Available separators: pipe (│), dashed (┊), dot (•), square (▪), none (blank), or any custom text (e.g. agybar sep "›")`)
		return
	}
	sep, err := normalizeSeparatorArg(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid separator: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Separator = sep
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Separator set to %s (%s)\n", sep, config.Path())
}

func runShowConfig() {
	cfg := config.Load()
	fmt.Printf("Config file: %s\n", config.Path())
	fmt.Printf("  theme:       %s\n", cfg.Theme)
	fmt.Printf("  style:       %s\n", cfg.Style)
	fmt.Printf("  bar_style:   %s\n", cfg.BarStyle)
	fmt.Printf("  separator:   %s\n", cfg.Separator)
	fmt.Printf("  show_tokens: %v\n", cfg.ShowTokens)
	fmt.Printf("  show_quota:  %v\n", cfg.ShowQuota)
	fmt.Printf("  show_second: %v\n", cfg.ShowSecond)
	fmt.Printf("  compact_at:  %d\n", cfg.CompactAt)
}

func printHelp() {
	fmt.Printf(`agybar %s — A fast, aesthetic statusline & usage bar for Antigravity CLI (agy)

Usage:
  agybar                         Read agy statusLine JSON from stdin and render
  agybar preview                 Preview all themes, styles, and responsive layouts
  agybar theme <name>            Switch theme (%s)
  agybar style <minimal|badge|pill>  Switch visual style
  agybar bar <%s>  Switch progress bar style
  agybar sep <preset|text>       Switch minimal separator (pipe | dashed | dot | square | none | <custom>)
  agybar config                  Show active configuration
  agybar version                 Print version

Debug:
  Set AGYBAR_DEBUG=1 to dump raw stdin JSON to /tmp/agybar-last-payload.json
`, version, strings.Join(theme.Names(), " | "), strings.Join(render.BarStyleNames(), "|"))
}
