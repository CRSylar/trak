package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CRSylar/trak/internal/config"
)

type Segment struct {
	Project string    `json:"project"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

type Session struct {
	Date          string    `json:"date"`
	Closed        bool      `json:"closed"`
	ActiveProject string    `json:"active_project"`
	DayStart      time.Time `json:"day_start"`
	Segments      []Segment `json:"segments"`
}

func New(activeProject string) *Session {
	now := time.Now()
	return &Session{
		Date:          now.Format("2006-01-02"),
		Closed:        false,
		DayStart:      now,
		ActiveProject: activeProject,
		Segments: []Segment{
			{Project: activeProject, Start: now},
		},
	}
}

func Save(s Session, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	tmpFile, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmp := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		_ = os.Remove(tmp)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}

func Close() error {
	today := time.Now().Format("2006-01-02")

	sess, err := LoadSession(today)
	if err != nil {
		return err
	}

	sess.Closed = true
	sess.ActiveProject = ""
	sess.Segments[len(sess.Segments)-1].End = time.Now()

	if err := Save(*sess, FilePath(today)); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s", buildReport(sess.DayStart, time.Now(), sess.Segments))
	return nil
}

func LoadSession(day string) (*Session, error) {
	path := FilePath(day)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupt or invalid session file %s: %w", path, err)
	}

	return &s, nil
}

func FilePath(today string) string {
	return filepath.Join(config.GetSessionsDir(), today+".json")
}

func buildReport(dayStart, dayEnd time.Time, segments []Segment) string {
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
			dur:  FormatDuration(d),
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
		FormatDuration(totalDay),
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

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh %02dm", h, m)
}
