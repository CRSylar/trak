package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "config.json"

var conf *Config

type Config struct {
	SessionsDir       string `json:"sessions_dir"`
	ReminderStartTime string `json:"reminder_start_time,omitempty"`
	ReminderEndTime   string `json:"reminder_end_time,omitempty"`
}

func Load() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	cfg, err := defaults()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			conf = cfg
			return save(cfg, path)
		}
		return err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("config: failed to read %s: %w", path, err)
	}

	if len(cfg.SessionsDir) == 0 {
		return fmt.Errorf("conifg: SessionsDir (sessions_dir) is empty or not set; session file directory must be configured")
	}

	cfg.SessionsDir, err = expandHome(cfg.SessionsDir)
	if err != nil {
		return err
	}

	conf = cfg
	return nil
}

func GetSessionsDir() string {
	return conf.SessionsDir
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

	data, err := json.MarshalIndent(cfg, "", " ")
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
