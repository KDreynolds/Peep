package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// RetentionConfig defines automatic log retention policies
type RetentionConfig struct {
	// MaxLogs - keep only the N most recent logs (0 = disabled)
	MaxLogs int `json:"max_logs"`

	// MaxAge - delete logs older than this duration (0 = disabled)
	MaxAge time.Duration `json:"max_age"`

	// MaxSizeMB - trigger cleanup when database exceeds this size (0 = disabled)
	MaxSizeMB float64 `json:"max_size_mb"`

	// CheckInterval - how often to run cleanup checks
	CheckInterval time.Duration `json:"check_interval"`

	// Enabled - whether automatic cleanup is enabled
	Enabled bool `json:"enabled"`
}

// DefaultRetentionConfig returns sensible defaults for daemon mode
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		MaxLogs:       100000,              // Keep last 100k logs
		MaxAge:        30 * 24 * time.Hour, // Delete logs older than 30 days
		MaxSizeMB:     500,                 // Cleanup when DB > 500MB
		CheckInterval: 10 * time.Minute,    // Check every 10 minutes
		Enabled:       true,
	}
}

// AutoRetentionManager handles automatic cleanup
type AutoRetentionManager struct {
	storage *Storage
	config  RetentionConfig
	ticker  *time.Ticker
	stop    chan bool
}

// NewAutoRetentionManager creates a new retention manager
func NewAutoRetentionManager(storage *Storage, config RetentionConfig) *AutoRetentionManager {
	return &AutoRetentionManager{
		storage: storage,
		config:  config,
		stop:    make(chan bool),
	}
}

// Start begins automatic retention checking
func (arm *AutoRetentionManager) Start() {
	if !arm.config.Enabled {
		return
	}

	log.Printf("🧹 Starting automatic retention manager (check every %v)", arm.config.CheckInterval)
	arm.ticker = time.NewTicker(arm.config.CheckInterval)

	go func() {
		for {
			select {
			case <-arm.ticker.C:
				arm.performCleanup()
			case <-arm.stop:
				return
			}
		}
	}()
}

// Stop stops the automatic retention manager
func (arm *AutoRetentionManager) Stop() {
	if arm.ticker != nil {
		arm.ticker.Stop()
	}
	close(arm.stop)
}

// performCleanup runs the actual cleanup logic
func (arm *AutoRetentionManager) performCleanup() {
	db := arm.storage.GetDB()

	// Check if cleanup is needed
	shouldCleanup, reason := arm.shouldCleanup(db)
	if !shouldCleanup {
		return
	}

	log.Printf("🧹 Auto-cleanup triggered: %s", reason)

	var deletedCount int
	var err error

	// Priority order: MaxLogs > MaxAge > Size-based cleanup
	if arm.config.MaxLogs > 0 {
		deletedCount, err = arm.cleanupByCount(db)
	} else if arm.config.MaxAge > 0 {
		deletedCount, err = arm.cleanupByAge(db)
	}

	if err != nil {
		log.Printf("❌ Auto-cleanup failed: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("🗑️  Auto-cleanup: removed %d logs", deletedCount)

		// Vacuum database to reclaim space
		_, err = db.Exec("VACUUM")
		if err != nil {
			log.Printf("⚠️  Warning: Failed to vacuum database: %v", err)
		} else {
			log.Printf("✅ Database optimized after cleanup")
		}
	}
}

// shouldCleanup determines if cleanup is needed
func (arm *AutoRetentionManager) shouldCleanup(db *sql.DB) (bool, string) {
	// Check log count
	if arm.config.MaxLogs > 0 {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
		if err == nil && count > arm.config.MaxLogs {
			return true, fmt.Sprintf("log count (%d) exceeds limit (%d)", count, arm.config.MaxLogs)
		}
	}

	// Check database size
	if arm.config.MaxSizeMB > 0 {
		size := arm.getDatabaseSizeMB()
		if size > arm.config.MaxSizeMB {
			return true, fmt.Sprintf("database size (%.1f MB) exceeds limit (%.1f MB)", size, arm.config.MaxSizeMB)
		}
	}

	// Check age-based cleanup
	if arm.config.MaxAge > 0 {
		cutoff := time.Now().Add(-arm.config.MaxAge)
		cutoffStr := cutoff.Format("2006-01-02 15:04:05")

		var oldCount int
		err := db.QueryRow("SELECT COUNT(*) FROM logs WHERE timestamp < ?", cutoffStr).Scan(&oldCount)
		if err == nil && oldCount > 0 {
			return true, fmt.Sprintf("found %d logs older than %v", oldCount, arm.config.MaxAge)
		}
	}

	return false, ""
}

// cleanupByCount keeps only the most recent N logs
func (arm *AutoRetentionManager) cleanupByCount(db *sql.DB) (int, error) {
	result, err := db.Exec(`
		DELETE FROM logs 
		WHERE id NOT IN (
			SELECT id FROM logs 
			ORDER BY timestamp DESC 
			LIMIT ?
		)`, arm.config.MaxLogs)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup by count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// cleanupByAge removes logs older than MaxAge
func (arm *AutoRetentionManager) cleanupByAge(db *sql.DB) (int, error) {
	cutoff := time.Now().Add(-arm.config.MaxAge)
	cutoffStr := cutoff.Format("2006-01-02 15:04:05")

	result, err := db.Exec("DELETE FROM logs WHERE timestamp < ?", cutoffStr)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup by age: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// getDatabaseSizeMB returns the database file size in MB
func (arm *AutoRetentionManager) getDatabaseSizeMB() float64 {
	// For simplicity, we'll estimate based on log count
	// In production, you'd want to check actual file size
	var count int
	err := arm.storage.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		return 0
	}

	// Rough estimate: ~350 bytes per log entry
	estimatedBytes := float64(count) * 350
	return estimatedBytes / (1024 * 1024)
}

// TriggerCleanupIfNeeded can be called during ingestion to check if cleanup is needed
func (arm *AutoRetentionManager) TriggerCleanupIfNeeded() {
	if !arm.config.Enabled {
		return
	}

	db := arm.storage.GetDB()
	shouldCleanup, reason := arm.shouldCleanup(db)
	if shouldCleanup {
		log.Printf("🧹 Triggering immediate cleanup: %s", reason)
		arm.performCleanup()
	}
}
