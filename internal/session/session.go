package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Segment is a single contiguous block of work on one project
type Segment struct {
	Project string    `json:"project"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

// Session represents a full workday session file
type Session struct {
	Date          string    `json:"date"`   // YYYY-MM-DD
	Closed        bool      `json:"closed"` // true = day ended, immutable
	DayStart      time.Time `json:"day_start"`
	ActiveProject string    `json:"active_project"` // only meaningful when !Closed
	SegmentStart  time.Time `json:"segment_start"`  // start of the current open segment
	Segments      []Segment `json:"segments"`       // completed segments only
}

// New creates a fresh session for today
func New(activeProject string) *Session {
	now := time.Now()
	return &Session{
		Date:          now.Format("2006-01-02"),
		Closed:        false,
		DayStart:      now,
		ActiveProject: activeProject,
		SegmentStart:  now,
		Segments:      []Segment{},
	}
}

// FilePath returns the path for a session file given a sessions dir and date string
func FilePath(sessionsDir, date string) string {
	return filepath.Join(sessionsDir, date+".json")
}

// ExtraFilePath returns an alternative path for the same day (crash recovery new session)
func ExtraFilePath(sessionsDir, date string) string {
	return filepath.Join(sessionsDir, date+"-extra.json")
}

// Load reads a session from disk
func Load(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupt session file %s: %w", path, err)
	}
	return &s, nil
}

// Save writes the session to disk atomically (write to temp, rename)
func Save(s *Session, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temp file then rename for atomicity
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

	if err := tmpFile.Sync(); err != nil {
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

// FindActiveToday looks for an unclosed session file for today in sessionsDir.
// Returns the session and its path, or nil if none found.
func FindActiveToday(sessionsDir string) (*Session, string, error) {
	today := time.Now().Format("2006-01-02")
	path := FilePath(sessionsDir, today)

	s, err := Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}

	if s.Closed {

		// if the <day>.json file is closed, do a lookup for <date>-extra.json
		// this is needed to handle cases where the daemon crashed (or similar)
		// after an extra session was created
		extraPath := ExtraFilePath(sessionsDir, today)
		s, err = Load(extraPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, "", nil
			}
			return nil, "", err
		}
		if !s.Closed {
			return s, extraPath, nil
		}

		return nil, "", nil
	}

	return s, path, nil
}

// Close marks the session as closed and saves it.
func Close(s *Session, now time.Time, path string) error {
	s.Closed = true

	return Save(s, path)
}
