package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"argos/agent/probes"
	"argos/shared"
)

type Probe interface {
	Collect(ctx context.Context) []shared.Metric
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Argos Agent: %s", cfg.AgentID)
	log.Printf("Pushing metrics to: %s every %s", cfg.PushEndpoint, cfg.PushInterval)

	pusher := shared.NewPusher(cfg.PushEndpoint)

	probeList := createProbes(cfg)
	log.Printf("Initialized %d probes", len(probeList))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(signal.Channel, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.PushInterval)
	defer ticker.Stop()

	log.Println("Agent started, collecting metrics...")

	for {
		select {
		case <-ticker.C:
			metrics := collectAllMetrics(ctx, probeList)

			if len(metrics) > 0 {
				if err := pusher.Push(cfg.AgentID, metrics); err != nil {
					log.Printf("Failed to push metrics: %v", err)
				} else {
					log.Printf("Pushed %d metrics successfully", len(metrics))
				}
			}

		case <-sigChan:
			log.Println("Shutting down agent...")
			return
		}
	}
}

func createProbes(cfg *Config) []Probe {
	var probeList []Probe

	for _, target := range cfg.Targets.HTTP {
		p := probes.NewHTTPProbe(target.Name, target.URL, target.Method, target.Timeout)
		probeList = append(probeList, p)
		log.Printf("  HTTP probe: %s -> %s", target.Name, target.URL)
	}

	for _, target := range cfg.Targets.DNS {
		p := probes.NewDNSProbe(target.Name, target.FQDN, target.Server)
		probeList = append(probeList, p)
		log.Printf("  DNS probe: %s -> %s @ %s", target.Name, target.FQDN, target.Server)
	}

	for _, target := range cfg.Targets.SMTP {
		p := probes.NewSMTPProbe(target.Name, target.Host, target.Port, target.StartTLS, target.Timeout)
		probeList = append(probeList, p)
		log.Printf("  SMTP probe: %s -> %s:%d", target.Name, target.Host, target.Port)
	}

	for _, target := range cfg.Targets.ICMP {
		p := probes.NewICMPProbe(target.Name, target.Host, target.Timeout)
		probeList = append(probeList, p)
		log.Printf("  ICMP probe: %s -> %s", target.Name, target.Host)
	}

	for _, target := range cfg.Targets.Postgres {
		p := probes.NewPostgresProbe(target.Name, target.DSN, target.SlowMS, target.PingSQL)
		probeList = append(probeList, p)
		log.Printf("  Postgres probe: %s", target.Name)
	}

	return probeList
}

func collectAllMetrics(ctx context.Context, probeList []Probe) []shared.Metric {
	var wg sync.WaitGroup
	metricsChan := make(chan []shared.Metric, len(probeList))

	for _, probe := range probeList {
		wg.Add(1)
		go func(p Probe) {
			defer wg.Done()
			metrics := p.Collect(ctx)
			metricsChan <- metrics
		}(probe)
	}

	wg.Wait()
	close(metricsChan)

	var allMetrics []shared.Metric
	for metrics := range metricsChan {
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics
}
