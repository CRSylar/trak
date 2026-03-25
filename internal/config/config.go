package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "config.json"

// Config holds user-configurable settings
type Config struct {
	SessionsDir string `json:"sessions_dir"`
}

// Load reads the config file, creating it with defaults if it doesn't exist
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg, err := defaults()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First run — write defaults and return them
			return cfg, save(cfg, path)
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if len(cfg.SessionsDir) == 0 {
		return nil, fmt.Errorf("config: SessionsDir (sessions_dir) is empty or not set; session files directory must be configured")
	}

	// Expand ~ in sessions_dir
	cfg.SessionsDir, err = expandHome(cfg.SessionsDir)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaults() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		SessionsDir: filepath.Join(home, ".trak", "sessions"),
	}, nil
}

func save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".trak", configFileName), nil
}

func expandHome(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	rest := path[1:]
	if len(rest) == 0 {
		return home, nil
	}

	if rest[0] == os.PathSeparator {
		rest = rest[1:]
	}

	if len(rest) == 0 {
		return home, nil
	}

	return filepath.Join(home, rest), nil
}
