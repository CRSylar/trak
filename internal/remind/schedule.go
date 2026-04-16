package remind

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/CRSylar/trak/internal/config"
)

var parseTimeRegexp = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)

// ParseTime parses a HH:MM string and returns hour and minute.
// Returns error if format is invalid.
func ParseTime(s string) (hour, minute int, err error) {
	match := parseTimeRegexp.FindStringSubmatch(s)
	if match == nil {
		return 0, 0, fmt.Errorf("invalid time format: %q (expected HH:MM)", s)
	}
	hour, err1 := strconv.Atoi(match[1])
	minute, err2 := strconv.Atoi(match[2])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid hour/minute: %q", s)
	}
	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("hour must be 0-23: %q", s)
	}
	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("minute must be 0-59: %q", s)
	}
	return hour, minute, nil
}

// CronEntry returns a cron line for the given command at hour:minute.
func CronEntry(hour, minute int, command string) string {
	return fmt.Sprintf("%d %d * * * %s", minute, hour, command)
}

// Schedule holds the parsed reminder times.
type Schedule struct {
	StartHour, StartMinute int
	EndHour, EndMinute     int
}

// LoadSchedule loads reminder times from config.
// Returns nil if either time is not set.
func LoadSchedule(cfg *config.Config) (*Schedule, error) {
	if cfg.ReminderStartTime == "" || cfg.ReminderEndTime == "" {
		return nil, nil
	}
	startHour, startMin, err := ParseTime(cfg.ReminderStartTime)
	if err != nil {
		return nil, fmt.Errorf("reminder_start_time: %w", err)
	}
	endHour, endMin, err := ParseTime(cfg.ReminderEndTime)
	if err != nil {
		return nil, fmt.Errorf("reminder_end_time: %w", err)
	}
	return &Schedule{
		StartHour:   startHour,
		StartMinute: startMin,
		EndHour:     endHour,
		EndMinute:   endMin,
	}, nil
}

// CronLines returns the two cron lines for start and stop reminders.
// Uses the absolute path to the current trak executable when available.
func (s *Schedule) CronLines() (startCron, stopCron string) {
	commandPrefix := "trak"
	if executablePath, err := os.Executable(); err == nil && executablePath != "" {
		commandPrefix = fmt.Sprintf("%q", executablePath)
	}

	startCron = CronEntry(s.StartHour, s.StartMinute, commandPrefix+" remind start")
	stopCron = CronEntry(s.EndHour, s.EndMinute, commandPrefix+" remind stop")
	return startCron, stopCron
}

// Platform represents the current operating system scheduler.
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformDarwin  Platform = "darwin"
	PlatformWindows Platform = "windows"
)

// DetectPlatform returns the current platform.
func DetectPlatform() Platform {
	switch strings.ToLower(runtime.GOOS) {
	case "linux":
		return PlatformLinux
	case "darwin":
		return PlatformDarwin
	case "windows":
		return PlatformWindows
	default:
		return PlatformLinux
	}
}

// InstallInstructions returns human-readable instructions for installing reminders.
func (s *Schedule) InstallInstructions(platform Platform) string {
	startCron, stopCron := s.CronLines()
	var sb strings.Builder
	sb.WriteString("Reminder schedule:\n")
	sb.WriteString(fmt.Sprintf("  Start: %02d:%02d\n", s.StartHour, s.StartMinute))
	sb.WriteString(fmt.Sprintf("  End:   %02d:%02d\n", s.EndHour, s.EndMinute))
	sb.WriteString("\n")

	switch platform {
	case PlatformLinux, PlatformDarwin:
		sb.WriteString("To install with cron, add these lines to your crontab (crontab -e):\n")
		sb.WriteString(fmt.Sprintf("  %s\n", startCron))
		sb.WriteString(fmt.Sprintf("  %s\n", stopCron))
		if platform == PlatformDarwin {
			sb.WriteString("\nNote: macOS also supports launchd; cron instructions provided for simplicity.\n")
		}
	case PlatformWindows:
		exeCommand := "trak"
		if exePath, err := os.Executable(); err == nil && exePath != "" {
			exeCommand = strconv.Quote(exePath)
		}

		sb.WriteString("Windows: create scheduled tasks using Task Scheduler.\n")
		sb.WriteString(fmt.Sprintf("Command: %s remind start at %02d:%02d\n", exeCommand, s.StartHour, s.StartMinute))
		sb.WriteString(fmt.Sprintf("Command: %s remind stop at %02d:%02d\n", exeCommand, s.EndHour, s.EndMinute))
	}
	return sb.String()
}
