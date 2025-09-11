# Rules of thumb:
# - Use up-* commands for normal development (Air handles live reload)
# - Use *-build commands when Dockerfile or go.mod dependencies change
# - Use down-reset to get fresh database with migrations
# - up-prod is for local production testing only (K8s doesn't use compose)

.PHONY: up-dev up-dev-build up-prod up-auth up-auth-build up-core up-core-build up-files up-files-build down-dev down-reset down-clean sqlc test

# Start commands
up-dev:
	cd deployments && docker compose -f compose.dev.yml up -d

up-dev-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d

up-prod:
	cd deployments && docker compose up --build -d

# Individual services (direct access, no nginx)
up-auth:
	cd deployments && docker compose -f compose.dev.yml up -d auth-service main-db

up-auth-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d auth-service main-db

up-core:
	cd deployments && docker compose -f compose.dev.yml up -d core-service main-db

up-core-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d core-service main-db

up-files:
	cd deployments && docker compose -f compose.dev.yml up -d files-service main-db

up-files-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d files-service main-db

# Stop commands
down-dev:
	cd deployments && docker compose -f compose.dev.yml down

down-reset:
	cd deployments && docker compose -f compose.dev.yml down -v

down-clean:
	cd deployments && docker compose -f compose.dev.yml down -v
	docker container prune -f
	docker volume prune -f
	docker system prune -f

# Database  
sqlc:
	sqlc generate

# Testing
test:
	curl -s http://localhost/healthz/auth || echo "Auth service down"
	curl -s http://localhost/healthz/core || echo "Core service down"  
	curl -s http://localhost/healthz/files || echo "Files service down"