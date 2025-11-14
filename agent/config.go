package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AgentID      string        `yaml:"agent_id"`
	PushEndpoint string        `yaml:"push_endpoint"`
	PushInterval time.Duration `yaml:"push_interval"`
	Targets      Targets       `yaml:"targets"`
}

type Targets struct {
	HTTP     []HTTPTarget     `yaml:"http"`
	DNS      []DNSTarget      `yaml:"dns"`
	SMTP     []SMTPTarget     `yaml:"smtp"`
	ICMP     []ICMPTarget     `yaml:"icmp"`
	Postgres []PostgresTarget `yaml:"postgres"`
}

type HTTPTarget struct {
	Name    string        `yaml:"name"`
	URL     string        `yaml:"url"`
	Method  string        `yaml:"method"`
	Timeout time.Duration `yaml:"timeout"`
}

type DNSTarget struct {
	Name   string `yaml:"name"`
	FQDN   string `yaml:"fqdn"`
	Server string `yaml:"server"`
}

type SMTPTarget struct {
	Name     string        `yaml:"name"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	StartTLS bool          `yaml:"starttls"`
	Timeout  time.Duration `yaml:"timeout"`
}

type ICMPTarget struct {
	Name    string        `yaml:"name"`
	Host    string        `yaml:"host"`
	Timeout time.Duration `yaml:"timeout"`
}

type PostgresTarget struct {
	Name    string `yaml:"name"`
	DSN     string `yaml:"dsn"`
	SlowMS  int    `yaml:"slow_ms"`
	PingSQL string `yaml:"ping_sql"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.PushInterval == 0 {
		cfg.PushInterval = 10 * time.Second
	}

	for i := range cfg.Targets.HTTP {
		if cfg.Targets.HTTP[i].Method == "" {
			cfg.Targets.HTTP[i].Method = "GET"
		}
		if cfg.Targets.HTTP[i].Timeout == 0 {
			cfg.Targets.HTTP[i].Timeout = 5 * time.Second
		}
	}

	for i := range cfg.Targets.SMTP {
		if cfg.Targets.SMTP[i].Timeout == 0 {
			cfg.Targets.SMTP[i].Timeout = 5 * time.Second
		}
	}

	for i := range cfg.Targets.ICMP {
		if cfg.Targets.ICMP[i].Timeout == 0 {
			cfg.Targets.ICMP[i].Timeout = 2 * time.Second
		}
	}

	for i := range cfg.Targets.Postgres {
		if cfg.Targets.Postgres[i].PingSQL == "" {
			cfg.Targets.Postgres[i].PingSQL = "SELECT 1"
		}
		if cfg.Targets.Postgres[i].SlowMS == 0 {
			cfg.Targets.Postgres[i].SlowMS = 100
		}
	}

	return &cfg, nil
}
