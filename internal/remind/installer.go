package remind

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var trakCronStartRegex = regexp.MustCompile(`^\d+\s+\d+\s+\*\s+\*\s+\*\s+.*trak\s+remind\s+start`)
var trakCronStopRegex = regexp.MustCompile(`^\d+\s+\d+\s+\*\s+\*\s+\*\s+.*trak\s+remind\s+stop`)

func GetInstaller(platform Platform) Installer {
	switch platform {
	case PlatformLinux, PlatformDarwin:
		return &cronInstaller{}
	case PlatformWindows:
		return &windowsInstaller{}
	default:
		return &cronInstaller{}
	}
}

type Installer interface {
	Install(startCron, stopCron string) error
	Uninstall() error
	CheckExisting() (startExists, stopExists bool, err error)
	IsSupported() bool
	Name() string
}

type cronInstaller struct{}

func (c *cronInstaller) Name() string {
	return "cron"
}

func (c *cronInstaller) IsSupported() bool {
	_, err := exec.LookPath("crontab")
	return err == nil
}

func (c *cronInstaller) CheckExisting() (startExists, stopExists bool, err error) {
	lines, err := c.readCrontab()
	if err != nil {
		return false, false, err
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if trakCronStartRegex.MatchString(line) {
			startExists = true
		}
		if trakCronStopRegex.MatchString(line) {
			stopExists = true
		}
	}

	return startExists, stopExists, nil
}

func (c *cronInstaller) readCrontab() ([]string, error) {
	cmd := exec.Command("crontab", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return []string{}, nil
			}
		}
		return nil, fmt.Errorf("failed to read crontab: %w", err)
	}

	scanner := bufio.NewScanner(&out)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading crontab: %w", err)
	}

	return lines, nil
}

func (c *cronInstaller) Install(startCron, stopCron string) error {
	if !c.IsSupported() {
		return fmt.Errorf("crontab command not found; cannot install reminders automatically")
	}

	lines, err := c.readCrontab()
	if err != nil {
		return err
	}

	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# trak:") {
			continue
		}
		if trakCronStartRegex.MatchString(line) || trakCronStopRegex.MatchString(line) {
			continue
		}
		newLines = append(newLines, line)
	}

	newLines = append(newLines, "", "# trak: start reminder", startCron)
	newLines = append(newLines, "# trak: stop reminder", stopCron)

	content := strings.Join(newLines, "\n")
	if len(newLines) > 0 {
		content += "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install crontab: %s", stderr.String())
	}

	return nil
}

func (c *cronInstaller) Uninstall() error {
	if !c.IsSupported() {
		return fmt.Errorf("crontab command not found; cannot uninstall reminders automatically")
	}

	lines, err := c.readCrontab()
	if err != nil {
		return err
	}

	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# trak:") {
			continue
		}
		if trakCronStartRegex.MatchString(line) || trakCronStopRegex.MatchString(line) {
			continue
		}
		newLines = append(newLines, line)
	}

	content := strings.Join(newLines, "\n")
	if len(newLines) > 0 {
		content += "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall crontab: %s", stderr.String())
	}

	return nil
}

type windowsInstaller struct{}

func (w *windowsInstaller) Name() string {
	return "Windows Task Scheduler"
}

func (w *windowsInstaller) IsSupported() bool {
	return false
}

func (w *windowsInstaller) CheckExisting() (startExists, stopExists bool, err error) {
	return false, false, fmt.Errorf("Windows Task Scheduler automation not supported; use manual instructions")
}

func (w *windowsInstaller) Install(startCron, stopCron string) error {
	return fmt.Errorf("Windows Task Scheduler automation not supported; use manual instructions (run 'trak install-reminders' for details)")
}

func (w *windowsInstaller) Uninstall() error {
	return fmt.Errorf("Windows Task Scheduler automation not supported; remove tasks manually via Task Scheduler")
}
