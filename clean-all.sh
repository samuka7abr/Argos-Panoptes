#!/bin/bash
# Script para limpar TUDO (containers, volumes, networks, imagens)

set -e

echo "🧹 Limpando toda a infraestrutura..."
echo ""
echo "⚠️  ATENÇÃO: Isso vai remover TODOS os dados!"
read -p "   Continuar? (s/N): " confirm

if [[ $confirm != "s" && $confirm != "S" ]]; then
    echo "❌ Cancelado."
    exit 0
fi

echo ""
echo "🔴 Parando e removendo API-EXEMPLO (com volumes)..."
cd API-EXEMPLO
docker compose down -v
cd ..

echo ""
echo "🔵 Parando e removendo Argos (com volumes)..."
docker compose down -v

echo ""
echo "📡 Removendo rede externa..."
docker network rm external-monitoring 2>/dev/null || echo "   ℹ️  Rede já foi removida"

echo ""
echo "🗑️  Removendo imagens do Argos..."
docker images | grep -E "poc9p|argos" | awk '{print $3}' | xargs -r docker rmi -f || true

echo ""
echo "✅ Tudo limpo!"
echo ""

