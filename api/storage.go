package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"argos/shared"

	_ "github.com/lib/pq"
)

type StorageInterface interface {
	InsertMetrics(agentID string, metrics []shared.Metric) error
	QueryLatest(name, service, target string) (*shared.Metric, error)
	QueryRange(name, service, target string, start, end time.Time, step string) ([]shared.DataPoint, error)
	GetLatestMetrics() ([]shared.Metric, error)
	ListServices() ([]string, error)
	ListTargets(service string) ([]string, error)
	GetMetricsCount() (int64, error)
	GetLastIngestTime() (time.Time, error)
	GetActiveAlerts() ([]shared.Alert, error)
	GetAlertRules() ([]AlertRule, error)
	GetAlertRule(id int) (*AlertRule, error)
	CreateAlertRule(rule *AlertRule) error
	UpdateAlertRule(rule *AlertRule) error
	DeleteAlertRule(id int) error
	// Security methods
	GetSecurityEvents(limit int) ([]SecurityEvent, error)
	CreateSecurityEvent(event *SecurityEvent) error
	GetFailedLoginsByIP(limit int) ([]struct {
		IPAddress string `json:"ip_address"`
		Count     int    `json:"count"`
	}, error)
	GetTotalFailedLogins() (int64, error)
	RecordFailedLogin(ip, username, service, userAgent string) error
	GetConfigChanges(limit int) ([]ConfigChange, error)
	RecordConfigChange(change *ConfigChange) error
	GetVulnerabilities() ([]Vulnerability, error)
	GetTrafficAnomalies(limit int) (int64, error)
	Close() error
}

type Storage struct {
	db *sql.DB
}

// Expose db for direct access when needed (for vulnerabilities)
func (s *Storage) DB() *sql.DB {
	return s.db
}

func NewStorage(databaseURL string) (*Storage, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) InsertMetrics(agentID string, metrics []shared.Metric) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO metrics (ts, service, target, name, value, labels, agent_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range metrics {
		labelsJSON, _ := json.Marshal(m.Labels)
		_, err := stmt.Exec(m.TS, m.Service, m.Target, m.Name, m.Value, labelsJSON, agentID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) QueryLatest(name, service, target string) (*shared.Metric, error) {
	query := `
		SELECT ts, service, target, name, value, labels
		FROM metrics
		WHERE name = $1
	`
	args := []interface{}{name}

	if service != "" {
		query += " AND service = $2"
		args = append(args, service)
	}
	if target != "" {
		if service != "" {
			query += " AND target = $3"
		} else {
			query += " AND target = $2"
		}
		args = append(args, target)
	}

	query += " ORDER BY ts DESC LIMIT 1"

	var m shared.Metric
	var labelsJSON []byte

	err := s.db.QueryRow(query, args...).Scan(
		&m.TS, &m.Service, &m.Target, &m.Name, &m.Value, &labelsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(labelsJSON, &m.Labels)
	return &m, nil
}

func (s *Storage) QueryRange(name, service, target string, start, end time.Time, step string) ([]shared.DataPoint, error) {
	query := `
		SELECT
			date_trunc('minute', ts) AS bucket,
			AVG(value) AS avg_value
		FROM metrics
		WHERE name = $1
			AND ts >= $2
			AND ts <= $3
	`
	args := []interface{}{name, start, end}

	if service != "" {
		query += " AND service = $4"
		args = append(args, service)
	}
	if target != "" {
		if service != "" {
			query += " AND target = $5"
		} else {
			query += " AND target = $4"
		}
		args = append(args, target)
	}

	query += " GROUP BY bucket ORDER BY bucket ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []shared.DataPoint
	for rows.Next() {
		var ts time.Time
		var value float64
		if err := rows.Scan(&ts, &value); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, shared.DataPoint{
			Timestamp: ts.Unix(),
			Value:     value,
		})
	}

	return dataPoints, nil
}

func (s *Storage) GetLatestMetrics() ([]shared.Metric, error) {
	rows, err := s.db.Query(`
		SELECT name, service, target, value, ts, labels
		FROM latest_metrics
		ORDER BY service, target, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []shared.Metric
	for rows.Next() {
		var m shared.Metric
		var labelsJSON []byte
		if err := rows.Scan(&m.Name, &m.Service, &m.Target, &m.Value, &m.TS, &labelsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal(labelsJSON, &m.Labels)
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (s *Storage) ListServices() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT service FROM metrics ORDER BY service
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var svc string
		if err := rows.Scan(&svc); err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	return services, nil
}

func (s *Storage) ListTargets(service string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT target FROM metrics WHERE service = $1 ORDER BY target
	`, service)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []string
	for rows.Next() {
		var tgt string
		if err := rows.Scan(&tgt); err != nil {
			return nil, err
		}
		targets = append(targets, tgt)
	}

	return targets, nil
}

func (s *Storage) GetMetricsCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM metrics").Scan(&count)
	return count, err
}

func (s *Storage) GetLastIngestTime() (time.Time, error) {
	var ts time.Time
	err := s.db.QueryRow("SELECT MAX(ts) FROM metrics").Scan(&ts)
	return ts, err
}

func (s *Storage) GetActiveAlerts() ([]shared.Alert, error) {
	rows, err := s.db.Query(`
		SELECT name, rule, severity, service, target, labels, message, fired_at
		FROM alerts
		WHERE resolved_at IS NULL
		ORDER BY fired_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []shared.Alert
	for rows.Next() {
		var a shared.Alert
		var labelsJSON []byte
		if err := rows.Scan(&a.Name, &a.Rule, &a.Severity, &a.Service, &a.Target, &labelsJSON, &a.Message, &a.Since); err != nil {
			return nil, err
		}
		json.Unmarshal(labelsJSON, &a.Labels)
		alerts = append(alerts, a)
	}

	return alerts, nil
}

type AlertRule struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Expr        string    `json:"expr"`
	Service     string    `json:"service"`
	Target      string    `json:"target"`
	ForDuration string    `json:"for_duration"`
	Severity    string    `json:"severity"`
	EmailTo     []string  `json:"email_to"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Storage) GetAlertRules() ([]AlertRule, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, expr, service, target, for_duration, 
		       severity, email_to, enabled, created_at, updated_at
		FROM alert_rules
		WHERE enabled = true
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		var emailJSON []byte
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Expr, &r.Service,
			&r.Target, &r.ForDuration, &r.Severity, &emailJSON, &r.Enabled,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(emailJSON, &r.EmailTo)
		rules = append(rules, r)
	}

	return rules, nil
}

func (s *Storage) GetAlertRule(id int) (*AlertRule, error) {
	var r AlertRule
	var emailJSON []byte

	err := s.db.QueryRow(`
		SELECT id, name, description, expr, service, target, for_duration,
		       severity, email_to, enabled, created_at, updated_at
		FROM alert_rules
		WHERE id = $1
	`, id).Scan(&r.ID, &r.Name, &r.Description, &r.Expr, &r.Service,
		&r.Target, &r.ForDuration, &r.Severity, &emailJSON, &r.Enabled,
		&r.CreatedAt, &r.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(emailJSON, &r.EmailTo)
	return &r, nil
}

func (s *Storage) CreateAlertRule(rule *AlertRule) error {
	emailJSON, _ := json.Marshal(rule.EmailTo)

	return s.db.QueryRow(`
		INSERT INTO alert_rules (name, description, expr, service, target, 
		                         for_duration, severity, email_to, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, rule.Name, rule.Description, rule.Expr, rule.Service, rule.Target,
		rule.ForDuration, rule.Severity, emailJSON, rule.Enabled).
		Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (s *Storage) UpdateAlertRule(rule *AlertRule) error {
	emailJSON, _ := json.Marshal(rule.EmailTo)

	result, err := s.db.Exec(`
		UPDATE alert_rules
		SET name = $1, description = $2, expr = $3, service = $4, target = $5,
		    for_duration = $6, severity = $7, email_to = $8, enabled = $9,
		    updated_at = NOW()
		WHERE id = $10
	`, rule.Name, rule.Description, rule.Expr, rule.Service, rule.Target,
		rule.ForDuration, rule.Severity, emailJSON, rule.Enabled, rule.ID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

func (s *Storage) DeleteAlertRule(id int) error {
	result, err := s.db.Exec("DELETE FROM alert_rules WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

// Security methods
type SecurityEvent struct {
	ID          int                    `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Service     string                 `json:"service"`
	Target      string                 `json:"target"`
	IPAddress   string                 `json:"ip_address"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

type FailedLogin struct {
	ID        int       `json:"id"`
	IPAddress string    `json:"ip_address"`
	Username  string    `json:"username"`
	Service   string    `json:"service"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type ConfigChange struct {
	ID         int       `json:"id"`
	FilePath   string    `json:"file_path"`
	ChangeType string    `json:"change_type"`
	OldHash    string    `json:"old_hash"`
	NewHash    string    `json:"new_hash"`
	Service    string    `json:"service"`
	DetectedAt time.Time `json:"detected_at"`
}

type Vulnerability struct {
	ID          int        `json:"id"`
	Service     string     `json:"service"`
	CVE         string     `json:"cve"`
	Severity    string     `json:"severity"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	DetectedAt  time.Time  `json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

func (s *Storage) CreateSecurityEvent(event *SecurityEvent) error {
	metadataJSON, _ := json.Marshal(event.Metadata)
	return s.db.QueryRow(`
		INSERT INTO security_events (type, severity, description, service, target, ip_address, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, event.Type, event.Severity, event.Description, event.Service, event.Target, event.IPAddress, metadataJSON).
		Scan(&event.ID, &event.CreatedAt)
}

func (s *Storage) GetSecurityEvents(limit int) ([]SecurityEvent, error) {
	rows, err := s.db.Query(`
		SELECT id, type, severity, description, service, target, ip_address, metadata, created_at
		FROM security_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SecurityEvent
	for rows.Next() {
		var e SecurityEvent
		var metadataJSON []byte
		if err := rows.Scan(&e.ID, &e.Type, &e.Severity, &e.Description, &e.Service, &e.Target, &e.IPAddress, &metadataJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &e.Metadata)
		events = append(events, e)
	}
	return events, nil
}

func (s *Storage) RecordFailedLogin(ip, username, service, userAgent string) error {
	_, err := s.db.Exec(`
		INSERT INTO failed_logins (ip_address, username, service, user_agent)
		VALUES ($1, $2, $3, $4)
	`, ip, username, service, userAgent)
	return err
}

func (s *Storage) GetFailedLoginsByIP(limit int) ([]struct {
	IPAddress string `json:"ip_address"`
	Count     int    `json:"count"`
}, error) {
	rows, err := s.db.Query(`
		SELECT ip_address, COUNT(*) as count
		FROM failed_logins
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY ip_address
		ORDER BY count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		IPAddress string `json:"ip_address"`
		Count     int    `json:"count"`
	}
	for rows.Next() {
		var r struct {
			IPAddress string `json:"ip_address"`
			Count     int    `json:"count"`
		}
		if err := rows.Scan(&r.IPAddress, &r.Count); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (s *Storage) GetTotalFailedLogins() (int64, error) {
	var count int64
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM failed_logins
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&count)
	return count, err
}

func (s *Storage) RecordConfigChange(change *ConfigChange) error {
	return s.db.QueryRow(`
		INSERT INTO config_changes (file_path, change_type, old_hash, new_hash, service)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, detected_at
	`, change.FilePath, change.ChangeType, change.OldHash, change.NewHash, change.Service).
		Scan(&change.ID, &change.DetectedAt)
}

func (s *Storage) GetConfigChanges(limit int) ([]ConfigChange, error) {
	rows, err := s.db.Query(`
		SELECT id, file_path, change_type, old_hash, new_hash, service, detected_at
		FROM config_changes
		ORDER BY detected_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ConfigChange
	for rows.Next() {
		var c ConfigChange
		if err := rows.Scan(&c.ID, &c.FilePath, &c.ChangeType, &c.OldHash, &c.NewHash, &c.Service, &c.DetectedAt); err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	return changes, nil
}

func (s *Storage) GetVulnerabilities() ([]Vulnerability, error) {
	rows, err := s.db.Query(`
		SELECT id, service, cve, severity, description, version, detected_at, resolved_at
		FROM vulnerabilities
		WHERE resolved_at IS NULL
		ORDER BY 
			CASE severity
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
				ELSE 5
			END,
			detected_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vulns []Vulnerability
	for rows.Next() {
		var v Vulnerability
		if err := rows.Scan(&v.ID, &v.Service, &v.CVE, &v.Severity, &v.Description, &v.Version, &v.DetectedAt, &v.ResolvedAt); err != nil {
			return nil, err
		}
		vulns = append(vulns, v)
	}
	return vulns, nil
}

func (s *Storage) GetTrafficAnomalies(limit int) (int64, error) {
	// Contar eventos de anomalia de tráfego nas últimas 24h
	var count int64
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM security_events
		WHERE type IN ('traffic_spike', 'ddos_attack', 'anomaly_detected')
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&count)
	return count, err
}
