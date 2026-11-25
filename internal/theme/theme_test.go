package theme

import "testing"

func TestNamesResolveAndGet(t *testing.T) {
	names := Names()
	if len(names) != 7 {
		t.Fatalf("len(Names()) = %d, want 7", len(names))
	}
	for _, n := range names {
		canonical, ok := ResolveName(n)
		if !ok || canonical != n {
			t.Errorf("ResolveName(%q) = (%q, %v), want (%q, true)", n, canonical, ok, n)
		}
		pal := Get(n)
		if pal.Name != n {
			t.Errorf("Get(%q).Name = %q, want %q", n, pal.Name, n)
		}
	}
}

func TestResolveNormalizationAndFallback(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{input: " Rose-Pine ", want: "rose-pine"},
		{input: "GRUVBOX", want: "gruvbox"},
		{input: "  Everforest  ", want: "everforest"},
		{input: "TOKYO-NIGHT", want: "tokyo-night"},
	}
	for _, tc := range cases {
		canonical, ok := ResolveName(tc.input)
		if !ok || canonical != tc.want {
			t.Errorf("ResolveName(%q) = (%q, %v), want (%q, true)", tc.input, canonical, ok, tc.want)
		}
		if got := Get(tc.input).Name; got != tc.want {
			t.Errorf("Get(%q).Name = %q, want %q", tc.input, got, tc.want)
		}
	}

	for _, invalid := range []string{"unknown-theme", "latte", "light", "catppuccin-latte"} {
		if canonical, ok := ResolveName(invalid); ok || canonical != "" {
			t.Errorf("ResolveName(%q) = (%q, %v), want (\"\", false)", invalid, canonical, ok)
		}
		if got := Get(invalid).Name; got != "catppuccin" {
			t.Errorf("Get(%q).Name = %q, want catppuccin", invalid, got)
		}
	}
}
