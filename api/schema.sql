CREATE TABLE IF NOT EXISTS metrics (
    ts TIMESTAMPTZ NOT NULL,
    service TEXT NOT NULL,
    target TEXT NOT NULL,
    name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    agent_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics (ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_by_name ON metrics (name, service, target, ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_by_service ON metrics (service, ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_labels ON metrics USING GIN (labels);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    rule TEXT NOT NULL,
    severity TEXT NOT NULL,
    service TEXT NOT NULL,
    target TEXT NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    message TEXT,
    fired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    notified BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts (resolved_at) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts (severity, fired_at DESC);

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER REFERENCES alerts(id),
    channel TEXT NOT NULL,
    recipient TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL,
    error TEXT
);

CREATE OR REPLACE VIEW latest_metrics AS
SELECT DISTINCT ON (service, target, name)
    ts,
    service,
    target,
    name,
    value,
    labels,
    agent_id
FROM metrics
ORDER BY service, target, name, ts DESC;

CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    expr TEXT NOT NULL,
    service TEXT,
    target TEXT,
    for_duration TEXT NOT NULL DEFAULT '1m',
    severity TEXT NOT NULL DEFAULT 'warning',
    email_to JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules (enabled);

CREATE OR REPLACE FUNCTION cleanup_old_metrics()
RETURNS void AS $$
BEGIN
    DELETE FROM metrics WHERE ts < NOW() - INTERVAL '30 days';
    DELETE FROM alerts WHERE resolved_at < NOW() - INTERVAL '90 days';
    DELETE FROM notifications WHERE sent_at < NOW() - INTERVAL '90 days';
    DELETE FROM security_events WHERE created_at < NOW() - INTERVAL '90 days';
    DELETE FROM failed_logins WHERE created_at < NOW() - INTERVAL '30 days';
    DELETE FROM config_changes WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Tabelas de Segurança
CREATE TABLE IF NOT EXISTS security_events (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    severity TEXT NOT NULL,
    description TEXT NOT NULL,
    service TEXT,
    target TEXT,
    ip_address TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_created ON security_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events (severity, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events (type, created_at DESC);

CREATE TABLE IF NOT EXISTS failed_logins (
    id SERIAL PRIMARY KEY,
    ip_address TEXT NOT NULL,
    username TEXT,
    service TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_failed_logins_ip ON failed_logins (ip_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_failed_logins_created ON failed_logins (created_at DESC);

CREATE TABLE IF NOT EXISTS config_changes (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL,
    change_type TEXT NOT NULL,
    old_hash TEXT,
    new_hash TEXT,
    service TEXT,
    detected_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_config_changes_detected ON config_changes (detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_config_changes_file ON config_changes (file_path, detected_at DESC);

CREATE TABLE IF NOT EXISTS vulnerabilities (
    id SERIAL PRIMARY KEY,
    service TEXT NOT NULL,
    cve TEXT,
    severity TEXT NOT NULL,
    description TEXT,
    version TEXT,
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vulnerabilities_service ON vulnerabilities (service, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities (severity, resolved_at);
