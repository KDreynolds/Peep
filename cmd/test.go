package cmd

import (
	"fmt"

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

func init() {
	testCmd.AddCommand(testSlackCmd)
	testCmd.AddCommand(testDesktopCmd)
}
