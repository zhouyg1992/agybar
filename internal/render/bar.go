package render

import (
	"math"
	"strings"

	"github.com/zhouyg1992/agybar/internal/theme"
)

// ProgressBar renders a smooth gradient progress bar for pct in [0, 100].
func ProgressBar(pct float64, width int, style string, pal theme.Palette) string {
	if width <= 0 {
		return ""
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filledExact := (pct / 100.0) * float64(width)
	filledFull := int(math.Floor(filledExact))
	frac := filledExact - float64(filledFull)

	var fillChar, halfChar, emptyChar string
	switch strings.ToLower(style) {
	case "smooth":
		fillChar, halfChar, emptyChar = "━", "╸", "─"
	case "braille":
		fillChar, halfChar, emptyChar = "⣿", "⣤", "⣀"
	default: // blocks
		fillChar, halfChar, emptyChar = "▰", "▰", "▱"
	}

	// Build a subtle per-cell gradient across the filled portion.
	startCol := pal.Good
	endCol := pal.UsageColor(pct)

	var b strings.Builder
	for i := 0; i < width; i++ {
		t := 0.0
		if width > 1 {
			t = float64(i) / float64(width-1)
		}
		cellCol := lerpRGB(startCol, endCol, t)

		switch {
		case i < filledFull:
			b.WriteString(fg(cellCol, fillChar))
		case i == filledFull && frac >= 0.35:
			b.WriteString(fg(cellCol, halfChar))
		default:
			b.WriteString(fg(pal.Muted, emptyChar))
		}
	}
	return b.String()
}
