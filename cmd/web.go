package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface on localhost:8080",
	Long: `Start the minimal web interface for browsing logs and creating dashboards.
Access it at http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🌐 Starting Peep Web Interface...")
		fmt.Println("📍 Will be available at: http://localhost:8080")
		fmt.Println("🚧 Web interface coming soon!")
		fmt.Println("This will include:")
		fmt.Println("  • Log viewer and search")
		fmt.Println("  • Simple dashboard creation")
		fmt.Println("  • Alert management UI")
	},
}
