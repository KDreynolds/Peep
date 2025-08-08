# Peep TODO List

## High Priority - Production Readiness

### 🗂️ Log Management & Retention
- [ ] **Log Retention Commands**
  - `peep clean --older-than 7d` (delete logs older than 7 days)
  - `peep clean --keep-last 1000` (keep only last N logs)
  - `peep clean --all` (purge all logs)
  - `peep clean --level info` (delete specific log levels)

- [ ] **Automatic Retention Policies**
  - Config option for max log age
  - Config option for max log count
  - Background cleanup during ingestion

### 🎛️ Ingestion Filtering
- [ ] **Log Level Filtering**
  - `--exclude-levels info,debug` flag for ingest command
  - `--include-levels error,warn` flag for ingest command
  - Config file support for default exclusions

- [ ] **Pattern-Based Filtering**
  - `--exclude-pattern "health.*check"` for noisy patterns
  - `--include-pattern "error|exception"` for important patterns

### 📊 Performance & Maintenance
- [ ] **Database Optimization**
  - SQLite VACUUM command integration
  - Index optimization for large datasets
  - WAL mode for concurrent access

- [x] **Monitoring & Stats**
  - `peep stats` command (log count, size, oldest/newest)
  - Performance metrics for ingestion rate
  - Memory usage monitoring

### 🚀 Service Deployment & Operations
- [ ] **Daemon Mode**
  - Background service mode (`peep daemon`)
  - Auto-restart on crashes with exponential backoff
  - Graceful shutdown handling (SIGTERM/SIGINT)
  - Health check endpoints

- [ ] **Resource Management**
  - Memory usage limits and monitoring
  - CPU usage monitoring and throttling
  - Disk space monitoring and alerts
  - Log rotation and archiving

- [ ] **Deployment Packaging**
  - Systemd service files
  - Docker containers and docker-compose
  - Kubernetes manifests and Helm charts
  - Process supervision (supervisor, pm2)

## Current Sprint - HTTP Error Detection

### 🚨 HTTP Error Alerts
- [x] Create HTTP error detection alert rule
- [ ] Test with production 4xx/5xx scenarios
- [ ] Add HTTP status code parsing improvements
- [ ] Create dashboard for HTTP error rates

### 🧪 Production Testing
- [x] Multi-pod real-time ingestion (5,467+ logs)
- [x] Streaming service health monitoring
- [x] Web-map HTTP access log monitoring
- [ ] Create comprehensive alert rule set
- [ ] Test notification channels with production data

## Future Enhancements

### 🔧 CLI Improvements
- [ ] Better progress indicators for large ingestions
- [ ] Colored output for different log levels
- [ ] Interactive mode for common tasks

### 🌐 Web Interface
- [ ] Real-time log streaming in browser
- [ ] Advanced filtering UI
- [ ] Export capabilities (CSV, JSON)

### 🔗 Integrations
- [ ] Kubernetes integration (`kubectl logs` wrapper)
- [ ] Docker Compose integration
- [ ] CI/CD pipeline examples

---

*Last Updated: August 7, 2025*
