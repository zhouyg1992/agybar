package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config controls the appearance and enabled modules of agybar.
type Config struct {
	Theme      string `json:"theme"`       // catppuccin | tokyo-night | nord | mono
	Style      string `json:"style"`       // minimal | badge | pill
	BarStyle   string `json:"bar_style"`   // smooth | blocks | braille
	Separator  string `json:"separator"`   // pipe | dashed | dot | square | none | <custom>
	ShowTokens bool   `json:"show_tokens"` // show ↑in ↓out token counters
	ShowQuota  bool   `json:"show_quota"`  // show model quota buckets
	ShowSecond bool   `json:"show_second"` // allow 2-line layout on wide terminals
	CompactAt  int    `json:"compact_at"`  // collapse to 1-line when terminal_width < CompactAt
}

// Default returns the default aesthetic configuration.
func Default() Config {
	return Config{
		Theme:      "catppuccin",
		Style:      "minimal",
		BarStyle:   "blocks",
		Separator:  "pipe",
		ShowTokens: true,
		ShowQuota:  true,
		ShowSecond: true,
		CompactAt:  88,
	}
}

// Path returns ~/.config/agybar/config.json.
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "agybar", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "agybar.json"
	}
	return filepath.Join(home, ".config", "agybar", "config.json")
}

// Load reads config from disk, falling back to defaults for any missing fields.
func Load() Config {
	cfg := Default()
	data, err := os.ReadFile(Path())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.Theme == "" {
		cfg.Theme = "catppuccin"
	}
	if cfg.Style == "" {
		cfg.Style = "minimal"
	}
	if cfg.BarStyle == "" {
		cfg.BarStyle = "blocks"
	}
	if cfg.Separator == "" {
		cfg.Separator = "pipe"
	}
	if cfg.CompactAt <= 0 {
		cfg.CompactAt = 88
	}
	return cfg
}

// Save writes config to ~/.config/agybar/config.json.
func Save(cfg Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}
