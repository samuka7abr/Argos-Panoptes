package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AlertState struct {
	FiredAt  time.Time
	Notified bool
	Value    float64
}

var activeAlerts = make(map[string]*AlertState)

func main() {
	rulesPath := os.Getenv("RULES_PATH")
	if rulesPath == "" {
		rulesPath = "rules.yaml"
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8081"
	}

	cfg, err := LoadRules(rulesPath)
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	log.Printf("Loaded %d alert rules", len(cfg.Rules))

	evaluator := NewEvaluator(apiURL)
	notifier := NewNotifier(cfg.Email)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("Alert engine started, evaluating rules every 30s...")

	evaluateAllRules(ctx, cfg.Rules, evaluator, notifier)

	for {
		select {
		case <-ticker.C:
			evaluateAllRules(ctx, cfg.Rules, evaluator, notifier)

		case <-sigChan:
			log.Println("Shutting down alert engine...")
			return
		}
	}
}

func evaluateAllRules(ctx context.Context, rules []Rule, evaluator *Evaluator, notifier *Notifier) {
	for _, rule := range rules {
		go func(r Rule) {
			if err := evaluateRule(ctx, &r, evaluator, notifier); err != nil {
				log.Printf("Error evaluating rule %s: %v", r.Name, err)
			}
		}(rule)
	}
}

func evaluateRule(ctx context.Context, rule *Rule, evaluator *Evaluator, notifier *Notifier) error {
	triggered, value, err := evaluator.Evaluate(ctx, rule)
	if err != nil {
		return err
	}

	alertKey := rule.Name

	if triggered {
		state, exists := activeAlerts[alertKey]

		if !exists {
			activeAlerts[alertKey] = &AlertState{
				FiredAt:  time.Now(),
				Notified: false,
				Value:    value,
			}
			log.Printf("[%s] Alert triggered: %s (value: %.2f)", rule.Severity, rule.Name, value)
			return nil
		}

		duration := time.Since(state.FiredAt)
		forDuration := rule.ForDuration()

		if duration >= forDuration && !state.Notified {
			log.Printf("[%s] Alert firing: %s (value: %.2f, duration: %s)",
				rule.Severity, rule.Name, value, duration)

			if err := notifier.SendAlert(rule, value); err != nil {
				log.Printf("Failed to send alert notification: %v", err)
				return err
			}

			state.Notified = true
			state.Value = value
			log.Printf("Alert notification sent for: %s", rule.Name)
		}

	} else {
		if state, exists := activeAlerts[alertKey]; exists {
			if state.Notified {
				log.Printf("Alert resolved: %s", rule.Name)
			}
			delete(activeAlerts, alertKey)
		}
	}

	return nil
}
