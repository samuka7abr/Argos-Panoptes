#!/bin/bash
# Script para iniciar a API-EXEMPLO (sistema monitorado)

set -e

echo "🟢 Iniciando API-EXEMPLO..."
echo ""

# Verifica se a rede externa existe
if ! docker network inspect external-monitoring >/dev/null 2>&1; then
    echo "❌ Erro: A rede 'external-monitoring' não existe!"
    echo "   Execute primeiro: ./start-argos.sh"
    exit 1
fi

echo "🏗️  Construindo e iniciando API-EXEMPLO..."
cd API-EXEMPLO && docker compose -f docker-compose-simple.yml up --build -d
cd ..

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 10

echo ""
echo "✅ API-EXEMPLO está rodando!"
echo ""
echo "🌐 Endpoints disponíveis:"
echo "   - API Web:      http://localhost:8888"
echo "   - Health:       http://localhost:8888/health"
echo "   - Users:        http://localhost:8888/users"
echo "   - Database:     localhost:5434"
echo ""
echo "📊 Para ver os logs:"
echo "   cd API-EXEMPLO && docker compose -f docker-compose-simple.yml logs -f"
echo ""
echo "🛑 Para parar:"
echo "   cd API-EXEMPLO && docker compose -f docker-compose-simple.yml down"
echo ""

