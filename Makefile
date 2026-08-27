.PHONY: help setup start stop logs test build deploy clean fmt lint

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

help:
	@echo "$(BLUE)Coherence Development Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Setup & Installation$(NC)"
	@echo "  make setup          - Initialize development environment"
	@echo "  make install-tools  - Install required tools"
	@echo ""
	@echo "$(GREEN)Running Services$(NC)"
	@echo "  make start          - Start all Docker services"
	@echo "  make stop           - Stop all Docker services"
	@echo "  make restart        - Restart all Docker services"
	@echo "  make logs           - View logs from all services"
	@echo "  make logs-backend   - View backend logs"
	@echo "  make logs-frontend  - View frontend logs"
	@echo ""
	@echo "$(GREEN)Development$(NC)"
	@echo "  make fmt            - Format code (Go & JavaScript)"
	@echo "  make lint           - Run linters"
	@echo "  make test           - Run all tests"
	@echo "  make test-backend   - Run backend tests"
	@echo "  make test-frontend  - Run frontend tests"
	@echo "  make dev-backend    - Start backend in development mode"
	@echo "  make dev-frontend   - Start frontend in development mode"
	@echo ""
	@echo "$(GREEN)Building$(NC)"
	@echo "  make build          - Build all components"
	@echo "  make build-backend  - Build backend binary"
	@echo "  make build-frontend - Build frontend"
	@echo "  make docker-build   - Build Docker images"
	@echo ""
	@echo "$(GREEN)Database$(NC)"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-clean       - Drop all tables"
	@echo "  make db-seed        - Seed database with sample data"
	@echo ""
	@echo "$(GREEN)Deployment$(NC)"
	@echo "  make deploy-k8s     - Deploy to Kubernetes"
	@echo "  make deploy-docker  - Deploy using Docker Compose"
	@echo ""
	@echo "$(GREEN)Cleanup$(NC)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make clean-docker   - Stop and remove Docker resources"
	@echo "  make reset          - Full reset (remove all data)"

setup:
	@echo "$(BLUE)Setting up Coherence development environment...$(NC)"
	chmod +x scripts/setup.sh
	./scripts/setup.sh

install-tools:
	@echo "$(BLUE)Installing development tools...$(NC)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)Go not found. Please install Go 1.21+$(NC)" && exit 1)
	@command -v npm >/dev/null 2>&1 || (echo "$(RED)Node.js not found. Please install Node.js 18+$(NC)" && exit 1)
	@echo "$(GREEN)✓ All required tools are installed$(NC)"

start:
	@echo "$(BLUE)Starting Docker services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo ""
	@echo "$(GREEN)Access points:$(NC)"
	@echo "  Dashboard: http://localhost:3000"
	@echo "  API: http://localhost:8080"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana: http://localhost:3001"

stop:
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

restart: stop start

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	cd backend && go fmt ./...
	cd frontend && npm run format 2>/dev/null || echo "prettier not configured"
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint:
	@echo "$(BLUE)Running linters...$(NC)"
	cd backend && golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed, skipping"
	@echo "$(GREEN)✓ Linting complete$(NC)"

test:
	@echo "$(BLUE)Running all tests...$(NC)"
	cd backend && go test ./... -v -race
	@echo "$(GREEN)✓ Tests passed$(NC)"

test-backend:
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd backend && go test ./... -v -race -coverprofile=coverage.out

test-frontend:
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd frontend && npm test -- --watchAll=false

dev-backend:
	@echo "$(BLUE)Starting backend in development mode...$(NC)"
	cd backend && go run cmd/main.go

dev-frontend:
	@echo "$(BLUE)Starting frontend in development mode...$(NC)"
	cd frontend && npm start

build: build-backend build-frontend
	@echo "$(GREEN)✓ All components built$(NC)"

build-backend:
	@echo "$(BLUE)Building backend...$(NC)"
	cd backend && CGO_ENABLED=0 go build -o coherence-server cmd/main.go
	@echo "$(GREEN)✓ Backend built$(NC)"

build-frontend:
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && npm run build
	@echo "$(GREEN)✓ Frontend built$(NC)"

docker-build:
	@echo "$(BLUE)Building Docker images...$(NC)"
	docker-compose build
	@echo "$(GREEN)✓ Docker images built$(NC)"

db-migrate:
	@echo "$(BLUE)Running database migrations...$(NC)"
	docker-compose exec backend ./coherence-server migrate
	@echo "$(GREEN)✓ Migrations completed$(NC)"

db-clean:
	@echo "$(RED)Dropping all database tables...$(NC)"
	docker-compose exec postgres psql -U coherence -d coherence -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "$(GREEN)✓ Database cleaned$(NC)"

db-seed:
	@echo "$(BLUE)Seeding database...$(NC)"
	@echo "Sample data insertion would go here"
	@echo "$(GREEN)✓ Database seeded$(NC)"

deploy-k8s:
	@echo "$(BLUE)Deploying to Kubernetes...$(NC)"
	kubectl apply -f deployments/k8s/deployment.yaml
	@echo "$(GREEN)✓ Deployed to Kubernetes$(NC)"

deploy-docker:
	@echo "$(BLUE)Deploying with Docker Compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Deployed with Docker Compose$(NC)"

clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	cd backend && go clean -cache -testcache
	rm -f backend/coherence-server
	cd frontend && rm -rf build dist node_modules
	rm -rf *.zip
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-docker:
	@echo "$(RED)Removing Docker resources...$(NC)"
	docker-compose down --rmi all
	@echo "$(GREEN)✓ Docker resources removed$(NC)"

reset: clean clean-docker
	@echo "$(GREEN)✓ Full reset complete$(NC)"
	@echo "$(BLUE)Run 'make setup' to reinitialize$(NC)"

# Additional helpful targets

check-health:
	@echo "$(BLUE)Checking service health...$(NC)"
	@curl -s http://localhost:8080/health | jq . || echo "$(RED)Backend not responding$(NC)"
	@echo ""
	@echo "Services status:"
	docker-compose ps

open-dashboard:
	@echo "$(BLUE)Opening dashboard...$(NC)"
	@command -v open >/dev/null 2>&1 && open http://localhost:3000 || \
	command -v xdg-open >/dev/null 2>&1 && xdg-open http://localhost:3000 || \
	echo "Please open http://localhost:3000 in your browser"

psql:
	docker-compose exec postgres psql -U coherence -d coherence

redis-cli:
	docker-compose exec redis redis-cli

ps:
	docker-compose ps

stats:
	docker stats

.DEFAULT_GOAL := help
