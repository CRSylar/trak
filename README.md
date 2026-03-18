# track
 
A minimal CLI time tracker for freelancers. Register projects, switch between them instantly via a hotkey, and get a clean end-of-day report.
 
Built with Go. No cloud, no accounts, no bloat — just a fast local daemon and two keystrokes.
 
---
 
## How it works
 
```
CLI (track) ──── unix socket ────► Daemon (trackd)
                                     holds session state in memory
```
 
`trackd` runs in the background for the duration of your workday. `track` is the CLI you interact with — or don't, once your hotkeys are set up. Project registrations are persisted to `~/.track/projects.json`.
 
---
 
## Install
 
**Requirements:** Go 1.22+, macOS (Apple Silicon)
 
> Linux support is planned for v0.2. See [CONTRIBUTING.md](CONTRIBUTING.md).
 
```bash
git clone https://github.com/you/track
cd track
make install
```
 
Installs both binaries to `~/.local/bin/`. Make sure that's in your `$PATH`:
 
```bash
# add to ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```
 
---
 
## Hotkey setup (macOS + Raycast)
 
1. Open Raycast → **Settings → Script Commands → Add Directory**
2. Select the `scripts/raycast/` folder from this repo
3. Find **"Next Work Project"** → assign a hotkey (e.g. `⌥N`)
4. Find **"Rest"** → assign a hotkey (e.g. `⌥R`)
 
That's it. You'll never need to touch the mouse to switch projects.
 
---
 
## Daily usage
 
```bash
# Morning — start the workday
track start
 
# First time only — register your projects (saved permanently)
track register client-alpha
track register client-beta
track register internal
 
# Switch projects
track next          # cycles: client-alpha → client-beta → internal → client-alpha
track rest          # jump to rest immediately
track switch <name> # go to a specific project by name
 
# Check where you are
track status
 
# Evening — end the workday and print the report
track end
```
 
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
| `track start` | Start workday, launch daemon |
| `track end` | End workday, print report, stop daemon |
| `track next` | Cycle to the next work project (skips rest) |
| `track rest` | Switch to rest immediately |
| `track switch <name>` | Switch to a specific project |
| `track status` | Current project + elapsed time |
| `track projects` | List all registered projects |
| `track register <name>` | Register a new project |
| `track unregister <name>` | Remove a project |
 
---
 
## Notes
 
- `rest` is a built-in project — always available, cannot be unregistered
- Session state lives **in memory only** and resets on `track end` — persistence is planned for v0.3
- If the daemon crashes or you need to force-stop it: `pkill trackd`
- The `track next` cycle order is alphabetical and excludes `rest`
 
---
 
## Roadmap
 
| Milestone | Highlights |
|---|---|
| **v0.2** | Linux support (amd64 + arm64), CI build matrix |
| **v0.3** | JSON session persistence, crash recovery, `track edit` |
| **v0.4** | `track report [date]`, weekly summary, CSV export |
| **v1.0** | Idle detection, shell prompt integration, Homebrew tap |
 
---
 
## Contributing
 
See [CONTRIBUTING.md](CONTRIBUTING.md). All contributions welcome.
 
## License
 
MIT — see [LICENSE](LICENSE).
