package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/zhouyg1992/agybar/internal/theme"
)

const ansiReset = "\x1b[0m"

func fg(c theme.RGB, s string) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s%s", c.R, c.G, c.B, s, ansiReset)
}

func fgBold(c theme.RGB, s string) string {
	return fmt.Sprintf("\x1b[1;38;2;%d;%d;%dm%s%s", c.R, c.G, c.B, s, ansiReset)
}

func bgFg(bg, fgCol theme.RGB, s string) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%d;38;2;%d;%d;%dm%s%s", bg.R, bg.G, bg.B, fgCol.R, fgCol.G, fgCol.B, s, ansiReset)
}

// pill wraps inner text in a rounded Nerd Font capsule with a colored background.
func pill(bg, fgCol theme.RGB, inner string) string {
	leftCap := fg(bg, "\ue0b6")  // 
	rightCap := fg(bg, "\ue0b4") // 
	body := fmt.Sprintf("\x1b[48;2;%d;%d;%d;38;2;%d;%d;%dm %s %s", bg.R, bg.G, bg.B, fgCol.R, fgCol.G, fgCol.B, inner, ansiReset)
	return leftCap + body + rightCap
}

// lerpRGB linearly interpolates between two RGB colors for t in [0, 1].
func lerpRGB(a, b theme.RGB, t float64) theme.RGB {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	return theme.RGB{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
	}
}

// stripANSI removes ANSI CSI escape sequences for width calculations.
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			inEsc = true
			i++
			continue
		}
		if inEsc {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// visibleLen estimates terminal column width of a styled string.
func visibleLen(s string) int {
	plain := stripANSI(s)
	return utf8.RuneCountInString(plain)
}
