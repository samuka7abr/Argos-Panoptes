#!/bin/bash
# Script de testes de integração do Argos

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

passed=0
failed=0

test_endpoint() {
    local name=$1
    local url=$2
    local expected_code=${3:-200}
    
    echo -n "   🧪 $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✓${NC} (HTTP $response)"
        ((passed++))
        return 0
    else
        echo -e "${RED}✗${NC} (HTTP $response, esperado $expected_code)"
        ((failed++))
        return 1
    fi
}

test_json_endpoint() {
    local name=$1
    local url=$2
    local jq_filter=$3
    
    echo -n "   🧪 $name... "
    
    result=$(curl -s "$url" | jq -r "$jq_filter" 2>/dev/null || echo "error")
    
    if [ "$result" != "error" ] && [ "$result" != "null" ] && [ -n "$result" ]; then
        echo -e "${GREEN}✓${NC} ($result)"
        ((passed++))
        return 0
    else
        echo -e "${RED}✗${NC} (falhou ao processar JSON)"
        ((failed++))
        return 1
    fi
}

echo ""
echo "🧪 Testes de Integração do Argos Panoptes"
echo "=========================================="
echo ""

# Verifica se os serviços estão rodando
echo -e "${BLUE}1. Verificando containers...${NC}"
required_containers=("argos-api" "argos-agent" "argos-alert" "argos-web" "argos-postgres" "argos-mailhog" "api-example" "api-database" "api-nginx")
all_running=true

for container in "${required_containers[@]}"; do
    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        echo -e "   ${GREEN}✓${NC} $container está rodando"
    else
        echo -e "   ${RED}✗${NC} $container NÃO está rodando"
        all_running=false
    fi
done

if [ "$all_running" = false ]; then
    echo ""
    echo -e "${RED}❌ Alguns containers não estão rodando!${NC}"
    echo "   Execute: ./start-all.sh"
    exit 1
fi

echo ""
echo -e "${BLUE}2. Testando API-EXEMPLO...${NC}"
test_endpoint "GET /" "http://localhost:8888/" 200
test_endpoint "GET /health" "http://localhost:8888/health" 200
test_json_endpoint "GET /users" "http://localhost:8888/users" ".[0].name"

echo ""
echo -e "${BLUE}3. Testando Argos API...${NC}"
test_endpoint "GET /health" "http://localhost:8082/health" 200
test_json_endpoint "GET /api/metrics/services" "http://localhost:8082/api/metrics/services" ".services | length"

echo ""
echo -e "${BLUE}4. Verificando métricas coletadas...${NC}"
echo "   ⏳ Aguardando 30s para coleta de métricas..."
sleep 30

test_json_endpoint "Métricas da API-EXEMPLO" "http://localhost:8082/api/metrics/latest" '[.[] | select(.target | contains("api"))] | length'
test_json_endpoint "Métricas HTTP" "http://localhost:8082/api/metrics/latest" '[.[] | select(.metrics.http_up)] | length'
test_json_endpoint "Métricas Postgres" "http://localhost:8082/api/metrics/latest" '[.[] | select(.metrics.postgres_up)] | length'

echo ""
echo -e "${BLUE}5. Testando Dashboard Web...${NC}"
test_endpoint "Dashboard" "http://localhost:3000/" 200

echo ""
echo -e "${BLUE}6. Testando MailHog...${NC}"
test_endpoint "MailHog UI" "http://localhost:8025/" 200

echo ""
echo -e "${BLUE}7. Verificando rede externa...${NC}"
if docker network inspect external-monitoring >/dev/null 2>&1; then
    echo -e "   ${GREEN}✓${NC} Rede 'external-monitoring' existe"
    ((passed++))
else
    echo -e "   ${RED}✗${NC} Rede 'external-monitoring' não existe"
    ((failed++))
fi

echo ""
echo "=========================================="
echo -e "${GREEN}Passou: $passed${NC} | ${RED}Falhou: $failed${NC}"
echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}🎉 Todos os testes passaram!${NC}"
    echo ""
    echo "📊 Próximos passos:"
    echo "   - Dashboard:    http://localhost:3000"
    echo "   - API-EXEMPLO:  http://localhost:8888"
    echo "   - MailHog:      http://localhost:8025"
    echo ""
    echo "🔥 Para testar alertas:"
    echo "   cd API-EXEMPLO && docker compose stop api-example"
    echo "   Aguarde 2-3 minutos e verifique o MailHog"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Alguns testes falharam!${NC}"
    echo ""
    echo "📋 Para investigar:"
    echo "   docker compose logs -f"
    echo "   cd API-EXEMPLO && docker compose logs -f"
    echo ""
    exit 1
fi

