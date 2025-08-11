package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/kylereynolds/peep/internal/alerts"
	"github.com/kylereynolds/peep/internal/storage"
)

type Server struct {
	storage *storage.Storage
	engine  *alerts.Engine
}

type PageData struct {
	Title   string
	Content interface{}
}

type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	RawLog    string    `json:"raw_log"`
}

type DashboardData struct {
	TotalLogs    int64
	ErrorCount   int64
	WarningCount int64
	RecentAlerts []*alerts.AlertInstance
	AlertRules   []*alerts.AlertRule
	Channels     []*alerts.NotificationChannel
}

func NewServer(storage *storage.Storage, engine *alerts.Engine) *Server {
	return &Server{
		storage: storage,
		engine:  engine,
	}
}

func (s *Server) Start(port int) error {
	// Static files and templates
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/logs", s.handleLogs)
	http.HandleFunc("/logs/search", s.handleLogsSearch)
	http.HandleFunc("/logs/stream", s.handleLogsStream)
	http.HandleFunc("/query", s.handleQuery)
	http.HandleFunc("/query/execute", s.handleQueryExecute)
	http.HandleFunc("/alerts", s.handleAlerts)
	http.HandleFunc("/alerts/rules", s.handleAlertRules)
	http.HandleFunc("/alerts/rules/add", s.handleAddAlertRule)
	http.HandleFunc("/alerts/channels", s.handleAlertChannels)
	http.HandleFunc("/alerts/channels/add", s.handleAddAlertChannel)
	http.HandleFunc("/alerts/tab/rules", s.handleAlertsTabRules)
	http.HandleFunc("/alerts/tab/channels", s.handleAlertsTabChannels)
	http.HandleFunc("/api/stats", s.handleAPIStats)
	http.HandleFunc("/api/debug/channels", s.handleDebugChannels)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🌐 Starting web server on http://localhost%s\n", addr)
	fmt.Println("📊 Dashboard: http://localhost" + addr)
	fmt.Println("📋 Logs: http://localhost" + addr + "/logs")
	fmt.Println("🚨 Alerts: http://localhost" + addr + "/alerts")

	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get dashboard data
	data, err := s.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Peep - Observability Dashboard</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/hyperscript.org@0.9.12"></script>
    <style>
        :root {
            --primary: #2563eb;
            --primary-hover: #1d4ed8;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 1rem;
        }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 1.5rem;
            font-weight: bold;
            color: var(--primary);
        }
        
        .tagline {
            font-size: 0.875rem;
            color: var(--gray-500);
            margin-left: 0.5rem;
        }
        
        nav {
            display: flex;
            gap: 1rem;
        }
        
        nav a {
            text-decoration: none;
            color: var(--gray-700);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        
        nav a:hover, nav a.active {
            background: var(--gray-100);
        }
        
        .grid {
            display: grid;
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        
        .grid-cols-4 {
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        }
        
        .card {
            background: white;
            border-radius: 0.5rem;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        
        .stat-card {
            text-align: center;
        }
        
        .stat-number {
            font-size: 2rem;
            font-weight: bold;
            margin-bottom: 0.5rem;
        }
        
        .stat-label {
            color: var(--gray-500);
            font-size: 0.875rem;
        }
        
        .text-primary { color: var(--primary); }
        .text-success { color: var(--success); }
        .text-warning { color: var(--warning); }
        .text-danger { color: var(--danger); }
        
        .btn {
            display: inline-block;
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            text-decoration: none;
            font-weight: 500;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .btn-primary {
            background: var(--primary);
            color: white;
        }
        
        .btn-primary:hover {
            background: var(--primary-hover);
        }
        
        .section-title {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 1rem;
        }
        
        .alert-item {
            padding: 0.75rem;
            border-left: 4px solid var(--warning);
            background: var(--gray-50);
            margin-bottom: 0.5rem;
            border-radius: 0 0.375rem 0.375rem 0;
        }
        
        .alert-critical {
            border-left-color: var(--danger);
        }
        
        .alert-title {
            font-weight: 600;
            margin-bottom: 0.25rem;
        }
        
        .alert-meta {
            font-size: 0.875rem;
            color: var(--gray-500);
        }
        
        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
        }
        
        .status-enabled {
            background: var(--success);
            color: white;
        }
        
        .status-disabled {
            background: var(--gray-300);
            color: var(--gray-700);
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/" class="active">Dashboard</a>
                    <a href="/logs">Logs</a>
                    <a href="/query">Query</a>
                    <a href="/alerts">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <!-- Stats Grid -->
        <div class="grid grid-cols-4">
            <div class="card stat-card">
                <div class="stat-number text-primary">{{.TotalLogs}}</div>
                <div class="stat-label">Total Logs</div>
            </div>
            <div class="card stat-card">
                <div class="stat-number text-danger">{{.ErrorCount}}</div>
                <div class="stat-label">Errors</div>
            </div>
            <div class="card stat-card">
                <div class="stat-number text-warning">{{.WarningCount}}</div>
                <div class="stat-label">Warnings</div>
            </div>
            <div class="card stat-card">
                <div class="stat-number text-success">{{len .AlertRules}}</div>
                <div class="stat-label">Alert Rules</div>
            </div>
        </div>

        <!-- Recent Alerts -->
        <div class="card">
            <div class="section-title">🚨 Recent Alerts</div>
            {{if .RecentAlerts}}
                {{range .RecentAlerts}}
                <div class="alert-item {{if ge .Count (mul .Threshold 2)}}alert-critical{{end}}">
                    <div class="alert-title">{{.RuleName}}</div>
                    <div class="alert-meta">
                        {{.Count}}/{{.Threshold}} events • {{.FiredAt.Format "2006-01-02 15:04:05"}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <p style="color: var(--gray-500); text-align: center; padding: 2rem;">
                    No recent alerts. Your system is running smoothly! 🎉
                </p>
            {{end}}
        </div>

        <!-- Alert Rules Status -->
        <div class="card">
            <div class="section-title">📋 Alert Rules</div>
            <div style="margin-bottom: 1rem;">
                <a href="/alerts/rules/add" class="btn btn-primary">+ Add Rule</a>
            </div>
            {{if .AlertRules}}
                {{range .AlertRules}}
                <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem; border-bottom: 1px solid var(--gray-200);">
                    <div>
                        <strong>{{.Name}}</strong>
                        <div style="font-size: 0.875rem; color: var(--gray-500);">{{.Description}}</div>
                    </div>
                    <div>
                        {{if .Enabled}}
                            <span class="status-badge status-enabled">Enabled</span>
                        {{else}}
                            <span class="status-badge status-disabled">Disabled</span>
                        {{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <p style="color: var(--gray-500); text-align: center; padding: 2rem;">
                    No alert rules configured yet.
                </p>
            {{end}}
        </div>

        <!-- Notification Channels -->
        <div class="card">
            <div class="section-title">📢 Notification Channels</div>
            <div style="margin-bottom: 1rem;">
                <a href="/alerts/channels/add" class="btn btn-primary">+ Add Channel</a>
            </div>
            {{if .Channels}}
                {{range .Channels}}
                <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem; border-bottom: 1px solid var(--gray-200);">
                    <div>
                        <strong>{{.Name}}</strong>
                        <div style="font-size: 0.875rem; color: var(--gray-500);">{{.Type}}</div>
                    </div>
                    <div>
                        {{if .Enabled}}
                            <span class="status-badge status-enabled">Enabled</span>
                        {{else}}
                            <span class="status-badge status-disabled">Disabled</span>
                        {{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <p style="color: var(--gray-500); text-align: center; padding: 2rem;">
                    No notification channels configured yet.
                </p>
            {{end}}
        </div>
    </div>

    <script>
        // Auto-refresh dashboard stats every 30 seconds
        setInterval(function() {
            htmx.ajax('GET', '/api/stats', {
                target: '.grid-cols-4',
                swap: 'innerHTML'
            });
        }, 30000);
    </script>
</body>
</html>`

	t, err := template.New("dashboard").Funcs(template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
	}).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getDashboardData() (*DashboardData, error) {
	db := s.storage.GetDB()

	// Get total logs count
	var totalLogs int64
	err := db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&totalLogs)
	if err != nil {
		return nil, err
	}

	// Get error count (last 24 hours)
	var errorCount int64
	err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE level = 'error' AND timestamp >= datetime('now', '-24 hours')").Scan(&errorCount)
	if err != nil {
		errorCount = 0
	}

	// Get warning count (last 24 hours)
	var warningCount int64
	err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE level = 'warning' AND timestamp >= datetime('now', '-24 hours')").Scan(&warningCount)
	if err != nil {
		warningCount = 0
	}

	// Get recent alerts (last 10)
	recentAlerts := make([]*alerts.AlertInstance, 0)
	rows, err := db.Query(`
		SELECT id, rule_id, rule_name, count, threshold, query, fired_at, resolved
		FROM alert_instances 
		ORDER BY fired_at DESC 
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			alert := &alerts.AlertInstance{}
			err := rows.Scan(&alert.ID, &alert.RuleID, &alert.RuleName, &alert.Count, &alert.Threshold, &alert.Query, &alert.FiredAt, &alert.Resolved)
			if err == nil {
				recentAlerts = append(recentAlerts, alert)
			}
		}
	}

	return &DashboardData{
		TotalLogs:    totalLogs,
		ErrorCount:   errorCount,
		WarningCount: warningCount,
		RecentAlerts: recentAlerts,
		AlertRules:   s.engine.GetRules(),
		Channels:     s.engine.GetChannels(),
	}, nil
}

func (s *Server) getFilteredLogs(search, level, service string, limit int) ([]*LogEntry, error) {
	db := s.storage.GetDB()

	// Build query with filters
	query := "SELECT id, timestamp, level, message, service, raw_log FROM logs WHERE 1=1"
	args := []interface{}{}

	if search != "" {
		query += " AND message LIKE ?"
		args = append(args, "%"+search+"%")
	}

	if level != "" {
		query += " AND level = ?"
		args = append(args, level)
	}

	if service != "" {
		query += " AND service = ?"
		args = append(args, service)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogEntry
	for rows.Next() {
		log := &LogEntry{}
		var serviceStr sql.NullString

		err := rows.Scan(&log.ID, &log.Timestamp, &log.Level, &log.Message, &serviceStr, &log.RawLog)
		if err != nil {
			continue
		}

		if serviceStr.Valid {
			log.Service = serviceStr.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (s *Server) getUniqueServices() ([]string, error) {
	db := s.storage.GetDB()

	rows, err := db.Query("SELECT DISTINCT service FROM logs WHERE service IS NOT NULL AND service != '' ORDER BY service")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err == nil {
			services = append(services, service)
		}
	}

	return services, nil
}

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	data, err := s.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return just the stats cards HTML for HTMX updates
	statsHTML := fmt.Sprintf(`
        <div class="card stat-card">
            <div class="stat-number text-primary">%d</div>
            <div class="stat-label">Total Logs</div>
        </div>
        <div class="card stat-card">
            <div class="stat-number text-danger">%d</div>
            <div class="stat-label">Errors</div>
        </div>
        <div class="card stat-card">
            <div class="stat-number text-warning">%d</div>
            <div class="stat-label">Warnings</div>
        </div>
        <div class="card stat-card">
            <div class="stat-number text-success">%d</div>
            <div class="stat-label">Alert Rules</div>
        </div>
    `, data.TotalLogs, data.ErrorCount, data.WarningCount, len(data.AlertRules))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(statsHTML))
}

func (s *Server) handleDebugChannels(w http.ResponseWriter, r *http.Request) {
	channels := s.engine.GetChannels()
	w.Header().Set("Content-Type", "application/json")

	// Simple JSON output for debugging
	fmt.Fprintf(w, `{"count": %d, "channels": [`, len(channels))
	for i, ch := range channels {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id": %d, "name": "%s", "type": "%s", "enabled": %t}`, ch.ID, ch.Name, ch.Type, ch.Enabled)
	}
	fmt.Fprint(w, "]}")
}

func (s *Server) handleAlertsTabRules(w http.ResponseWriter, r *http.Request) {
	rules := s.engine.GetRules()

	tmpl := `<div class="card">
		<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
			<h2 style="font-size: 1.25rem;">📋 Alert Rules</h2>
			<a href="/alerts/rules/add" class="btn btn-primary">+ Add Rule</a>
		</div>
		
		{{if .Rules}}
			{{range .Rules}}
			<div class="rule-item">
				<div class="rule-header">
					<div class="rule-title">{{.Name}}</div>
					<div>
						{{if .Enabled}}
							<span class="status-badge status-enabled">Enabled</span>
						{{else}}
							<span class="status-badge status-disabled">Disabled</span>
						{{end}}
					</div>
				</div>
				<div class="rule-description">{{.Description}}</div>
				<div class="rule-query">{{.Query}}</div>
				<div class="rule-meta">
					<span>Threshold: {{.Threshold}}</span>
					<span>Window: {{.Window}}</span>
				</div>
			</div>
			{{end}}
		{{else}}
			<div style="text-align: center; padding: 3rem; color: var(--gray-500);">
				<div style="font-size: 3rem; margin-bottom: 1rem;">📝</div>
				<h3>No alert rules configured</h3>
				<p>Create your first alert rule to start monitoring your logs.</p>
			</div>
		{{end}}
	</div>`

	data := struct {
		Rules []*alerts.AlertRule
	}{
		Rules: rules,
	}

	t, err := template.New("rulesTab").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleAlertsTabChannels(w http.ResponseWriter, r *http.Request) {
	channels := s.engine.GetChannels()

	tmpl := `<div class="card">
		<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
			<h2 style="font-size: 1.25rem;">📢 Notification Channels</h2>
			<a href="/alerts/channels/add" class="btn btn-primary">+ Add Channel</a>
		</div>
		
		{{if .Channels}}
			{{range .Channels}}
			<div class="channel-item">
				<div class="channel-header">
					<div class="channel-title">{{.Name}}</div>
					<div>
						{{if .Enabled}}
							<span class="status-badge status-enabled">Enabled</span>
						{{else}}
							<span class="status-badge status-disabled">Disabled</span>
						{{end}}
					</div>
				</div>
				<div class="channel-meta">
					<span><strong>Type:</strong> {{.Type}}</span>
					{{if eq .Type "slack"}}
						<span><strong>Webhook:</strong> {{if index .Config "webhook_url"}}Configured{{else}}Not set{{end}}</span>
					{{else if eq .Type "email"}}
						<span><strong>SMTP:</strong> {{index .Config "smtp_host"}}:{{index .Config "smtp_port"}}</span>
					{{else if eq .Type "shell"}}
						<span><strong>Script:</strong> {{index .Config "script_path"}}</span>
					{{end}}
				</div>
			</div>
			{{end}}
		{{else}}
			<div style="text-align: center; padding: 3rem; color: var(--gray-500);">
				<div style="font-size: 3rem; margin-bottom: 1rem;">📬</div>
				<h3>No notification channels configured</h3>
				<p>Add channels to receive alert notifications.</p>
			</div>
		{{end}}
	</div>`

	data := struct {
		Channels []*alerts.NotificationChannel
	}{
		Channels: channels,
	}

	t, err := template.New("channelsTab").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	search := r.URL.Query().Get("search")
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")
	limit := 50 // Default page size

	logs, err := s.getFilteredLogs(search, level, service, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unique services for filter dropdown
	services, _ := s.getUniqueServices()

	data := struct {
		Logs     []*LogEntry
		Search   string
		Level    string
		Service  string
		Services []string
	}{
		Logs:     logs,
		Search:   search,
		Level:    level,
		Service:  service,
		Services: services,
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Logs - Peep</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 1rem;
        }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 1.5rem;
            font-weight: bold;
            color: var(--primary);
        }
        
        .tagline {
            font-size: 0.875rem;
            color: var(--gray-500);
            margin-left: 0.5rem;
        }
        
        nav {
            display: flex;
            gap: 1rem;
        }
        
        nav a {
            text-decoration: none;
            color: var(--gray-700);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        
        nav a:hover, nav a.active {
            background: var(--gray-100);
        }
        
        .card {
            background: white;
            border-radius: 0.5rem;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            margin-bottom: 1.5rem;
        }
        
        .filters {
            display: flex;
            gap: 1rem;
            margin-bottom: 1.5rem;
            flex-wrap: wrap;
        }
        
        .filter-group {
            display: flex;
            flex-direction: column;
            gap: 0.25rem;
        }
        
        .filter-group label {
            font-size: 0.875rem;
            font-weight: 500;
            color: var(--gray-700);
        }
        
        .filter-group input, .filter-group select {
            padding: 0.5rem;
            border: 1px solid var(--gray-300);
            border-radius: 0.375rem;
            font-size: 0.875rem;
        }
        
        .filter-group input:focus, .filter-group select:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
        }
        
        .btn {
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            font-weight: 500;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
            font-size: 0.875rem;
        }
        
        .btn-primary {
            background: var(--primary);
            color: white;
        }
        
        .btn-secondary {
            background: var(--gray-200);
            color: var(--gray-700);
        }
        
        .log-table {
            width: 100%;
            border-collapse: collapse;
        }
        
        .log-table th {
            background: var(--gray-50);
            padding: 0.75rem;
            text-align: left;
            font-weight: 600;
            border-bottom: 1px solid var(--gray-200);
            font-size: 0.875rem;
        }
        
        .log-table td {
            padding: 0.75rem;
            border-bottom: 1px solid var(--gray-200);
            font-size: 0.875rem;
            vertical-align: top;
        }
        
        .log-table tr:hover {
            background: var(--gray-50);
        }
        
        .level-badge {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
        }
        
        .level-info { background: #dbeafe; color: #1e40af; }
        .level-warning { background: #fef3c7; color: #92400e; }
        .level-error { background: #fee2e2; color: #dc2626; }
        .level-debug { background: #f3f4f6; color: #6b7280; }
        
        .log-message {
            max-width: 400px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .log-raw {
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.75rem;
            color: var(--gray-600);
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .timestamp {
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.75rem;
            color: var(--gray-600);
        }
        
        .empty-state {
            text-align: center;
            padding: 3rem;
            color: var(--gray-500);
        }
        
        .loading {
            text-align: center;
            padding: 2rem;
            color: var(--gray-500);
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/">Dashboard</a>
                    <a href="/logs" class="active">Logs</a>
                    <a href="/query">Query</a>
                    <a href="/alerts">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <div class="card">
            <h1 style="margin-bottom: 1.5rem; font-size: 1.5rem;">📋 Log Viewer</h1>
            
            <!-- Filters -->
            <form hx-get="/logs/search" hx-target="#log-results" hx-trigger="input delay:300ms, change" class="filters">
                <div class="filter-group">
                    <label for="search">Search</label>
                    <input type="text" id="search" name="search" value="{{.Search}}" placeholder="Search messages..." style="width: 300px;">
                </div>
                <div class="filter-group">
                    <label for="level">Level</label>
                    <select id="level" name="level">
                        <option value="">All Levels</option>
                        <option value="debug" {{if eq .Level "debug"}}selected{{end}}>Debug</option>
                        <option value="info" {{if eq .Level "info"}}selected{{end}}>Info</option>
                        <option value="warning" {{if eq .Level "warning"}}selected{{end}}>Warning</option>
                        <option value="error" {{if eq .Level "error"}}selected{{end}}>Error</option>
                    </select>
                </div>
                <div class="filter-group">
                    <label for="service">Service</label>
                    <select id="service" name="service">
                        <option value="">All Services</option>
                        {{range .Services}}
                        <option value="{{.}}" {{if eq $.Service .}}selected{{end}}>{{.}}</option>
                        {{end}}
                    </select>
                </div>
                <div class="filter-group" style="justify-content: end;">
                    <label>&nbsp;</label>
                    <button type="button" class="btn btn-secondary" onclick="document.querySelector('form').reset(); htmx.trigger(document.querySelector('form'), 'change');">Clear</button>
                </div>
            </form>
        </div>

        <!-- Log Results -->
        <div class="card">
            <div id="log-results">
                {{template "logTable" .}}
            </div>
        </div>
    </div>
</body>
</html>

{{define "logTable"}}
{{if .Logs}}
<table class="log-table">
    <thead>
        <tr>
            <th style="width: 150px;">Timestamp</th>
            <th style="width: 80px;">Level</th>
            <th style="width: 100px;">Service</th>
            <th>Message</th>
            <th style="width: 200px;">Raw Log</th>
        </tr>
    </thead>
    <tbody>
        {{range .Logs}}
        <tr>
            <td class="timestamp">{{.Timestamp.Format "01-02 15:04:05"}}</td>
            <td>
                <span class="level-badge level-{{.Level}}">{{.Level}}</span>
            </td>
            <td>{{if .Service}}{{.Service}}{{else}}-{{end}}</td>
            <td class="log-message" title="{{.Message}}">{{.Message}}</td>
            <td class="log-raw" title="{{.RawLog}}">{{.RawLog}}</td>
        </tr>
        {{end}}
    </tbody>
</table>
{{else}}
<div class="empty-state">
    <div style="font-size: 3rem; margin-bottom: 1rem;">📝</div>
    <h3>No logs found</h3>
    <p>Try adjusting your search filters or ingest some logs first.</p>
</div>
{{end}}
{{end}}`

	t, err := template.New("logs").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleLogsSearch(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	search := r.URL.Query().Get("search")
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")
	limit := 50

	logs, err := s.getFilteredLogs(search, level, service, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unique services for filter dropdown
	services, _ := s.getUniqueServices()

	data := struct {
		Logs     []*LogEntry
		Search   string
		Level    string
		Service  string
		Services []string
	}{
		Logs:     logs,
		Search:   search,
		Level:    level,
		Service:  service,
		Services: services,
	}

	// Return just the table for HTMX
	tmpl := `{{if .Logs}}
<table class="log-table">
    <thead>
        <tr>
            <th style="width: 150px;">Timestamp</th>
            <th style="width: 80px;">Level</th>
            <th style="width: 100px;">Service</th>
            <th>Message</th>
            <th style="width: 200px;">Raw Log</th>
        </tr>
    </thead>
    <tbody>
        {{range .Logs}}
        <tr>
            <td class="timestamp">{{.Timestamp.Format "01-02 15:04:05"}}</td>
            <td>
                <span class="level-badge level-{{.Level}}">{{.Level}}</span>
            </td>
            <td>{{if .Service}}{{.Service}}{{else}}-{{end}}</td>
            <td class="log-message" title="{{.Message}}">{{.Message}}</td>
            <td class="log-raw" title="{{.RawLog}}">{{.RawLog}}</td>
        </tr>
        {{end}}
    </tbody>
</table>
{{else}}
<div class="empty-state">
    <div style="font-size: 3rem; margin-bottom: 1rem;">📝</div>
    <h3>No logs found</h3>
    <p>Try adjusting your search filters or ingest some logs first.</p>
</div>
{{end}}`

	t, err := template.New("logTable").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	rules := s.engine.GetRules()
	channels := s.engine.GetChannels()

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alerts - Peep</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container { max-width: 1200px; margin: 0 auto; padding: 0 1rem; }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo { font-size: 1.5rem; font-weight: bold; color: var(--primary); }
        .tagline { font-size: 0.875rem; color: var(--gray-500); margin-left: 0.5rem; }
        
        nav { display: flex; gap: 1rem; }
        nav a {
            text-decoration: none;
            color: var(--gray-700);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        nav a:hover, nav a.active { background: var(--gray-100); }
        
        .card {
            background: white;
            border-radius: 0.5rem;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            margin-bottom: 1.5rem;
        }
        
        .btn {
            display: inline-block;
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            text-decoration: none;
            font-weight: 500;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
            font-size: 0.875rem;
        }
        
        .btn-primary { background: var(--primary); color: white; }
        .btn-danger { background: var(--danger); color: white; }
        .btn-secondary { background: var(--gray-200); color: var(--gray-700); }
        
        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
        }
        
        .status-enabled { background: var(--success); color: white; }
        .status-disabled { background: var(--gray-300); color: var(--gray-700); }
        
        .rule-item, .channel-item {
            border: 1px solid var(--gray-200);
            border-radius: 0.5rem;
            padding: 1rem;
            margin-bottom: 1rem;
        }
        
        .rule-header, .channel-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
        }
        
        .rule-title, .channel-title { font-weight: 600; font-size: 1.1rem; }
        .rule-description { color: var(--gray-600); margin-bottom: 0.5rem; }
        .rule-query { 
            font-family: 'Monaco', 'Consolas', monospace; 
            background: var(--gray-100); 
            padding: 0.5rem; 
            border-radius: 0.25rem; 
            font-size: 0.875rem;
            margin: 0.5rem 0;
        }
        
        .rule-meta, .channel-meta {
            display: flex;
            gap: 1rem;
            font-size: 0.875rem;
            color: var(--gray-600);
        }
        
        .tab-nav {
            display: flex;
            border-bottom: 1px solid var(--gray-200);
            margin-bottom: 1.5rem;
        }
        
        .tab-btn {
            background: none;
            border: none;
            padding: 0.75rem 1.5rem;
            font-size: 0.875rem;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        
        .tab-btn:hover {
            background: var(--gray-50);
        }
        
        .tab-btn.active {
            color: var(--primary);
            border-bottom-color: var(--primary);
        }
        
        .tab-content { display: none; }
        .tab-content.active { display: block; }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/">Dashboard</a>
                    <a href="/logs">Logs</a>
                    <a href="/query">Query</a>
                    <a href="/alerts" class="active">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <h1 style="margin-bottom: 1.5rem; font-size: 1.75rem;">🚨 Alert Management</h1>
        
        <div class="tab-nav">
            <button class="tab-btn active" 
                    hx-get="/alerts/tab/rules" 
                    hx-target="#tab-container" 
                    hx-swap="innerHTML"
                    hx-on:click="
                        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
                        this.classList.add('active');
                    ">Alert Rules</button>
            <button class="tab-btn" 
                    hx-get="/alerts/tab/channels" 
                    hx-target="#tab-container" 
                    hx-swap="innerHTML"
                    hx-on:click="
                        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
                        this.classList.add('active');
                    ">Notification Channels</button>
        </div>

        <!-- Tab Container -->
        <div id="tab-container">
            <!-- Default content will be loaded via HTMX -->
            <div class="card">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                    <h2 style="font-size: 1.25rem;">📋 Alert Rules</h2>
                    <a href="/alerts/rules/add" class="btn btn-primary">+ Add Rule</a>
                </div>
                
                {{if .Rules}}
                    {{range .Rules}}
                    <div class="rule-item">
                        <div class="rule-header">
                            <div class="rule-title">{{.Name}}</div>
                            <div>
                                {{if .Enabled}}
                                    <span class="status-badge status-enabled">Enabled</span>
                                {{else}}
                                    <span class="status-badge status-disabled">Disabled</span>
                                {{end}}
                            </div>
                        </div>
                        <div class="rule-description">{{.Description}}</div>
                        <div class="rule-query">{{.Query}}</div>
                        <div class="rule-meta">
                            <span>Threshold: {{.Threshold}}</span>
                            <span>Interval: {{.Interval}}s</span>
                            {{if .Channels}}
                                <span>Channels: {{range $i, $ch := .Channels}}{{if $i}}, {{end}}{{$ch}}{{end}}</span>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                {{else}}
                    <div style="text-align: center; padding: 3rem; color: var(--gray-500);">
                        <div style="font-size: 3rem; margin-bottom: 1rem;">📝</div>
                        <h3>No alert rules configured</h3>
                        <p>Create your first alert rule to start monitoring your logs.</p>
                    </div>
                {{end}}
            </div>
        </div>

    <script>
        // Tab functionality handled by HTMX inline events
    </script>
</body>
</html>`

	data := struct {
		Rules    []*alerts.AlertRule
		Channels []*alerts.NotificationChannel
	}{
		Rules:    rules,
		Channels: channels,
	}

	t, err := template.New("alerts").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleAlertRules(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alert rules management coming soon!"))
}

func (s *Server) handleAddAlertRule(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show the form
		channels := s.engine.GetChannels()

		data := struct {
			Channels []*alerts.NotificationChannel
		}{
			Channels: channels,
		}

		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Add Alert Rule - Peep</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container { max-width: 800px; margin: 0 auto; padding: 0 1rem; }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo { font-size: 1.5rem; font-weight: bold; color: var(--primary); }
        .tagline { font-size: 0.875rem; color: var(--gray-500); margin-left: 0.5rem; }
        
        nav { display: flex; gap: 1rem; }
        nav a {
            text-decoration: none;
            color: var(--gray-700);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        nav a:hover, nav a.active { background: var(--gray-100); }
        
        .card {
            background: white;
            border-radius: 0.5rem;
            padding: 2rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            margin-bottom: 1.5rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-group label {
            display: block;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: var(--gray-700);
        }
        
        .form-group input, .form-group textarea, .form-group select {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid var(--gray-300);
            border-radius: 0.375rem;
            font-size: 0.875rem;
        }
        
        .form-group input:focus, .form-group textarea:focus, .form-group select:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
        }
        
        .form-group textarea {
            resize: vertical;
            min-height: 100px;
            font-family: 'Monaco', 'Consolas', monospace;
        }
        
        .form-help {
            font-size: 0.875rem;
            color: var(--gray-600);
            margin-top: 0.25rem;
        }
        
        .checkbox-group {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
            margin-top: 0.5rem;
        }
        
        .checkbox-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .checkbox-item input[type="checkbox"] {
            width: auto;
            margin: 0;
        }
        
        .btn {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            border-radius: 0.375rem;
            text-decoration: none;
            font-weight: 500;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
            font-size: 0.875rem;
            margin-right: 0.5rem;
        }
        
        .btn-primary { background: var(--primary); color: white; }
        .btn-primary:hover { background: #1d4ed8; }
        .btn-secondary { background: var(--gray-200); color: var(--gray-700); }
        .btn-secondary:hover { background: var(--gray-300); }
        
        .breadcrumb {
            margin-bottom: 1.5rem;
            font-size: 0.875rem;
            color: var(--gray-600);
        }
        
        .breadcrumb a {
            color: var(--primary);
            text-decoration: none;
        }
        
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 1rem;
        }
        
        .query-preview {
            background: var(--gray-50);
            border: 1px solid var(--gray-200);
            border-radius: 0.375rem;
            padding: 1rem;
            margin-top: 1rem;
        }
        
        .query-preview h4 {
            margin-bottom: 0.5rem;
            font-size: 0.875rem;
            color: var(--gray-700);
        }
        
        .query-examples {
            margin-top: 0.5rem;
        }
        
        .query-example {
            background: var(--gray-100);
            padding: 0.5rem;
            border-radius: 0.25rem;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.75rem;
            margin-bottom: 0.25rem;
            cursor: pointer;
        }
        
        .query-example:hover {
            background: var(--gray-200);
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/">Dashboard</a>
                    <a href="/logs">Logs</a>
                    <a href="/query">Query</a>
                    <a href="/alerts" class="active">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <div class="breadcrumb">
            <a href="/alerts">Alerts</a> / Add Rule
        </div>
        
        <div class="card">
            <h1 style="margin-bottom: 1.5rem; font-size: 1.5rem;">📝 Add Alert Rule</h1>
            
            <form hx-post="/alerts/rules/add" hx-target="#form-result">
                <div class="form-group">
                    <label for="name">Rule Name *</label>
                    <input type="text" id="name" name="name" required placeholder="e.g., High Error Rate">
                    <div class="form-help">A descriptive name for this alert rule</div>
                </div>

                <div class="form-group">
                    <label for="description">Description</label>
                    <input type="text" id="description" name="description" placeholder="e.g., Alert when error rate exceeds threshold">
                    <div class="form-help">Optional description of what this rule monitors</div>
                </div>

                <div class="form-group">
                    <label for="query">SQL Query *</label>
                    <textarea id="query" name="query" required placeholder="SELECT COUNT(*) FROM logs WHERE level='error' AND timestamp > datetime('now', '-5 minutes')"></textarea>
                    <div class="form-help">SQL query that returns a count. The result will be compared against the threshold.</div>
                    
                    <div class="query-preview">
                        <h4>Example Queries:</h4>
                        <div class="query-examples">
                            <div class="query-example" onclick="setQuery(this)">SELECT COUNT(*) FROM logs WHERE level='error' AND timestamp > datetime('now', '-5 minutes')</div>
                            <div class="query-example" onclick="setQuery(this)">SELECT COUNT(*) FROM logs WHERE message LIKE '%timeout%' AND timestamp > datetime('now', '-10 minutes')</div>
                            <div class="query-example" onclick="setQuery(this)">SELECT COUNT(*) FROM logs WHERE service='api' AND level IN ('error', 'warning') AND timestamp > datetime('now', '-15 minutes')</div>
                        </div>
                    </div>
                </div>

                <div class="form-row">
                    <div class="form-group">
                        <label for="threshold">Threshold *</label>
                        <input type="number" id="threshold" name="threshold" required min="1" value="5">
                        <div class="form-help">Alert fires when query result >= this value</div>
                    </div>

                    <div class="form-group">
                        <label for="interval">Check Interval (seconds) *</label>
                        <input type="number" id="interval" name="interval" required min="10" value="60">
                        <div class="form-help">How often to run the query</div>
                    </div>
                </div>

                <div class="form-group">
                    <label>Notification Channels</label>
                    <div style="padding: 1rem; background: var(--gray-100); border-radius: 0.375rem; color: var(--gray-600);">
                        📢 Channel assignment will be available in the next update. For now, all channels will receive alerts.
                    </div>
                </div>

                <div class="form-group">
                    <div class="checkbox-item">
                        <input type="checkbox" id="enabled" name="enabled" checked>
                        <label for="enabled">Enable this rule</label>
                    </div>
                </div>

                <div style="margin-top: 2rem;">
                    <button type="submit" class="btn btn-primary">Create Alert Rule</button>
                    <a href="/alerts" class="btn btn-secondary">Cancel</a>
                </div>

                <div id="form-result" style="margin-top: 1rem;"></div>
            </form>
        </div>
    </div>

    <script>
        function setQuery(element) {
            document.getElementById('query').value = element.textContent;
        }
    </script>
</body>
</html>`

		t, err := template.New("addRule").Parse(tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if r.Method == "POST" {
		// Handle form submission
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Extract form data
		name := r.FormValue("name")
		description := r.FormValue("description")
		query := r.FormValue("query")
		threshold := r.FormValue("threshold")
		interval := r.FormValue("interval")
		enabled := r.FormValue("enabled") == "on"

		// Validate required fields
		if name == "" || query == "" || threshold == "" || interval == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Please fill in all required fields.
			</div>`))
			return
		}

		// Convert string values to integers and create window
		thresholdInt := 0
		intervalInt := 0
		if _, err := fmt.Sscanf(threshold, "%d", &thresholdInt); err != nil || thresholdInt <= 0 {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Threshold must be a positive number.
			</div>`))
			return
		}

		if _, err := fmt.Sscanf(interval, "%d", &intervalInt); err != nil || intervalInt < 10 {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Interval must be at least 10 seconds.
			</div>`))
			return
		}

		// Convert interval to window format (e.g., "60s", "5m")
		window := fmt.Sprintf("%ds", intervalInt)
		if intervalInt >= 60 && intervalInt%60 == 0 {
			window = fmt.Sprintf("%dm", intervalInt/60)
		}

		// Create the alert rule
		rule := &alerts.AlertRule{
			Name:        name,
			Description: description,
			Query:       query,
			Threshold:   thresholdInt,
			Window:      window,
			Enabled:     enabled,
		}

		// Add the rule via the engine
		err = s.engine.AddRule(rule)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(fmt.Sprintf(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Error creating rule: %s
			</div>`, err.Error())))
			return
		}

		// Success response with redirect
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div style="color: var(--success); padding: 1rem; background: #d1fae5; border-radius: 0.375rem;">
			✅ Alert rule created successfully! <a href="/alerts">View all rules</a>
		</div>`))
	}
}

func (s *Server) handleAlertChannels(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alert channels management coming soon!"))
}

func (s *Server) handleAddAlertChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show the form
		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Add Notification Channel - Peep</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container { max-width: 800px; margin: 0 auto; padding: 0 1rem; }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo { font-size: 1.5rem; font-weight: bold; color: var(--primary); }
        .tagline { font-size: 0.875rem; color: var(--gray-500); margin-left: 0.5rem; }
        
        nav { display: flex; gap: 1rem; }
        nav a {
            text-decoration: none;
            color: var(--gray-700);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        nav a:hover, nav a.active { background: var(--gray-100); }
        
        .card {
            background: white;
            border-radius: 0.5rem;
            padding: 2rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            margin-bottom: 1.5rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-group label {
            display: block;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: var(--gray-700);
        }
        
        .form-group input, .form-group textarea, .form-group select {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid var(--gray-300);
            border-radius: 0.375rem;
            font-size: 0.875rem;
        }
        
        .form-group input:focus, .form-group textarea:focus, .form-group select:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
        }
        
        .form-help {
            font-size: 0.875rem;
            color: var(--gray-600);
            margin-top: 0.25rem;
        }
        
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 1rem;
        }
        
        .checkbox-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .checkbox-item input[type="checkbox"] {
            width: auto;
            margin: 0;
        }
        
        .btn {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            border-radius: 0.375rem;
            text-decoration: none;
            font-weight: 500;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
            font-size: 0.875rem;
            margin-right: 0.5rem;
        }
        
        .btn-primary { background: var(--primary); color: white; }
        .btn-primary:hover { background: #1d4ed8; }
        .btn-secondary { background: var(--gray-200); color: var(--gray-700); }
        .btn-secondary:hover { background: var(--gray-300); }
        
        .breadcrumb {
            margin-bottom: 1.5rem;
            font-size: 0.875rem;
            color: var(--gray-600);
        }
        
        .breadcrumb a {
            color: var(--primary);
            text-decoration: none;
        }
        
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        
        .channel-config {
            display: none;
            background: var(--gray-50);
            border: 1px solid var(--gray-200);
            border-radius: 0.375rem;
            padding: 1rem;
            margin-top: 1rem;
        }
        
        .channel-config.active {
            display: block;
        }
        
        .channel-config h4 {
            margin-bottom: 1rem;
            color: var(--gray-700);
        }
        
        .config-example {
            background: var(--gray-100);
            padding: 0.5rem;
            border-radius: 0.25rem;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.75rem;
            margin-top: 0.5rem;
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/">Dashboard</a>
                    <a href="/logs">Logs</a>
                    <a href="/query">Query</a>
                    <a href="/alerts" class="active">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <div class="breadcrumb">
            <a href="/alerts">Alerts</a> / Add Channel
        </div>
        
        <div class="card">
            <h1 style="margin-bottom: 1.5rem; font-size: 1.5rem;">📢 Add Notification Channel</h1>
            
            <form hx-post="/alerts/channels/add" hx-target="#form-result">
                <div class="form-group">
                    <label for="name">Channel Name *</label>
                    <input type="text" id="name" name="name" required placeholder="e.g., Team Slack, DevOps Email">
                    <div class="form-help">A descriptive name for this notification channel</div>
                </div>

                <div class="form-group">
                    <label for="type">Channel Type *</label>
                    <select id="type" name="type" required onchange="showChannelConfig(this.value)">
                        <option value="">Select channel type...</option>
                        <option value="slack">Slack (Webhook)</option>
                        <option value="email">Email (SMTP)</option>
                        <option value="shell">Shell Script</option>
                        <option value="desktop">Desktop Notifications</option>
                    </select>
                    <div class="form-help">Choose how you want to receive notifications</div>
                </div>

                <!-- Slack Configuration -->
                <div id="slack-config" class="channel-config">
                    <h4>🔗 Slack Configuration</h4>
                    <div class="form-group">
                        <label for="slack-webhook">Webhook URL *</label>
                        <input type="url" id="slack-webhook" name="slack-webhook" placeholder="https://hooks.slack.com/services/...">
                        <div class="form-help">Get this from your Slack app's "Incoming Webhooks" settings</div>
                        <div class="config-example">Example: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX</div>
                    </div>
                    <div class="form-group">
                        <label for="slack-channel">Channel (optional)</label>
                        <input type="text" id="slack-channel" name="slack-channel" placeholder="#alerts">
                        <div class="form-help">Override default channel (include # for channels, @ for users)</div>
                    </div>
                </div>

                <!-- Email Configuration -->
                <div id="email-config" class="channel-config">
                    <h4>📧 Email Configuration</h4>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="email-smtp-host">SMTP Host *</label>
                            <input type="text" id="email-smtp-host" name="email-smtp-host" placeholder="smtp.gmail.com">
                        </div>
                        <div class="form-group">
                            <label for="email-smtp-port">SMTP Port *</label>
                            <input type="number" id="email-smtp-port" name="email-smtp-port" placeholder="587" value="587">
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="email-username">Username *</label>
                            <input type="text" id="email-username" name="email-username" placeholder="your-email@gmail.com">
                        </div>
                        <div class="form-group">
                            <label for="email-password">Password *</label>
                            <input type="password" id="email-password" name="email-password" placeholder="app-password">
                            <div class="form-help">Use app password for Gmail</div>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="email-from">From Email *</label>
                            <input type="email" id="email-from" name="email-from" placeholder="alerts@yourcompany.com">
                        </div>
                        <div class="form-group">
                            <label for="email-to">To Email(s) *</label>
                            <input type="text" id="email-to" name="email-to" placeholder="team@yourcompany.com">
                            <div class="form-help">Comma-separated for multiple recipients</div>
                        </div>
                    </div>
                    <div class="checkbox-item">
                        <input type="checkbox" id="email-tls" name="email-tls" checked>
                        <label for="email-tls">Use TLS encryption</label>
                    </div>
                </div>

                <!-- Shell Script Configuration -->
                <div id="shell-config" class="channel-config">
                    <h4>⚡ Shell Script Configuration</h4>
                    <div class="form-group">
                        <label for="shell-script">Script Path *</label>
                        <input type="text" id="shell-script" name="shell-script" placeholder="/path/to/alert-handler.sh">
                        <div class="form-help">Absolute path to your alert handler script</div>
                        <div class="config-example">Script will receive: ALERT_TITLE, ALERT_MESSAGE, ALERT_COUNT, ALERT_THRESHOLD environment variables</div>
                    </div>
                    <div class="form-group">
                        <label for="shell-args">Arguments (optional)</label>
                        <input type="text" id="shell-args" name="shell-args" placeholder="--format json --urgent">
                        <div class="form-help">Space-separated arguments to pass to the script</div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="shell-timeout">Timeout (seconds)</label>
                            <input type="number" id="shell-timeout" name="shell-timeout" value="30" min="1" max="300">
                        </div>
                        <div class="form-group">
                            <label for="shell-workdir">Working Directory</label>
                            <input type="text" id="shell-workdir" name="shell-workdir" placeholder="/opt/peep">
                        </div>
                    </div>
                </div>

                <!-- Desktop Configuration -->
                <div id="desktop-config" class="channel-config">
                    <h4>🖥️ Desktop Notifications</h4>
                    <p style="color: var(--gray-600); margin-bottom: 1rem;">
                        Desktop notifications work out of the box on macOS, Linux, and Windows. 
                        No additional configuration required.
                    </p>
                    <div class="config-example">
                        ✅ Notifications will appear in your system notification area<br>
                        🔔 Make sure Peep has notification permissions on your system
                    </div>
                </div>

                <div class="form-group">
                    <div class="checkbox-item">
                        <input type="checkbox" id="enabled" name="enabled" checked>
                        <label for="enabled">Enable this channel</label>
                    </div>
                </div>

                <div style="margin-top: 2rem;">
                    <button type="submit" class="btn btn-primary">Create Notification Channel</button>
                    <a href="/alerts" class="btn btn-secondary">Cancel</a>
                </div>

                <div id="form-result" style="margin-top: 1rem;"></div>
            </form>
        </div>
    </div>

    <script>
        function showChannelConfig(channelType) {
            // Hide all config sections
            document.querySelectorAll('.channel-config').forEach(config => {
                config.classList.remove('active');
            });
            
            // Show selected config section
            if (channelType) {
                const configElement = document.getElementById(channelType + '-config');
                if (configElement) {
                    configElement.classList.add('active');
                }
            }
        }
    </script>
</body>
</html>`

		t, err := template.New("addChannel").Parse(tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if r.Method == "POST" {
		// Handle form submission
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Extract common fields
		name := r.FormValue("name")
		channelType := r.FormValue("type")
		enabled := r.FormValue("enabled") == "on"

		// Validate required fields
		if name == "" || channelType == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Please fill in channel name and type.
			</div>`))
			return
		}

		// Build config based on channel type
		config := make(map[string]string)

		switch channelType {
		case "slack":
			webhookURL := r.FormValue("slack-webhook")
			channel := r.FormValue("slack-channel")

			if webhookURL == "" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
					❌ Slack webhook URL is required.
				</div>`))
				return
			}

			config["webhook_url"] = webhookURL
			if channel != "" {
				config["channel"] = channel
			}

		case "email":
			smtpHost := r.FormValue("email-smtp-host")
			smtpPort := r.FormValue("email-smtp-port")
			username := r.FormValue("email-username")
			password := r.FormValue("email-password")
			fromEmail := r.FormValue("email-from")
			toEmail := r.FormValue("email-to")
			useTLS := r.FormValue("email-tls") == "on"

			if smtpHost == "" || smtpPort == "" || username == "" || password == "" || fromEmail == "" || toEmail == "" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
					❌ Please fill in all required email fields.
				</div>`))
				return
			}

			config["smtp_host"] = smtpHost
			config["smtp_port"] = smtpPort
			config["username"] = username
			config["password"] = password
			config["from_email"] = fromEmail
			config["to_emails"] = toEmail
			if useTLS {
				config["use_tls"] = "true"
			}

		case "shell":
			scriptPath := r.FormValue("shell-script")
			args := r.FormValue("shell-args")
			timeout := r.FormValue("shell-timeout")
			workdir := r.FormValue("shell-workdir")

			if scriptPath == "" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
					❌ Script path is required for shell notifications.
				</div>`))
				return
			}

			config["script_path"] = scriptPath
			if args != "" {
				config["args"] = args
			}
			if timeout != "" {
				config["timeout"] = timeout
			}
			if workdir != "" {
				config["working_dir"] = workdir
			}

		case "desktop":
			// Desktop notifications need no additional config
			config["enabled"] = "true"
		}

		// Create the notification channel
		channel := &alerts.NotificationChannel{
			Name:    name,
			Type:    channelType,
			Config:  config,
			Enabled: enabled,
		}

		// Add the channel via the engine
		err = s.engine.AddNotificationChannel(channel)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(fmt.Sprintf(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
				❌ Error creating channel: %s
			</div>`, err.Error())))
			return
		}

		// Success response with redirect
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div style="color: var(--success); padding: 1rem; background: #d1fae5; border-radius: 0.375rem;">
			✅ Notification channel created successfully! <a href="/alerts">View all channels</a>
		</div>`))
	}
}

// handleLogsStream provides real-time log streaming via Server-Sent Events
func (s *Server) handleLogsStream(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the latest log ID to start streaming from
	lastID := r.URL.Query().Get("lastId")
	if lastID == "" {
		lastID = "0"
	}

	// Send initial ping
	fmt.Fprintf(w, "data: {\"type\":\"ping\"}\n\n")
	w.(http.Flusher).Flush()

	// TODO: Implement actual streaming - for now, just acknowledge the endpoint
	fmt.Fprintf(w, "data: {\"type\":\"info\",\"message\":\"Stream endpoint ready\"}\n\n")
	w.(http.Flusher).Flush()
}

// handleQuery shows the SQL query interface
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Query Interface - Peep</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-500: #6b7280;
            --gray-700: #374151;
            --gray-900: #111827;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--gray-50);
            color: var(--gray-900);
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 1rem;
        }
        
        header {
            background: white;
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem 0;
            margin-bottom: 2rem;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--primary);
        }
        
        .tagline {
            font-size: 0.875rem;
            color: var(--gray-500);
            margin-left: 0.5rem;
        }
        
        nav {
            display: flex;
            gap: 1rem;
        }
        
        nav a {
            color: var(--gray-700);
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            transition: all 0.2s;
        }
        
        nav a:hover, nav a.active {
            background: var(--gray-100);
            color: var(--primary);
        }
        
        .query-container {
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
            margin-bottom: 2rem;
        }
        
        .query-header {
            background: var(--gray-50);
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem;
        }
        
        .query-header h2 {
            margin-bottom: 0.5rem;
        }
        
        .query-examples {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
        }
        
        .example-query {
            background: var(--primary);
            color: white;
            border: none;
            padding: 0.25rem 0.75rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .example-query:hover {
            background: #1d4ed8;
        }
        
        .query-form {
            padding: 1rem;
        }
        
        .query-textarea {
            width: 100%;
            min-height: 120px;
            padding: 0.75rem;
            border: 1px solid var(--gray-300);
            border-radius: 0.375rem;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.875rem;
            resize: vertical;
            margin-bottom: 1rem;
        }
        
        .query-actions {
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        
        .btn {
            padding: 0.5rem 1rem;
            border: none;
            border-radius: 0.375rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .btn-primary {
            background: var(--primary);
            color: white;
        }
        
        .btn-primary:hover {
            background: #1d4ed8;
        }
        
        .results-container {
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .results-header {
            background: var(--gray-50);
            border-bottom: 1px solid var(--gray-200);
            padding: 1rem;
        }
        
        .results-content {
            padding: 1rem;
        }
        
        .query-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.875rem;
        }
        
        .query-table th,
        .query-table td {
            padding: 0.75rem;
            text-align: left;
            border-bottom: 1px solid var(--gray-200);
        }
        
        .query-table th {
            background: var(--gray-50);
            font-weight: 600;
            color: var(--gray-700);
        }
        
        .query-table tr:hover {
            background: var(--gray-50);
        }
        
        .empty-state {
            text-align: center;
            padding: 3rem;
            color: var(--gray-500);
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <span class="logo">🔍 Peep</span>
                    <span class="tagline">Observability for humans</span>
                </div>
                <nav>
                    <a href="/">Dashboard</a>
                    <a href="/logs">Logs</a>
                    <a href="/query" class="active">Query</a>
                    <a href="/alerts">Alerts</a>
                </nav>
            </div>
        </div>
    </header>

    <div class="container">
        <div class="query-container">
            <div class="query-header">
                <h2>📊 SQL Query Interface</h2>
                <p>Run SQL queries against your log data for custom analytics and insights.</p>
                <div class="query-examples" style="margin-top: 1rem;">
                    <button class="example-query" onclick="setQuery('SELECT COUNT(*) as total_logs FROM logs')">Total Logs</button>
                    <button class="example-query" onclick="setQuery('SELECT level, COUNT(*) as count FROM logs GROUP BY level ORDER BY count DESC')">Logs by Level</button>
                    <button class="example-query" onclick="setQuery('SELECT service, COUNT(*) as count FROM logs WHERE service IS NOT NULL GROUP BY service ORDER BY count DESC LIMIT 10')">Top Services</button>
                    <button class="example-query" onclick="setQuery('SELECT DATE(timestamp) as date, COUNT(*) as logs FROM logs GROUP BY DATE(timestamp) ORDER BY date DESC LIMIT 7')">Daily Log Counts</button>
                    <button class="example-query" onclick="setQuery('SELECT * FROM logs WHERE level = \'error\' ORDER BY timestamp DESC LIMIT 50')">Recent Errors</button>
                    <button class="example-query" onclick="setQuery('SELECT COUNT(*) as http_errors FROM logs WHERE raw_log LIKE \'%\" 5__ %\' AND timestamp > datetime(\'now\', \'-1 hour\')')">HTTP 5xx Errors (1h)</button>
                </div>
            </div>
            <div class="query-form">
                <form hx-post="/query/execute" hx-target="#query-results" hx-indicator="#loading">
                    <textarea name="query" id="query-input" class="query-textarea" placeholder="SELECT * FROM logs WHERE level = 'error' ORDER BY timestamp DESC LIMIT 10"></textarea>
                    <div class="query-actions">
                        <button type="submit" class="btn btn-primary">Execute Query</button>
                        <span id="loading" class="htmx-indicator">⏳ Executing...</span>
                    </div>
                </form>
            </div>
        </div>

        <div class="results-container">
            <div class="results-header">
                <h3>Query Results</h3>
            </div>
            <div class="results-content">
                <div id="query-results" class="empty-state">
                    <div style="font-size: 3rem; margin-bottom: 1rem;">📊</div>
                    <h3>Ready to query</h3>
                    <p>Enter a SQL query above and click "Execute Query" to see results.</p>
                </div>
            </div>
        </div>
    </div>

    <script>
        function setQuery(query) {
            document.getElementById('query-input').value = query;
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

// handleQueryExecute executes custom SQL queries
func (s *Server) handleQueryExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.FormValue("query")
	if query == "" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="empty-state">
			<div style="font-size: 3rem; margin-bottom: 1rem;">⚠️</div>
			<h3>No query provided</h3>
			<p>Please enter a SQL query to execute.</p>
		</div>`))
		return
	}

	// Execute the query
	db := s.storage.GetDB()
	rows, err := db.Query(query)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf(`<div style="color: var(--danger); padding: 1rem; background: #fee2e2; border-radius: 0.375rem;">
			❌ Query Error: %s
		</div>`, err.Error())))
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare to scan results
	var results [][]interface{}
	for rows.Next() {
		// Create a slice of interfaces to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Convert to strings for display
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		results = append(results, row)
	}

	// Generate HTML table
	if len(results) == 0 {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="empty-state">
			<div style="font-size: 3rem; margin-bottom: 1rem;">📊</div>
			<h3>No results</h3>
			<p>The query executed successfully but returned no rows.</p>
		</div>`))
		return
	}

	html := `<div style="margin-bottom: 1rem; color: var(--success);">
		✅ Query executed successfully - ` + fmt.Sprintf("%d", len(results)) + ` rows returned
	</div>
	<div style="overflow-x: auto;">
		<table class="query-table">
			<thead>
				<tr>`

	// Add column headers
	for _, col := range columns {
		html += fmt.Sprintf("<th>%s</th>", col)
	}
	html += "</tr></thead><tbody>"

	// Add data rows
	for _, row := range results {
		html += "<tr>"
		for _, val := range row {
			html += fmt.Sprintf("<td>%v</td>", val)
		}
		html += "</tr>"
	}

	html += "</tbody></table></div>"

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
