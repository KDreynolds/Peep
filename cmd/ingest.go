package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kylereynolds/peep/internal/ingestion"
	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [file]",
	Short: "Ingest logs from a file or stdin",
	Long: `Ingest logs from a file or stdin and store them in the SQLite database.
	
Examples:
  peep ingest app.log           # Ingest from file
  docker logs myapp | peep      # Ingest from stdin
  tail -f app.log | peep        # Real-time ingestion`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage
		store, err := storage.NewStorage("logs.db")
		if err != nil {
			fmt.Printf("❌ Error initializing storage: %v\n", err)
			return
		}
		defer store.Close()

		parser := &ingestion.LogParser{}

		if len(args) == 0 {
			// Read from stdin
			fmt.Println("📥 Reading logs from stdin...")
			scanner := bufio.NewScanner(os.Stdin)
			lineCount := 0
			for scanner.Scan() {
				line := scanner.Text()
				entry := parser.ParseLine(line)

				if err := store.InsertLog(entry); err != nil {
					fmt.Printf("❌ Error storing log: %v\n", err)
					continue
				}

				fmt.Printf("📝 [%d] %s | %s | %s\n", lineCount, entry.Level, entry.Service, entry.Message)
				lineCount++
			}
			fmt.Printf("✅ Processed %d log lines\n", lineCount)
		} else {
			// Read from file
			filename := args[0]
			fmt.Printf("📥 Ingesting logs from %s...\n", filename)

			file, err := os.Open(filename)
			if err != nil {
				fmt.Printf("❌ Error opening file: %v\n", err)
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineCount := 0
			for scanner.Scan() {
				line := scanner.Text()
				entry := parser.ParseLine(line)

				if err := store.InsertLog(entry); err != nil {
					fmt.Printf("❌ Error storing log: %v\n", err)
					continue
				}

				fmt.Printf("📝 [%d] %s | %s | %s\n", lineCount, entry.Level, entry.Service, entry.Message)
				lineCount++
			}
			fmt.Printf("✅ Processed %d log lines from %s\n", lineCount, filename)
		}
	},
}
