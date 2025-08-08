#!/bin/bash

# Notification Testing Script
# Tests all Peep notification channels

echo "🔔 Peep Notification System Test"
echo "================================"

echo ""
echo "1. 🖥️  Testing Desktop Notifications..."

# First, let's create a simple test program to test notifications directly
cat > test_notifications.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/kylereynolds/peep/internal/notifications"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test_notifications <type> [args...]")
		fmt.Println("Types: desktop, shell")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "desktop":
		title := "Peep Alert Test"
		message := "This is a test desktop notification from Peep!"
		if len(os.Args) >= 4 {
			title = os.Args[2]
			message = os.Args[3]
		}
		err := notifications.SendDesktopNotification(title, message)
		if err != nil {
			log.Fatalf("Desktop notification failed: %v", err)
		}
		fmt.Println("✅ Desktop notification sent!")

	case "shell":
		if len(os.Args) < 3 {
			fmt.Println("Usage: test_notifications shell <script_path> [args...]")
			os.Exit(1)
		}
		
		config := notifications.ShellConfig{
			ScriptPath: os.Args[2],
			Args:       os.Args[3:],
		}
		
		notification := notifications.NewShellNotification(config)
		err := notification.Send("Test Alert", "This is a test shell notification from Peep!")
		if err != nil {
			log.Fatalf("Shell notification failed: %v", err)
		}
		fmt.Println("✅ Shell notification executed!")

	default:
		fmt.Printf("Unknown notification type: %s\n", os.Args[1])
		os.Exit(1)
	}
}
EOF

# Build the test program
echo "Building notification test program..."
go build -o test_notifications test_notifications.go

if [ $? -eq 0 ]; then
    echo "✅ Test program built successfully"
    
    echo ""
    echo "Testing desktop notification..."
    ./test_notifications desktop "Peep Test" "Desktop notifications are working! 🎉"
    
    if [ $? -eq 0 ]; then
        echo "✅ Desktop notification test passed"
        echo "💡 You should have seen a desktop notification appear"
    else
        echo "❌ Desktop notification test failed"
    fi
else
    echo "❌ Failed to build test program"
fi

echo ""
echo "2. 🐚 Testing Shell Script Notifications..."

# Create a test shell script
cat > test_alert_script.sh << 'EOF'
#!/bin/bash
echo "🚨 ALERT RECEIVED 🚨"
echo "Title: $1"
echo "Message: $2"
echo "Timestamp: $(date)"
echo "Args: $@"
echo "Environment:"
env | grep PEEP_ || echo "No PEEP environment variables set"
echo "Test script executed successfully!"
EOF

chmod +x test_alert_script.sh

echo "Testing shell script notification..."
if [ -f test_notifications ]; then
    ./test_notifications shell ./test_alert_script.sh "Shell Test" "Shell notifications working!"
    
    if [ $? -eq 0 ]; then
        echo "✅ Shell notification test passed"
    else
        echo "❌ Shell notification test failed"
    fi
else
    echo "❌ Test program not available"
fi

echo ""
echo "3. 🔗 Testing HTTP Error Detection Alerts..."

# Test our HTTP error detection with notifications
echo "Injecting HTTP errors to trigger alerts..."
./test-http-errors.sh

echo ""
echo "📊 Current alert rules:"
./peep alerts list

echo ""
echo "4. 💻 Testing with Real Alert System..."

# Check if alert daemon is running
if pgrep -f "peep.*daemon" > /dev/null; then
    echo "✅ Peep daemon is running"
else
    echo "🚀 Starting alert daemon for testing..."
    ./peep daemon --max-logs 10000 --check-mins 60 &
    DAEMON_PID=$!
    echo "Started daemon with PID: $DAEMON_PID"
    sleep 3
fi

echo ""
echo "🎯 Summary:"
echo "- Desktop notifications: Test completed"
echo "- Shell script notifications: Test completed" 
echo "- HTTP error alerts: Injected test data"
echo "- Alert daemon: Running for monitoring"

echo ""
echo "💡 Next steps:"
echo "1. Check if desktop notification appeared"
echo "2. Verify shell script output above"
echo "3. Configure Slack webhook for Slack notifications"
echo "4. Configure SMTP for email notifications"

# Cleanup
echo ""
echo "🧹 Cleaning up test files..."
rm -f test_notifications test_notifications.go test_alert_script.sh

if [ ! -z "$DAEMON_PID" ]; then
    echo "Daemon is still running (PID: $DAEMON_PID) for continued testing"
    echo "Stop it with: kill $DAEMON_PID"
fi
