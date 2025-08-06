package notifications

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SendDesktopNotification sends a desktop notification using the OS notification system
func SendDesktopNotification(title, message string) error {
	switch runtime.GOOS {
	case "darwin": // macOS
		return sendMacOSNotification(title, message)
	case "linux":
		return sendLinuxNotification(title, message)
	case "windows":
		return sendWindowsNotification(title, message)
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}
}

// sendMacOSNotification sends a notification on macOS using osascript
func sendMacOSNotification(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// sendLinuxNotification sends a notification on Linux using notify-send
func sendLinuxNotification(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	return cmd.Run()
}

// sendWindowsNotification sends a notification on Windows using PowerShell
func sendWindowsNotification(title, message string) error {
	script := fmt.Sprintf(`
	[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
	[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
	[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

	$template = @"
	<toast>
		<visual>
			<binding template="ToastText02">
				<text id="1">%s</text>
				<text id="2">%s</text>
			</binding>
		</visual>
	</toast>
	"@

	$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
	$xml.LoadXml($template)
	$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
	$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Peep")
	$notifier.Show($toast)
	`, title, message)

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}
