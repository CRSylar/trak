package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const RestProject = "rest"
const projectsFileName = "projects.json"

type projectConfig struct {
	Projects []string `json:"projects"`
}

func GetRegisteredProjects() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projPath := filepath.Join(home, ".trak", projectsFileName)

	data, err := os.ReadFile(projPath)
	if err != nil {
		return nil, err
	}

	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("getRegisteredProjects: cannot Unmarshal data in file %s: %w", projPath, err)
	}

	return cfg.Projects, nil
}

func SaveProjectConfig(p []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	projPath := filepath.Join(home, ".trak", projectsFileName)

	cfg := projectConfig{Projects: p}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(projPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(projPath, data, 0644)
}
