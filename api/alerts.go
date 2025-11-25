package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func alertsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		if strings.HasPrefix(r.URL.Path, "/api/alert-rules/") {
			getAlertRuleHandler(w, r)
		} else {
			listAlertRulesHandler(w, r)
		}
	case "POST":
		createAlertRuleHandler(w, r)
	case "PUT":
		updateAlertRuleHandler(w, r)
	case "DELETE":
		deleteAlertRuleHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	rules, err := storage.GetAlertRules()
	if err != nil {
		log.Printf("Error getting alert rules: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

func getAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rule, err := storage.GetAlertRule(id)
	if err != nil {
		log.Printf("Error getting alert rule: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rule == nil {
		http.Error(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func createAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if rule.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if rule.Expr == "" {
		http.Error(w, "Expression is required", http.StatusBadRequest)
		return
	}

	if len(rule.EmailTo) == 0 {
		http.Error(w, "At least one email recipient is required", http.StatusBadRequest)
		return
	}

	if rule.ForDuration == "" {
		rule.ForDuration = "1m"
	}

	if rule.Severity == "" {
		rule.Severity = "warning"
	}

	rule.Enabled = true

	if err := storage.CreateAlertRule(&rule); err != nil {
		log.Printf("Error creating alert rule: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "Alert rule with this name already exists", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Created alert rule: %s (ID: %d)", rule.Name, rule.ID)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func updateAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	rule.ID = id

	if rule.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if rule.Expr == "" {
		http.Error(w, "Expression is required", http.StatusBadRequest)
		return
	}

	if err := storage.UpdateAlertRule(&rule); err != nil {
		log.Printf("Error updating alert rule: %v", err)
		if err.Error() == "alert rule not found" {
			http.Error(w, "Alert rule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Updated alert rule: %s (ID: %d)", rule.Name, rule.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func deleteAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := storage.DeleteAlertRule(id); err != nil {
		log.Printf("Error deleting alert rule: %v", err)
		if err.Error() == "alert rule not found" {
			http.Error(w, "Alert rule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Deleted alert rule ID: %d", id)

	w.WriteHeader(http.StatusNoContent)
}
