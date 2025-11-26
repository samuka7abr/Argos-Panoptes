#!/bin/bash
# Script para iniciar o Argos Panoptes

set -e

echo "🔵 Iniciando Argos Panoptes..."
echo ""

# Cria a rede externa se não existir
echo "📡 Criando rede externa 'external-monitoring'..."
docker network create external-monitoring 2>/dev/null || echo "   ℹ️  Rede já existe"

echo ""
echo "🏗️  Construindo e iniciando serviços do Argos..."
docker compose up --build -d

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 10

echo ""
echo "✅ Argos Panoptes está rodando!"
echo ""
echo "🌐 Interfaces disponíveis:"
echo "   - Dashboard:    http://localhost:3000"
echo "   - API:          http://localhost:8082"
echo "   - MailHog UI:   http://localhost:8025"
echo ""
echo "📊 Para ver os logs:"
echo "   docker compose logs -f"
echo ""
echo "🛑 Para parar:"
echo "   docker compose down"
echo ""

