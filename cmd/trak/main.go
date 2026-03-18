package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/CRSylar/trak/internal/client"
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
		fmt.Printf("track %s\n", version)

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

	case "status":
		msg, err := client.Send(protocol.CmdStatus, "")
		dieOnErr(err)
		fmt.Println(msg)

	case "projects":
		// Optional --names flag for machine-readable output (used by Raycast)
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

// startWorkday launches the daemon then sends the start command
func startWorkday() {
	// Find trakd next to the trak binary
	trakdPath, err := findDaemonBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "trak: cannot find trakd binary: %v\n", err)
		os.Exit(1)
	}

	// Launch daemon in background
	daemonCmd := exec.Command(trakdPath)
	daemonCmd.Stdout = nil
	daemonCmd.Stderr = nil
	if err := daemonCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "trak: failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	// Give it a moment to bind the socket
	// We retry the send a few times to handle the startup delay
	var msg string
	var sendErr error
	for range 10 {
		msg, sendErr = client.Send(protocol.CmdStart, "")
		if sendErr == nil {
			break
		}
		// Small sleep via a busy wait 
		time.Sleep(time.Millisecond * 100)
	}
	dieOnErr(sendErr)
	fmt.Println(msg)
}

func findDaemonBinary() (string, error) {
	// 1. Look next to the running trak binary
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "trakd")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 2. Fall back to PATH
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
	fmt.Printf(`trak — WorkDay time tracker (%s)

USAGE:
  trak start                  Start the workday (launches background daemon)
  trak end                    End the workday, print report, stop daemon
  trak next                   Cycle to the next work project (skips rest)
  trak rest                   Switch to rest immediately
  trak switch <project>       Switch to a specific project by name
  trak status                 Show current project and elapsed time
  trak projects               List registered projects
  trak projects --names       List project names (machine-readable, for scripts)
  trak register <project>     Register a new project
  trak unregister <project>   Remove a project

NOTES:
  'rest' is a built-in project always available for breaks.
  Project registrations are saved to ~/.trak/projects.json.
  Timer state lives only in memory and resets at 'trak end'.
`, version)
}
