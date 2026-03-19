# trak
 
A minimal CLI time traker for freelancers. Register projects, switch between them instantly via a hotkey, and get a clean end-of-day report.
 
Built with Go. No cloud, no accounts, no bloat — just a fast local daemon and two keystrokes.
 
---
 
## How it works
 
```
CLI (trak) ──── unix socket ────► Daemon (trakd)
                                     holds session state in memory
```
 
`trakd` runs in the background for the duration of your workday. `trak` is the CLI you interact with — or don't, once your hotkeys are set up. Project registrations are persisted to `~/.trak/projects.json`.
 
---
 
## Install
 
**Requirements:** Go 1.22+, macOS (Apple Silicon)
 
> Linux support is planned for v0.2. See [CONTRIBUTING.md](CONTRIBUTING.md).
 
```bash
git clone https://github.com/you/trak
cd trak
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
trak start
 
# First time only — register your projects (saved permanently)
trak register client-alpha
trak register client-beta
trak register internal
 
# Switch projects
trak next          # cycles: client-alpha → client-beta → internal → client-alpha
trak rest          # jump to rest immediately
trak switch <name> # go to a specific project by name
 
# Check where you are
trak status
 
# Evening — end the workday and print the report
trak end
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
| `trak start` | Start workday, launch daemon |
| `trak end` | End workday, print report, stop daemon |
| `trak next` | Cycle to the next work project (skips rest) |
| `trak rest` | Switch to rest immediately |
| `trak switch <name>` | Switch to a specific project |
| `trak status` | Current project + elapsed time |
| `trak projects` | List all registered projects |
| `trak register <name>` | Register a new project |
| `trak unregister <name>` | Remove a project |
 
---
 
## Notes
 
- `rest` is a built-in project — always available, cannot be unregistered
- Session state lives **in memory only** and resets on `trak end` — persistence is planned for v0.3
- If the daemon crashes or you need to force-stop it: `pkill trakd`
- The `trak next` cycle order is alphabetical and excludes `rest`
 
---
 
## Roadmap
 
| Milestone | Highlights |
|---|---|
| **v0.3** | JSON session persistence, crash recovery, `trak edit` |
| **v0.4** | `trak report [date]`, weekly summary, CSV export |
| **v1.0** | Idle detection, shell prompt integration, Homebrew tap |
 
---
 
## Contributing
 
See [CONTRIBUTING.md](CONTRIBUTING.md). All contributions welcome.
 
## License
 
MIT — see [LICENSE](LICENSE).
