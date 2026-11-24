package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSeparator(t *testing.T) {
	cfg := Default()
	if cfg.Separator != "pipe" {
		t.Fatalf("Default().Separator = %q, want %q", cfg.Separator, "pipe")
	}
}

func TestLoadLegacyConfigDefaultsSeparator(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfgPath := Path()
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	legacy := []byte(`{"theme":"nord","style":"minimal","bar_style":"smooth"}`)
	if err := os.WriteFile(cfgPath, legacy, 0o644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	cfg := Load()
	if cfg.Theme != "nord" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "nord")
	}
	if cfg.Separator != "pipe" {
		t.Errorf("Separator = %q, want %q for legacy config", cfg.Separator, "pipe")
	}
}

func TestSaveAndLoadSeparatorRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	for _, sep := range []string{"dashed", "dot", "square", "none", "›", "//"} {
		cfg := Default()
		cfg.Separator = sep
		if err := Save(cfg); err != nil {
			t.Fatalf("Save(%q): %v", sep, err)
		}
		got := Load()
		if got.Separator != sep {
			t.Errorf("round-trip Separator = %q, want %q", got.Separator, sep)
		}
	}
}
