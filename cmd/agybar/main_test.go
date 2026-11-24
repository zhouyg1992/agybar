package main

import "testing"

func TestNormalizeSeparatorArg(t *testing.T) {
	validCases := []struct {
		raw  string
		want string
	}{
		{raw: "pipe", want: "pipe"},
		{raw: "PIPE", want: "pipe"},
		{raw: "dashed", want: "dashed"},
		{raw: " Dashed ", want: "dashed"},
		{raw: "dot", want: "dot"},
		{raw: "DoT", want: "dot"},
		{raw: "square", want: "square"},
		{raw: "SQUARE", want: "square"},
		{raw: "none", want: "none"},
		{raw: "None", want: "none"},
		{raw: "›", want: "›"},
		{raw: " // ", want: "//"},
		{raw: "X", want: "X"},
		{raw: "::", want: "::"},
	}

	for _, tc := range validCases {
		got, err := normalizeSeparatorArg(tc.raw)
		if err != nil {
			t.Errorf("normalizeSeparatorArg(%q) unexpected error: %v", tc.raw, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeSeparatorArg(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}

	invalidCases := []string{
		"",
		"   ",
		"\t",
		"a\nb",
		"x\x7fy",
		"foo\x00bar",
	}
	for _, raw := range invalidCases {
		got, err := normalizeSeparatorArg(raw)
		if err == nil {
			t.Errorf("normalizeSeparatorArg(%q) = %q, expected error", raw, got)
		}
	}
}
