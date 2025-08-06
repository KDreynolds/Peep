package cmd

import (
	"fmt"

	"github.com/kylereynolds/peep/internal/storage"
	"github.com/kylereynolds/peep/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the Terminal UI for browsing logs",
	Long: `Start the interactive Terminal User Interface for browsing, filtering,
and searching through your logs in real-time.

Controls:
  q, Ctrl+C  - Quit
  /          - Search mode
  r          - Manual refresh
  esc        - Cancel search
  ↑/↓        - Navigate logs
  enter      - Apply search filter`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		// Check if we have any logs
		logs, err := store.GetLogs(1)
		if err != nil {
			fmt.Printf("❌ Error checking logs: %v\n", err)
			return
		}

		if len(logs) == 0 {
			fmt.Println("📭 No logs found!")
			fmt.Println("💡 Try ingesting some logs first:")
			fmt.Println("   echo '{\"level\":\"info\",\"message\":\"Hello!\"}' | peep")
			fmt.Println("   peep ingest sample.log")
			return
		}

		fmt.Println("🖥️  Starting Peep TUI...")

		// Start the TUI
		if err := tui.Start(store); err != nil {
			fmt.Printf("❌ Error starting TUI: %v\n", err)
			return
		}
	},
}
