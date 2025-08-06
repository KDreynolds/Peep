package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	Context   string    `json:"context"` // JSON string
	RawLog    string    `json:"raw_log"`
	CreatedAt time.Time `json:"created_at"`
}

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (s *Storage) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		level TEXT,
		message TEXT,
		service TEXT,
		context TEXT, -- JSON
		raw_log TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) InsertLog(entry LogEntry) error {
	query := `
	INSERT INTO logs (timestamp, level, message, service, context, raw_log)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		entry.Timestamp,
		entry.Level,
		entry.Message,
		entry.Service,
		entry.Context,
		entry.RawLog,
	)

	return err
}

func (s *Storage) GetLogs(limit int) ([]LogEntry, error) {
	query := `
	SELECT id, timestamp, level, message, service, context, raw_log, created_at
	FROM logs
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Level,
			&entry.Message,
			&entry.Service,
			&entry.Context,
			&entry.RawLog,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database connection for advanced operations
func (s *Storage) GetDB() *sql.DB {
	return s.db
}
