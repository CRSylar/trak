package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	sess.Segments[len(sess.Segments) -1].End = time.Now()

	return Save(*sess, FilePath(today))
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
