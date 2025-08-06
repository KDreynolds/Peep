package cmd

import (
	"fmt"
	"strings"

	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent logs from the database",
	Long:  `Display the most recent logs stored in the SQLite database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		limit, _ := cmd.Flags().GetInt("limit")
		logs, err := store.GetLogs(limit)
		if err != nil {
			fmt.Printf("❌ Error retrieving logs: %v\n", err)
			return
		}

		if len(logs) == 0 {
			fmt.Println("📭 No logs found. Try ingesting some logs first!")
			fmt.Println("Example: echo '{\"level\":\"info\",\"message\":\"Hello!\"}' | peep")
			return
		}

		fmt.Printf("📋 Recent logs (showing %d):\n\n", len(logs))

		for _, log := range logs {
			levelIcon := getLevelIcon(log.Level)
			fmt.Printf("%s %s [%s] %s\n",
				levelIcon,
				log.Timestamp.Format("15:04:05"),
				log.Service,
				log.Message,
			)
		}
	},
}

func getLevelIcon(level string) string {
	switch strings.ToLower(level) {
	case "error", "err":
		return "🔴"
	case "warn", "warning":
		return "🟡"
	case "info":
		return "🔵"
	case "debug":
		return "🟣"
	default:
		return "⚪"
	}
}

func init() {
	listCmd.Flags().IntP("limit", "l", 20, "Number of recent logs to display")
}
