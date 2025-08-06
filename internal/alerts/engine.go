package alerts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kylereynolds/peep/internal/notifications"
	"github.com/kylereynolds/peep/internal/storage"
)

// AlertRule represents a SQL-based alert rule
type AlertRule struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Query       string    `json:"query"`     // SQL query that returns count
	Threshold   int       `json:"threshold"` // Alert if count >= threshold
	Window      string    `json:"window"`    // Time window (e.g., "5m", "1h")
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	LastCheck   time.Time `json:"last_check"`
	LastAlert   time.Time `json:"last_alert"`
}

// AlertInstance represents a triggered alert
type AlertInstance struct {
	ID        int64     `json:"id"`
	RuleID    int64     `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Count     int       `json:"count"`
	Threshold int       `json:"threshold"`
	Query     string    `json:"query"`
	FiredAt   time.Time `json:"fired_at"`
	Resolved  bool      `json:"resolved"`
}

// NotificationChannel represents a way to send alerts
type NotificationChannel struct {
	ID      int64             `json:"id"`
	Name    string            `json:"name"`
	Type    string            `json:"type"` // "desktop", "slack", "email", "shell"
	Config  map[string]string `json:"config"`
	Enabled bool              `json:"enabled"`
}

// Engine manages alert rules and notifications
type Engine struct {
	storage   *storage.Storage
	db        *sql.DB
	rules     map[int64]*AlertRule
	channels  map[int64]*NotificationChannel
	stopChan  chan struct{}
	isRunning bool
}

// NewEngine creates a new alert engine
func NewEngine(store *storage.Storage) (*Engine, error) {
	engine := &Engine{
		storage:  store,
		db:       store.GetDB(),
		rules:    make(map[int64]*AlertRule),
		channels: make(map[int64]*NotificationChannel),
		stopChan: make(chan struct{}),
	}

	if err := engine.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create alert tables: %w", err)
	}

	if err := engine.loadRules(); err != nil {
		return nil, fmt.Errorf("failed to load alert rules: %w", err)
	}

	if err := engine.loadChannels(); err != nil {
		return nil, fmt.Errorf("failed to load notification channels: %w", err)
	}

	// Create default desktop notification channel if none exist
	if len(engine.channels) == 0 {
		defaultChannel := &NotificationChannel{
			Name:    "Desktop Notifications",
			Type:    "desktop",
			Config:  map[string]string{},
			Enabled: true,
		}
		if err := engine.AddNotificationChannel(defaultChannel); err != nil {
			return nil, fmt.Errorf("failed to create default notification channel: %w", err)
		}
	}

	return engine, nil
}

// createTables creates the necessary database tables
func (e *Engine) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS alert_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		query TEXT NOT NULL,
		threshold INTEGER NOT NULL DEFAULT 1,
		window TEXT NOT NULL DEFAULT '5m',
		enabled BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_check DATETIME,
		last_alert DATETIME
	);

	CREATE TABLE IF NOT EXISTS alert_instances (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rule_id INTEGER NOT NULL,
		rule_name TEXT NOT NULL,
		count INTEGER NOT NULL,
		threshold INTEGER NOT NULL,
		query TEXT NOT NULL,
		fired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved BOOLEAN DEFAULT 0,
		FOREIGN KEY (rule_id) REFERENCES alert_rules (id)
	);

	CREATE TABLE IF NOT EXISTS notification_channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		config TEXT NOT NULL, -- JSON
		enabled BOOLEAN NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS alert_notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		alert_id INTEGER NOT NULL,
		channel_id INTEGER NOT NULL,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		success BOOLEAN NOT NULL,
		error_message TEXT,
		FOREIGN KEY (alert_id) REFERENCES alert_instances (id),
		FOREIGN KEY (channel_id) REFERENCES notification_channels (id)
	);

	CREATE INDEX IF NOT EXISTS idx_alert_instances_rule_id ON alert_instances(rule_id);
	CREATE INDEX IF NOT EXISTS idx_alert_instances_fired_at ON alert_instances(fired_at);
	`

	_, err := e.db.Exec(schema)
	return err
}

// AddRule adds a new alert rule
func (e *Engine) AddRule(rule *AlertRule) error {
	query := `
	INSERT INTO alert_rules (name, description, query, threshold, window, enabled)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := e.db.Exec(query, rule.Name, rule.Description, rule.Query, rule.Threshold, rule.Window, rule.Enabled)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	rule.ID = id
	rule.CreatedAt = time.Now()
	e.rules[id] = rule

	return nil
}

// GetChannels returns all notification channels
func (e *Engine) GetChannels() []*NotificationChannel {
	channels := make([]*NotificationChannel, 0, len(e.channels))
	for _, channel := range e.channels {
		channels = append(channels, channel)
	}
	return channels
}

// GetRules returns all alert rules
func (e *Engine) GetRules() []*AlertRule {
	rules := make([]*AlertRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// AddNotificationChannel adds a new notification channel
func (e *Engine) AddNotificationChannel(channel *NotificationChannel) error {
	configJSON, err := json.Marshal(channel.Config)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO notification_channels (name, type, config, enabled)
	VALUES (?, ?, ?, ?)
	`

	result, err := e.db.Exec(query, channel.Name, channel.Type, string(configJSON), channel.Enabled)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	channel.ID = id
	e.channels[id] = channel

	return nil
}

// loadRules loads all alert rules from the database
func (e *Engine) loadRules() error {
	query := `
	SELECT id, name, description, query, threshold, window, enabled, created_at, last_check, last_alert
	FROM alert_rules
	`

	rows, err := e.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		rule := &AlertRule{}
		var lastCheck, lastAlert sql.NullTime

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Query,
			&rule.Threshold, &rule.Window, &rule.Enabled, &rule.CreatedAt,
			&lastCheck, &lastAlert,
		)
		if err != nil {
			return err
		}

		if lastCheck.Valid {
			rule.LastCheck = lastCheck.Time
		}
		if lastAlert.Valid {
			rule.LastAlert = lastAlert.Time
		}

		e.rules[rule.ID] = rule
	}

	return nil
}

// loadChannels loads all notification channels from the database
func (e *Engine) loadChannels() error {
	query := `
	SELECT id, name, type, config, enabled
	FROM notification_channels
	`

	rows, err := e.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		channel := &NotificationChannel{}
		var configJSON string

		err := rows.Scan(&channel.ID, &channel.Name, &channel.Type, &configJSON, &channel.Enabled)
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(configJSON), &channel.Config); err != nil {
			return err
		}

		e.channels[channel.ID] = channel
	}

	return nil
}

// Start begins the alert monitoring loop
func (e *Engine) Start() {
	if e.isRunning {
		return
	}

	e.isRunning = true
	go e.monitorLoop()
}

// Stop stops the alert monitoring
func (e *Engine) Stop() {
	if !e.isRunning {
		return
	}

	e.stopChan <- struct{}{}
	e.isRunning = false
}

// monitorLoop runs the alert checking loop
func (e *Engine) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.checkAlerts()
		case <-e.stopChan:
			return
		}
	}
}

// checkAlerts evaluates all enabled alert rules
func (e *Engine) checkAlerts() {
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		if err := e.evaluateRule(rule); err != nil {
			fmt.Printf("Error evaluating rule %s: %v\n", rule.Name, err)
		}
	}
}

// evaluateRule checks a single alert rule
func (e *Engine) evaluateRule(rule *AlertRule) error {
	// Parse time window and create time-bounded query
	timeQuery := e.buildTimeQuery(rule.Query, rule.Window)

	var count int
	err := e.db.QueryRow(timeQuery).Scan(&count)
	if err != nil {
		return err
	}

	// Update last check time
	rule.LastCheck = time.Now()
	e.updateRuleLastCheck(rule)

	// Check if threshold is exceeded
	if count >= rule.Threshold {
		return e.fireAlert(rule, count)
	}

	return nil
}

// buildTimeQuery adds time window constraints to the alert query
func (e *Engine) buildTimeQuery(query, window string) string {
	// Parse window duration
	duration, err := time.ParseDuration(window)
	if err != nil {
		duration = 5 * time.Minute // Default to 5 minutes
	}

	since := time.Now().Add(-duration).Format("2006-01-02 15:04:05")

	// Add time constraint to the query
	if !containsWhere(query) {
		return query + fmt.Sprintf(" WHERE timestamp >= '%s'", since)
	} else {
		return query + fmt.Sprintf(" AND timestamp >= '%s'", since)
	}
} // containsWhere checks if query already has a WHERE clause
func containsWhere(query string) bool {
	return strings.Contains(strings.ToUpper(query), "WHERE")
}

// fireAlert creates an alert instance and sends notifications
func (e *Engine) fireAlert(rule *AlertRule, count int) error {
	// Create alert instance
	instance := &AlertInstance{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Count:     count,
		Threshold: rule.Threshold,
		Query:     rule.Query,
		FiredAt:   time.Now(),
	}

	if err := e.saveAlertInstance(instance); err != nil {
		return err
	}

	// Update rule last alert time
	rule.LastAlert = time.Now()
	e.updateRuleLastAlert(rule)

	// Send notifications to all enabled channels
	for _, channel := range e.channels {
		if channel.Enabled {
			e.sendNotification(instance, channel)
		}
	}

	return nil
}

// saveAlertInstance saves an alert instance to the database
func (e *Engine) saveAlertInstance(instance *AlertInstance) error {
	query := `
	INSERT INTO alert_instances (rule_id, rule_name, count, threshold, query, fired_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := e.db.Exec(query, instance.RuleID, instance.RuleName, instance.Count, instance.Threshold, instance.Query, instance.FiredAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	instance.ID = id
	return nil
}

// updateRuleLastCheck updates the last check time for a rule
func (e *Engine) updateRuleLastCheck(rule *AlertRule) {
	query := `UPDATE alert_rules SET last_check = ? WHERE id = ?`
	e.db.Exec(query, rule.LastCheck, rule.ID)
}

// updateRuleLastAlert updates the last alert time for a rule
func (e *Engine) updateRuleLastAlert(rule *AlertRule) {
	query := `UPDATE alert_rules SET last_alert = ? WHERE id = ?`
	e.db.Exec(query, rule.LastAlert, rule.ID)
}

// sendNotification sends an alert to a notification channel
func (e *Engine) sendNotification(instance *AlertInstance, channel *NotificationChannel) {
	var err error

	switch channel.Type {
	case "desktop":
		err = e.sendDesktopNotification(instance, channel)
	case "slack":
		err = e.sendSlackNotification(instance, channel)
	case "email":
		err = e.sendEmailNotification(instance, channel)
	case "shell":
		err = e.sendShellNotification(instance, channel)
	default:
		err = fmt.Errorf("unknown notification type: %s", channel.Type)
	}

	// Log notification result
	e.logNotification(instance.ID, channel.ID, err == nil, err)
}

// sendDesktopNotification sends a desktop notification
func (e *Engine) sendDesktopNotification(instance *AlertInstance, channel *NotificationChannel) error {
	title := fmt.Sprintf("🚨 Peep Alert: %s", instance.RuleName)
	message := fmt.Sprintf("Threshold exceeded: %d events (limit: %d)", instance.Count, instance.Threshold)

	if err := notifications.SendDesktopNotification(title, message); err != nil {
		// Fallback to console if desktop notification fails
		fmt.Printf("🚨 ALERT: %s - Count: %d (threshold: %d)\n", instance.RuleName, instance.Count, instance.Threshold)
		return err
	}

	fmt.Printf("🚨 ALERT: %s - Count: %d (threshold: %d) [Desktop notification sent]\n", instance.RuleName, instance.Count, instance.Threshold)
	return nil
}

// sendSlackNotification sends a Slack notification
func (e *Engine) sendSlackNotification(instance *AlertInstance, channel *NotificationChannel) error {
	webhookURL, exists := channel.Config["webhook_url"]
	if !exists {
		return fmt.Errorf("slack channel missing webhook_url in config")
	}

	title := instance.RuleName
	message := fmt.Sprintf("Alert threshold exceeded: **%d events** detected (limit: %d)", instance.Count, instance.Threshold)

	if err := notifications.SendSlackNotification(webhookURL, title, message, instance.Count, instance.Threshold); err != nil {
		fmt.Printf("❌ Failed to send Slack notification: %v\n", err)
		return err
	}

	fmt.Printf("📱 Slack notification sent: %s [%d/%d]\n", instance.RuleName, instance.Count, instance.Threshold)
	return nil
}

// sendEmailNotification sends an email notification
func (e *Engine) sendEmailNotification(instance *AlertInstance, channel *NotificationChannel) error {
	// Extract email configuration from channel config
	emailConfig := notifications.EmailConfig{
		SMTPHost:  channel.Config["smtp_host"],
		Username:  channel.Config["username"],
		Password:  channel.Config["password"],
		FromEmail: channel.Config["from_email"],
		FromName:  channel.Config["from_name"],
		ToEmails:  strings.Split(channel.Config["to_emails"], ","),
	}

	// Parse SMTP port
	if portStr, exists := channel.Config["smtp_port"]; exists {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			emailConfig.SMTPPort = port
		} else {
			emailConfig.SMTPPort = 587 // Default SMTP port
		}
	} else {
		emailConfig.SMTPPort = 587
	}

	// Clean up email addresses (trim spaces)
	for i, email := range emailConfig.ToEmails {
		emailConfig.ToEmails[i] = strings.TrimSpace(email)
	}

	emailNotifier := notifications.NewEmailNotification(emailConfig)

	title := fmt.Sprintf("Alert: %s", instance.RuleName)
	message := fmt.Sprintf("Alert threshold exceeded!\n\nRule: %s\nQuery: %s\nCount: %d\nThreshold: %d\nTime: %s",
		instance.RuleName,
		instance.Query,
		instance.Count,
		instance.Threshold,
		instance.FiredAt.Format("2006-01-02 15:04:05"),
	)

	severity := "warning"
	if instance.Count >= instance.Threshold*2 {
		severity = "critical"
	}

	if err := emailNotifier.Send(title, message, severity); err != nil {
		fmt.Printf("❌ Failed to send email notification: %v\n", err)
		return err
	}

	fmt.Printf("📧 Email notification sent: %s\n", instance.RuleName)
	return nil
}

// sendShellNotification executes a shell script
func (e *Engine) sendShellNotification(instance *AlertInstance, channel *NotificationChannel) error {
	scriptPath, exists := channel.Config["script_path"]
	if !exists {
		return fmt.Errorf("shell channel missing script_path in config")
	}

	// Parse timeout (optional)
	timeout := 30 * time.Second
	if timeoutStr, exists := channel.Config["timeout"]; exists {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	// Parse args (optional)
	var args []string
	if argsStr, exists := channel.Config["args"]; exists && argsStr != "" {
		args = strings.Split(argsStr, " ")
	}

	// Parse working directory (optional)
	workingDir := channel.Config["working_dir"]

	// Parse custom environment variables (optional)
	environment := make(map[string]string)
	if envStr, exists := channel.Config["environment"]; exists && envStr != "" {
		// Parse environment as comma-separated KEY=VALUE pairs
		for _, pair := range strings.Split(envStr, ",") {
			if parts := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(parts) == 2 {
				environment[parts[0]] = parts[1]
			}
		}
	}

	shellConfig := notifications.ShellConfig{
		ScriptPath:  scriptPath,
		Args:        args,
		Timeout:     timeout,
		WorkingDir:  workingDir,
		Environment: environment,
	}

	shellNotifier := notifications.NewShellNotification(shellConfig)

	title := instance.RuleName
	message := fmt.Sprintf("Alert threshold exceeded!\n\nRule: %s\nQuery: %s\nCount: %d\nThreshold: %d\nTime: %s",
		instance.RuleName,
		instance.Query,
		instance.Count,
		instance.Threshold,
		instance.FiredAt.Format("2006-01-02 15:04:05"),
	)

	severity := "warning"
	if instance.Count >= instance.Threshold*2 {
		severity = "critical"
	}

	if err := shellNotifier.Execute(title, message, severity, instance.Count, instance.Threshold); err != nil {
		fmt.Printf("❌ Failed to execute shell notification: %v\n", err)
		return err
	}

	fmt.Printf("🖥️  Shell script executed: %s [%s]\n", instance.RuleName, scriptPath)
	return nil
}

// logNotification logs the result of sending a notification
func (e *Engine) logNotification(alertID, channelID int64, success bool, err error) {
	query := `
	INSERT INTO alert_notifications (alert_id, channel_id, success, error_message)
	VALUES (?, ?, ?, ?)
	`

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	e.db.Exec(query, alertID, channelID, success, errorMsg)
}
