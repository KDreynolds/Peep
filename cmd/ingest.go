package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kylereynolds/peep/internal/ingestion"
	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var (
	excludeLevels   []string
	includeLevels   []string
	excludePatterns []string
	includePatterns []string
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [file]",
	Short: "Ingest logs from a file or stdin",
	Long: `Ingest logs from a file or stdin and store them in the SQLite database.
	
Examples:
  peep ingest app.log                              # Ingest from file
  docker logs myapp | peep                         # Ingest from stdin
  tail -f app.log | peep                           # Real-time ingestion
  docker logs myapp | peep --exclude-levels info,debug  # Skip noisy logs
  kubectl logs pod | peep --exclude-patterns "health.*check"`,
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
			filteredCount := 0
			for scanner.Scan() {
				line := scanner.Text()
				entry := parser.ParseLine(line)

				// Apply filtering
				if shouldSkipLog(entry, line) {
					filteredCount++
					continue
				}

				if err := store.InsertLog(entry); err != nil {
					fmt.Printf("❌ Error storing log: %v\n", err)
					continue
				}

				fmt.Printf("📝 [%d] %s | %s | %s\n", lineCount, entry.Level, entry.Service, entry.Message)
				lineCount++
			}
			fmt.Printf("✅ Processed %d log lines", lineCount)
			if filteredCount > 0 {
				fmt.Printf(" (filtered %d)", filteredCount)
			}
			fmt.Println()
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
			filteredCount := 0
			for scanner.Scan() {
				line := scanner.Text()
				entry := parser.ParseLine(line)

				// Apply filtering
				if shouldSkipLog(entry, line) {
					filteredCount++
					continue
				}

				if err := store.InsertLog(entry); err != nil {
					fmt.Printf("❌ Error storing log: %v\n", err)
					continue
				}

				fmt.Printf("📝 [%d] %s | %s | %s\n", lineCount, entry.Level, entry.Service, entry.Message)
				lineCount++
			}
			fmt.Printf("✅ Processed %d log lines from %s", lineCount, filename)
			if filteredCount > 0 {
				fmt.Printf(" (filtered %d)", filteredCount)
			}
			fmt.Println()
		}
	},
}

func shouldSkipLog(entry storage.LogEntry, rawLine string) bool {
	// Check exclude levels
	if len(excludeLevels) > 0 {
		for _, level := range excludeLevels {
			if strings.EqualFold(entry.Level, level) {
				return true
			}
		}
	}

	// Check include levels (if specified, only allow these levels)
	if len(includeLevels) > 0 {
		found := false
		for _, level := range includeLevels {
			if strings.EqualFold(entry.Level, level) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check exclude patterns
	for _, pattern := range excludePatterns {
		if matched, _ := regexp.MatchString(pattern, rawLine); matched {
			return true
		}
	}

	// Check include patterns (if specified, only allow lines matching these patterns)
	if len(includePatterns) > 0 {
		found := false
		for _, pattern := range includePatterns {
			if matched, _ := regexp.MatchString(pattern, rawLine); matched {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

func init() {
	ingestCmd.Flags().StringSliceVar(&excludeLevels, "exclude-levels", []string{}, "Skip logs with these levels (comma-separated)")
	ingestCmd.Flags().StringSliceVar(&includeLevels, "include-levels", []string{}, "Only process logs with these levels (comma-separated)")
	ingestCmd.Flags().StringSliceVar(&excludePatterns, "exclude-patterns", []string{}, "Skip logs matching these regex patterns (comma-separated)")
	ingestCmd.Flags().StringSliceVar(&includePatterns, "include-patterns", []string{}, "Only process logs matching these regex patterns (comma-separated)")
}
