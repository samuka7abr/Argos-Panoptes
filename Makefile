.PHONY: help build up down logs restart clean test

help:
	@echo "Argos Panoptes - Comandos Disponíveis:"
	@echo ""
	@echo "  make build    - Build de todas as imagens Docker"
	@echo "  make up       - Sobe todos os serviços"
	@echo "  make down     - Para todos os serviços"
	@echo "  make logs     - Mostra logs de todos os serviços"
	@echo "  make restart  - Reinicia todos os serviços"
	@echo "  make clean    - Remove volumes e containers"
	@echo "  make test     - Testa os endpoints da API"
	@echo ""

build:
	docker-compose build

up:
	docker-compose up -d
	@echo ""
	@echo "✅ Argos Panoptes iniciado!"
	@echo ""
	@echo "📊 Frontend:  http://localhost:3000"
	@echo "🔌 API:       http://localhost:8080"
	@echo "📧 MailHog:   http://localhost:8025"
	@echo "🌐 Test Web:  http://localhost:8081"
	@echo ""

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

clean:
	docker-compose down -v
	docker system prune -f

test:
	@echo "🧪 Testando endpoints..."
	@echo ""
	@echo "Health API:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ API não respondeu"
	@echo ""
	@echo "Latest Metrics:"
	@curl -s http://localhost:8080/api/metrics/latest | jq '.[0]' || echo "❌ Sem métricas ainda"
	@echo ""
	@echo "Alert Rules:"
	@curl -s http://localhost:8080/api/alert-rules | jq '.count' || echo "❌ Sem regras"

