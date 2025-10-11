package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// IsSSH returns true if the session is running over SSH
func IsSSH() bool {
	return os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != ""
}

// IsInteractiveEnvironment returns true if the environment appears to be
// an interactive TTY session (and not CI)
func IsInteractiveEnvironment() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	if fi, err := os.Stderr.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// OpenBrowser opens the specified URL in the default browser
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = "open"
		args = []string{url}
	case "linux":
		// Try different commands in order of preference
		for _, openCmd := range []string{"xdg-open", "gnome-open", "kde-open"} {
			if _, err := exec.LookPath(openCmd); err == nil {
				cmd = openCmd
				args = []string{url}
				break
			}
		}
		if cmd == "" {
			return fmt.Errorf("no suitable browser opener found for Linux")
		}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Use Start() instead of Run() to avoid waiting for the browser to close
	if err := exec.Command(cmd, args...).Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}
