package command

import (
	"fmt"
	"os"
	"time"

	"github.com/CRSylar/trak/internal/session"
)

func EditLastSwitch() {
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

	now := time.Now()
	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot load session file: %w\n", err))
		return
	}

	segLen := len(sess.Segments)
	if segLen <= 1 {
		fmt.Fprint(os.Stderr, "no completed segments to edit yet - switch projects at least one time first\n")
		return
	}

	currStart := sess.Segments[segLen-1].Start
	lastStart := sess.Segments[segLen-2].Start
	maxShift := currStart.Sub(lastStart)
	if total > maxShift {
		total = maxShift
	}

	if total < 0 {
		total = 0
	}

	newBoundary := currStart.Add(-total)
	sess.Segments[segLen-2].End = newBoundary
	sess.Segments[segLen-1].Start = newBoundary

	if err := session.Save(*sess, session.FilePath(now.Format("2006-01-02"))); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("failed to save edited session: %w\n", err))
		return
	}

	fmt.Fprintf(os.Stdout, "Shifted last switch back by %s → boundary now at %s\n", session.FormatDuration(total), newBoundary.Format("15:04"))
}
