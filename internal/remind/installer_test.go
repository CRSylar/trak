package remind

import (
	"strings"
	"testing"
)

func TestCronEntry(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		minute   int
		command  string
		expected string
	}{
		{
			name:     "basic entry",
			hour:     9,
			minute:   30,
			command:  "/usr/bin/trak remind start",
			expected: "30 9 * * * /usr/bin/trak remind start",
		},
		{
			name:     "evening time",
			hour:     18,
			minute:   0,
			command:  "trak remind stop",
			expected: "0 18 * * * trak remind stop",
		},
		{
			name:     "single digit hour",
			hour:     8,
			minute:   5,
			command:  "trak",
			expected: "5 8 * * * trak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cronEntry(tt.hour, tt.minute, tt.command)
			if result != tt.expected {
				t.Errorf("cronEntry(%d, %d, %q) = %q, want %q",
					tt.hour, tt.minute, tt.command, result, tt.expected)
			}
		})
	}
}

func TestTrakCronRegexStart(t *testing.T) {
	tests := []struct {
		name    string
		input  string
		want   bool
	}{
		{
			name:    "valid start cron line",
			input:  "30 9 * * * /usr/local/bin/trak remind start",
			want:   true,
		},
		{
			name:    "valid start cron with different path",
			input:  "30 9 * * * trak remind start",
			want:   true,
		},
		{
			name:    "valid start with path containing spaces",
			input:  "30 9 * * * /Users/cromalde/.local/bin/trak remind start",
			want:   true,
		},
		{
			name:    "stop cron should not match start",
			input:  "0 18 * * * trak remind stop",
			want:   false,
		},
		{
			name:    "comment line",
			input:  "# 30 9 * * * trak remind start",
			want:   false,
		},
		{
			name:    "empty line",
			input:  "",
			want:   false,
		},
		{
			name:    "random cron",
			input:  "0 * * * * some-command",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trakCronStartRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("trakCronStartRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrakCronRegexStop(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid stop cron line",
			input: "0 18 * * * /usr/local/bin/trak remind stop",
			want:  true,
		},
		{
			name:  "valid stop cron with different path",
			input: "0 18 * * * trak remind stop",
			want:  true,
		},
		{
			name:  "start cron should not match stop",
			input: "30 9 * * * trak remind start",
			want:  false,
		},
		{
			name:  "comment line",
			input: "# 0 18 * * * trak remind stop",
			want:  false,
		},
		{
			name:  "random cron",
			input: "*/5 * * * * some-command",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trakCronStopRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("trakCronStopRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCronInstallerName(t *testing.T) {
	ci := &cronInstaller{}
	if got := ci.Name(); got != "cron" {
		t.Errorf("cronInstaller.Name() = %q, want %q", got, "cron")
	}
}

func TestWindowsInstallerName(t *testing.T) {
	wi := &windowsInstaller{}
	if got := wi.Name(); got != "Windows Task Scheduler" {
		t.Errorf("windowsInstaller.Name() = %q, want %q", got, "Windows Task Scheduler")
	}
}

func TestWindowsInstallerIsSupported(t *testing.T) {
	wi := &windowsInstaller{}
	if got := wi.IsSupported(); got != false {
		t.Errorf("windowsInstaller.IsSupported() = %v, want false", got)
	}
}

func TestGetInstaller(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		wantName string
	}{
		{
			name:     "linux platform",
			platform: PlatformLinux,
			wantName: "cron",
		},
		{
			name:     "darwin platform",
			platform: PlatformDarwin,
			wantName: "cron",
		},
		{
			name:     "windows platform",
			platform: PlatformWindows,
			wantName: "Windows Task Scheduler",
		},
		{
			name:     "unknown platform defaults to cron",
			platform: "unknown",
			wantName: "cron",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := GetInstaller(tt.platform)
			if got := installer.Name(); got != tt.wantName {
				t.Errorf("GetInstaller(%q).Name() = %q, want %q", tt.platform, got, tt.wantName)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input    string
		wantHour int
		wantMin  int
		wantErr bool
	}{
		{
			name:      "valid time morning",
			input:    "09:30",
			wantHour: 9,
			wantMin:  30,
			wantErr:  false,
		},
		{
			name:      "valid time evening",
			input:    "18:00",
			wantHour: 18,
			wantMin:  0,
			wantErr:  false,
		},
		{
			name:      "single digit hour",
			input:    "8:05",
			wantHour: 8,
			wantMin:  5,
			wantErr:  false,
		},
		{
			name:      "midnight",
			input:    "00:00",
			wantHour: 0,
			wantMin:  0,
			wantErr:  false,
		},
		{
			name:      "invalid format missing colon",
			input:    "0930",
			wantErr:  true,
		},
		{
			name:      "invalid minute",
			input:    "09:60",
			wantErr:  true,
		},
		{
			name:      "invalid hour",
			input:    "24:00",
			wantErr:  true,
		},
		{
			name:      "invalid format",
			input:    "9:30:00",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour, min, err := parseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if hour != tt.wantHour || min != tt.wantMin {
					t.Errorf("parseTime(%q) = (%d, %d), want (%d, %d)",
						tt.input, hour, min, tt.wantHour, tt.wantMin)
				}
			}
		})
	}
}

func TestScheduleCronLines(t *testing.T) {
	sched := &Schedule{
		StartHour:   9,
		StartMinute: 30,
		EndHour:     18,
		EndMinute:  0,
	}

	startCron, stopCron := sched.CronLines()

	if !strings.Contains(startCron, "remind start") {
		t.Errorf("start cron does not contain 'remind start': %s", startCron)
	}
	if !strings.Contains(stopCron, "remind stop") {
		t.Errorf("stop cron does not contain 'remind stop': %s", stopCron)
	}

	startParts := strings.Fields(startCron)
	if startParts[0] != "30" || startParts[1] != "9" {
		t.Errorf("start cron expected '30 9', got '%s %s'", startParts[0], startParts[1])
	}

	stopParts := strings.Fields(stopCron)
	if stopParts[0] != "0" || stopParts[1] != "18" {
		t.Errorf("stop cron expected '0 18', got '%s %s'", stopParts[0], stopParts[1])
	}
}

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()
	if platform == "" {
		t.Error("DetectPlatform() returned empty platform")
	}

	validPlatforms := map[Platform]bool{
		PlatformLinux:   true,
		PlatformDarwin: true,
		PlatformWindows: true,
	}

	if !validPlatforms[platform] {
		t.Errorf("DetectPlatform() = %q, expected a valid platform", platform)
	}
}

func TestInstallInstructions(t *testing.T) {
	sched := &Schedule{
		StartHour:   9,
		StartMinute: 30,
		EndHour:     18,
		EndMinute:  0,
	}

	instructions := sched.InstallInstructions(PlatformLinux)
	if !strings.Contains(instructions, "09:30") {
		t.Errorf("Linux instructions should contain '09:30', got: %s", instructions)
	}
	if !strings.Contains(instructions, "18:00") {
		t.Errorf("Linux instructions should contain '18:00', got: %s", instructions)
	}

	instructionsDarwin := sched.InstallInstructions(PlatformDarwin)
	if !strings.Contains(instructionsDarwin, "09:30") {
		t.Errorf("Darwin instructions should contain '09:30', got: %s", instructionsDarwin)
	}

	instructionsWindows := sched.InstallInstructions(PlatformWindows)
	if !strings.Contains(instructionsWindows, "18:00") {
		t.Errorf("Windows instructions should contain '18:00', got: %s", instructionsWindows)
	}
}