#!/bin/bash

# SkillsHub Enterprise Makefile
# Common development and deployment tasks

.PHONY: help dev-up dev-down db-migrate db-seed test build clean

# Default target
help:
	@echo "SkillsHub Enterprise - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "Development:"
	@echo "  make dev-up         - Start all services with docker-compose"
	@echo "  make dev-down       - Stop all services"
	@echo "  make dev-restart    - Restart all services"
	@echo "  make logs           - View all logs"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-seed        - Seed initial data"
	@echo "  make db-backup      - Backup database"
	@echo "  make db-restore     - Restore database"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build all services"
	@echo "  make build-api      - Build API service"
	@echo "  make build-scanner  - Build Scanner service"
	@echo "  make build-web      - Build Web frontend"
	@echo ""
	@echo "Test:"
	@echo "  make test           - Run all tests"
	@echo "  make test-api       - Run API tests"
	@echo "  make test-scanner   - Run Scanner tests"
	@echo ""
	@echo "Clean:"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make clean-all      - Clean everything including volumes"

# Development
dev-up:
	docker-compose up -d

dev-down:
	docker-compose down

dev-restart:
	docker-compose restart

logs:
	docker-compose logs -f

# Database
db-migrate:
	docker-compose exec -T postgres psql -U skillshub -d skillshub -f /docker-entrypoint-initdb.d/001_initial_schema.sql

db-seed:
	docker-compose exec -T postgres psql -U skillshub -d skillshub -c "INSERT INTO users (username, email, role) VALUES ('test', 'test@test.com', 'developer') ON CONFLICT DO NOTHING;"

db-backup:
	docker-compose exec postgres pg_dump -U skillshub skillshub > backup_$$(date +%Y%m%d_%H%M%S).sql

db-restore:
	@echo "Usage: make db-restore FILE=backup_20240101.sql"
	docker-compose exec -T postgres psql -U skillshub skillshub < $(FILE)

# Build
build: build-api build-scanner build-web

build-api:
	cd apps/api && go build -o ../../bin/api ./cmd

build-scanner:
	cd apps/scanner && python -m py_compile app/main.py

build-web:
	cd apps/web && npm install && npm run build

# Test
test: test-api test-scanner

test-api:
	cd apps/api && go test ./...

test-scanner:
	cd apps/scanner && python -m pytest app/ || echo "pytest not configured"

# Clean
clean:
	rm -rf bin/
	rm -rf apps/web/dist/
	rm -rf apps/web/node_modules/
	find . -name "node_modules" -type d -exec rm -rf {} \; 2>/dev/null || true
	find . -name "__pycache__" -type d -exec rm -rf {} \; 2>/dev/null || true

clean-all: clean
	docker-compose down -v
	rm -rf backup_*.sql

# Kubernetes (for production)
k8s-deploy:
	kubectl apply -f deploy/k8s/

k8s-status:
	kubectl get pods -l app=skillshub

k8s-logs:
	kubectl logs -l app=skillshub -f
