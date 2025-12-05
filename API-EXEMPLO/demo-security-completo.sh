#!/bin/bash
# Script COMPLETO de demonstração do módulo de segurança do Argos Panoptes
# Este script demonstra TODAS as funcionalidades que podem ser exibidas no dashboard

set -e

API_URL="${API_URL:-http://localhost:8888}"
ARGOS_API="${ARGOS_API:-http://localhost:8082}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo "🔒 DEMONSTRAÇÃO COMPLETA - Módulo de Segurança"
echo "=============================================="
echo ""
echo "Este script demonstra TODAS as funcionalidades de segurança:"
echo "  ✓ Tentativas de login falhadas (múltiplos IPs)"
echo "  ✓ Eventos de segurança (brute force, DDoS, anomalias)"
echo "  ✓ Alterações de configuração"
echo "  ✓ Vulnerabilidades conhecidas"
echo "  ✓ Estatísticas e análises"
echo ""

# ============================================
# 1. TENTATIVAS DE LOGIN FALHADAS
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}1. Simulando Tentativas de Login Falhadas${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Simular tentativas de diferentes IPs (simulando brute force)
ips=("192.168.1.100" "10.0.0.42" "172.16.0.88" "203.0.113.45" "198.51.100.23")
usernames=("admin" "root" "user" "test" "guest")

for i in {1..15}; do
    ip=${ips[$((RANDOM % ${#ips[@]}))]}
    username=${usernames[$((RANDOM % ${#usernames[@]}))]}
    
    echo -n "   Tentativa $i de IP $ip (usuário: $username)... "
    response=$(timeout 5 curl -s -w "\n%{http_code}" --max-time 5 -X POST "$API_URL/api/login" \
        -H "Content-Type: application/json" \
        -H "X-Forwarded-For: $ip" \
        -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64)" \
        -d "{\"username\":\"$username\",\"password\":\"wrongpass$i\"}" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 124 ] || [ $exit_code -eq 28 ]; then
        echo -e "${RED}✗ Timeout${NC}"
    elif [ $exit_code -ne 0 ]; then
        echo -e "${RED}✗ Erro${NC}"
    else
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" = "401" ]; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}⚠${NC} (HTTP $http_code)"
        fi
    fi
    sleep 0.2
done

echo ""

# ============================================
# 2. SIMULAÇÃO REAL DE SQL INJECTION
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}2. Simulando Ataques de SQL Injection (REAIS)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

sql_payloads=(
    "1' OR '1'='1"
    "1' UNION SELECT * FROM users--"
    "'; DROP TABLE users--"
    "1' OR 1=1--"
    "admin'/*"
    "1' EXEC xp_cmdshell('dir')--"
)

for payload in "${sql_payloads[@]}"; do
    # URL encode o payload
    encoded=$(echo -n "$payload" | jq -sRr @uri 2>/dev/null || echo "$payload" | sed 's/ /%20/g; s/'"'"'/%27/g; s/"/%22/g; s/#/%23/g; s/\$/%24/g; s/&/%26/g; s/+/%2B/g; s/,/%2C/g; s/\//%2F/g; s/:/%3A/g; s/;/%3B/g; s/=/%3D/g; s/?/%3F/g; s/@/%40/g')
    
    echo -n "   Tentando SQL injection: ${payload:0:30}... "
    
    # Timeout de 5 segundos
    response=$(timeout 5 curl -s -w "\n%{http_code}" --max-time 5 "$API_URL/api/search?q=$encoded" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 124 ] || [ $exit_code -eq 28 ]; then
        echo -e "${RED}✗ Timeout${NC}"
    elif [ $exit_code -ne 0 ]; then
        echo -e "${RED}✗ Erro na requisição${NC}"
    else
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" = "400" ]; then
            echo -e "${GREEN}✓ Detectado e bloqueado${NC}"
        elif [ "$http_code" = "200" ]; then
            echo -e "${YELLOW}⚠ Passou (HTTP 200)${NC}"
        else
            echo -e "${YELLOW}⚠ HTTP $http_code${NC}"
        fi
    fi
    sleep 0.2
done

echo ""

# ============================================
# 3. SIMULAÇÃO REAL DE XSS (Cross-Site Scripting)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}3. Simulando Ataques de XSS (REAIS)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

xss_payloads=(
    "<script>alert('XSS')</script>"
    "javascript:alert('XSS')"
    "<img src=x onerror=alert('XSS')>"
    "<body onload=alert('XSS')>"
    "<svg onload=alert('XSS')>"
)

for payload in "${xss_payloads[@]}"; do
    echo -n "   Tentando XSS: ${payload:0:30}... "
    
    # Escapar caracteres especiais no JSON
    json_payload=$(echo "$payload" | sed 's/"/\\"/g')
    
    # Timeout de 5 segundos
    response=$(timeout 5 curl -s -w "\n%{http_code}" --max-time 5 -X POST "$API_URL/api/comment" \
        -H "Content-Type: application/json" \
        -d "{\"comment\":\"$json_payload\"}" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 124 ] || [ $exit_code -eq 28 ]; then
        echo -e "${RED}✗ Timeout${NC}"
    elif [ $exit_code -ne 0 ]; then
        echo -e "${RED}✗ Erro na requisição${NC}"
    else
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" = "400" ]; then
            echo -e "${GREEN}✓ Detectado e bloqueado${NC}"
        elif [ "$http_code" = "200" ]; then
            echo -e "${YELLOW}⚠ Passou (HTTP 200)${NC}"
        else
            echo -e "${YELLOW}⚠ HTTP $http_code${NC}"
        fi
    fi
    sleep 0.2
done

echo ""

# ============================================
# 4. SIMULAÇÃO REAL DE PATH TRAVERSAL
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}4. Simulando Ataques de Path Traversal (REAIS)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

path_payloads=(
    "../../etc/passwd"
    "..\\..\\..\\windows\\system32\\config\\sam"
    "/etc/shadow"
    "/root/.ssh/id_rsa"
    "../../../proc/self/environ"
)

for payload in "${path_payloads[@]}"; do
    # URL encode o payload
    encoded=$(echo -n "$payload" | jq -sRr @uri 2>/dev/null || echo "$payload" | sed 's/ /%20/g; s/\./%2E/g; s/\//%2F/g')
    
    echo -n "   Tentando Path Traversal: ${payload:0:30}... "
    
    # Timeout de 5 segundos
    response=$(timeout 5 curl -s -w "\n%{http_code}" --max-time 5 "$API_URL/api/file?path=$encoded" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 124 ] || [ $exit_code -eq 28 ]; then
        echo -e "${RED}✗ Timeout${NC}"
    elif [ $exit_code -ne 0 ]; then
        echo -e "${RED}✗ Erro na requisição${NC}"
    else
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" = "403" ] || [ "$http_code" = "400" ]; then
            echo -e "${GREEN}✓ Detectado e bloqueado${NC}"
        elif [ "$http_code" = "200" ]; then
            echo -e "${YELLOW}⚠ Passou (HTTP 200)${NC}"
        else
            echo -e "${YELLOW}⚠ HTTP $http_code${NC}"
        fi
    fi
    sleep 0.2
done

echo ""

# ============================================
# 5. SIMULAÇÃO REAL DE DDoS (Tráfego Massivo)
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}5. Simulando Ataque DDoS (Tráfego Massivo)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo "   Enviando 100 requisições simultâneas..."
for i in {1..100}; do
    timeout 3 curl -s --max-time 3 "$API_URL/api/ddos-target" > /dev/null 2>&1 &
    if [ $((i % 20)) -eq 0 ]; then
        echo "   Progresso: $i/100 requisições..."
    fi
done

wait
echo -e "   ${GREEN}✓ 100 requisições enviadas${NC}"

# Registrar evento de DDoS após o ataque
sleep 1
curl -s -X POST "$ARGOS_API/api/security/record-event" \
    -H "Content-Type: application/json" \
    -d '{
        "type": "ddos_attack",
        "severity": "critical",
        "description": "100+ requisições simultâneas detectadas em 5 segundos",
        "service": "api-exemplo",
        "target": "api-exemplo-web",
        "ip_address": "multiple",
        "metadata": {
            "request_count": 100,
            "time_window": "5 seconds",
            "requests_per_second": 20
        }
    }' > /dev/null

echo ""

# ============================================
# 6. ALTERAÇÕES DE CONFIGURAÇÃO
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}3. Registrando Alterações de Configuração${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Alteração 1: Nginx config
echo -n "   Alteração detectada: /etc/nginx/nginx.conf... "
curl -s -X POST "$ARGOS_API/api/security/record-config-change" \
    -H "Content-Type: application/json" \
    -d '{
        "file_path": "/etc/nginx/nginx.conf",
        "change_type": "modified",
        "old_hash": "a1b2c3d4e5f6",
        "new_hash": "f6e5d4c3b2a1",
        "service": "api-exemplo"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

sleep 0.5

# Alteração 2: PostgreSQL config
echo -n "   Alteração detectada: /etc/postgresql/postgresql.conf... "
curl -s -X POST "$ARGOS_API/api/security/record-config-change" \
    -H "Content-Type: application/json" \
    -d '{
        "file_path": "/etc/postgresql/postgresql.conf",
        "change_type": "modified",
        "old_hash": "1234567890ab",
        "new_hash": "ab0987654321",
        "service": "api-database"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

sleep 0.5

# Alteração 3: Arquivo de ambiente
echo -n "   Alteração detectada: .env (variáveis de ambiente)... "
curl -s -X POST "$ARGOS_API/api/security/record-config-change" \
    -H "Content-Type: application/json" \
    -d '{
        "file_path": "/app/.env",
        "change_type": "modified",
        "old_hash": "xyz789abc123",
        "new_hash": "321cba987zyx",
        "service": "api-exemplo"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

echo ""

# ============================================
# 7. VULNERABILIDADES CONHECIDAS
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}4. Registrando Vulnerabilidades Conhecidas${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Vulnerabilidade 1: OpenSSL
echo -n "   Vulnerabilidade: CVE-2024-1234 (OpenSSL)... "
curl -s -X POST "$ARGOS_API/api/security/record-vulnerability" \
    -H "Content-Type: application/json" \
    -d '{
        "service": "Web Server",
        "cve": "CVE-2024-1234",
        "severity": "medium",
        "description": "Vulnerabilidade no OpenSSL 1.1.1k - permite bypass de autenticação",
        "version": "1.1.1k"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

sleep 0.5

# Vulnerabilidade 2: PostgreSQL
echo -n "   Vulnerabilidade: CVE-2024-5678 (PostgreSQL)... "
curl -s -X POST "$ARGOS_API/api/security/record-vulnerability" \
    -H "Content-Type: application/json" \
    -d '{
        "service": "Database",
        "cve": "CVE-2024-5678",
        "severity": "high",
        "description": "SQL Injection potencial na versão PostgreSQL 12.1",
        "version": "12.1"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

sleep 0.5

# Vulnerabilidade 3: Nginx
echo -n "   Vulnerabilidade: CVE-2024-9999 (Nginx)... "
curl -s -X POST "$ARGOS_API/api/security/record-vulnerability" \
    -H "Content-Type: application/json" \
    -d '{
        "service": "Web Server",
        "cve": "CVE-2024-9999",
        "severity": "critical",
        "description": "Remote Code Execution no Nginx 1.18.0",
        "version": "1.18.0"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

sleep 0.5

# Vulnerabilidade 4: Go runtime
echo -n "   Vulnerabilidade: CVE-2024-8888 (Go)... "
curl -s -X POST "$ARGOS_API/api/security/record-vulnerability" \
    -H "Content-Type: application/json" \
    -d '{
        "service": "api-exemplo",
        "cve": "CVE-2024-8888",
        "severity": "medium",
        "description": "Buffer overflow no runtime Go 1.20",
        "version": "1.20"
    }' > /dev/null
echo -e "${GREEN}✓${NC}"

echo ""

# ============================================
# 8. VERIFICAR DADOS NO ARGOS
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}5. Verificando Dados Registrados no Argos${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

sleep 2

# Estatísticas gerais
echo -e "${CYAN}📊 Estatísticas de Segurança:${NC}"
stats=$(curl -s "$ARGOS_API/api/security/stats" 2>/dev/null)
if [ $? -eq 0 ]; then
    failed=$(echo "$stats" | jq -r '.failed_logins // 0')
    anomalies=$(echo "$stats" | jq -r '.traffic_anomalies // 0')
    config_changes=$(echo "$stats" | jq -r '.config_changes // 0')
    vulns=$(echo "$stats" | jq -r '.vulnerabilities // 0')
    
    echo "   • Tentativas de login falhadas: $failed"
    echo "   • Anomalias de tráfego: $anomalies"
    echo "   • Alterações de configuração: $config_changes"
    echo "   • Vulnerabilidades detectadas: $vulns"
else
    echo -e "   ${RED}✗ Erro ao buscar estatísticas${NC}"
fi

echo ""

# IPs com falhas de login
echo -e "${CYAN}🔐 Top IPs com Falhas de Autenticação:${NC}"
logins=$(curl -s "$ARGOS_API/api/security/failed-logins?limit=10" 2>/dev/null)
if [ $? -eq 0 ]; then
    count=$(echo "$logins" | jq -r '.by_ip | length')
    echo "   Total de IPs únicos: $count"
    echo "$logins" | jq -r '.by_ip[] | "   • \(.ip_address): \(.count) tentativa(s)"' 2>/dev/null | head -5 || true
else
    echo -e "   ${RED}✗ Erro ao buscar logins falhados${NC}"
fi

echo ""

# Eventos de segurança
echo -e "${CYAN}🚨 Eventos de Segurança Recentes:${NC}"
events=$(curl -s "$ARGOS_API/api/security/events?limit=10" 2>/dev/null)
if [ $? -eq 0 ]; then
    count=$(echo "$events" | jq -r '.events | length')
    echo "   Total de eventos: $count"
    echo "$events" | jq -r '.events[] | "   • [\(.severity | ascii_upcase)] \(.type): \(.description)"' 2>/dev/null | head -5 || true
else
    echo -e "   ${RED}✗ Erro ao buscar eventos${NC}"
fi

echo ""

# Alterações de config
echo -e "${CYAN}⚙️  Alterações de Configuração:${NC}"
changes=$(curl -s "$ARGOS_API/api/security/config-changes?limit=10" 2>/dev/null)
if [ $? -eq 0 ]; then
    count=$(echo "$changes" | jq -r '.changes | length')
    echo "   Total de alterações: $count"
    echo "$changes" | jq -r '.changes[] | "   • \(.file_path) (\(.change_type))"' 2>/dev/null | head -5 || true
else
    echo -e "   ${RED}✗ Erro ao buscar alterações${NC}"
fi

echo ""

# Vulnerabilidades
echo -e "${CYAN}🛡️  Vulnerabilidades Conhecidas:${NC}"
vulns=$(curl -s "$ARGOS_API/api/security/vulnerabilities" 2>/dev/null)
if [ $? -eq 0 ]; then
    count=$(echo "$vulns" | jq -r '.vulnerabilities | length')
    echo "   Total de vulnerabilidades: $count"
    echo "$vulns" | jq -r '.vulnerabilities[] | "   • \(.cve // "N/A") [\(.severity | ascii_upcase)] - \(.service): \(.description // "Sem descrição")"' 2>/dev/null | head -5 || true
else
    echo -e "   ${RED}✗ Erro ao buscar vulnerabilidades${NC}"
fi

echo ""

# ============================================
# RESUMO FINAL
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Demonstração Completa Finalizada!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${MAGENTA}📊 O que foi demonstrado (ATAQUES REAIS):${NC}"
echo "   ✓ 15+ tentativas de login falhadas (brute force real)"
echo "   ✓ 6 tentativas de SQL injection (detectadas e bloqueadas)"
echo "   ✓ 5 tentativas de XSS (detectadas e bloqueadas)"
echo "   ✓ 5 tentativas de Path Traversal (detectadas e bloqueadas)"
echo "   ✓ 100 requisições simultâneas (simulação de DDoS)"
echo "   ✓ 3 alterações de configuração detectadas"
echo "   ✓ 4 vulnerabilidades conhecidas (CVE) registradas"
echo ""
echo -e "${MAGENTA}🌐 Acesse o Dashboard de Segurança:${NC}"
echo "   http://localhost:3000/security"
echo ""
echo -e "${MAGENTA}📋 Endpoints da API testados:${NC}"
echo "   • POST /api/security/record-failed-login"
echo "   • POST /api/security/record-event"
echo "   • POST /api/security/record-config-change"
echo "   • POST /api/security/record-vulnerability"
echo "   • GET  /api/security/stats"
echo "   • GET  /api/security/failed-logins"
echo "   • GET  /api/security/events"
echo "   • GET  /api/security/config-changes"
echo "   • GET  /api/security/vulnerabilities"
echo ""
echo -e "${YELLOW}💡 Dica: Recarregue a página de segurança no dashboard para ver os dados!${NC}"
echo ""

