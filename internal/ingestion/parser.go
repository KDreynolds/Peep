package ingestion

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/kylereynolds/peep/internal/storage"
)

// LogParser handles parsing different log formats
type LogParser struct{}

// ParseLine attempts to parse a log line and extract structured information
func (p *LogParser) ParseLine(line string) storage.LogEntry {
	// Try JSON first
	if entry := p.tryParseJSON(line); entry != nil {
		return *entry
	}

	// Try common log patterns
	if entry := p.tryParseCommonFormat(line); entry != nil {
		return *entry
	}

	// Fallback to plain text
	return storage.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   line,
		Service:   "unknown",
		Context:   "{}",
		RawLog:    line,
	}
}

func (p *LogParser) tryParseJSON(line string) *storage.LogEntry {
	var jsonLog map[string]interface{}
	if err := json.Unmarshal([]byte(line), &jsonLog); err != nil {
		return nil
	}

	entry := storage.LogEntry{
		RawLog: line,
	}

	// Extract timestamp
	if ts, ok := jsonLog["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
	} else if ts, ok := jsonLog["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
	} else {
		entry.Timestamp = time.Now()
	}

	// Extract level
	if level, ok := jsonLog["level"].(string); ok {
		entry.Level = level
	} else if level, ok := jsonLog["severity"].(string); ok {
		entry.Level = level
	} else {
		entry.Level = "info"
	}

	// Extract message
	if msg, ok := jsonLog["message"].(string); ok {
		entry.Message = msg
	} else if msg, ok := jsonLog["msg"].(string); ok {
		entry.Message = msg
	} else {
		entry.Message = line
	}

	// Extract service
	if svc, ok := jsonLog["service"].(string); ok {
		entry.Service = svc
	} else if svc, ok := jsonLog["app"].(string); ok {
		entry.Service = svc
	} else {
		entry.Service = "unknown"
	}

	// Store full context as JSON
	if contextBytes, err := json.Marshal(jsonLog); err == nil {
		entry.Context = string(contextBytes)
	} else {
		entry.Context = "{}"
	}

	return &entry
}

func (p *LogParser) tryParseCommonFormat(line string) *storage.LogEntry {
	// Common patterns like: "2023-08-06 10:30:45 INFO [service] message"
	patterns := []struct {
		regex *regexp.Regexp
		parse func([]string) *storage.LogEntry
	}{
		{
			// ISO timestamp with level and optional service
			regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)\s+(\w+)\s+(?:\[([^\]]+)\])?\s*(.*)$`),
			func(matches []string) *storage.LogEntry {
				timestamp, _ := time.Parse("2006-01-02T15:04:05", strings.Replace(matches[1], " ", "T", 1))
				service := "unknown"
				if matches[3] != "" {
					service = matches[3]
				}
				return &storage.LogEntry{
					Timestamp: timestamp,
					Level:     strings.ToLower(matches[2]),
					Message:   matches[4],
					Service:   service,
					Context:   "{}",
					RawLog:    line,
				}
			},
		},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(line); matches != nil {
			return pattern.parse(matches)
		}
	}

	return nil
}
