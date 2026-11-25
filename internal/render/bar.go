package render

import (
	"math"
	"strings"

	"github.com/zhouyg1992/agybar/internal/theme"
)

// BarStyleNames returns all built-in progress bar style names in display order.
func BarStyleNames() []string {
	return []string{"blocks", "smooth", "braille", "slider", "dots", "micro"}
}

// ResolveBarStyle normalizes a bar style name to its canonical built-in name.
func ResolveBarStyle(name string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	for _, s := range BarStyleNames() {
		if key == s {
			return s, true
		}
	}
	return "", false
}

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

	// Build a subtle per-cell gradient across the filled portion.
	startCol := pal.Good
	endCol := pal.UsageColor(pct)
	normStyle := strings.ToLower(strings.TrimSpace(style))

	if normStyle == "slider" {
		thumbIdx := 0
		if width > 1 {
			thumbIdx = int(math.Round((pct / 100.0) * float64(width-1)))
			if thumbIdx < 0 {
				thumbIdx = 0
			}
			if thumbIdx >= width {
				thumbIdx = width - 1
			}
		}

		var b strings.Builder
		for i := 0; i < width; i++ {
			if pct == 0 {
				if i == 0 {
					b.WriteString(fg(pal.Muted, "•"))
				} else {
					b.WriteString(fg(pal.Muted, "─"))
				}
				continue
			}

			t := 0.0
			if width > 1 {
				t = float64(i) / float64(width-1)
			}
			cellCol := lerpRGB(startCol, endCol, t)

			switch {
			case i < thumbIdx:
				b.WriteString(fg(cellCol, "─"))
			case i == thumbIdx:
				b.WriteString(fgBold(cellCol, "•"))
			default:
				b.WriteString(fg(pal.Muted, "─"))
			}
		}
		return b.String()
	}

	filledExact := (pct / 100.0) * float64(width)
	filledFull := int(math.Floor(filledExact))
	frac := filledExact - float64(filledFull)

	var fillChar, halfChar, emptyChar string
	switch normStyle {
	case "smooth":
		fillChar, halfChar, emptyChar = "━", "╸", "─"
	case "braille":
		fillChar, halfChar, emptyChar = "⣿", "⣤", "⣀"
	case "dots":
		fillChar, halfChar, emptyChar = "•", "•", "•"
	case "micro":
		fillChar, halfChar, emptyChar = "▪", "▪", "▪"
	default: // blocks
		fillChar, halfChar, emptyChar = "▰", "▰", "▰"
	}

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
