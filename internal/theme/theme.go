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
	"gruvbox": {
		Name:      "gruvbox",
		Text:      RGB{235, 219, 178}, // fg1
		Subtext:   RGB{189, 174, 147}, // fg3
		Muted:     RGB{124, 111, 100}, // bg4
		Surface:   RGB{60, 56, 54},    // bg1
		SurfaceHi: RGB{80, 73, 69},    // bg2
		Accent:    RGB{254, 128, 25},  // orange
		Accent2:   RGB{131, 165, 152}, // blue
		Cyan:      RGB{142, 192, 124}, // aqua
		Good:      RGB{184, 187, 38},  // green
		Warn:      RGB{250, 189, 47},  // yellow
		High:      RGB{254, 128, 25},  // orange
		Crit:      RGB{251, 73, 52},   // red
	},
	"rose-pine": {
		Name:      "rose-pine",
		Text:      RGB{224, 222, 244}, // Text
		Subtext:   RGB{144, 140, 170}, // Subtle
		Muted:     RGB{110, 106, 134}, // Muted
		Surface:   RGB{38, 35, 58},    // Overlay
		SurfaceHi: RGB{64, 61, 82},    // Highlight Med
		Accent:    RGB{235, 188, 186}, // Rose
		Accent2:   RGB{196, 167, 231}, // Iris
		Cyan:      RGB{156, 207, 216}, // Foam
		Good:      RGB{126, 196, 166}, // Sage green
		Warn:      RGB{246, 193, 119}, // Gold
		High:      RGB{234, 154, 151}, // Rose coral
		Crit:      RGB{235, 111, 146}, // Love
	},
	"everforest": {
		Name:      "everforest",
		Text:      RGB{211, 198, 170}, // fg
		Subtext:   RGB{157, 169, 160}, // grey2
		Muted:     RGB{122, 132, 120}, // grey0
		Surface:   RGB{52, 63, 68},    // bg2
		SurfaceHi: RGB{67, 79, 85},    // bg4
		Accent:    RGB{167, 192, 128}, // green
		Accent2:   RGB{127, 187, 179}, // blue
		Cyan:      RGB{131, 192, 146}, // aqua
		Good:      RGB{167, 192, 128}, // green
		Warn:      RGB{219, 188, 127}, // yellow
		High:      RGB{230, 152, 117}, // orange
		Crit:      RGB{230, 126, 128}, // red
	},
}

// ResolveName normalizes a theme name to its canonical built-in theme name.
func ResolveName(name string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	if _, ok := presets[key]; ok {
		return key, true
	}
	return "", false
}

// Get returns the named palette, defaulting to catppuccin if unknown.
func Get(name string) Palette {
	if canonical, ok := ResolveName(name); ok {
		return presets[canonical]
	}
	return presets["catppuccin"]
}

// Names returns all built-in theme names in display order.
func Names() []string {
	return []string{"catppuccin", "tokyo-night", "nord", "gruvbox", "rose-pine", "everforest", "mono"}
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
