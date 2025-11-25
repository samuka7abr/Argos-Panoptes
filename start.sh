#!/bin/bash

set -e

echo "🔥 Argos Panoptes - Inicializando..."
echo ""

if ! command -v docker compose &> /dev/null; then
    echo "❌ docker compose não encontrado. Instale Docker Compose primeiro."
    exit 1
fi

echo "📦 Building imagens Docker..."
docker compose build

echo ""
echo "🚀 Iniciando serviços..."
docker      compose up -d

echo ""
echo "⏳ Aguardando serviços iniciarem..."
sleep 15

echo ""
echo "🧪 Testando conectividade..."
echo ""

if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ API: http://localhost:8080"
else
    echo "❌ API não respondeu"
fi

if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend: http://localhost:3000"
else
    echo "⚠️  Frontend ainda não respondeu (pode levar mais tempo)"
fi

if curl -s http://localhost:8025 > /dev/null; then
    echo "✅ MailHog: http://localhost:8025"
else
    echo "❌ MailHog não respondeu"
fi

if curl -s http://localhost:8081 > /dev/null; then
    echo "✅ Test Web: http://localhost:8081"
else
    echo "❌ Test Web não respondeu"
fi

echo ""
echo "════════════════════════════════════════════════"
echo "🎉 Argos Panoptes está rodando!"
echo "════════════════════════════════════════════════"
echo ""
echo "📊 Dashboard:     http://localhost:3000"
echo "🔌 API:           http://localhost:8080/health"
echo "📧 E-mails:       http://localhost:8025"
echo "🌐 Serviço Teste: http://localhost:8081"
echo ""
echo "💡 Dicas:"
echo "  - Ver logs:     docker-compose logs -f"
echo "  - Parar tudo:   docker-compose down"
echo "  - Limpar tudo:  make clean"
echo ""
echo "⏱️  Aguarde ~1-2 minutos para as primeiras métricas aparecerem"
echo ""

