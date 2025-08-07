package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var (
	olderThan   string
	keepLast    int
	cleanLevels []string
	cleanAll    bool
	dryRun      bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old logs to manage database size",
	Long: `Remove old logs from the database to prevent unlimited growth.
	
Examples:
  peep clean --older-than 7d           # Delete logs older than 7 days
  peep clean --keep-last 1000          # Keep only the 1000 most recent logs
  peep clean --levels info,debug       # Delete logs with specific levels
  peep clean --all                     # Delete all logs (with confirmation)
  peep clean --older-than 30d --dry-run  # Show what would be deleted`,
	RunE: runClean,
}

func init() {
	cleanCmd.Flags().StringVar(&olderThan, "older-than", "", "Delete logs older than duration (e.g., 7d, 24h, 30m)")
	cleanCmd.Flags().IntVar(&keepLast, "keep-last", 0, "Keep only the N most recent logs")
	cleanCmd.Flags().StringSliceVar(&cleanLevels, "levels", []string{}, "Delete logs with specific levels (comma-separated)")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Delete all logs (requires confirmation)")
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")
}

func runClean(cmd *cobra.Command, args []string) error {
	store, err := storage.NewStorage("logs.db")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Get the database handle (we'll need to add a method for this)
	db := store.GetDB()

	// Count total logs before cleanup
	var totalBefore int
	err = db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&totalBefore)
	if err != nil {
		return fmt.Errorf("failed to count logs: %w", err)
	}

	if totalBefore == 0 {
		fmt.Println("📭 No logs found in database")
		return nil
	}

	fmt.Printf("📊 Found %d logs in database\n", totalBefore)

	var deleted int

	// Handle different cleanup modes
	if cleanAll {
		deleted, err = cleanAllLogs(db)
	} else if olderThan != "" {
		deleted, err = cleanOlderThan(db, olderThan)
	} else if keepLast > 0 {
		deleted, err = cleanKeepLast(db, keepLast)
	} else if len(cleanLevels) > 0 {
		deleted, err = cleanByLevels(db, cleanLevels)
	} else {
		return fmt.Errorf("please specify a cleanup mode: --older-than, --keep-last, --levels, or --all")
	}

	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("🔍 [DRY RUN] Would delete %d logs\n", deleted)
		fmt.Printf("📊 Would keep %d logs\n", totalBefore-deleted)
	} else {
		fmt.Printf("🗑️  Deleted %d logs\n", deleted)
		fmt.Printf("📊 %d logs remaining\n", totalBefore-deleted)

		// Vacuum the database to reclaim space
		fmt.Println("🧹 Optimizing database...")
		_, err = db.Exec("VACUUM")
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to vacuum database: %v\n", err)
		} else {
			fmt.Println("✅ Database optimized")
		}
	}

	return nil
}

func cleanAllLogs(db *sql.DB) (int, error) {
	if !dryRun {
		fmt.Print("⚠️  This will delete ALL logs. Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("❌ Cancelled")
			return 0, nil
		}
	}

	if dryRun {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
		return count, err
	}

	result, err := db.Exec("DELETE FROM logs")
	if err != nil {
		return 0, fmt.Errorf("failed to delete logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

func cleanOlderThan(db *sql.DB, duration string) (int, error) {
	// Parse duration
	dur, err := parseDuration(duration)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	cutoff := time.Now().Add(-dur)
	cutoffStr := cutoff.Format("2006-01-02 15:04:05")

	if dryRun {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM logs WHERE timestamp < ?", cutoffStr).Scan(&count)
		return count, err
	}

	result, err := db.Exec("DELETE FROM logs WHERE timestamp < ?", cutoffStr)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

func cleanKeepLast(db *sql.DB, keep int) (int, error) {
	if dryRun {
		var total int
		err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&total)
		if err != nil {
			return 0, err
		}
		if total <= keep {
			return 0, nil
		}
		return total - keep, nil
	}

	result, err := db.Exec(`
		DELETE FROM logs 
		WHERE id NOT IN (
			SELECT id FROM logs 
			ORDER BY timestamp DESC 
			LIMIT ?
		)`, keep)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

func cleanByLevels(db *sql.DB, levels []string) (int, error) {
	// Build the WHERE clause for levels
	placeholders := make([]string, len(levels))
	args := make([]interface{}, len(levels))
	for i, level := range levels {
		placeholders[i] = "?"
		args[i] = level
	}
	whereClause := fmt.Sprintf("level IN (%s)", strings.Join(placeholders, ","))

	if dryRun {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM logs WHERE %s", whereClause)
		err := db.QueryRow(query, args...).Scan(&count)
		return count, err
	}

	query := fmt.Sprintf("DELETE FROM logs WHERE %s", whereClause)
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete logs by level: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

func parseDuration(s string) (time.Duration, error) {
	// Handle common duration formats: 7d, 24h, 30m, 60s
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// For other formats, use standard time.ParseDuration
	return time.ParseDuration(s)
}
