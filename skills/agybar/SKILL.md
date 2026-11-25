---
name: agybar
description: Install, activate, configure, preview, and customize themes or visual styles for the agybar statusline in Antigravity CLI (agy). Use when the user asks to enable, set up, configure, preview, or change the theme/style of agybar.
---

# agybar

`agybar` is a fast, zero-dependency Go statusline & real-time usage bar for Antigravity CLI (`agy`).

## Auto-Activation Workflow (after `agy plugin install`)

When the user asks to **enable**, **activate**, or **set up** `agybar`:

1. **Ensure the `agybar` binary is installed** (prefer `~/.local/bin/agybar`):
   - Check if `agybar` is already on `PATH` or at `~/.local/bin/agybar`.
   - If not found, run `curl -fsSL https://raw.githubusercontent.com/zhouyg1992/agybar/main/install.sh | sh` (downloads the prebuilt GitHub Release binary with checksum verification).
2. **Update `~/.gemini/antigravity-cli/settings.json`** (preserve all existing keys!):
   Merge the following `statusLine` block into `~/.gemini/antigravity-cli/settings.json`:
   ```json
   {
     "statusLine": {
       "type": "command",
       "command": "/ABSOLUTE/PATH/TO/agybar",
       "enabled": true,
       "padding": 0,
       "stack_with_default": false
     }
   }
   ```
3. **Verify**: Run `agybar preview` or `agybar config` and tell the user to restart `agy` to see `agybar` active at the bottom of the terminal.

## Common Customization Commands
- Preview all themes and layouts: `agybar preview`
- Switch theme (`catppuccin`, `tokyo-night`, `nord`, `gruvbox`, `rose-pine`, `everforest`, `mono`): `agybar theme <name>`
- Switch style (`minimal`, `badge`, `pill`): `agybar style <style>`
- Switch progress bar (`blocks`, `smooth`, `braille`): `agybar bar <bar>`
- Switch `minimal` separator (`pipe`, `dashed`, `dot`, `square`, `none`, or custom text): `agybar sep <preset|text>`
- Inspect current config: `agybar config`
