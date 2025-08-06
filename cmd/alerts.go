package cmd

import (
	"fmt"

	"github.com/kylereynolds/peep/internal/alerts"
	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage alert rules and notifications",
	Long: `Create, list, and manage SQL-based alert rules that monitor your logs.
	
Examples:
  peep alerts list                           # List all alert rules
  peep alerts add "High Errors" "SELECT COUNT(*) FROM logs WHERE level='error'" --threshold 5 --window 5m
  peep alerts channels list                  # List notification channels
  peep alerts channels add desktop "Desktop Notifications"
  peep alerts channels add email "Team Alerts" --smtp-host smtp.gmail.com --username user@gmail.com --password app-password --from user@gmail.com --to team@company.com`,
}

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all alert rules",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		rules := engine.GetRules()
		if len(rules) == 0 {
			fmt.Println("📭 No alert rules configured.")
			fmt.Println("💡 Add one with: peep alerts add \"Rule Name\" \"SELECT COUNT(*) FROM logs WHERE level='error'\"")
			return
		}

		fmt.Printf("🚨 Alert Rules (%d):\n\n", len(rules))
		for _, rule := range rules {
			status := "🔴 Disabled"
			if rule.Enabled {
				status = "🟢 Enabled"
			}

			fmt.Printf("%s %s\n", status, rule.Name)
			fmt.Printf("   Query: %s\n", rule.Query)
			fmt.Printf("   Threshold: %d in %s\n", rule.Threshold, rule.Window)
			if !rule.LastCheck.IsZero() {
				fmt.Printf("   Last Check: %s\n", rule.LastCheck.Format("2006-01-02 15:04:05"))
			}
			if !rule.LastAlert.IsZero() {
				fmt.Printf("   Last Alert: %s\n", rule.LastAlert.Format("2006-01-02 15:04:05"))
			}
			fmt.Println()
		}
	},
}

var alertsAddCmd = &cobra.Command{
	Use:   "add [name] [query]",
	Short: "Add a new alert rule",
	Long: `Add a new SQL-based alert rule.

The query should return a count that will be compared against the threshold.

Examples:
  peep alerts add "High Errors" "SELECT COUNT(*) FROM logs WHERE level='error'"
  peep alerts add "DB Issues" "SELECT COUNT(*) FROM logs WHERE service='db' AND level='error'"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		query := args[1]

		threshold, _ := cmd.Flags().GetInt("threshold")
		window, _ := cmd.Flags().GetString("window")
		description, _ := cmd.Flags().GetString("description")

		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		rule := &alerts.AlertRule{
			Name:        name,
			Description: description,
			Query:       query,
			Threshold:   threshold,
			Window:      window,
			Enabled:     true,
		}

		if err := engine.AddRule(rule); err != nil {
			fmt.Printf("❌ Error adding alert rule: %v\n", err)
			return
		}

		fmt.Printf("✅ Alert rule '%s' added successfully!\n", name)
		fmt.Printf("   Query: %s\n", query)
		fmt.Printf("   Threshold: %d events in %s\n", threshold, window)
	},
}

var alertsChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage notification channels",
}

var alertsChannelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification channels",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		channels := engine.GetChannels()
		if len(channels) == 0 {
			fmt.Println("� No notification channels configured.")
			fmt.Println("� Add one with: peep alerts channels add slack \"Team Alerts\" --webhook https://hooks.slack.com/...")
			return
		}

		fmt.Printf("📢 Notification Channels (%d):\n\n", len(channels))
		for _, channel := range channels {
			status := "🔴 Disabled"
			if channel.Enabled {
				status = "🟢 Enabled"
			}

			icon := getChannelIcon(channel.Type)
			fmt.Printf("%s %s %s (%s)\n", status, icon, channel.Name, channel.Type)

			// Show relevant config (without sensitive data)
			if channel.Type == "slack" {
				if webhookURL, exists := channel.Config["webhook_url"]; exists && webhookURL != "" {
					// Mask webhook URL for security
					maskedURL := maskWebhookURL(webhookURL)
					fmt.Printf("   Webhook: %s\n", maskedURL)
				}
			}
			fmt.Println()
		}
	},
}

var alertsChannelsAddCmd = &cobra.Command{
	Use:   "add [type] [name]",
	Short: "Add a notification channel",
	Long: `Add a notification channel for alerts.

Supported types:
  desktop - Desktop notifications
  slack   - Slack webhook (requires --webhook flag)
  email   - Email notifications (requires SMTP config)
  shell   - Execute shell script (requires script path)

Examples:
  peep alerts channels add slack "Team Alerts" --webhook https://hooks.slack.com/services/...
  peep alerts channels add desktop "Local Notifications"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		channelType := args[0]
		name := args[1]

		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		config := make(map[string]string)

		// Handle type-specific configuration
		switch channelType {
		case "slack":
			webhook, _ := cmd.Flags().GetString("webhook")
			if webhook == "" {
				fmt.Println("❌ Slack channels require a webhook URL")
				fmt.Println("💡 Use: --webhook https://hooks.slack.com/services/...")
				return
			}
			config["webhook_url"] = webhook

		case "desktop":
			// Desktop notifications don't need additional config

		case "email":
			// Get email configuration from flags
			smtpHost, _ := cmd.Flags().GetString("smtp-host")
			smtpPort, _ := cmd.Flags().GetString("smtp-port")
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			fromEmail, _ := cmd.Flags().GetString("from")
			fromName, _ := cmd.Flags().GetString("from-name")
			toEmails, _ := cmd.Flags().GetString("to")

			if smtpHost == "" || username == "" || password == "" || fromEmail == "" || toEmails == "" {
				fmt.Println("❌ Email channels require SMTP configuration")
				fmt.Println("💡 Required flags: --smtp-host, --username, --password, --from, --to")
				fmt.Println("💡 Example: peep alerts channels add email \"Team Alerts\" \\")
				fmt.Println("    --smtp-host smtp.gmail.com --smtp-port 587 \\")
				fmt.Println("    --username your-email@gmail.com --password your-app-password \\")
				fmt.Println("    --from your-email@gmail.com --from-name \"Peep Alerts\" \\")
				fmt.Println("    --to team@company.com,admin@company.com")
				return
			}

			config["smtp_host"] = smtpHost
			config["smtp_port"] = smtpPort
			config["username"] = username
			config["password"] = password
			config["from_email"] = fromEmail
			config["from_name"] = fromName
			config["to_emails"] = toEmails

		case "shell":
			fmt.Println("🚧 Shell script notifications coming soon!")
			return

		default:
			fmt.Printf("❌ Unknown channel type: %s\n", channelType)
			fmt.Println("💡 Supported types: slack, desktop, email, shell")
			return
		}

		channel := &alerts.NotificationChannel{
			Name:    name,
			Type:    channelType,
			Config:  config,
			Enabled: true,
		}

		if err := engine.AddNotificationChannel(channel); err != nil {
			fmt.Printf("❌ Error adding notification channel: %v\n", err)
			return
		}

		icon := getChannelIcon(channelType)
		fmt.Printf("✅ %s %s channel '%s' added successfully!\n", icon, channelType, name)

		if channelType == "slack" {
			fmt.Println("� Test it with: peep alerts start")
		}
	},
}

// getChannelIcon returns an icon for the channel type
func getChannelIcon(channelType string) string {
	switch channelType {
	case "slack":
		return "💬"
	case "desktop":
		return "🖥️"
	case "email":
		return "📧"
	case "shell":
		return "🖥️"
	default:
		return "📢"
	}
}

// maskWebhookURL masks sensitive parts of webhook URLs
func maskWebhookURL(url string) string {
	if len(url) < 20 {
		return "***"
	}

	// Show first 30 chars and last 10 chars, mask the middle
	if len(url) > 40 {
		return url[:30] + "..." + url[len(url)-10:]
	}

	return url[:len(url)/2] + "***"
}

var alertsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the alert monitoring daemon",
	Long: `Start monitoring your logs for alert conditions in the background.
	
This will continuously check your alert rules and send notifications when thresholds are exceeded.`,
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		rules := engine.GetRules()
		enabledRules := 0
		for _, rule := range rules {
			if rule.Enabled {
				enabledRules++
			}
		}

		if enabledRules == 0 {
			fmt.Println("⚠️  No enabled alert rules found!")
			fmt.Println("💡 Add some rules first:")
			fmt.Println("   peep alerts add \"High Errors\" \"SELECT COUNT(*) FROM logs WHERE level='error'\"")
			return
		}

		fmt.Printf("🚨 Starting alert monitoring with %d enabled rules...\n", enabledRules)
		fmt.Println("📊 Checking every 30 seconds")
		fmt.Println("Press Ctrl+C to stop")

		engine.Start()
		defer engine.Stop()

		// Keep running until interrupted
		select {}
	},
}

func init() {
	// Add flags to the add command
	alertsAddCmd.Flags().IntP("threshold", "t", 1, "Alert threshold (number of matching events)")
	alertsAddCmd.Flags().StringP("window", "w", "5m", "Time window (e.g., 5m, 1h, 30s)")
	alertsAddCmd.Flags().StringP("description", "d", "", "Alert rule description")

	// Add flags to the channels add command
	alertsChannelsAddCmd.Flags().StringP("webhook", "", "", "Slack webhook URL (required for slack channels)")

	// Email notification flags
	alertsChannelsAddCmd.Flags().StringP("smtp-host", "", "", "SMTP server hostname (e.g., smtp.gmail.com)")
	alertsChannelsAddCmd.Flags().StringP("smtp-port", "", "587", "SMTP server port (default: 587)")
	alertsChannelsAddCmd.Flags().StringP("username", "", "", "SMTP username/email")
	alertsChannelsAddCmd.Flags().StringP("password", "", "", "SMTP password (use app password for Gmail)")
	alertsChannelsAddCmd.Flags().StringP("from", "", "", "From email address")
	alertsChannelsAddCmd.Flags().StringP("from-name", "", "Peep Alerts", "From display name")
	alertsChannelsAddCmd.Flags().StringP("to", "", "", "Recipient email addresses (comma-separated)")

	// Build command hierarchy
	alertsChannelsCmd.AddCommand(alertsChannelsListCmd)
	alertsChannelsCmd.AddCommand(alertsChannelsAddCmd)

	alertsCmd.AddCommand(alertsListCmd)
	alertsCmd.AddCommand(alertsAddCmd)
	alertsCmd.AddCommand(alertsChannelsCmd)
	alertsCmd.AddCommand(alertsStartCmd)
}
