# agybar

> A fast, aesthetic, zero-dependency Go statusline & real-time usage bar for **Google Antigravity CLI (`agy`)**.

```text
 ✦ Gemini 4 Neon [med] │ ▰▰▰▰▰▱▱▱▱▱ 52% (104.0k/200.0k) │ ↑92.4k ↓11.6k │ ⚡ 7d:64%(3d) · 5h:82%(2h15m)
 ◉ thinking │ ⎇ main* │ ~/…/github/agybar │ ⬡ 2 agents │ ⏳ 1
```

## Features

- **Zero dependencies, millisecond startup** — Single static Go binary; renders in a few milliseconds without slowing down `agy`.
- **4 curated 24-bit TrueColor themes** — `catppuccin` (default), `tokyo-night`, `nord`, and `mono`.
- **2 visual styles** — `minimal` (pure Unicode, works out of the box in any terminal) and `pill` (Nerd Font rounded capsules).
- **3 gradient progress bar styles** — `blocks` (`▰▱`), `smooth` (`━╸─`), and `braille` (`⣿⣤⣀`) with semantic color transitions (`green → yellow → peach → red`, plus an automatic `→ /compact` hint at `≥90%`).
- **Responsive layout** — Uses `terminal_width` to render 2 spacious lines on wide terminals and automatically collapse to 1 compact line on narrow panes (`<88` columns).
- **Rich `agy` telemetry** — Model & effort level, context window usage & token counters (`↑in ↓out`), model quota buckets with reset countdowns, agent state, Git branch/dirty status, active subagents, background tasks, artifacts, and sandbox indicator.

---

## Installation

### Option 1: `go install` (Recommended)

```bash
go install github.com/zhouyg1992/agybar/cmd/agybar@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

### Option 2: Build from Source

```bash
git clone git@github.com:zhouyg1992/agybar.git
cd agybar
go build -o bin/agybar ./cmd/agybar
```

### Optional: Install as an `agy` Plugin

```bash
agy plugin install https://github.com/zhouyg1992/agybar
```

---

## Quick Start

### 1. Preview themes & layouts in your terminal

```bash
agybar preview
```

### 2. Enable in Antigravity CLI (`agy`)

Add the `statusLine` block to `~/.gemini/antigravity-cli/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "agybar",
    "enabled": true,
    "padding": 0,
    "stack_with_default": false
  }
}
```

*(If `agybar` is not in your `PATH`, use the full binary path such as `/path/to/agybar/bin/agybar`.)*

Restart `agy`, and `agybar` will appear at the bottom of your terminal.

---

## Customization

### CLI Commands

| Command | Description |
| :--- | :--- |
| `agybar preview` | Preview all themes, high-usage warnings, compact mode, and pill style |
| `agybar theme <name>` | Switch theme: `catppuccin`, `tokyo-night`, `nord`, `mono` |
| `agybar style <style>` | Switch style: `minimal` (default) or `pill` (requires Nerd Font) |
| `agybar bar <style>` | Switch progress bar: `blocks`, `smooth`, `braille` |
| `agybar config` | Show active configuration and config file path |
| `agybar version` | Print version |

### Configuration File

Settings are persisted at `~/.config/agybar/config.json` (respects `XDG_CONFIG_HOME`):

```json
{
  "theme": "catppuccin",
  "style": "minimal",
  "bar_style": "blocks",
  "show_tokens": true,
  "show_quota": true,
  "show_second": true,
  "compact_at": 88
}
```

---

## Debugging

To inspect the raw JSON payload that `agy` pipes into `agybar`, run with `AGYBAR_DEBUG=1`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "AGYBAR_DEBUG=1 agybar",
    "enabled": true
  }
}
```

The latest payload will be dumped to `/tmp/agybar-last-payload.json`.

---

## License

[MIT](./LICENSE)
