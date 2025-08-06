package web

import (
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
	http.HandleFunc("/alerts", s.handleAlerts)
	http.HandleFunc("/alerts/rules", s.handleAlertRules)
	http.HandleFunc("/alerts/rules/add", s.handleAddAlertRule)
	http.HandleFunc("/alerts/channels", s.handleAlertChannels)
	http.HandleFunc("/alerts/channels/add", s.handleAddAlertChannel)
	http.HandleFunc("/api/stats", s.handleAPIStats)

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

// Placeholder handlers - we'll implement these next
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Logs page coming soon with HTMX!"))
}

func (s *Server) handleLogsSearch(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Log search with HTMX coming soon!"))
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alerts management page coming soon!"))
}

func (s *Server) handleAlertRules(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alert rules management coming soon!"))
}

func (s *Server) handleAddAlertRule(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Add alert rule form coming soon!"))
}

func (s *Server) handleAlertChannels(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alert channels management coming soon!"))
}

func (s *Server) handleAddAlertChannel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Add alert channel form coming soon!"))
}
