package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/CRSylar/trak/internal/client"
	"github.com/CRSylar/trak/internal/daemon"
	"github.com/CRSylar/trak/internal/protocol"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0"
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("trak %s\n", version)

	case "start":
		startWorkday()

	case "end":
		msg, err := client.Send(protocol.CmdEnd, "")
		dieOnErr(err)
		fmt.Println(msg)

	case "next":
		msg, err := client.Send(protocol.CmdNext, "")
		dieOnErr(err)
		fmt.Println(msg)

	case "rest":
		msg, err := client.Send(protocol.CmdRest, "")
		dieOnErr(err)
		fmt.Println(msg)

	case "switch":
		requireArg("switch", "<project-name>")
		msg, err := client.Send(protocol.CmdSwitch, os.Args[2])
		dieOnErr(err)
		fmt.Println(msg)

	case "edit":
		editLastSwitch()

	case "status":
		msg, err := client.Send(protocol.CmdStatus, "")
		dieOnErr(err)
		fmt.Println(msg)

	case "projects":
		payload := ""
		if len(os.Args) > 2 && os.Args[2] == "--names" {
			payload = "names"
		}
		msg, err := client.Send(protocol.CmdProjects, payload)
		dieOnErr(err)
		fmt.Println(msg)

	case "register":
		requireArg("register", "<project-name>")
		msg, err := client.Send(protocol.CmdRegister, os.Args[2])
		dieOnErr(err)
		fmt.Println(msg)

	case "unregister":
		requireArg("unregister", "<project-name>")
		msg, err := client.Send(protocol.CmdUnregister, os.Args[2])
		dieOnErr(err)
		fmt.Println(msg)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "trak: unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// ---------- start with resume flow ----------

func startWorkday() {
	daemonPath, err := findDaemonBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "trak: cannot find trakd binary: %v\n", err)
		os.Exit(1)
	}

	daemonCmd := exec.Command(daemonPath)
	daemonCmd.Stdout = os.Stdout
	daemonCmd.Stderr = os.Stderr
	if err := daemonCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "trak: failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	// Wait until the daemon socket is ready
	for range 9 {
		_, sendErr := client.Send(protocol.CmdProjects, "")
		if sendErr == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	_, sendErr := client.Send(protocol.CmdProjects, "")
	if sendErr != nil {
		fmt.Fprintf(os.Stderr, "trak: failed to connect to daemon, use `pkill trakd` and retry ")
		os.Exit(1)
	}

	// Check for an unfinished session from today
	raw, err := client.Send(protocol.CmdCheckResume, "")
	dieOnErr(err)

	if raw == "" {
		// No unfinished session — start fresh
		msg, err := client.Send(protocol.CmdStart, "")
		dieOnErr(err)
		fmt.Println(msg)
		return
	}

	// Parse the resume candidate
	var candidate daemon.ResumeCandidate
	if err := json.Unmarshal([]byte(raw), &candidate); err != nil {
		dieOnErr(fmt.Errorf("failed to parse resume candidate: %w", err))
	}

	fmt.Printf("⚠️  Unfinished session found — last active: %s (at %s)\n",
		candidate.ActiveProject, candidate.SessAt)
	fmt.Print("Resume it? [y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "trak: failed to read input: %v\n", err)
		os.Exit(1)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		msg, err := client.Send(protocol.CmdResume, candidate.SessPath)
		dieOnErr(err)
		fmt.Println(msg)
	} else {
		msg, err := client.Send(protocol.CmdDiscardAndStart, candidate.SessPath)
		dieOnErr(err)
		fmt.Println(msg)
	}
}

// ---------- trak edit ----------

func editLastSwitch() {
	args := os.Args[2:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: trak edit --last <duration> [--last <duration> ...]\n")
		fmt.Fprintf(os.Stderr, "       duration examples: 15m  1h  1h15m\n")
		os.Exit(1)
	}

	var total time.Duration
	argLen := len(args)
	for i := 0; i < argLen; i++ {
		if args[i] == "--last" {
			if i+1 >= argLen {
				fmt.Fprintf(os.Stderr, "trak edit: --last requires a value\n")
				os.Exit(1)
			}
			i++
			d, err := time.ParseDuration(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "trak edit: invalid duration %q: %v\n", args[i], err)
				os.Exit(1)
			}
			total += d
		} else {
			fmt.Fprintf(os.Stderr, "trak edit: unknown flag %q\n", args[i])
			os.Exit(1)
		}
	}

	if total <= 0 {
		fmt.Fprintf(os.Stderr, "trak edit: total duration must be positive\n")
		os.Exit(1)
	}

	msg, err := client.Send(protocol.CmdEdit, total.String())
	dieOnErr(err)
	fmt.Println(msg)
}

// ---------- helpers ----------

func findDaemonBinary() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "trakd")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	path, err := exec.LookPath("trakd")
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("trakd not found next to trak binary or in PATH (OS: %s)", runtime.GOOS)
}

func dieOnErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func requireArg(cmd, argName string) {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "trak %s requires an argument: %s\n", cmd, argName)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`trak — time tracker for freelancers (%s)

USAGE:
  trak start                         Start the workday (launches trakd daemon)
  trak end                           End the workday, print report, stop daemon
  trak next                          Cycle to the next work project (skips rest)
  trak rest                          Switch to rest immediately
  trak switch <project>              Switch to a specific project by name
  trak edit --last <duration>        Shift the last switch back by duration
  trak status                        Show current project and elapsed time
  trak projects                      List registered projects
  trak projects --names              List project names (machine-readable)
  trak register <project>            Register a new project
  trak unregister <project>          Remove a project
  trak version                       Print version

DURATION FORMAT (for trak edit):
  15m                    15 minutes
  1h                     1 hour
  1h15m                  1 hour and 15 minutes
  --last 1h --last 15m   chained flags (summed to 1h15m)

NOTES:
  'rest' is a built-in project always available for breaks.
  Project registrations are saved to ~/.trak/projects.json.
  Session data is saved to sessions_dir (see ~/.trak/config.json) after every switch.
  If trakd crashes, session data is recovered on next 'trak start'.
`, version)
}
