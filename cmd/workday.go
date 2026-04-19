package command

import (
	"fmt"
	"os"
	"time"

	"github.com/CRSylar/trak/internal/session"
)

func StartWorkday(project string) {
	today := time.Now()

	sess := session.New(project)
	if err := session.Save(*sess, session.FilePath(today.Format("2006-01-02"))); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("failed to star workday: %w\n", err))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Workday started at %s. Active project: %s\n", today.Format("15:04"), project)
}

func EndWorkday() {
	session.Close()
}

func GetStatus() {
	now := time.Now()
	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, "failed to load session: %w\n", err)
		return
	}

	if sess.Closed {
		fmt.Fprint(os.Stdout, "today session closed\n")
		return
	}

	fmt.Fprintf(os.Stdout, "Active: %s (%s on this task) | Day: %s since %s\n",
		sess.ActiveProject,
		session.FormatDuration(time.Since(sess.Segments[len(sess.Segments)-1].Start)),
		session.FormatDuration(time.Since(sess.DayStart)),
		sess.DayStart.Format("15:04"),
	)
}


