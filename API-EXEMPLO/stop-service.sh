#!/bin/bash
# Script interativo para parar serviços específicos da API-EXEMPLO

set -e

COMPOSE_FILE="docker-compose-simple.yml"
COMPOSE_DIR="$(cd "$(dirname "$0")" && pwd)"

cd "$COMPOSE_DIR"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}🛑 Parar Serviços da API-EXEMPLO${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Listar serviços rodando
echo -e "${CYAN}📋 Serviços em execução:${NC}"
echo ""

services=()
index=1

# Obter lista de serviços do compose
all_services=$(docker compose -f "$COMPOSE_FILE" config --services 2>/dev/null)

for service in $all_services; do
    # Verificar se o serviço está rodando usando docker compose ps
    status=$(docker compose -f "$COMPOSE_FILE" ps "$service" --format "{{.Status}}" 2>/dev/null | head -n1 || echo "")
    
    if echo "$status" | grep -qiE "up|running"; then
        # Obter informações do container
        container_name=$(docker compose -f "$COMPOSE_FILE" ps "$service" --format "{{.Name}}" 2>/dev/null | head -n1 || echo "$service")
        ports=$(docker compose -f "$COMPOSE_FILE" ps "$service" --format "{{.Ports}}" 2>/dev/null | head -n1 || echo "")
        
        # Extrair porta publicada se existir
        published_port=$(echo "$ports" | grep -oE '[0-9]+:[0-9]+' | head -n1 | cut -d: -f1 || echo "")
        
        echo -e "   ${GREEN}[$index]${NC} $service"
        echo -e "       Container: ${CYAN}$container_name${NC}"
        if [ -n "$published_port" ]; then
            echo -e "       Porta: ${CYAN}$published_port${NC}"
        fi
        echo -e "       Status: ${CYAN}$status${NC}"
        echo ""
        
        services+=("$service")
        index=$((index + 1))
    fi
done

if [ ${#services[@]} -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Nenhum serviço está rodando no momento.${NC}"
    echo ""
    exit 0
fi

echo -e "   ${YELLOW}[0]${NC} Cancelar"
echo ""

# Solicitar seleção
read -p "Selecione o serviço para parar (0 para cancelar): " choice

# Validar entrada
if ! [[ "$choice" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}✗ Entrada inválida. Use um número.${NC}"
    exit 1
fi

if [ "$choice" -eq 0 ]; then
    echo -e "${YELLOW}Operação cancelada.${NC}"
    exit 0
fi

if [ "$choice" -lt 1 ] || [ "$choice" -gt ${#services[@]} ]; then
    echo -e "${RED}✗ Número inválido. Selecione entre 1 e ${#services[@]}.${NC}"
    exit 1
fi

# Obter serviço selecionado
selected_service=${services[$((choice - 1))]}

echo ""
echo -e "${YELLOW}⏸️  Parando serviço: ${CYAN}$selected_service${NC}"
echo ""

# Parar o serviço
if docker compose -f "$COMPOSE_FILE" stop "$selected_service" 2>/dev/null; then
    echo -e "${GREEN}✅ Serviço '$selected_service' parado com sucesso!${NC}"
    echo ""
    echo -e "${CYAN}📊 Status atual:${NC}"
    docker compose -f "$COMPOSE_FILE" ps "$selected_service" --format "table {{.Service}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    echo -e "${BLUE}💡 Para iniciar novamente:${NC}"
    echo "   docker compose -f $COMPOSE_FILE start $selected_service"
    echo ""
else
    echo -e "${RED}✗ Erro ao parar o serviço '$selected_service'${NC}"
    exit 1
fi

