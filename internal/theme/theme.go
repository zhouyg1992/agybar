package theme

import "strings"

// RGB represents a 24-bit TrueColor value.
type RGB struct {
	R, G, B uint8
}

// Palette defines all semantic colors for agybar.
type Palette struct {
	Name string

	// Text & surfaces
	Text      RGB // primary text
	Subtext   RGB // secondary info (paths, separators)
	Muted     RGB // dim separators & empty bar track
	Surface   RGB // pill background
	SurfaceHi RGB // accented pill background

	// Brand & accents
	Accent  RGB // model name / primary highlight
	Accent2 RGB // git branch / secondary highlight
	Cyan    RGB // tool_use / info

	// Semantic usage levels
	Good RGB // 0-50% usage / high quota remaining
	Warn RGB // 50-75% usage / medium quota
	High RGB // 75-90% usage / low quota
	Crit RGB // >90% usage / critical quota
}

var presets = map[string]Palette{
	"catppuccin": {
		Name:      "catppuccin",
		Text:      RGB{205, 214, 244}, // Text
		Subtext:   RGB{166, 173, 200}, // Subtext0
		Muted:     RGB{88, 91, 112},   // Surface2
		Surface:   RGB{49, 50, 68},    // Surface0
		SurfaceHi: RGB{69, 71, 90},    // Surface1
		Accent:    RGB{203, 166, 247}, // Mauve
		Accent2:   RGB{137, 180, 250}, // Blue
		Cyan:      RGB{148, 226, 213}, // Teal
		Good:      RGB{166, 227, 161}, // Green
		Warn:      RGB{249, 226, 175}, // Yellow
		High:      RGB{250, 179, 135}, // Peach
		Crit:      RGB{243, 139, 168}, // Red
	},
	"tokyo-night": {
		Name:      "tokyo-night",
		Text:      RGB{192, 202, 245},
		Subtext:   RGB{154, 165, 206},
		Muted:     RGB{86, 95, 137},
		Surface:   RGB{36, 40, 59},
		SurfaceHi: RGB{52, 59, 88},
		Accent:    RGB{187, 154, 247}, // Purple
		Accent2:   RGB{122, 162, 247}, // Blue
		Cyan:      RGB{125, 207, 255}, // Cyan
		Good:      RGB{158, 206, 106}, // Green
		Warn:      RGB{224, 175, 104}, // Yellow
		High:      RGB{255, 158, 100}, // Orange
		Crit:      RGB{247, 118, 142}, // Red
	},
	"nord": {
		Name:      "nord",
		Text:      RGB{236, 239, 244},
		Subtext:   RGB{216, 222, 233},
		Muted:     RGB{76, 86, 106},
		Surface:   RGB{59, 66, 82},
		SurfaceHi: RGB{67, 76, 94},
		Accent:    RGB{136, 192, 208}, // Frost cyan
		Accent2:   RGB{129, 161, 193}, // Frost blue
		Cyan:      RGB{143, 188, 187}, // Frost teal
		Good:      RGB{163, 190, 140}, // Aurora green
		Warn:      RGB{235, 203, 139}, // Aurora yellow
		High:      RGB{208, 135, 112}, // Aurora orange
		Crit:      RGB{191, 97, 106},  // Aurora red
	},
	"mono": {
		Name:      "mono",
		Text:      RGB{235, 235, 235},
		Subtext:   RGB{165, 165, 165},
		Muted:     RGB{90, 90, 90},
		Surface:   RGB{42, 42, 42},
		SurfaceHi: RGB{62, 62, 62},
		Accent:    RGB{255, 255, 255},
		Accent2:   RGB{200, 200, 200},
		Cyan:      RGB{215, 215, 215},
		Good:      RGB{210, 210, 210},
		Warn:      RGB{235, 203, 139},
		High:      RGB{250, 179, 135},
		Crit:      RGB{243, 139, 168},
	},
}

// Get returns the named palette, defaulting to catppuccin if unknown.
func Get(name string) Palette {
	if p, ok := presets[strings.ToLower(strings.TrimSpace(name))]; ok {
		return p
	}
	return presets["catppuccin"]
}

// Names returns all built-in theme names in display order.
func Names() []string {
	return []string{"catppuccin", "tokyo-night", "nord", "mono"}
}

// UsageColor picks a semantic color based on used percentage [0, 100].
func (p Palette) UsageColor(usedPct float64) RGB {
	switch {
	case usedPct >= 90:
		return p.Crit
	case usedPct >= 75:
		return p.High
	case usedPct >= 50:
		return p.Warn
	default:
		return p.Good
	}
}

// QuotaRemainingColor picks a semantic color based on remaining percentage [0, 100].
func (p Palette) QuotaRemainingColor(remPct float64) RGB {
	switch {
	case remPct <= 10:
		return p.Crit
	case remPct <= 25:
		return p.High
	case remPct <= 50:
		return p.Warn
	default:
		return p.Good
	}
}
