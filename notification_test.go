package main

import (
	"fmt"
	"os"

	"github.com/kylereynolds/peep/internal/notifications"
)

func main() {
	fmt.Println("🔔 Testing Peep Notifications")
	fmt.Println("=============================")

	// Test 1: Desktop Notification
	fmt.Println("1. Testing Desktop Notification...")
	err := notifications.SendDesktopNotification("Peep Alert Test", "Desktop notifications are working! 🎉")
	if err != nil {
		fmt.Printf("❌ Desktop notification failed: %v\n", err)
	} else {
		fmt.Println("✅ Desktop notification sent successfully!")
		fmt.Println("💡 You should see a notification on your desktop")
	}

	// Test 2: Shell Script Notification
	fmt.Println("\n2. Testing Shell Script Notification...")

	// Create a simple test script
	script := `#!/bin/bash
echo "🚨 PEEP ALERT RECEIVED 🚨"
echo "Title: $1"
echo "Message: $2"
echo "Severity: $3"
echo "Count: $4"
echo "Threshold: $5"
echo "Timestamp: $(date)"
echo "Environment Variables:"
env | grep PEEP_ | sort || echo "No PEEP environment variables"
echo "Shell notification test completed!"
`

	// Write test script
	err = os.WriteFile("test_shell_alert.sh", []byte(script), 0755)
	if err != nil {
		fmt.Printf("❌ Failed to create test script: %v\n", err)
	} else {
		// Test shell notification
		config := notifications.ShellConfig{
			ScriptPath: "./test_shell_alert.sh",
		}

		shellNotifier := notifications.NewShellNotification(config)
		err = shellNotifier.Execute("Test Alert", "Shell notification working!", "medium", 5, 3)
		if err != nil {
			fmt.Printf("❌ Shell notification failed: %v\n", err)
		} else {
			fmt.Println("✅ Shell notification executed successfully!")
		}

		// Clean up
		os.Remove("test_shell_alert.sh")
	}

	fmt.Println("\n🎯 Notification Tests Complete!")
	fmt.Println("Next: Configure Slack/Email for production alerts")
}
