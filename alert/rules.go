package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type RulesConfig struct {
	Rules []Rule      `yaml:"rules"`
	Email EmailConfig `yaml:"email"`
}

type Rule struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Expr        string   `yaml:"expr"`
	Service     string   `yaml:"service"`
	Target      string   `yaml:"target"`
	For         string   `yaml:"for"`
	Severity    string   `yaml:"severity"`
	EmailTo     []string `yaml:"email_to"`
}

type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUser     string `yaml:"smtp_user"`
	SMTPPassword string `yaml:"smtp_password"`
	From         string `yaml:"from"`
	UseTLS       bool   `yaml:"use_tls"`
}

func LoadRules(path string) (*RulesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg RulesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for i := range cfg.Rules {
		if cfg.Rules[i].For == "" {
			cfg.Rules[i].For = "1m"
		}
		if cfg.Rules[i].Severity == "" {
			cfg.Rules[i].Severity = "warning"
		}
	}

	if cfg.Email.SMTPPort == 0 {
		cfg.Email.SMTPPort = 587
	}

	return &cfg, nil
}

func (r *Rule) ForDuration() time.Duration {
	d, err := time.ParseDuration(r.For)
	if err != nil {
		return 1 * time.Minute
	}
	return d
}
