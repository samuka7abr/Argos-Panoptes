package main

import (
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


