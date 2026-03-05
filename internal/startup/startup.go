package startup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InitSystem represents an initialization system
type InitSystem string

const (
	Systemd  InitSystem = "systemd"
	Upstart  InitSystem = "upstart"
	Launchd  InitSystem = "launchd"
	RC       InitSystem = "rc"
	Unknown  InitSystem = "unknown"
)

// DetectInitSystem detects the current initialization system
func DetectInitSystem() InitSystem {
	// Check for systemd
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return Systemd
	}

	// Check for upstart
	if _, err := os.Stat("/etc/init"); err == nil {
		return Upstart
	}

	// Check for launchd (macOS)
	if _, err := exec.LookPath("launchctl"); err == nil {
		return Launchd
	}

	// Check for rc.d
	if _, err := os.Stat("/etc/rc.d"); err == nil {
		return RC
	}

	return Unknown
}

// GenerateStartupScript generates a startup script for the current init system
func GenerateStartupScript() error {
	initSystem := DetectInitSystem()

	switch initSystem {
	case Systemd:
		return generateSystemdScript()
	case Upstart:
		return generateUpstartScript()
	case Launchd:
		return generateLaunchdScript()
	case RC:
		return generateRCScript()
	default:
		return fmt.Errorf("unsupported init system: %s", initSystem)
	}
}

// generateSystemdScript generates a systemd service file
func generateSystemdScript() error {
	serviceFile := "/etc/systemd/system/pm3.service"

	// Check if we have root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges are required to generate systemd service file")
	}

	// Get pm3 executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=PM3 Process Manager
After=network.target

[Service]
Type=forking
ExecStart=%s startup
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, execPath)

	// Write service file
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}

	// Enable and start the service
	if err := exec.Command("systemctl", "enable", "pm3.service").Run(); err != nil {
		return err
	}

	if err := exec.Command("systemctl", "start", "pm3.service").Run(); err != nil {
		return err
	}

	return nil
}

// generateUpstartScript generates an upstart job file
func generateUpstartScript() error {
	jobFile := "/etc/init/pm3.conf"

	// Check if we have root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges are required to generate upstart job file")
	}

	// Get pm3 executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create job file content
	jobContent := fmt.Sprintf(`description "PM3 Process Manager"
author "PM3 Team"

start on runlevel [2345]
stop on runlevel [016]

respawn

script
  exec %s startup
end script
`, execPath)

	// Write job file
	if err := os.WriteFile(jobFile, []byte(jobContent), 0644); err != nil {
		return err
	}

	// Start the job
	if err := exec.Command("initctl", "start", "pm3").Run(); err != nil {
		return err
	}

	return nil
}

// generateLaunchdScript generates a launchd plist file
func generateLaunchdScript() error {
	plistFile := "~/Library/LaunchAgents/com.pm3.plist"
	plistFile = strings.Replace(plistFile, "~", os.Getenv("HOME"), 1)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(plistFile), 0755); err != nil {
		return err
	}

	// Get pm3 executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.pm3</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>startup</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`, execPath)

	// Write plist file
	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return err
	}

	// Load the plist
	if err := exec.Command("launchctl", "load", plistFile).Run(); err != nil {
		return err
	}

	return nil
}

// generateRCScript generates an rc.d script
func generateRCScript() error {
	rcFile := "/etc/rc.d/init.d/pm3"

	// Check if we have root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges are required to generate rc.d script")
	}

	// Get pm3 executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create rc script content
	rcContent := fmt.Sprintf(`#!/bin/sh
# PM3 Process Manager

case "$1" in
  start)
    %s startup
    ;;
  stop)
    %s stop all
    ;;
  restart)
    %s restart all
    ;;
  *)
    echo "Usage: $0 {start|stop|restart}"
    exit 1
    ;;
esac
`, execPath, execPath, execPath)

	// Write rc script
	if err := os.WriteFile(rcFile, []byte(rcContent), 0755); err != nil {
		return err
	}

	// Add to startup
	if err := exec.Command("chkconfig", "--add", "pm3").Run(); err != nil {
		return err
	}

	if err := exec.Command("chkconfig", "pm3", "on").Run(); err != nil {
		return err
	}

	// Start the service
	if err := exec.Command("service", "pm3", "start").Run(); err != nil {
		return err
	}

	return nil
}
