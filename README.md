# 🔍 Peep - Observability for Humans

*One binary. No boilerplate. No YAML cults.*

## Quick Start

```bash
# Install dependencies and build
make build

# Try it out with some logs
echo '{"level":"info","message":"Hello from Peep!"}' | ./peep

# Or ingest from a file
./peep ingest my-app.log

# Start the TUI (coming soon)
./peep tui

# Start the web interface (coming soon)  
./peep web
```

## Current Status: 🚧 Foundation Phase

✅ **Completed:**
- Basic Go project structure
- CLI framework with cobra
- SQLite storage layer
- Log parsing (JSON and common formats)
- Basic ingestion from stdin and files

🚧 **In Progress:**
- Integrating parser with storage
- TUI interface with bubbletea
- Web interface

📋 **Next Up:**
- Real-time log tailing
- Filtering and search
- Alert system

## Development

```bash
# Install dependencies
make deps

# Build and run
make run

# Run tests
make test

# Development mode (auto-rebuild on changes)
make dev

# Quick demo
make demo
```

## Architecture

- **Single Binary:** Written in Go, cross-compiled for all platforms
- **SQLite Backend:** Local storage in `logs.db`, transparent schema
- **Dual Interface:** TUI for real-time monitoring, Web UI for dashboards
- **Logs-First:** Metrics derived from log entries, no complex TSDB

See [`Roadmap.md`](Roadmap.md) for the full development plan.

---

*"Grug-brained compatible: read logs, see logs, click logs."*
