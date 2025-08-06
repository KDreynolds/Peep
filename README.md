# 🔍 Peep - Observability for Humans

*One binary. No boilerplate. No YAML cults.*

A lightweight, powerful observability tool built for developers who want to understand their logs without the enterprise complexity.

## ✨ Features

- **📊 Real-time Dashboard** - Beautiful HTMX-powered web interface
- **🚨 Smart Alerts** - SQL-based rules with multiple notification channels
- **🖥️ TUI Interface** - Terminal UI for real-time log monitoring
- **📝 Multiple Formats** - JSON, plain text, and custom log parsing
- **🔔 Notifications** - Desktop, Slack, Email, and Shell script integrations
- **💾 SQLite Backend** - Local storage with transparent, queryable schema

## 🚀 Quick Start

```bash
# Build Peep
make build

# Ingest some logs
echo '{"level":"info","message":"Hello from Peep!","service":"api"}' | ./peep
./peep ingest my-app.log

# Start the web dashboard
./peep web
# Visit http://localhost:8080

# Launch the TUI
./peep tui

# Set up alerts
./peep alerts add "High Errors" "SELECT COUNT(*) FROM logs WHERE level='error'" --threshold 5

# Test notifications
./peep test desktop
```

## � Notification Channels

```bash
# Desktop notifications (built-in)
./peep test desktop

# Slack webhooks
./peep alerts channels add slack "Team Alerts" --webhook https://hooks.slack.com/...

# Email (SMTP)
./peep alerts channels add email "Alerts" --smtp-host smtp.gmail.com --username user@gmail.com --password app-password --from user@gmail.com --to team@company.com

# Custom shell scripts
./peep alerts channels add shell "Custom Handler" --script ./alert-handler.sh
```

## 🎯 Current Status

✅ **Phase 1 - Foundation (Complete)**
- CLI framework and log ingestion
- SQLite storage with schema
- TUI interface with Bubble Tea
- Multi-format log parsing

✅ **Phase 2 - Intelligence (Complete)**  
- SQL-based alert engine
- 4 notification channels (Desktop, Slack, Email, Shell)
- Real-time alert monitoring
- Web dashboard with HTMX

🚧 **Phase 3 - In Progress**
- Enhanced web interface (logs viewer, alert management)
- Performance optimization
- Documentation and examples

## 🛠️ Development

```bash
# Install dependencies
make deps

# Build and run
make run

# Run tests  
make test

# Development mode (auto-rebuild on changes)
make dev

# Try the demos
./demo.sh         # Basic log ingestion
./tui-demo.sh     # TUI interface  
./slack-demo.sh   # Slack notifications
./email-demo.sh   # Email alerts
./shell-demo.sh   # Custom shell scripts
```

## 🏗️ Architecture

- **Single Binary:** Cross-compiled Go, runs anywhere
- **SQLite Backend:** Local `logs.db` file, SQL-queryable
- **Dual Interface:** TUI for monitoring, Web UI for dashboards  
- **HTMX Web:** Progressive enhancement, no complex JavaScript
- **Plugin System:** Shell scripts for custom integrations

## 📚 Examples

**Simple monitoring:**
```bash
# Watch logs in real-time
tail -f app.log | ./peep

# Set up error alerting
./peep alerts add "API Errors" "SELECT COUNT(*) FROM logs WHERE service='api' AND level='error' AND timestamp > datetime('now', '-5 minutes')" --threshold 3
```

**Advanced usage:**
```bash
# Custom log format
./peep ingest --format "{{.timestamp}} [{{.level}}] {{.service}}: {{.message}}" custom.log

# Multi-channel alerts
./peep alerts add "Critical Errors" "SELECT COUNT(*) FROM logs WHERE level='error' AND message LIKE '%database%'" --threshold 1 --channels "slack,email,desktop"
```

See [`Roadmap.md`](Roadmap.md) for the full development plan and [`docs/`](docs/) for detailed guides.

---

*"Observability for the 99% - because not everyone needs Kubernetes."*
