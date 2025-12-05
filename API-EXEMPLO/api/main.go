package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func main() {
	// Configuração do banco
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "apiuser")
	dbPass := getEnv("DB_PASSWORD", "apipass")
	dbName := getEnv("DB_NAME", "apidb")
	port := getEnv("PORT", "8080")

	// Conectar ao banco
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Tentar conectar com retry
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (%d/30)", i+1)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Fatalf("Database not available after 30 seconds: %v", err)
	}

	log.Println("Connected to database successfully")

	// Criar tabela de usuários
	initDB()

	// Rotas
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/slow", slowHandler)
	http.HandleFunc("/api/error", errorHandler)

	// Security endpoints
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/security/traffic-spike", trafficSpikeHandler)
	http.HandleFunc("/api/security/test", securityTestHandler)

	// Attack simulation endpoints (para demonstração)
	http.HandleFunc("/api/search", searchHandler)          // SQL injection simulation
	http.HandleFunc("/api/comment", commentHandler)        // XSS simulation
	http.HandleFunc("/api/file", fileHandler)              // Path traversal simulation
	http.HandleFunc("/api/ddos-target", ddosTargetHandler) // DDoS target

	// Iniciar servidor
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("API-EXEMPLO listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initDB() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Warning: Failed to create table: %v", err)
		return
	}

	// Inserir dados de exemplo
	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
			('João Silva', 'joao@example.com'),
			('Maria Santos', 'maria@example.com'),
			('Pedro Oliveira', 'pedro@example.com')
		ON CONFLICT (email) DO NOTHING
	`)
	if err != nil {
		log.Printf("Warning: Failed to insert sample data: %v", err)
	}

	log.Println("Database initialized with sample data")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>API-EXEMPLO</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		h1 { color: #333; }
		.endpoint { background: #f4f4f4; padding: 10px; margin: 10px 0; border-radius: 5px; }
		code { background: #333; color: #fff; padding: 2px 5px; border-radius: 3px; }
	</style>
</head>
<body>
	<h1>🚀 API-EXEMPLO</h1>
	<p>Sistema de demonstração para monitoramento do Argos Panoptes</p>
	
	<h2>Endpoints Disponíveis:</h2>
	<div class="endpoint">
		<strong>GET /health</strong> - Status de saúde da API
	</div>
	<div class="endpoint">
		<strong>GET /users</strong> - Lista de usuários
	</div>
	<div class="endpoint">
		<strong>GET /api/slow</strong> - Endpoint lento (5s de delay)
	</div>
	<div class="endpoint">
		<strong>GET /api/error</strong> - Simula erro 500
	</div>
	
	<p><small>Uptime: ` + time.Since(startTime).Round(time.Second).String() + `</small></p>
</body>
</html>
`
	fmt.Fprint(w, html)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	if err := db.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	health := HealthResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Slow endpoint called - sleeping 5 seconds...")
	time.Sleep(5 * time.Second)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "This endpoint intentionally takes 5 seconds to respond",
		"delay":   "5s",
	})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Error endpoint called - returning 500")
	http.Error(w, "Internal Server Error - This is intentional for testing", http.StatusInternalServerError)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simular login: sempre falha para demonstrar segurança
	// Em produção, isso seria verificado contra banco de dados
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	// Registrar tentativa falhada no Argos
	argosURL := getEnv("ARGOS_API_URL", "http://api:8082")
	failedLoginReq := map[string]interface{}{
		"ip_address": ip,
		"username":   req.Username,
		"service":    "api-exemplo",
		"user_agent": userAgent,
	}

	reqBody, _ := json.Marshal(failedLoginReq)
	http.Post(argosURL+"/api/security/record-failed-login", "application/json", bytes.NewReader(reqBody))

	// Retornar erro 401
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func trafficSpikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	argosURL := getEnv("ARGOS_API_URL", "http://api:8082")

	// Criar evento de anomalia de tráfego
	event := map[string]interface{}{
		"type":        "traffic_spike",
		"severity":    "warning",
		"description": "Pico súbito de tráfego detectado - aumento de 300% nas requisições",
		"service":     "api-exemplo",
		"target":      "api-exemplo-web",
		"metadata": map[string]interface{}{
			"spike_percentage": 300,
			"duration":         "5 minutes",
		},
	}

	reqBody, _ := json.Marshal(event)
	resp, err := http.Post(argosURL+"/api/security/record-event", "application/json", bytes.NewReader(reqBody))
	if err == nil {
		resp.Body.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "recorded",
		"message": "Traffic spike event recorded",
	})
}

func securityTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"endpoints": map[string]string{
			"login":         "POST /api/login - Simula tentativa de login (sempre falha)",
			"traffic_spike": "POST /api/security/traffic-spike - Registra pico de tráfego",
			"search":        "GET /api/search?q=... - Simula SQL injection",
			"comment":       "POST /api/comment - Simula XSS",
			"file":          "GET /api/file?path=... - Simula Path Traversal",
			"ddos":          "GET /api/ddos-target - Alvo para simulação de DDoS",
		},
	})
}

// searchHandler - Simula SQL injection (detecta e registra)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing parameter 'q'", http.StatusBadRequest)
		return
	}

	// Detectar padrões de SQL injection
	sqlInjectionPatterns := []string{
		"UNION SELECT", "OR 1=1", "'; DROP", "1' OR '1'='1",
		"'; --", "/*", "*/", "xp_", "EXEC", "EXECUTE",
	}

	detected := false
	for _, pattern := range sqlInjectionPatterns {
		if contains(query, pattern) {
			detected = true
			break
		}
	}

	if detected {
		// Registrar evento de segurança
		ip := getClientIP(r)
		recordSecurityEvent("sql_injection_attempt", "critical",
			fmt.Sprintf("Tentativa de SQL injection detectada: %s", query),
			"api-exemplo", "api-exemplo-web", ip,
			map[string]interface{}{
				"query":      query,
				"endpoint":   "/api/search",
				"user_agent": r.UserAgent(),
			})
		http.Error(w, "Invalid query detected", http.StatusBadRequest)
		return
	}

	// Query normal (simulada)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": []string{"result1", "result2"},
		"query":   query,
	})
}

// commentHandler - Simula XSS (detecta e registra)
func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Detectar padrões de XSS
	xssPatterns := []string{"<script>", "javascript:", "onerror=", "onload=", "<img src=x", "alert("}

	detected := false
	for _, pattern := range xssPatterns {
		if contains(req.Comment, pattern) {
			detected = true
			break
		}
	}

	if detected {
		ip := getClientIP(r)
		recordSecurityEvent("xss_attempt", "high",
			fmt.Sprintf("Tentativa de XSS detectada: %s", req.Comment),
			"api-exemplo", "api-exemplo-web", ip,
			map[string]interface{}{
				"comment":    req.Comment,
				"endpoint":   "/api/comment",
				"user_agent": r.UserAgent(),
			})
		http.Error(w, "XSS attempt detected", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "comment saved"})
}

// fileHandler - Simula Path Traversal (detecta e registra)
func fileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing parameter 'path'", http.StatusBadRequest)
		return
	}

	// Detectar path traversal
	if contains(path, "..") || contains(path, "/etc/") || contains(path, "/root/") || contains(path, "/proc/") {
		ip := getClientIP(r)
		recordSecurityEvent("path_traversal_attempt", "high",
			fmt.Sprintf("Tentativa de Path Traversal detectada: %s", path),
			"api-exemplo", "api-exemplo-web", ip,
			map[string]interface{}{
				"path":       path,
				"endpoint":   "/api/file",
				"user_agent": r.UserAgent(),
			})
		http.Error(w, "Path traversal attempt detected", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"file":    path,
		"content": "file content here",
	})
}

// ddosTargetHandler - Endpoint alvo para simulação de DDoS
func ddosTargetHandler(w http.ResponseWriter, r *http.Request) {
	// Simular processamento
	time.Sleep(10 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "DDoS target endpoint",
	})
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func recordSecurityEvent(eventType, severity, description, service, target, ip string, metadata map[string]interface{}) {
	argosURL := getEnv("ARGOS_API_URL", "http://api:8082")
	event := map[string]interface{}{
		"type":        eventType,
		"severity":    severity,
		"description": description,
		"service":     service,
		"target":      target,
		"ip_address":  ip,
		"metadata":    metadata,
	}

	reqBody, _ := json.Marshal(event)
	http.Post(argosURL+"/api/security/record-event", "application/json", bytes.NewReader(reqBody))
}
