# Peep - Observability for Humans
*One binary. No boilerplate. No YAML cults.*

## 🎯 Project Vision

**Working Title:** `peep` (glimpse, beacon, watchdog, telltale)
**Tagline:** "Observability for humans. One binary. No boilerplate. No YAML cults."

## 🧩 Core Architecture

### 1. 🚀 Single Binary (Go)
- Cross-compiled static builds for all platforms
- Install via `curl | sh`, `brew install`, or GitHub releases
- Runs locally or in CI/CD environments

### 2. 📦 SQLite Backend
- Local storage in `logs.db` by default
- Transparent, user-editable schema
- Optional push to hosted backends (Supabase, Turso, Postgres)

### 3. 🖥 Dual Interface
- **TUI:** Real-time filtering, regex highlighting, saved searches
- **Web UI:** Minimal localhost:8080 interface for dashboards

### 4. 📉 Logs-First Approach
- Metrics derived from log entries
- Dashboards use SQLite views
- No pre-aggregated time series complexity

### 5. 🧠 Simple Alerts
- SQL-based alert definitions
- Multiple targets: Slack, email, desktop, shell scripts

## 🛣️ Development Roadmap

### Phase 1: Foundation (Weeks 1-2)
**Week 1: Core CLI & Ingestion**
- [x] Set up Go project structure
- [x] Implement SQLite schema and models
- [x] Build basic CLI with ingestion commands
  - [x] `peep ingest file.log`
  - [x] `docker logs container | peep`
- [x] Auto-detect log formats (JSON, plain text, ndjson)
- [x] Basic field extraction (timestamp, level, message, service)

**Week 2: TUI Interface**
- [x] Implement TUI using `bubbletea` or `tview`
- [x] Real-time log tailing
- [x] Filtering and search functionality
- [x] Regex highlighting
- [x] Saved search presets

### Phase 2: Intelligence (Weeks 3-4)
**Week 3: Alerts & Notifications**
- [x] SQL-based alert engine with timezone-aware queries
- [x] Alert rule configuration with time windows
- [x] Alert suppression with cooldown periods (5-min default)
- [x] Escalation detection for increasing error counts
- [x] Notification channels (all tested in production):
  - [x] Desktop notifications ✅ *Tested and working*
  - [x] Slack webhooks ✅ *Tested with real Slack workspace*
  - [x] Email notifications ✅ *SMTP integration working*
  - [x] Shell script execution ✅ *Custom integrations working*
- [x] Real-time alert monitoring with daemon mode
- [x] HTTP status code monitoring (4xx, 5xx, 304 cache hits)

**Week 4: Web Interface**
- [x] Minimal web server (localhost:8080)
- [x] Basic React/Svelte frontend
- [x] Log viewer and search
- [x] Simple dashboard creation
- [x] Alert management UI

**Week 4 Additional: Production Readiness**
- [x] HTTP error detection alerts (4xx, 5xx status codes)
- [x] HTTP 304 cache hit monitoring for performance insights
- [x] Log retention and cleanup commands with auto-vacuum
- [x] Auto-retention system with configurable policies
- [x] Ingestion filtering by log level
- [x] Database maintenance utilities
- [x] Service deployment and daemon mode ✅ *Background monitoring active*
- [x] Kubernetes integration with real-time log streaming
- [x] Stats and monitoring commands
- [x] Timezone-aware time window queries
- [x] Alert suppression and escalation detection

### Phase 3: Polish & Launch (Weeks 5-6)
**Week 5: Dogfooding & Feedback**
- [ ] Use peep on own projects
- [ ] Performance optimization
- [ ] Documentation and examples
- [ ] CI/CD integration examples
- [ ] Error handling and edge cases

**Week 6: Launch Preparation**
- [ ] Comprehensive README
- [ ] Installation scripts
- [ ] Demo videos/GIFs
- [ ] Package for multiple platforms
- [ ] Launch on HN/Product Hunt

## 🧰 Technical Specifications

### Ingestion Sources
- [x] stdin (`docker logs | peep`)
- [x] File ingestion (`peep ingest app.log`)
- [x] Kubernetes pod log streaming (`kubectl logs -f | peep ingest`)
- [x] Real-time log tailing with automatic reconnection
- [ ] HTTP endpoints for log pushing
- [ ] Directory watching
- [ ] Syslog integration

### Log Format Support
- [x] JSON logs
- [x] Plain text
- [x] NDJSON
- [ ] Common formats (Apache, Nginx, etc.)
- [ ] Custom format definitions

### Storage Schema
```sql
CREATE TABLE logs (
  id INTEGER PRIMARY KEY,
  timestamp DATETIME,
  level TEXT,
  message TEXT,
  service TEXT,
  context JSON,
  raw_log TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 🎯 MVP Use Cases

| Persona | Use Case |
|---------|----------|
| Indie Hacker | Monitor Flask app without AWS complexity |
| QA Engineer | Crash/error visibility during e2e tests |
| Consultant | Quick insight into legacy applications |
| Small Startup Dev | Replace multiple Grafana panels |

## 🏗️ Project Structure
```
peep/
├── cmd/                 # CLI commands
├── internal/
│   ├── storage/        # SQLite operations
│   ├── ingestion/      # Log parsing and ingestion
│   ├── tui/           # Terminal UI
│   ├── web/           # Web server and UI
│   ├── alerts/        # Alert engine
│   └── config/        # Configuration management
├── web/               # Frontend assets
├── docs/              # Documentation
└── scripts/           # Build and install scripts
```

## 💡 Inspiration & Anti-Goals

### Inspired By
- `lazygit` - TUI model and UX
- `logtail` - Simplicity (but lighter)
- Healthchecks.io - Focused, single-purpose
- Woodpecker CI - Single binary utility

### Anti-Goals
- No dashboard hell like Grafana
- No YAML configuration complexity
- No vendor lock-in
- No cloud-first assumptions

## 🪜 Future Enhancements

### Phase 4: Advanced Features
- [ ] Templated dashboards (error rates, latency percentiles)
- [ ] GitHub CI/CD integration
- [ ] SQLite WAL mode for performance
- [ ] Plugin system (WASM-based)
- [ ] Log retention management and pruning
- [ ] Log level filtering during ingestion
- [ ] Performance optimization for large datasets
- [ ] Service deployment (systemd, Docker, K8s)
- [ ] Daemon mode with auto-restart and health checks
- [ ] Resource monitoring and limits

### Phase 5: Hosted Option
- [ ] Peep Cloud (freemium model)
- [ ] Multi-user support
- [ ] Team collaboration features
- [ ] Advanced analytics

## 🚀 Getting Started

1. **Prerequisites:** Go 1.21+, SQLite
2. **Clone and build:** `make build`
3. **Try it:** `echo "hello world" | ./peep`
4. **Start TUI:** `./peep tui`
5. **Web interface:** `./peep web`

---

*"Grug-brained compatible: read logs, see logs, click logs."*