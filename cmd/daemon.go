package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kylereynolds/peep/internal/storage"
	"github.com/spf13/cobra"
)

var (
	maxLogs     int
	maxAgeDays  int
	maxSizeMB   float64
	checkMins   int
	disableAuto bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run Peep in daemon mode with automatic maintenance",
	Long: `Run Peep as a background daemon with automatic log retention,
alert monitoring, and health checks. Designed for production deployment.

Examples:
  peep daemon                                    # Run with default settings
  peep daemon --max-logs 50000                  # Keep max 50k logs
  peep daemon --max-age-days 7                  # Delete logs older than 7 days
  peep daemon --max-size-mb 100                 # Cleanup when DB > 100MB
  peep daemon --check-mins 5                    # Check every 5 minutes
  peep daemon --disable-auto                    # Disable auto-cleanup`,
	RunE: runDaemon,
}

func init() {
	daemonCmd.Flags().IntVar(&maxLogs, "max-logs", 100000, "Maximum number of logs to keep (0 = unlimited)")
	daemonCmd.Flags().IntVar(&maxAgeDays, "max-age-days", 30, "Delete logs older than N days (0 = unlimited)")
	daemonCmd.Flags().Float64Var(&maxSizeMB, "max-size-mb", 500, "Trigger cleanup when database exceeds size (0 = unlimited)")
	daemonCmd.Flags().IntVar(&checkMins, "check-mins", 10, "Minutes between retention checks")
	daemonCmd.Flags().BoolVar(&disableAuto, "disable-auto", false, "Disable automatic retention cleanup")
}

func runDaemon(cmd *cobra.Command, args []string) error {
	log.Println("🚀 Starting Peep daemon...")

	// Initialize storage
	store, err := storage.NewStorage("logs.db")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Configure auto-retention if enabled
	if !disableAuto {
		config := storage.RetentionConfig{
			MaxLogs:       maxLogs,
			MaxAge:        time.Duration(maxAgeDays) * 24 * time.Hour,
			MaxSizeMB:     maxSizeMB,
			CheckInterval: time.Duration(checkMins) * time.Minute,
			Enabled:       true,
		}

		log.Printf("🧹 Configuring auto-retention:")
		log.Printf("   Max logs: %d", config.MaxLogs)
		log.Printf("   Max age: %v", config.MaxAge)
		log.Printf("   Max size: %.1f MB", config.MaxSizeMB)
		log.Printf("   Check interval: %v", config.CheckInterval)

		store.EnableAutoRetention(config)
	} else {
		log.Println("⚠️  Auto-retention disabled")
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start health monitoring
	go healthMonitor(ctx, store)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("📡 Received signal: %v", sig)
	log.Println("🛑 Shutting down gracefully...")

	cancel()

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	log.Println("✅ Daemon stopped")
	return nil
}

func healthMonitor(ctx context.Context, store *storage.Storage) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("💓 Starting health monitor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("💓 Health monitor stopping...")
			return
		case <-ticker.C:
			checkHealth(store)
		}
	}
}

func checkHealth(store *storage.Storage) {
	db := store.GetDB()

	// Check database connectivity
	if err := db.Ping(); err != nil {
		log.Printf("❌ Database health check failed: %v", err)
		return
	}

	// Get basic stats
	var logCount int
	err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&logCount)
	if err != nil {
		log.Printf("⚠️  Failed to count logs: %v", err)
		return
	}

	// Check recent activity (logs in last hour)
	var recentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE timestamp > datetime('now', '-1 hour')").Scan(&recentCount)
	if err != nil {
		log.Printf("⚠️  Failed to count recent logs: %v", err)
		return
	}

	// Check active alerts
	var alertCount int
	err = db.QueryRow("SELECT COUNT(*) FROM alert_rules WHERE enabled = 1").Scan(&alertCount)
	if err != nil {
		// Alert table might not exist yet
		alertCount = 0
	}

	log.Printf("💓 Health: %d total logs, %d in last hour, %d active alerts",
		logCount, recentCount, alertCount)

	// Trigger retention check if needed
	store.TriggerRetentionCheck()
}
