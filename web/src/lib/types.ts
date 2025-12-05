export interface Metric {
  service: string;
  target: string;
  metric_name: string;
  value: number;
  timestamp: string;
  labels?: Record<string, string>;
}

export interface DataPoint {
  timestamp: string;
  value: number;
}

export interface QueryRangeResponse {
  service: string;
  target: string;
  metric_name: string;
  data: DataPoint[];
}

export interface LatestMetric {
  service: string;
  target: string;
  metrics: Record<string, number>;
  timestamp: string;
}

export interface AlertRule {
  id?: number;
  name: string;
  description?: string;
  expr: string;
  service?: string;
  target?: string;
  for_duration: string;
  severity: "info" | "warning" | "critical";
  email_to: string[];
  enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AlertRulesResponse {
  rules: AlertRule[];
  count: number;
}

export interface ActiveAlert {
  rule_name: string;
  service: string;
  target: string;
  severity: string;
  message: string;
  triggered_at: string;
}

export interface HealthResponse {
  status: string;
  uptime: number;
  metrics_count: number;
  services: string[];
  targets_count: number;
  last_ingest: string;
  active_alerts: ActiveAlert[];
}

export type ServiceType = "http" | "https" | "postgres" | "dns" | "smtp" | "icmp";

export type Status = "ok" | "warning" | "critical";

// Security types
export interface SecurityEvent {
  id: number;
  type: string;
  severity: string;
  description: string;
  service?: string;
  target?: string;
  ip_address?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface FailedLoginByIP {
  ip_address: string;
  count: number;
}

export interface ConfigChange {
  id: number;
  file_path: string;
  change_type: string;
  old_hash?: string;
  new_hash?: string;
  service?: string;
  detected_at: string;
}

export interface Vulnerability {
  id: number;
  service: string;
  cve?: string;
  severity: string;
  description?: string;
  version?: string;
  detected_at: string;
  resolved_at?: string;
}



