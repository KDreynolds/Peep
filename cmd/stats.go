package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var (
	detailed bool
	json     bool
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database and performance statistics",
	Long: `Display comprehensive statistics about the Peep database including
log counts, storage size, performance metrics, and system health.

Examples:
  peep stats                    # Basic stats
  peep stats --detailed         # Detailed breakdown by level and service
  peep stats --json             # JSON output for scripting`,
	RunE: runStats,
}

func init() {
	statsCmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed breakdown by log level and service")
	statsCmd.Flags().BoolVar(&json, "json", false, "Output stats in JSON format")
}

func runStats(cmd *cobra.Command, args []string) error {
	store, err := storage.NewStorage("logs.db")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	db := store.GetDB()

	if json {
		return printJSONStats(db)
	}

	return printHumanStats(db)
}

func printHumanStats(db *sql.DB) error {
	fmt.Println("📊 Peep Database Statistics")
	fmt.Println("========================================")

	// Database file info
	if info, err := os.Stat("logs.db"); err == nil {
		fmt.Printf("💾 Database Size: %.2f MB\n", float64(info.Size())/(1024*1024))
		fmt.Printf("📅 Last Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	}

	// Log counts
	var totalLogs int
	err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&totalLogs)
	if err != nil {
		return fmt.Errorf("failed to count logs: %w", err)
	}
	fmt.Printf("📝 Total Logs: %d\n", totalLogs)

	if totalLogs == 0 {
		fmt.Println("\n🔍 No logs found in database")
		return nil
	}

	// Time range
	var oldest, newest string
	err = db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM logs").Scan(&oldest, &newest)
	if err != nil {
		return fmt.Errorf("failed to get time range: %w", err)
	}

	if oldest != "" && newest != "" {
		oldestTime, err1 := time.Parse("2006-01-02 15:04:05-07:00", oldest)
		if err1 != nil {
			// Try without timezone
			oldestTime, err1 = time.Parse("2006-01-02 15:04:05", oldest)
		}
		if err1 != nil {
			// Try RFC3339 format
			oldestTime, err1 = time.Parse(time.RFC3339, oldest)
		}

		newestTime, err2 := time.Parse("2006-01-02 15:04:05-07:00", newest)
		if err2 != nil {
			// Try without timezone
			newestTime, err2 = time.Parse("2006-01-02 15:04:05", newest)
		}
		if err2 != nil {
			// Try RFC3339 format
			newestTime, err2 = time.Parse(time.RFC3339, newest)
		}

		fmt.Printf("⏰ Time Range: %s to %s\n", oldest, newest)

		if err1 == nil && err2 == nil {
			duration := newestTime.Sub(oldestTime)
			fmt.Printf("⏱️  Duration: %s\n", formatDuration(duration))
		}
	}

	// Log levels breakdown
	fmt.Println("\n📊 Log Levels:")
	rows, err := db.Query(`
		SELECT level, COUNT(*) as count, 
		ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM logs), 1) as percentage
		FROM logs 
		WHERE level != '' 
		GROUP BY level 
		ORDER BY count DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to get log levels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var level string
		var count int
		var percentage float64
		if err := rows.Scan(&level, &count, &percentage); err != nil {
			continue
		}
		fmt.Printf("  %s: %d (%.1f%%)\n", level, count, percentage)
	}

	// Services breakdown (if detailed)
	if detailed {
		fmt.Println("\n🔧 Services:")
		rows, err := db.Query(`
			SELECT service, COUNT(*) as count
			FROM logs 
			WHERE service != '' 
			GROUP BY service 
			ORDER BY count DESC
			LIMIT 10
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var service string
				var count int
				if err := rows.Scan(&service, &count); err != nil {
					continue
				}
				fmt.Printf("  %s: %d logs\n", service, count)
			}
		}

		// Recent activity
		fmt.Println("\n📈 Recent Activity (last 24 hours):")
		var recent24h int
		err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE timestamp > datetime('now', '-24 hours')").Scan(&recent24h)
		if err == nil {
			fmt.Printf("  Last 24h: %d logs\n", recent24h)
		}

		var recent1h int
		err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE timestamp > datetime('now', '-1 hour')").Scan(&recent1h)
		if err == nil {
			fmt.Printf("  Last 1h: %d logs\n", recent1h)
		}
	}

	// Performance info
	fmt.Println("\n⚡ Performance:")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  Memory Usage: %.2f MB\n", float64(m.Alloc)/(1024*1024))
	fmt.Printf("  Go Routines: %d\n", runtime.NumGoroutine())

	// Alert rules count
	var alertCount int
	err = db.QueryRow("SELECT COUNT(*) FROM alert_rules WHERE enabled = 1").Scan(&alertCount)
	if err == nil && alertCount > 0 {
		fmt.Printf("\n🚨 Active Alert Rules: %d\n", alertCount)
	}

	return nil
}

func printJSONStats(db *sql.DB) error {
	stats := make(map[string]interface{})

	// Database file info
	if info, err := os.Stat("logs.db"); err == nil {
		stats["database_size_bytes"] = info.Size()
		stats["database_size_mb"] = float64(info.Size()) / (1024 * 1024)
		stats["last_modified"] = info.ModTime().Unix()
	}

	// Log counts
	var totalLogs int
	if err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&totalLogs); err == nil {
		stats["total_logs"] = totalLogs
	}

	// Time range
	var oldest, newest string
	if err := db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM logs").Scan(&oldest, &newest); err == nil {
		stats["oldest_log"] = oldest
		stats["newest_log"] = newest
	}

	// Log levels
	levels := make(map[string]int)
	rows, err := db.Query("SELECT level, COUNT(*) FROM logs WHERE level != '' GROUP BY level")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var level string
			var count int
			if rows.Scan(&level, &count) == nil {
				levels[level] = count
			}
		}
		stats["levels"] = levels
	}

	// Performance
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats["memory_usage_bytes"] = m.Alloc
	stats["memory_usage_mb"] = float64(m.Alloc) / (1024 * 1024)
	stats["goroutines"] = runtime.NumGoroutine()

	// Alert rules
	var alertCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM alert_rules WHERE enabled = 1").Scan(&alertCount); err == nil {
		stats["active_alert_rules"] = alertCount
	}

	stats["timestamp"] = time.Now().Unix()

	// Print JSON
	fmt.Printf("{\n")
	first := true
	for key, value := range stats {
		if !first {
			fmt.Printf(",\n")
		}
		first = false
		switch v := value.(type) {
		case string:
			fmt.Printf("  \"%s\": \"%s\"", key, v)
		case int:
			fmt.Printf("  \"%s\": %d", key, v)
		case int64:
			fmt.Printf("  \"%s\": %d", key, v)
		case float64:
			fmt.Printf("  \"%s\": %.2f", key, v)
		case map[string]int:
			fmt.Printf("  \"%s\": {", key)
			firstLevel := true
			for k, count := range v {
				if !firstLevel {
					fmt.Printf(", ")
				}
				firstLevel = false
				fmt.Printf("\"%s\": %d", k, count)
			}
			fmt.Printf("}")
		default:
			fmt.Printf("  \"%s\": \"%v\"", key, v)
		}
	}
	fmt.Printf("\n}\n")

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	} else {
		days := d.Hours() / 24
		return fmt.Sprintf("%.1f days", days)
	}
}
