package notifications

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ShellConfig struct {
	ScriptPath  string
	Args        []string
	Timeout     time.Duration
	WorkingDir  string
	Environment map[string]string
}

type ShellNotification struct {
	config ShellConfig
}

func NewShellNotification(config ShellConfig) *ShellNotification {
	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &ShellNotification{
		config: config,
	}
}

func (s *ShellNotification) Execute(title, message, severity string, count, threshold int) error {
	if s.config.ScriptPath == "" {
		return fmt.Errorf("script path is required")
	}

	// Check if script exists and is executable
	if err := s.validateScript(); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	// Prepare command
	cmd := exec.Command(s.config.ScriptPath, s.config.Args...)

	// Set working directory if specified
	if s.config.WorkingDir != "" {
		cmd.Dir = s.config.WorkingDir
	}

	// Set environment variables
	cmd.Env = os.Environ()

	// Add alert-specific environment variables
	alertEnv := s.buildAlertEnvironment(title, message, severity, count, threshold)
	for key, value := range alertEnv {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add custom environment variables
	for key, value := range s.config.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute with timeout
	return s.executeWithTimeout(cmd)
}

func (s *ShellNotification) validateScript() error {
	// Check if file exists
	info, err := os.Stat(s.config.ScriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script file does not exist: %s", s.config.ScriptPath)
		}
		return fmt.Errorf("cannot access script file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("script path is not a regular file: %s", s.config.ScriptPath)
	}

	// Check if it's executable
	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("script file is not executable: %s (try: chmod +x %s)", s.config.ScriptPath, s.config.ScriptPath)
	}

	return nil
}

func (s *ShellNotification) buildAlertEnvironment(title, message, severity string, count, threshold int) map[string]string {
	timestamp := time.Now().Format("2006-01-02T15:04:05Z07:00")

	return map[string]string{
		"PEEP_ALERT_TITLE":     title,
		"PEEP_ALERT_MESSAGE":   message,
		"PEEP_ALERT_SEVERITY":  severity,
		"PEEP_ALERT_COUNT":     fmt.Sprintf("%d", count),
		"PEEP_ALERT_THRESHOLD": fmt.Sprintf("%d", threshold),
		"PEEP_ALERT_TIMESTAMP": timestamp,
		"PEEP_ALERT_RATIO":     fmt.Sprintf("%.2f", float64(count)/float64(threshold)),
	}
}

func (s *ShellNotification) executeWithTimeout(cmd *exec.Cmd) error {
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start script: %w", err)
	}

	// Create a channel to signal completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}
		return nil
	case <-time.After(s.config.Timeout):
		// Kill the process on timeout
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("script execution timed out after %v", s.config.Timeout)
	}
}

// TestScript executes the script with test data to verify it works
func (s *ShellNotification) TestScript() error {
	return s.Execute(
		"Peep Test Alert",
		"This is a test notification from Peep to verify your shell script integration is working correctly.\n\nScript: "+s.config.ScriptPath+"\nIf you can see this message, your shell script notifications are properly configured!",
		"info",
		5,
		3,
	)
}

// GetScriptInfo returns information about the configured script
func (s *ShellNotification) GetScriptInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Basic script info
	info["script_path"] = s.config.ScriptPath
	info["args"] = s.config.Args
	info["timeout"] = s.config.Timeout.String()
	info["working_dir"] = s.config.WorkingDir

	// File system info
	if stat, err := os.Stat(s.config.ScriptPath); err == nil {
		info["file_size"] = stat.Size()
		info["file_mode"] = stat.Mode().String()
		info["modified_time"] = stat.ModTime().Format("2006-01-02 15:04:05")
		info["is_executable"] = stat.Mode().Perm()&0111 != 0
	} else {
		info["file_error"] = err.Error()
	}

	return info, nil
}

// CreateExampleScript creates an example shell script for testing
func CreateExampleScript(path string) error {
	script := `#!/bin/bash

# Peep Alert Handler Example Script
# This script receives alert information via environment variables

echo "🚨 Peep Alert Received!"
echo "======================="
echo "Title: $PEEP_ALERT_TITLE"
echo "Severity: $PEEP_ALERT_SEVERITY"
echo "Count: $PEEP_ALERT_COUNT"
echo "Threshold: $PEEP_ALERT_THRESHOLD"
echo "Ratio: $PEEP_ALERT_RATIO"
echo "Timestamp: $PEEP_ALERT_TIMESTAMP"
echo ""
echo "Message:"
echo "$PEEP_ALERT_MESSAGE"
echo ""

# Example: Log to a file
echo "$(date): Alert - $PEEP_ALERT_TITLE ($PEEP_ALERT_COUNT/$PEEP_ALERT_THRESHOLD)" >> /tmp/peep-alerts.log

# Example: Send to a webhook (uncomment to use)
# curl -X POST https://your-webhook-url.com/alerts \
#   -H "Content-Type: application/json" \
#   -d "{\"title\":\"$PEEP_ALERT_TITLE\",\"severity\":\"$PEEP_ALERT_SEVERITY\",\"count\":$PEEP_ALERT_COUNT}"

# Example: Play a sound (macOS)
# if command -v afplay &> /dev/null; then
#   afplay /System/Library/Sounds/Glass.aiff
# fi

# Example: Send system notification (Linux)
# if command -v notify-send &> /dev/null; then
#   notify-send "Peep Alert" "$PEEP_ALERT_TITLE: $PEEP_ALERT_COUNT events"
# fi

echo "✅ Alert handled successfully!"
`

	// Write the script file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create script file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		return fmt.Errorf("failed to write script content: %w", err)
	}

	// Make it executable
	if err := os.Chmod(path, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	return nil
}

// ValidateConfig checks if the shell configuration is valid
func (s *ShellNotification) ValidateConfig() error {
	if s.config.ScriptPath == "" {
		return fmt.Errorf("script path is required")
	}

	// Expand relative paths
	if !strings.HasPrefix(s.config.ScriptPath, "/") && !strings.HasPrefix(s.config.ScriptPath, "~") {
		if wd, err := os.Getwd(); err == nil {
			s.config.ScriptPath = fmt.Sprintf("%s/%s", wd, s.config.ScriptPath)
		}
	}

	return s.validateScript()
}
