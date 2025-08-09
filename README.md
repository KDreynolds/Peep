# 🔍 Peep - Observability for Humans

*One binary. No ## 🔔 Notification Channels

```bash
# Desktop notifications (built-in, cross-platform)
./peep test desktop

# Slack webhooks (production tested)
./peep alerts channels add slack "Team Alerts" --webhook https://hooks.slack.com/...

# Email with SMTP (Gmail/Office365 compatible)
./peep alerts channels add email "Critical Alerts" --smtp-host smtp.gmail.com \
  --username user@gmail.com --password app-password \
  --from user@gmail.com --to team@company.com

# Custom shell scripts for advanced integrations
./peep alerts channels add shell "PagerDuty Integration" --script ./alert-handler.sh

# Test all channels at once
./peep test all
``` YAML cults.*

A lightweight, powerful observability tool built for developers who want to understand their logs without the enterprise complexity.

## ✨ Features

- **📊 Real-time Dashboard** - Beautiful HTMX-powered web interface
- **🚨 Smart Alerts** - SQL-based rules with timezone-aware time windows
- **🔔 Multi-channel Notifications** - Desktop, Slack, Email, and Shell integrations (all production-tested)
- **⚡ Alert Suppression** - Intelligent cooldown periods with escalation detection
- **🖥️ TUI Interface** - Terminal UI for real-time log monitoring
- **☸️ Kubernetes Integration** - Direct pod log streaming with auto-reconnection
- **� HTTP Monitoring** - 4xx/5xx error detection and 304 cache hit analysis
- **�📝 Multiple Formats** - JSON, plain text, and custom log parsing
- **🧹 Auto-retention** - Configurable log cleanup with database optimization
- **🕐 Daemon Mode** - Background monitoring with 30-second polling intervals
- **💾 SQLite Backend** - Local storage with transparent, queryable schema

## 🚀 Quick Start

```bash
# Build Peep
make build

# Ingest logs from various sources
echo '{"level":"info","message":"Hello from Peep!","service":"api"}' | ./peep
./peep ingest my-app.log
kubectl logs -f deployment/my-app | ./peep ingest  # Kubernetes integration

# Start the web dashboard
./peep web
# Visit http://localhost:8080

# Launch the TUI
./peep tui

# Set up intelligent alerts with time windows
./peep alerts add "High Errors" "SELECT COUNT(*) FROM logs WHERE level='error' AND timestamp > datetime('now', 'localtime', '-5 minutes')" --threshold 5

# Monitor HTTP status codes
./peep alerts add "4xx Errors" "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%\" 4__ %'" --threshold 10
./peep alerts add "Cache Efficiency" "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%\" 304 %'" --threshold 50

# Start daemon mode for background monitoring
./peep alerts start

# Test all notification channels
./peep test desktop
./peep test slack
./peep test email
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
- SQLite storage with optimized schema
- TUI interface with Bubble Tea
- Multi-format log parsing

✅ **Phase 2 - Intelligence (Complete)**  
- SQL-based alert engine with timezone-aware queries
- 4 notification channels (Desktop, Slack, Email, Shell) - all production tested
- Real-time alert monitoring with daemon mode
- Alert suppression with 5-minute cooldown and escalation detection
- Web dashboard with HTMX

✅ **Phase 2.5 - Production Features (Complete)**
- Kubernetes integration with real-time log streaming
- HTTP status code monitoring (4xx, 5xx, 304 cache hits)
- Auto-retention system with configurable cleanup policies
- Background daemon mode with 30-second polling
- Advanced time-window queries with proper timezone handling

🚧 **Phase 3 - Polish & Scale (In Progress)**
- Enhanced web interface (logs viewer, alert management UI)
- Performance optimization for high-volume logs
- Comprehensive documentation and deployment guides

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

# Try the comprehensive demos
./demo.sh                           # Basic log ingestion and TUI
./tui-demo.sh                       # Interactive TUI interface  
./slack-demo.sh                     # Slack webhook integration
./email-demo.sh                     # SMTP email notifications
./shell-demo.sh                     # Custom shell script alerts
./stats-demo.sh                     # Database statistics and cleanup
./retention-demo.sh                 # Auto-retention system
./comprehensive-notification-test.sh # Test all notification channels
```

## 🏗️ Architecture

- **Single Binary:** Cross-compiled Go, runs anywhere
- **SQLite Backend:** Local `logs.db` file, SQL-queryable with auto-optimization
- **Dual Interface:** TUI for monitoring, Web UI for dashboards  
- **HTMX Web:** Progressive enhancement, no complex JavaScript
- **Kubernetes Native:** Direct integration with kubectl and pod logs
- **Smart Alerting:** Timezone-aware queries with suppression and escalation
- **Plugin System:** Shell scripts for custom integrations

## 🚀 Production Features

- **🔄 Auto-Retention:** Configurable log cleanup policies with database vacuum
- **⏰ Timezone Handling:** Proper local time support for accurate time-window queries
- **🚫 Alert Suppression:** 5-minute cooldown periods prevent notification spam
- **📈 Escalation Detection:** Alerts on increasing error rates (>20% threshold growth)
- **🔌 Multi-Channel:** Simultaneous notifications across Slack, Email, Desktop
- **☸️ K8s Integration:** Real-time log streaming with automatic reconnection
- **📊 HTTP Monitoring:** Built-in 4xx/5xx error and 304 cache hit detection
- **⚙️ Daemon Mode:** Background monitoring with 30-second polling intervals

## 📚 Examples

**Kubernetes monitoring:**
```bash
# Stream logs from Kubernetes pods
kubectl logs -f deployment/web-app -n production | ./peep ingest

# Monitor HTTP errors in real-time
./peep alerts add "API 5xx Errors" \
  "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%\" 5__ %' AND timestamp > datetime('now', 'localtime', '-5 minutes')" \
  --threshold 5 --channels "slack,email"

# Track cache efficiency
./peep alerts add "Low Cache Hit Rate" \
  "SELECT COUNT(*) FROM logs WHERE raw_log LIKE '%\" 304 %' AND timestamp > datetime('now', 'localtime', '-10 minutes')" \
  --threshold 20
```

**Production alerting:**
```bash
# Database connection errors
./peep alerts add "DB Connection Failures" \
  "SELECT COUNT(*) FROM logs WHERE message LIKE '%database%' AND level='error' AND timestamp > datetime('now', 'localtime', '-2 minutes')" \
  --threshold 3 --channels "slack,email,desktop"

# Memory usage warnings
./peep alerts add "High Memory Usage" \
  "SELECT COUNT(*) FROM logs WHERE message LIKE '%memory%' AND level='warn'" \
  --threshold 10

# Start background monitoring
./peep alerts start  # Runs with 30s polling, 5min alert cooldowns
```

**Advanced usage:**
```bash
# Custom log format parsing
./peep ingest --format "{{.timestamp}} [{{.level}}] {{.service}}: {{.message}}" custom.log

# Auto-retention for log management
./peep clean --days 30 --vacuum  # Keep 30 days, optimize database

# Database statistics
./peep stats
```

See [`Roadmap.md`](Roadmap.md) for the full development plan and [`docs/`](docs/) for detailed guides.

---

*"Production-ready observability without the enterprise complexity - because monitoring should just work."*
