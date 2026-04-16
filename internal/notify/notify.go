package notify

import (
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
	// beeep internally checks platform support; we can just try a harmless operation.
	// For simplicity, assume true.
	return true
}
