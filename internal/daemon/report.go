package daemon

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/CRSylar/trak/internal/session"
)

func buildReport(dayStart, dayEnd time.Time, segments []session.Segment) string {
	// Accumulate time per project
	totals := make(map[string]time.Duration)
	for _, seg := range segments {
		totals[seg.Project] += seg.End.Sub(seg.Start)
	}

	totalDay := dayEnd.Sub(dayStart)

	// Sort project names
	var projects []string
	for p := range totals {
		projects = append(projects, p)
	}
	sort.Strings(projects)

	// Calculate column widths
	maxName := len("Project")
	for _, p := range projects {
		if len(p) > maxName {
			maxName = len(p)
		}
	}

	timeCol := len("Time")
	pctCol := len("%")

	// Build rows
	type row struct {
		name string
		dur  string
		pct  string
	}
	var rows []row
	for _, p := range projects {
		d := totals[p]
		pct := 0.0
		if totalDay > 0 {
			pct = float64(d) / float64(totalDay) * 100
		}
		rows = append(rows, row{
			name: p,
			dur:  formatDuration(d),
			pct:  fmt.Sprintf("%d%%", int(pct)),
		})
	}

	// Measure actual column widths from data
	for _, r := range rows {
		if len(r.dur) > timeCol {
			timeCol = len(r.dur)
		}
		if len(r.pct) > pctCol {
			pctCol = len(r.pct)
		}
	}

	// Separators
	sep := fmt.Sprintf("┼%s┼%s┼%s┤",
		strings.Repeat("─", maxName+2),
		strings.Repeat("─", timeCol+2),
		strings.Repeat("─", pctCol+2),
	)
	top := "┌" + sep[1:len(sep)-1] + "┐"
	top = fmt.Sprintf("┌%s┬%s┬%s┐",
		strings.Repeat("─", maxName+2),
		strings.Repeat("─", timeCol+2),
		strings.Repeat("─", pctCol+2),
	)
	mid := fmt.Sprintf("├%s┼%s┼%s┤",
		strings.Repeat("─", maxName+2),
		strings.Repeat("─", timeCol+2),
		strings.Repeat("─", pctCol+2),
	)
	bot := fmt.Sprintf("└%s┴%s┴%s┘",
		strings.Repeat("─", maxName+2),
		strings.Repeat("─", timeCol+2),
		strings.Repeat("─", pctCol+2),
	)

	fmtRow := func(name, dur, pct string) string {
		return fmt.Sprintf("│ %-*s │ %-*s │ %*s │",
			maxName, name,
			timeCol, dur,
			pctCol, pct,
		)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "\nWorkDay ended — %s\n", dayEnd.Format("02 Jan 2006"))
	fmt.Fprintf(&sb, "Total: %s  (%s → %s)\n\n",
		formatDuration(totalDay),
		dayStart.Format("15:04"),
		dayEnd.Format("15:04"))
	sb.WriteString(top + "\n")
	sb.WriteString(fmtRow("Project", "Time", "%") + "\n")
	sb.WriteString(mid + "\n")
	for _, r := range rows {
		sb.WriteString(fmtRow(r.name, r.dur, r.pct) + "\n")
	}
	sb.WriteString(bot + "\n")

	return sb.String()
}
