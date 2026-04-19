package notify

import (
	"os/exec"
	"runtime"

	"github.com/gen2brain/beeep"
)

// Notify sends a desktop notification with the given title and message.
// It returns an error if the notification fails.
func Notify(title, message string) error {
	return beeep.Notify(title, message, "")
}

// NotifyWithIcon sends a desktop notification with a custom icon path.
func NotifyWithIcon(title, message, iconPath string) error {
	return beeep.Notify(title, message, iconPath)
}

// NotifyIfNotStarted sends a reminder to start the workday.
func NotifyIfNotStarted() error {
	return Notify("trak — start your workday", "It's time to start tracking! Run 'trak start'")
}

// NotifyIfNotStopped sends a reminder to end the workday.
func NotifyIfNotStopped() error {
	return Notify("trak — end your workday", "Remember to stop tracking! Run 'trak end'")
}

// NotifyCustom sends a custom notification with the given message.
func NotifyCustom(msg string) error {
	return Notify("trak", msg)
}

// Test sends a test notification to verify the setup.
func Test() error {
	return Notify("trak test", "Notifications are working!")
}

// IsAvailable checks if notifications are supported on the current platform.
func IsAvailable() bool {
	switch runtime.GOOS {
	case "linux":
		_, err := exec.LookPath("notify-send")
		return err == nil
	case "darwin":
		_, err := exec.LookPath("osascript")
		return err == nil
	case "windows":
		_, err := exec.LookPath("powershell")
		return err == nil
	default:
		return false
	}
}
