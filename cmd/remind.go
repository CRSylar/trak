package command

import (
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

	fmt.Print(
		schedule.InstallInstructions(
			remind.DetectPlatform(),
		),
	)
}
