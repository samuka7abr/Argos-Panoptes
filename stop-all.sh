#!/bin/bash
# Script para parar TUDO e remover volumes

set -e

echo "🛑 Parando toda a infraestrutura..."
echo ""

echo "🔴 Parando API-EXEMPLO (com volumes)..."
cd API-EXEMPLO && docker compose -f docker-compose-simple.yml down -v
cd ..

echo ""
echo "🔵 Parando Argos (com volumes)..."
docker compose down -v

echo ""
echo "📡 Removendo rede externa..."
docker network rm external-monitoring 2>/dev/null || echo "   ℹ️  Rede já foi removida"

echo ""
echo "🗑️  Verificando volumes do PostgreSQL..."
# Remover volumes explicitamente (caso ainda existam)
docker volume rm argos-panoptes_postgres_data 2>/dev/null || echo "   ℹ️  Volume do Argos já removido"
docker volume rm api-exemplo_api-db-data 2>/dev/null || echo "   ℹ️  Volume da API-EXEMPLO já removido"
docker volume rm $(docker volume ls -q | grep postgres) 2>/dev/null || true

echo ""
echo "✅ Tudo parado e volumes removidos!"
echo ""
echo "💡 Nota: Todos os dados do banco foram apagados."
echo ""

