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
)

const restProject = "rest"
const configFileName = "projects.json"

// Segment represents a contiguous block of work on a single project
type Segment struct {
	Project string
	Start   time.Time
	End     time.Time
}

// State holds all runtime state for a workday
type State struct {
	mu                 sync.Mutex
	running            bool
	dayStart           time.Time
	activeProject      string
	segmentStart       time.Time
	segments           []Segment
	registeredProjects map[string]bool // name -> true
	configPath         string
	cycleIndex         int // index into sorted non-rest project list
}

func New() *State {
	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".trak", configFileName)

	s := &State{
		configPath:         cfgPath,
		registeredProjects: make(map[string]bool),
	}
	s.registeredProjects[restProject] = true
	s.loadProjects()
	return s
}

// ---------- project persistence ----------

type projectConfig struct {
	Projects []string `json:"projects"`
}

func (s *State) loadProjects() {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return // file doesn't exist yet, fine
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	for _, p := range cfg.Projects {
		s.registeredProjects[p] = true
	}
}

func (s *State) saveProjects() error {
	var names []string
	for name := range s.registeredProjects {
		if name == restProject {
			continue // rest is always implicit
		}
		names = append(names, name)
	}
	sort.Strings(names)

	cfg := projectConfig{Projects: names}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

// ---------- commands ----------

func (s *State) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return "", fmt.Errorf("workday already started (active project: %s)", s.activeProject)
	}

	s.running = true
	s.dayStart = time.Now()
	s.activeProject = restProject
	s.segmentStart = s.dayStart
	s.segments = nil

	return fmt.Sprintf("Workday started at %s. Active project: %s",
		s.dayStart.Format("15:04"), restProject), nil
}

func (s *State) End() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress")
	}

	now := time.Now()
	// Close the current segment
	s.segments = append(s.segments, Segment{
		Project: s.activeProject,
		Start:   s.segmentStart,
		End:     now,
	})

	report := buildReport(s.dayStart, now, s.segments)
	s.running = false
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
	if s.activeProject == project {
		return fmt.Sprintf("already on %s", project), nil
	}
	return s.doSwitch(project)
}

func (s *State) Status() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress")
	}

	elapsed := time.Since(s.segmentStart)
	dayElapsed := time.Since(s.dayStart)

	return fmt.Sprintf("Active: %s (%s on this task) | Day: %s since %s",
		s.activeProject,
		formatDuration(elapsed),
		formatDuration(dayElapsed),
		s.dayStart.Format("15:04"),
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
		marker := "  "
		if s.running && name == s.activeProject {
			marker := "▶ "
			_ = marker
			fmt.Fprintf(&result, "▶ %s (active)\n", name)
			continue
		}
		fmt.Fprintf(&result, "%s%s\n", marker, name)
	}
	return result.String(), nil
}

// ListProjectNames returns just the names (used by Raycast script)
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

// workProjects returns sorted project names excluding rest (no lock — caller must hold)
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

func (s *State) Next() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}

	projects := s.workProjects()
	if len(projects) == 0 {
		return "", fmt.Errorf("no work projects registered — add one with 'trak register <name>'")
	}

	// If currently on rest, resume at last cycleIndex; otherwise advance
	if s.activeProject != restProject {
		// find current index and advance
		for i, p := range projects {
			if p == s.activeProject {
				s.cycleIndex = (i + 1) % len(projects)
				break
			}
		}
	}
	// clamp in case projects were removed
	if s.cycleIndex >= len(projects) {
		s.cycleIndex = 0
	}

	next := projects[s.cycleIndex]
	return s.doSwitch(next)
}

func (s *State) Rest() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return "", fmt.Errorf("no workday in progress — run 'trak start' first")
	}
	if s.activeProject == restProject {
		return "already on rest", nil
	}
	return s.doSwitch(restProject)
}

// doSwitch performs the actual project switch — caller must hold the lock
func (s *State) doSwitch(project string) (string, error) {
	now := time.Now()
	s.segments = append(s.segments, Segment{
		Project: s.activeProject,
		Start:   s.segmentStart,
		End:     now,
	})
	s.activeProject = project
	s.segmentStart = now
	return fmt.Sprintf("⏱ %s", project), nil
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
	if s.running && s.activeProject == name {
		return "", fmt.Errorf("cannot unregister active project %q — switch first", name)
	}

	delete(s.registeredProjects, name)
	if err := s.saveProjects(); err != nil {
		return "", fmt.Errorf("unregistered in memory but failed to save: %w", err)
	}
	return fmt.Sprintf("Project %q unregistered", name), nil
}

// ---------- helpers ----------

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh %02dm", h, m)
}
