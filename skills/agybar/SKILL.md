---
name: agybar
description: Configure, preview, and switch themes or visual styles for the agybar statusline in Antigravity CLI (agy).
---

# agybar

`agybar` is a fast, zero-dependency Go statusline for Antigravity CLI (`agy`).

## Common Commands
- Preview all themes and layouts: `agybar preview`
- Switch theme (`catppuccin`, `tokyo-night`, `nord`, `mono`): `agybar theme <name>`
- Switch style (`minimal`, `pill`): `agybar style <style>`
- Switch progress bar (`blocks`, `smooth`, `braille`): `agybar bar <bar>`
- Inspect current config: `agybar config`

## Enable in `~/.gemini/antigravity-cli/settings.json`
```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/presence/code/github/agybar/bin/agybar",
    "enabled": true,
    "padding": 0,
    "stack_with_default": false
  }
}
```
