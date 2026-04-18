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
		fmt.Fprint(os.Stderr, fmt.Errorf("failed to star workday: %w", err))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Workday started at %s. Active project: %s", today.Format("15:04"), project)
}

func EndWorkday() {
	session.Close()
}

func GetStatus() {
	now := time.Now()
	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, "failed to load session: %w", err)
		return
	}

	if sess.Closed {
		fmt.Fprint(os.Stdout, "today session closed")
		return
	}

	fmt.Fprintf(os.Stdout, "Active: %s (%s on this task) | Day: %s since %s",
		sess.ActiveProject,
		formatDuration(time.Since(sess.Segments[len(sess.Segments)-1].Start)),
		formatDuration(time.Since(sess.DayStart)),
		sess.DayStart.Format("15:04"),
	)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh %02dm", h, m)
}
