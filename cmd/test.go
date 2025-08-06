package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kylereynolds/peep/internal/notifications"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test notification channels",
	Long:  `Send test notifications to verify your channels are working correctly.`,
}

var testSlackCmd = &cobra.Command{
	Use:   "slack [webhook-url]",
	Short: "Test Slack notification",
	Long: `Send a test notification to Slack to verify webhook is working.
	
Example:
  peep test slack https://hooks.slack.com/services/YOUR/WEBHOOK/URL`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhookURL := args[0]

		fmt.Println("📱 Sending test Slack notification...")

		title := "Test Alert"
		message := "This is a test notification from Peep! If you can see this, your Slack integration is working perfectly."

		err := notifications.SendSlackNotification(webhookURL, title, message, 5, 3)
		if err != nil {
			fmt.Printf("❌ Failed to send Slack notification: %v\n", err)
			fmt.Println("💡 Check your webhook URL and try again")
			return
		}

		fmt.Println("✅ Test notification sent successfully!")
		fmt.Println("🎉 Check your Slack channel to see the message")
	},
}

var testDesktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Test desktop notification",
	Long:  `Send a test desktop notification to verify it's working on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🖥️  Sending test desktop notification...")

		err := notifications.SendDesktopNotification("Peep Test", "This is a test notification from Peep!")
		if err != nil {
			fmt.Printf("❌ Failed to send desktop notification: %v\n", err)
			fmt.Println("💡 Desktop notifications may not be supported on your system")
			return
		}

		fmt.Println("✅ Test notification sent successfully!")
		fmt.Println("🎉 You should see a desktop notification now")
	},
}

var testEmailCmd = &cobra.Command{
	Use:   "email",
	Short: "Test email notification",
	Long: `Send a test email notification to verify SMTP configuration.
	
Example:
  peep test email --smtp-host smtp.gmail.com --username user@gmail.com --password app-password --from user@gmail.com --to recipient@example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get configuration from flags
		smtpHost, _ := cmd.Flags().GetString("smtp-host")
		smtpPort, _ := cmd.Flags().GetString("smtp-port")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		fromEmail, _ := cmd.Flags().GetString("from")
		fromName, _ := cmd.Flags().GetString("from-name")
		toEmail, _ := cmd.Flags().GetString("to")

		if smtpHost == "" || username == "" || password == "" || fromEmail == "" || toEmail == "" {
			fmt.Println("❌ Email test requires SMTP configuration")
			fmt.Println("💡 Required flags: --smtp-host, --username, --password, --from, --to")
			fmt.Println("💡 Example: peep test email --smtp-host smtp.gmail.com --username user@gmail.com --password app-password --from user@gmail.com --to recipient@example.com")
			return
		}

		fmt.Println("📧 Sending test email notification...")

		// Parse SMTP port
		port := 587 // default
		if smtpPort != "" {
			if parsedPort, err := strconv.Atoi(smtpPort); err == nil && parsedPort > 0 {
				port = parsedPort
			}
		}

		emailConfig := notifications.EmailConfig{
			SMTPHost:  smtpHost,
			SMTPPort:  port,
			Username:  username,
			Password:  password,
			FromEmail: fromEmail,
			FromName:  fromName,
			ToEmails:  []string{toEmail},
		}

		emailNotifier := notifications.NewEmailNotification(emailConfig)

		err := emailNotifier.TestConnection()
		if err != nil {
			fmt.Printf("❌ Failed to send email notification: %v\n", err)
			fmt.Println("💡 Check your SMTP configuration and try again")
			return
		}

		fmt.Println("✅ Test email sent successfully!")
		fmt.Printf("🎉 Check %s for the test message\n", toEmail)
	},
}

var testShellCmd = &cobra.Command{
	Use:   "shell [script-path]",
	Short: "Test shell script notification",
	Long: `Execute a shell script with test alert data to verify it works.
	
Example:
  peep test shell ./alert-handler.sh
  peep test shell /path/to/script.sh --timeout 60s --args "--verbose"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := args[0]

		// Get configuration from flags
		argsStr, _ := cmd.Flags().GetString("args")
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		workingDir, _ := cmd.Flags().GetString("working-dir")
		envStr, _ := cmd.Flags().GetString("env")

		fmt.Printf("🖥️  Testing shell script: %s\n", scriptPath)

		// Parse timeout
		timeout := 30 * time.Second
		if timeoutStr != "" {
			if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
				timeout = parsedTimeout
			}
		}

		// Parse args
		var scriptArgs []string
		if argsStr != "" {
			scriptArgs = strings.Split(argsStr, " ")
		}

		// Parse environment variables
		environment := make(map[string]string)
		if envStr != "" {
			for _, pair := range strings.Split(envStr, ",") {
				if parts := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(parts) == 2 {
					environment[parts[0]] = parts[1]
				}
			}
		}

		shellConfig := notifications.ShellConfig{
			ScriptPath:  scriptPath,
			Args:        scriptArgs,
			Timeout:     timeout,
			WorkingDir:  workingDir,
			Environment: environment,
		}

		shellNotifier := notifications.NewShellNotification(shellConfig)

		err := shellNotifier.TestScript()
		if err != nil {
			fmt.Printf("❌ Failed to execute shell script: %v\n", err)
			fmt.Println("💡 Check script path, permissions, and try again")
			return
		}

		fmt.Println("✅ Shell script executed successfully!")
		fmt.Printf("🎉 Script %s handled the test alert\n", scriptPath)
	},
}

func init() {
	// Add email test flags
	testEmailCmd.Flags().StringP("smtp-host", "", "", "SMTP server hostname (e.g., smtp.gmail.com)")
	testEmailCmd.Flags().StringP("smtp-port", "", "587", "SMTP server port (default: 587)")
	testEmailCmd.Flags().StringP("username", "", "", "SMTP username/email")
	testEmailCmd.Flags().StringP("password", "", "", "SMTP password (use app password for Gmail)")
	testEmailCmd.Flags().StringP("from", "", "", "From email address")
	testEmailCmd.Flags().StringP("from-name", "", "Peep Test", "From display name")
	testEmailCmd.Flags().StringP("to", "", "", "Recipient email address")

	// Add shell test flags
	testShellCmd.Flags().StringP("args", "", "", "Arguments to pass to script (space-separated)")
	testShellCmd.Flags().StringP("timeout", "", "30s", "Script execution timeout (e.g., 30s, 1m)")
	testShellCmd.Flags().StringP("working-dir", "", "", "Working directory for script execution")
	testShellCmd.Flags().StringP("env", "", "", "Environment variables (comma-separated KEY=VALUE pairs)")

	testCmd.AddCommand(testSlackCmd)
	testCmd.AddCommand(testDesktopCmd)
	testCmd.AddCommand(testEmailCmd)
	testCmd.AddCommand(testShellCmd)
}
