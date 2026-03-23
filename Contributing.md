# Contributing to trak
 
Thanks for your interest in contributing! trak is a small, focused tool and we want to keep it that way — contributions that add complexity without clear user value are unlikely to be merged. When in doubt, open an issue first.
 
---
 
## Getting started
 
**Requirements**
- Go 1.26+
- macOS or Linux
- `make`
 
```bash
git clone https://github.com/CRSylar/trak
cd trak
make dev       # builds native binaries into bin/ without cross-compiling
```
 
The two binaries are `bin/trak` (CLI) and `bin/trakd` (daemon). You can test them directly from `bin/` without installing.
 
**Project layout**
 
```
trak/
├── cmd/
│   ├── trak/      CLI entrypoint
│   └── trakd/     Daemon entrypoint
├── internal/
│   ├── daemon/     State, timer logic, unix socket server
│   ├── client/     Socket client used by the CLI
│   └── protocol/   Shared JSON message types
├── scripts/
│   └── raycast/    Raycast script commands (macOS)
└── Makefile
```
 
---
 
## Platform support
 
| Platform | Status | Notes |
|---|---|---|
| macOS arm64 (Apple Silicon) | ✅ Supported | Primary target |
| macOS amd64 (Intel) | ✅ Should work | Untested |
| Linux arm64 | 🔜 Planned v0.2 | |
| Linux amd64 | 🔜 Planned v0.2 | |
| Windows | ❌ Not planned | Unix sockets only |
 
For Linux contributions, the only platform-specific piece is the hotkey integration — Raycast is macOS-only. The equivalent on Linux is **rofi** (X11) or **wofi** (Wayland). A `scripts/rofi/` directory following the same pattern as `scripts/raycast/` is the right approach.
 
---
 
## Making changes
 
- **Keep the binary small and dependency-free.** The Go standard library covers everything we need. Do not add third-party dependencies without a strong reason.
- **No breaking changes to the CLI interface** without a major version bump — people script against `trak switch`, `trak next`, etc.
- **The protocol between `trak` and `trakd`** lives in `internal/protocol/`. If you add a command, add it there first, then wire it through the server dispatcher and the CLI.
- **All time math belongs in `internal/daemon/state.go`**, not scattered across the CLI.
 
---
 
## Submitting a pull request
 
1. Fork the repo and create a branch: `git checkout -b my-feature`
2. Make your changes
3. Run `make build` and verify both binaries compile cleanly for all targets
4. Update `README.md` if you've changed or added any commands
5. Open a PR with a clear description of what and why
 
There are no automated tests yet (planned for v0.3). Manual testing steps in your PR description are appreciated.
 
---
 
## Reporting bugs
 
Open a GitHub issue with:
- Your OS and architecture
- The `trak` version (`trak --version`, once that's implemented)
- The exact command you ran and the output you got
- Whether `trakd` was running (`pgrep trakd`)
 
---
 
## Roadmap and priorities
 
See the roadmap in [README.md](README.md). If you want to work on a roadmap item, comment on the relevant issue (or open one) so we can coordinate before you invest time in it.
 
