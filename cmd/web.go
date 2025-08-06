package cmd

import (
	"fmt"
	"log"

	"github.com/kylereynolds/peep/internal/alerts"
	"github.com/kylereynolds/peep/internal/storage"
	"github.com/kylereynolds/peep/internal/web"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface on localhost:8080",
	Long: `Start the web interface for browsing logs, managing alerts, and viewing dashboards.
	
Features:
  • Real-time dashboard with log statistics
  • Log viewer and search interface  
  • Alert rules and notification management
  • HTMX-powered interactivity
  
Access it at http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		// Initialize storage
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		// Initialize alert engine
		engine, err := alerts.NewEngine(store)
		if err != nil {
			fmt.Printf("❌ Error initializing alert engine: %v\n", err)
			return
		}

		// Create and start web server
		server := web.NewServer(store, engine)
		if err := server.Start(port); err != nil {
			log.Fatal("❌ Failed to start web server:", err)
		}
	},
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
}
