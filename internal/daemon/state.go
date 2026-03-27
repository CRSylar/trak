package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/CRSylar/trak/internal/config"
	"github.com/CRSylar/trak/internal/session"
)

const restProject = "rest"
const projectsFileName = "projects.json"

// State holds all runtime state for a workday
type State struct {
	mu                 sync.Mutex
	running            bool
	cfg                *config.Config
	sess               *session.Session
	sessPath           string
	registeredProjects map[string]bool
	projectsPath       string
	cycleIndex         int
}

func New() (*State, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	s := &State{
		cfg:                cfg,
		registeredProjects: make(map[string]bool),
		projectsPath:       filepath.Join(home, ".trak", projectsFileName),
	}
	s.registeredProjects[restProject] = true
	err = s.loadProjects()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *State) validSessPath(sessPath string) error {

	if !strings.HasPrefix(filepath.Clean(sessPath), filepath.Clean(s.cfg.SessionsDir)+string(os.PathSeparator)) {
		return fmt.Errorf("unsafe path: %s", sessPath)
	}
	return nil
}

// ---------- project persistence ----------

type projectConfig struct {
	Projects []string `json:"projects"`
}

func (s *State) loadProjects() error {
	data, err := os.ReadFile(s.projectsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	for _, p := range cfg.Projects {
		s.registeredProjects[p] = true
	}
	return nil
}

func (s *State) saveProjects() error {
	var names []string
	for name := range s.registeredProjects {
		if name == restProject {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	cfg := projectConfig{Projects: names}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.projectsPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.projectsPath, data, 0644)
}

// ---------- session checkpoint ----------

// checkpoint flushes current in-memory state to disk — caller must hold the lock
func (s *State) checkpoint() error {
	if !s.running || s.sess == nil {
		return nil
	}
	return session.Save(s.sess, s.sessPath)
}

// ---------- commands ----------

// ResumeCandidate is returned by CheckResume when an unfinished session exists
type ResumeCandidate struct {
	SessPath      string
	ActiveProject string
	SessAt        string // formatted last segment start
}

// CheckResume looks for an unclosed session for today. Returns nil if none found.
func (s *State) CheckResume() (*ResumeCandidate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil, fmt.Errorf("workday already started (active: %s)", s.sess.ActiveProject)
	}

	existing, path, err := session.FindActiveToday(s.cfg.SessionsDir)
	if err != nil {
		return nil, fmt.Errorf("error scanning sessions: %w", err)
	}
	if existing == nil {
		return nil, nil
	}

	return &ResumeCandidate{
		SessPath:      path,
		ActiveProject: existing.ActiveProject,
		SessAt:        existing.SegmentStart.Format("15:04"),
	}, nil
}

// Resume loads an existing session back into memory
func (s *State) Resume(sessPath string) (string, error) {

	if err := s.validSessPath(sessPath); err != nil {
		return "", fmt.Errorf("invalid sessionsPath %q: %w", sessPath, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := session.Load(sessPath)
	if err != nil {
		return "", err
	}

	if err := session.ValidateSession(*existing); err != nil {
		return "", fmt.Errorf("cannot resume session at %q, file contains invalid session data; errror: %w", sessPath, err)
	}

	s.running = true
	s.sess = existing
	s.sessPath = sessPath

	return fmt.Sprintf("Resumed session — active: %s (since %s)",
		existing.ActiveProject, existing.SegmentStart.Format("15:04")), nil
}

// DiscardAndStart closes the old session and starts a fresh one for today
func (s *State) DiscardAndStart(oldPath string) (string, error) {

	if err := s.validSessPath(oldPath); err != nil {
		return "", fmt.Errorf("invalid session path: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	old, err := session.Load(oldPath)
	if err != nil {
		return "", err
	}

	if err := session.ValidateSession(*old); err == nil {
		return "", fmt.Errorf("old session is not valid, discardAndRestart operation will be aborted; error: %w", err)
	}

	now := time.Now()
	activeBackup := old.ActiveProject
	segStartBackup := old.SegmentStart
	old.Segments = append(old.Segments, session.Segment{
		Project: old.ActiveProject,
		Start:   old.SegmentStart,
		End:     now,
	})

	old.ActiveProject = ""
	old.SegmentStart = time.Time{}

	if err := session.Close(old, oldPath); err != nil {
		old.ActiveProject = activeBackup
		old.SegmentStart = segStartBackup
		old.Segments = old.Segments[:len(old.Segments)-1]
		return "", fmt.Errorf("failed to close old session, that is loaded in memory, please try to close it manually with `trak end`; error: %w", err)
	}

	return s.doStart(true)
}

// Start begins a new workday — only called when CheckResume found nothing
func (s *State) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return "", fmt.Errorf("workday already started (active: %s)", s.sess.ActiveProject)
	}
	return s.doStart(false)
}

// doStart initialises a fresh session — caller must hold the lock
func (s *State) doStart(extra bool) (string, error) {
	today := time.Now().Format("2006-01-02")
	var path string
	if extra {
		path = session.ExtraFilePath(s.cfg.SessionsDir, today)
	} else {
		path = session.FilePath(s.cfg.SessionsDir, today)
	}

	sess := session.New(restProject)
	if err := session.Save(sess, path); err != nil {
		return "", fmt.Errorf("failed to create session file: %w", err)
	}

	s.running = true
	s.sess = sess
	s.sessPath = path

	return fmt.Sprintf("Workday started at %s. Active project: %s",
		sess.DayStart.Format("15:04"), restProject), nil
}

func (s *State) End() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress")
	}

	now := time.Now()
	activeBackup := s.sess.ActiveProject
	segStartBackup := s.sess.SegmentStart
	s.sess.Segments = append(s.sess.Segments, session.Segment{
		Project: s.sess.ActiveProject,
		Start:   s.sess.SegmentStart,
		End:     now,
	})

	s.sess.ActiveProject = ""
	s.sess.SegmentStart = time.Time{}

	err := session.Close(s.sess, s.sessPath)
	if err != nil {
		s.sess.ActiveProject = activeBackup
		s.sess.SegmentStart = segStartBackup
		s.sess.Segments = s.sess.Segments[:len(s.sess.Segments)-1]
		return "", fmt.Errorf("error saving sessions to file, last segment is still in daemon state and is not been written, please retry; error: %w", err)
	}

	report := buildReport(s.sess.DayStart, now, s.sess.Segments)

	s.running = false
	s.sess = nil

	return report, nil
}

func (s *State) Switch(project string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}
	if !s.registeredProjects[project] {
		return "", fmt.Errorf("unknown project %q — register it first with 'trak register %s'", project, project)
	}
	if s.sess.ActiveProject == project {
		return fmt.Sprintf("already on %s", project), nil
	}
	return s.doSwitch(project)
}

func (s *State) Next() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}

	projects := s.workProjects()
	if len(projects) == 0 {
		return "", fmt.Errorf("no work projects registered — add one with 'trak register <project>'")
	}

	if s.sess.ActiveProject != restProject {
		for i, p := range projects {
			if p == s.sess.ActiveProject {
				s.cycleIndex = (i + 1) % len(projects)
				break
			}
		}
	}
	if s.cycleIndex >= len(projects) {
		s.cycleIndex = 0
	}

	return s.doSwitch(projects[s.cycleIndex])
}

func (s *State) Rest() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}
	if s.sess.ActiveProject == restProject {
		return "already on rest", nil
	}
	return s.doSwitch(restProject)
}

func (s *State) Edit(shift time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}

	segs := s.sess.Segments
	if len(segs) == 0 {
		return "", fmt.Errorf("no completed segments to edit yet — switch projects at least once first")
	}

	// Shift the boundary between the last completed segment and the current open one.
	// Cap the shift so it never exceeds the last segment's own duration.
	currentStart := s.sess.SegmentStart
	lastStart := segs[len(segs)-1].Start
	maxShift := currentStart.Sub(lastStart)
	if shift > maxShift {
		shift = maxShift
	}

	// safety check for negative value.
	// since we'll use the minus sign later
	// a negative value here will actually move timers in the future
	if shift < 0 {
		shift = 0
	}

	newBoundary := currentStart.Add(-shift)
	s.sess.Segments[len(segs)-1].End = newBoundary
	s.sess.SegmentStart = newBoundary

	if err := s.checkpoint(); err != nil {
		return "", fmt.Errorf("edit applied in memory but failed to save: %w", err)
	}

	return fmt.Sprintf("Shifted last switch back by %s → boundary now at %s",
		formatDuration(shift), newBoundary.Format("15:04")), nil
}

func (s *State) Status() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress")
	}

	elapsed := time.Since(s.sess.SegmentStart)
	dayElapsed := time.Since(s.sess.DayStart)

	return fmt.Sprintf("Active: %s (%s on this task) | Day: %s since %s",
		s.sess.ActiveProject,
		formatDuration(elapsed),
		formatDuration(dayElapsed),
		s.sess.DayStart.Format("15:04"),
	), nil
}

func (s *State) ListProjects() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var names []string
	for name := range s.registeredProjects {
		names = append(names, name)
	}
	sort.Strings(names)

	var result strings.Builder
	result.WriteString("Registered projects:\n")
	for _, name := range names {
		if s.running && name == s.sess.ActiveProject {
			fmt.Fprintf(&result, "▶ %s (active)\n", name)
			continue
		}
		fmt.Fprintf(&result, "  %s\n", name)
	}
	return result.String(), nil
}

func (s *State) ListProjectNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var names []string
	for name := range s.registeredProjects {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *State) Register(name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == restProject {
		return "", fmt.Errorf("%q is a reserved project name", restProject)
	}
	if s.registeredProjects[name] {
		return "", fmt.Errorf("project %q already registered", name)
	}

	s.registeredProjects[name] = true
	if err := s.saveProjects(); err != nil {
		return "", fmt.Errorf("registered in memory but failed to save: %w", err)
	}
	return fmt.Sprintf("Project %q registered", name), nil
}

func (s *State) Unregister(name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == restProject {
		return "", fmt.Errorf("%q cannot be unregistered — it's a built-in project", restProject)
	}
	if !s.registeredProjects[name] {
		return "", fmt.Errorf("project %q not found", name)
	}
	if s.running && s.sess.ActiveProject == name {
		return "", fmt.Errorf("cannot unregister active project %q — switch first", name)
	}

	delete(s.registeredProjects, name)
	if err := s.saveProjects(); err != nil {
		return "", fmt.Errorf("unregistered in memory but failed to save: %w", err)
	}
	return fmt.Sprintf("Project %q unregistered", name), nil
}

// ---------- internal helpers ----------

// doSwitch performs the actual project switch and checkpoints — caller must hold the lock
func (s *State) doSwitch(project string) (string, error) {
	now := time.Now()
	oldSess := session.Session{
		ActiveProject: s.sess.ActiveProject,
		Date:          s.sess.Date,
		Closed:        s.sess.Closed,
		DayStart:      s.sess.DayStart,
		SegmentStart:  s.sess.SegmentStart,
		Segments:      s.sess.Segments[:],
	}
	s.sess.Segments = append(s.sess.Segments, session.Segment{
		Project: s.sess.ActiveProject,
		Start:   s.sess.SegmentStart,
		End:     now,
	})
	s.sess.ActiveProject = project
	s.sess.SegmentStart = now

	if err := s.checkpoint(); err != nil {
		s.sess = &oldSess
		return "", fmt.Errorf("failed to save switch to %q: %w; state rolled back to %q", project, err, oldSess.ActiveProject)
	}

	return fmt.Sprintf("⏱ %s", project), nil
}

func (s *State) workProjects() []string {
	var names []string
	for name := range s.registeredProjects {
		if name != restProject {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
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
