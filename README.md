oldSess.ActiveProject# trak

A minimal CLI time tracker for freelancers. Register projects, switch between them instantly via a hotkey, and get a clean end-of-day report.

Built with Go. No cloud, no accounts, no bloat — just a fast local daemon and two keystrokes.

---

## How it works

```
CLI (trak) ──── unix socket ────► Daemon (trakd)
                                     holds session state in memory
                                     checkpoints to disk on every action
```

`trakd` runs in the background for the duration of your workday. `trak` is the CLI you interact with — or don't, once your hotkeys are set up.

State is persisted after every action under the configured `sessions_dir` (default to `~/.trak/sessions/`). If `trakd` crashes mid-day, your data is safe and you'll be offered to resume on next `trak start`.

*NOTE*
The `sessions_dir` is configurable in `~/.trak/config.json` and its default is derived from `os.UserHomeDir()`, so the actual path may differ from the literal `~/.trak/sessions` depending on your system.
---

## Install

### Option 1 — Download binary (recommended)

Download the latest release for your platform from the [releases page](../../releases):

```bash
# example for macOS Apple Silicon
tar -xzf trak-darwin-arm64.tar.gz
mv trak-darwin-arm64/trak  ~/.local/bin/trak
mv trak-darwin-arm64/trakd ~/.local/bin/trakd
```

### Option 2 — Build from source

**Requirements:** Go 1.26+

```bash
git clone https://github.com/CRSylar/trak
cd trak
make install        # builds and installs to ~/.local/bin/
```

### Option 3 — go install

```bash
go install github.com/CRSylar/trak/cmd/trak@latest
go install github.com/CRSylar/trak/cmd/trakd@latest
```

> Both binaries are required — `trakd` is the background daemon that `trak` launches automatically.

### PATH setup

Make sure your install location is in your `$PATH`. Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# pick the line matching your install method
export PATH="$HOME/.local/bin:$PATH"   # make install / binary download
export PATH="$HOME/go/bin:$PATH"       # go install
```

---

## Hotkey setup — macOS (Raycast)

1. Open Raycast → **Settings → Script Commands → Add Directory**
2. Select the `scripts/raycast/` folder from this repo
3. Find **"Next Work Project"** → assign a hotkey (e.g. `⌥N`)
4. Find **"Rest"** → assign a hotkey (e.g. `⌥R`)

That's it. You'll never need to touch the mouse to switch projects.

---

## Hotkey setup — Linux (rofi / wofi)

Copy the scripts from `scripts/linux/` to a convenient location:

```bash
cp scripts/linux/*.sh ~/.local/bin/
chmod +x ~/.local/bin/*.sh
```

Then bind hotkeys in your WM/DE settings pointing to each script:

| Script | Suggested hotkey | Action |
|---|---|---|
| `next-project.sh` | `Super+Tab` | Cycle to next project |
| `rest.sh` | `Super+R` | Switch to rest |
| `switch-project.rofi.sh` | `Super+Space` | Open rofi picker |
| `switch-project.wofi.sh` | `Super+Space` | Open wofi picker (Wayland) |

**Dependencies:** `python3` (for JSON parsing in the scripts), `notify-send` for toast notifications, plus `rofi` or `wofi` for the picker:

```bash
# Debian / Ubuntu
sudo apt install python3 libnotify-bin rofi   # or wofi

# Arch
sudo pacman -S python libnotify rofi         # or wofi

# Fedora
sudo dnf install python3 libnotify rofi      # or wofi
```

**WM-specific examples:**

i3 / Sway — add to your config:
```
bindsym $mod+Tab   exec ~/.local/bin/next-project.sh
bindsym $mod+r     exec ~/.local/bin/rest.sh
bindsym $mod+space exec ~/.local/bin/switch-project.rofi.sh
```

GNOME — Settings → Keyboard → Custom Shortcuts → add each script.

KDE Plasma — System Settings → Shortcuts → Custom Shortcuts → add each script.

---

## Daily usage

```bash
# Morning — start the workday
trak start

# First time only — register your projects (saved permanently)
trak register client-alpha
trak register client-beta
trak register internal

# Switch projects
trak next          # cycles: client-alpha → client-beta → internal → client-alpha
trak rest          # jump to rest immediately
trak switch <n>    # go to a specific project by name

# Forgot to switch for 20 minutes? Shift the last boundary back
trak edit --last 20m

# Check where you are
trak status

# Evening — end the workday and print the report
trak end
```

### Crash recovery

If `trakd` crashes mid-day, your session is safe on disk. On next `trak start`:

```
⚠️  Unfinished session found — last active: client-alpha (at 11:30)
Resume it? [y/n]:
```

`y` — reloads the session and picks up where you left off.
`n` — closes the old session and starts a fresh one (`2026-03-25-extra.json`).

### End-of-day report

```
WorkDay ended — 18 Mar 2026
Total: 7h 42m  (09:00 → 16:42)

┌──────────────┬──────────┬──────┐
│ Project      │ Time     │    % │
├──────────────┼──────────┼──────┤
│ client-alpha │ 3h 15m   │  42% │
│ client-beta  │ 2h 10m   │  28% │
│ internal     │ 1h 05m   │  14% │
│ rest         │ 1h 12m   │  16% │
└──────────────┴──────────┴──────┘
```

---

## All commands

| Command | Description |
|---|---|
| `trak start` | Start workday, launch daemon |
| `trak end` | End workday, print report, stop daemon |
| `trak next` | Cycle to the next work project (skips rest) |
| `trak rest` | Switch to rest immediately |
| `trak switch <n>` | Switch to a specific project |
| `trak edit --last <duration>` | Shift the last switch back by duration |
| `trak status` | Current project + elapsed time |
| `trak projects` | List all registered projects |
| `trak register <n>` | Register a new project |
| `trak unregister <n>` | Remove a project |
| `trak version` | Print version |

### Duration format for `trak edit`

| Input | Meaning |
|---|---|
| `15m` | 15 minutes |
| `1h` | 1 hour |
| `1h15m` | 1 hour and 15 minutes |
| `--last 1h --last 15m` | chained flags, summed to 1h15m |

---

## Configuration

On first run, `~/.trak/config.json` is created with defaults (with your home directory expanded):

```json
{
  "sessions_dir": "/home/<user>/.trak/sessions"
}
```

Change `sessions_dir` to store session files anywhere you like (e.g. a synced folder).

---

## Session files

Each workday produces a JSON file in `sessions_dir`:

```
~/.trak/sessions/
  2026-03-25.json         # normal day
  2026-03-25-extra.json   # created if you discarded a crashed session and restarted
```

Files with `"closed": true` are finished days. Any file without it (or with `false`) is considered an active session and will trigger the resume prompt.

---

## Notes

- `rest` is a built-in project — always available, cannot be unregistered
- `trak edit` only affects the last switch boundary — closed days are immutable
- The `trak next` cycle order is alphabetical and excludes `rest`
- If you need to force-stop the daemon: `pkill trakd`

---

## Roadmap

| Milestone | Highlights |
|---|---|
| **v0.4** | `trak report [date]`, weekly summary, CSV export |
| **v1.0** | Idle detection, shell prompt integration, Homebrew tap |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). All contributions welcome.

## License

MIT — see [LICENSE](LICENSE.md).
