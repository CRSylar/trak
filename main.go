package main

import (
	"fmt"
	"os"

	command "github.com/CRSylar/trak/cmd"
	"github.com/CRSylar/trak/internal/projects"
)

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
		project := "rest"
		if len(os.Args) == 3 {
			project = os.Args[2]
		}
		command.StartWorkday(project)

	case "end":
		command.EndWorkday()
	case "next":
		command.CicleNextProject()
	case "rest":
		command.SwitchProject(projects.RestProject)
	case "switch":
		requireArg("switch", "<project-name>")
		command.SwitchProject(os.Args[2])
	case "edit":
		command.EditLastSwitch()
	case "status":
		command.GetStatus()
	case "projects":
		command.ListProjects()
	case "register":
		requireArg("register", "<project-name>")
		command.Register(os.Args[2])
	case "unregister":
		requireArg("unregister", "<project-name>")
		command.Unregister(os.Args[2])
	case "remind":
		//TODO:
	case "install-reminders":
		//TODO:
	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "trak: unknown command %q\n\n", cmd)
		printUsage()
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

Additional commands:")
	  remind <subcommand>")
	    start                Run a reminder check")
	    stop                 Stop reminder notifications")
	    custom <message>     Add a custom reminder")
	    test                 Send a test reminder")
	  install-reminders      Install reminder integration")

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
