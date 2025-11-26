#!/bin/bash
# Script para iniciar TUDO (Argos + API-EXEMPLO)

set -e

echo "🚀 Iniciando infraestrutura completa..."
echo "========================================"
echo ""

# Inicia o Argos
./start-argos.sh

echo ""
echo "========================================"
echo ""

# Inicia a API-EXEMPLO
./start-api-exemplo.sh

echo ""
echo "========================================"
echo ""
echo "🎉 Tudo está rodando!"
echo ""
echo "📊 Status dos containers:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(argos|api-)" || true
echo ""
echo "🧪 Para testar a integração:"
echo "   ./test-integration.sh"
echo ""

