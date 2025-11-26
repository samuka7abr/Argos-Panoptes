#!/bin/bash
# Script para parar TUDO

set -e

echo "🛑 Parando toda a infraestrutura..."
echo ""

echo "🔴 Parando API-EXEMPLO..."
cd API-EXEMPLO && docker compose -f docker-compose-simple.yml
docker compose down
cd ..

echo ""
echo "🔵 Parando Argos..."
docker compose down

echo ""
echo "📡 Removendo rede externa..."
docker network rm external-monitoring 2>/dev/null || echo "   ℹ️  Rede já foi removida"

echo ""
echo "✅ Tudo parado!"
echo ""

