package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "peep",
	Short: "Observability for humans. One binary. No boilerplate.",
	Long: `Peep is a simple, powerful observability tool that stores logs in SQLite
and provides both TUI and web interfaces for monitoring your applications.

No YAML configuration hell. No cloud vendor lock-in. Just logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if stdin has data (piped input)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin, use ingest command
			ingestCmd.Run(cmd, args)
			return
		}

		// No piped input, show help
		fmt.Println("🔍 Peep - Observability for humans")
		fmt.Println("Run 'peep --help' for available commands")
		fmt.Println("")
		fmt.Println("Quick examples:")
		fmt.Println("  echo '{\"level\":\"info\",\"message\":\"Hello!\"}' | peep")
		fmt.Println("  peep ingest app.log")
		fmt.Println("  peep list")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(alertsCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(cleanCmd)
}
