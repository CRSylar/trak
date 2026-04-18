package command

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/CRSylar/trak/internal/projects"
	"github.com/CRSylar/trak/internal/session"
)

func CicleNextProject() {
	now := time.Now()
	prjs, err := projects.GetRegisteredProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot retrieve registered projects: %w", err))
		return
	}
	// load the session file
	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot load session file: %w", err))
		return
	}

	cicleIndex := 0
	if sess.ActiveProject != projects.RestProject {
		for i, p := range prjs {
			if p == sess.ActiveProject {
				cicleIndex = (i + 1) % len(prjs)
				break
			}
		}
	}

	if cicleIndex >= len(prjs) {
		cicleIndex = 0
	}

	sess.ActiveProject = prjs[cicleIndex]
	sess.Segments[len(sess.Segments)-1].End = now
	sess.Segments = append(sess.Segments, session.Segment{
		Project: prjs[cicleIndex],
		Start:   now,
	})

	if err = session.Save(*sess, session.FilePath(now.Format("2006-01-02"))); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot switch to new project: %w", err))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "⏱ %s", prjs[cicleIndex])
}

func SwitchProject(project string) {
	now := time.Now()

	prjs, err := projects.GetRegisteredProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot retrieve registered projects: %w", err))
		return
	}

	if !slices.Contains(prjs, project) {
		fmt.Fprint(os.Stderr,
			fmt.Errorf("unknow project %s - register it first with 'trak register %s'", project, project),
		)
		return
	}

	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot load session file: %w", err))
		return
	}

	sess.ActiveProject = project
	sess.Segments[len(sess.Segments)-1].End = now
	sess.Segments = append(sess.Segments, session.Segment{
		Project: project,
		Start:   now,
	})

	if err = session.Save(*sess, session.FilePath(now.Format("2006-01-02"))); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot switch to new project: %w", err))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "⏱ %s", projects.RestProject)

}

func ListProjects() {
	now := time.Now()
	prjs, err := projects.GetRegisteredProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot retrieve registered projects: %w", err))
		return
	}

	sess, err := session.LoadSession(now.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot load session file: %w", err))
		return
	}

	var result strings.Builder
	result.WriteString("Registered projects:\n")
	for _, name := range prjs {
		if name == sess.ActiveProject {
			fmt.Fprintf(&result, "▶ %s (active)\n", name)
			continue
		}
		fmt.Fprintf(&result, "  %s\n", name)
	}
	fmt.Fprintf(os.Stdout, "%s", result.String())

}

func Register(newProj string) {
	prjs, err := projects.GetRegisteredProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot retrieve registered projects: %w", err))
		return
	}

	if slices.Contains(prjs, newProj) || newProj == projects.RestProject {
		fmt.Fprintf(os.Stderr, "project %s already registered", newProj)
		return
	}

	if err := projects.SaveProjectConfig(append(prjs, newProj)); err != nil {
		fmt.Fprint(os.Stdout, fmt.Errorf("failed to save projectConfig file: %w", err))
		return
	}
	fmt.Fprintf(os.Stdout, "Project %s registered", newProj)
}

func Unregister(prjToRemove string) {
	prjs, err := projects.GetRegisteredProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("cannot retrieve registered projects: %w", err))
		return
	}

	if prjToRemove == projects.RestProject {
		fmt.Fprintf(os.Stderr, "%s cannot be unregistered — it's a built-in project", prjToRemove)
		return
	}

	if !slices.Contains(prjs, prjToRemove) {
		fmt.Fprintf(os.Stderr, "project %s not found", prjToRemove)
		return
	}

	prjIndex := slices.Index(prjs, prjToRemove)

	if err := projects.SaveProjectConfig(slices.Delete(prjs, prjIndex, prjIndex+1)); err != nil {
		fmt.Fprint(os.Stdout, fmt.Errorf("failed to save projectConfig file: %w", err))
		return
	}
	fmt.Fprintf(os.Stdout, "Project %s unregistered", prjToRemove)
}
