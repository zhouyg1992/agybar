package render

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/zhouyg1992/agybar/internal/theme"
)

const ansiReset = "\x1b[0m"

func useTrueColor() bool {
	if override := strings.ToLower(strings.TrimSpace(os.Getenv("AGYBAR_COLOR"))); override != "" {
		switch override {
		case "truecolor", "24bit":
			return true
		case "256":
			return false
		}
	}

	termProg := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))
	if termProg == "apple_terminal" {
		return false
	}

	ct := strings.ToLower(strings.TrimSpace(os.Getenv("COLORTERM")))
	if ct == "truecolor" || ct == "24bit" {
		return true
	}

	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") || strings.Contains(term, "direct") {
		return true
	}

	switch termProg {
	case "iterm.app", "wezterm", "vscode", "warpterminal", "ghostty", "alacritty", "kitty", "hyper", "tabby", "rio":
		return true
	}
	for _, envKey := range []string{"ITERM_SESSION_ID", "WT_SESSION", "KITTY_WINDOW_ID", "GHOSTTY_RESOURCES_DIR", "WEZTERM_EXECUTABLE"} {
		if os.Getenv(envKey) != "" {
			return true
		}
	}

	if strings.HasSuffix(term, "256color") {
		return false
	}

	return true
}

func rgbToANSI256(c theme.RGB) uint8 {
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	nearestLevel := func(v uint8) int {
		bestIdx := 0
		bestDist := int(v)
		for i := 1; i < len(levels); i++ {
			d := int(v) - int(levels[i])
			if d < 0 {
				d = -d
			}
			if d < bestDist {
				bestDist = d
				bestIdx = i
			}
		}
		return bestIdx
	}

	ri := nearestLevel(c.R)
	gi := nearestLevel(c.G)
	bi := nearestLevel(c.B)
	cubeR, cubeG, cubeB := int(levels[ri]), int(levels[gi]), int(levels[bi])
	cubeDist := (int(c.R)-cubeR)*(int(c.R)-cubeR) +
		(int(c.G)-cubeG)*(int(c.G)-cubeG) +
		(int(c.B)-cubeB)*(int(c.B)-cubeB)
	cubeIdx := uint8(16 + 36*ri + 6*gi + bi)

	avg := (int(c.R) + int(c.G) + int(c.B)) / 3
	grayStep := (avg - 8 + 5) / 10
	if grayStep < 0 {
		grayStep = 0
	}
	if grayStep > 23 {
		grayStep = 23
	}
	grayVal := 8 + 10*grayStep
	grayDist := (int(c.R)-grayVal)*(int(c.R)-grayVal) +
		(int(c.G)-grayVal)*(int(c.G)-grayVal) +
		(int(c.B)-grayVal)*(int(c.B)-grayVal)
	grayIdx := uint8(232 + grayStep)

	if grayDist < cubeDist {
		return grayIdx
	}
	return cubeIdx
}

func fg(c theme.RGB, s string) string {
	if useTrueColor() {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s%s", c.R, c.G, c.B, s, ansiReset)
	}
	return fmt.Sprintf("\x1b[38;5;%dm%s%s", rgbToANSI256(c), s, ansiReset)
}

func fgBold(c theme.RGB, s string) string {
	if useTrueColor() {
		return fmt.Sprintf("\x1b[1;38;2;%d;%d;%dm%s%s", c.R, c.G, c.B, s, ansiReset)
	}
	return fmt.Sprintf("\x1b[1;38;5;%dm%s%s", rgbToANSI256(c), s, ansiReset)
}

func bgFg(bg, fgCol theme.RGB, s string) string {
	if useTrueColor() {
		return fmt.Sprintf("\x1b[48;2;%d;%d;%d;38;2;%d;%d;%dm%s%s", bg.R, bg.G, bg.B, fgCol.R, fgCol.G, fgCol.B, s, ansiReset)
	}
	return fmt.Sprintf("\x1b[48;5;%d;38;5;%dm%s%s", rgbToANSI256(bg), rgbToANSI256(fgCol), s, ansiReset)
}

// pill wraps inner text in a colored capsule; when rounded is true, Nerd Font / Powerline endcaps are added.
func pill(bg, fgCol theme.RGB, inner string, rounded bool) string {
	body := bgFg(bg, fgCol, " "+inner+" ")
	if !rounded {
		return body
	}
	leftCap := fg(bg, "\ue0b6")  // 
	rightCap := fg(bg, "\ue0b4") // 
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
