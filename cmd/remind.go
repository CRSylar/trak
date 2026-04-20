package command

import (
	"slices"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CRSylar/trak/internal/config"
	"github.com/CRSylar/trak/internal/notify"
	"github.com/CRSylar/trak/internal/remind"
	"github.com/CRSylar/trak/internal/session"
)

func Remind() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: trak remind <subcommand>\n")
		fmt.Fprintf(os.Stderr, "subcommands:\n")
		fmt.Fprintf(os.Stderr, "  start      - remind to start workday if daemon not running\n")
		fmt.Fprintf(os.Stderr, "  stop       - remind to stop workday if daemon is running\n")
		fmt.Fprintf(os.Stderr, "  custom <msg> - send a custom notification\n")
		fmt.Fprintf(os.Stderr, "  test       - send a test notification\n")
		os.Exit(1)
	}

	sub := os.Args[2]
	switch sub {
	case "start":
		// check if the daily session file exists, if so is a noop
		// else send reming notification
		now := time.Now()
		if _, err := os.Stat(session.FilePath(now.Format("2006-01-02"))); err != nil {
			if os.IsNotExist(err) {
				// file not exists, so day is not been started yet, send notification
				if err := notify.NotifyIfNotStarted(); err != nil {
					fmt.Fprintf(os.Stderr, "failed to send notification: %v\n", err)
					os.Exit(1)
				}
				fmt.Fprint(os.Stdout, "Sent reminder to start workday")
				return
			}
			fmt.Fprintf(os.Stderr, "cannot stat the session file: %v", err)
		}
		fmt.Fprint(os.Stdout, "Daily session already started, nothing to remind")
	case "stop":
		now := time.Now()
		sess, err := session.LoadSession(now.Format("2006-01-02"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load today's session: %v", err)
			os.Exit(1)
		}

		if !sess.Closed {
			if err := notify.NotifyIfNotStopped(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to send notification: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(os.Stdout, "Sent reminder to stop workday")
		}

		fmt.Fprint(os.Stdout, "Daily session already closed, nothing to do")
	case "custom":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: trak remind custom <message>\n")
			os.Exit(1)
		}
		msg := strings.Join(os.Args[3:], " ")
		if err := notify.NotifyCustom(msg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send notification: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Sent custom notification")

	case "test":
		if err := notify.Test(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send test notification: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Sent test notification")

	default:
		fmt.Fprintf(os.Stderr, "trak remind: unknown subcommand %q\n", sub)
		os.Exit(1)
	}
}

func InstallReminders() {
	schedule, err := remind.LoadSchedule(config.GetConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid reminder times: %v\n", err)
		fmt.Fprintf(os.Stderr, "Set reminder_start_time and reminder_end_time in ~/.trak/config.json\n")
		fmt.Fprintf(os.Stderr, "Format: HH:MM (24-hour)\n")
		os.Exit(1)
	}

	if schedule == nil {
		fmt.Fprintf(os.Stderr, "Reminder times not configured.\n")
		fmt.Fprintf(os.Stderr, "Add reminder_start_time and reminder_end_time to ~/.trak/config.json\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, `  "reminder_start_time": "09:30",`+"\n")
		fmt.Fprintf(os.Stderr, `  "reminder_end_time": "18:00"`+"\n")
		os.Exit(1)
	}

	platform := remind.DetectPlatform()
	installer := remind.GetInstaller(platform)

	if !installer.IsSupported() {
		fmt.Print(
			schedule.InstallInstructions(platform),
		)
		os.Exit(0)
	}

	startCron, stopCron := schedule.CronLines()

	startExists, stopExists, err := installer.CheckExisting()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check existing reminders: %v\n", err)
		fmt.Fprintf(os.Stderr, "Printing manual instructions instead...\n")
		fmt.Print(
			schedule.InstallInstructions(platform),
		)
		os.Exit(0)
	}

	if startExists || stopExists {
		fmt.Println("Found existing trak reminder entries:")
		if startExists {
			fmt.Println("  - Start reminder already exists")
		}
		if stopExists {
			fmt.Println("  - Stop reminder already exists")
		}
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  1) Replace existing entries with new schedule")
		fmt.Println("  2) Keep existing entries (skip installation)")
		fmt.Println("  3) Abort")
		fmt.Print("\nChoose [1-3]: ")

		choice := readChoice([]string{"1", "2", "3"})
		switch choice {
		case "1":
		case "2":
			fmt.Println("Skipping installation.")
			os.Exit(0)
		case "3":
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	startCron, stopCron = schedule.CronLines()

	fmt.Println("Will install reminders:")
	fmt.Printf("  %s\n", startCron)
	fmt.Printf("  %s\n", stopCron)
	fmt.Print("\nProceed? [y/N]: ")

	choice := readChoice([]string{"y", "Y", "n", "N"})
	if choice != "y" && choice != "Y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	if err := installer.Install(startCron, stopCron); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install reminders: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Reminders installed successfully!")
}

func UninstallReminders() {
	platform := remind.DetectPlatform()
	installer := remind.GetInstaller(platform)

	if !installer.IsSupported() {
		fmt.Fprintf(os.Stderr, "%s is not supported on this platform.\n", installer.Name())
		os.Exit(1)
	}

	startExists, stopExists, err := installer.CheckExisting()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check existing reminders: %v\n", err)
		os.Exit(1)
	}

	if !startExists && !stopExists {
		fmt.Println("No trak reminder entries found.")
		os.Exit(0)
	}

	fmt.Println("Found trak reminder entries:")
	if startExists {
		fmt.Println("  - Start reminder")
	}
	if stopExists {
		fmt.Println("  - Stop reminder")
	}
	fmt.Print("\nRemove them? [y/N]: ")

	choice := readChoice([]string{"y", "Y", "n", "N"})
	if choice != "y" && choice != "Y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	if err := installer.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to uninstall reminders: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Reminders uninstalled successfully!")
}

func readChoice(valid []string) string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())
		if slices.Contains(valid, choice) {
				return choice
			}
	}
	return ""
}
