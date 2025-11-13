package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"argos/shared"
)

type Storage struct {
	db *sql.DB
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
