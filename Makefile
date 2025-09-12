# Rules of thumb:
# - Use up-* commands for normal development (Air handles live reload)
# - Use *-build commands when Dockerfile or go.mod dependencies change
# - Use down-reset to get fresh database with migrations
# - up-prod is for local production testing only (K8s doesn't use compose)

.PHONY: up-dev up-dev-build up-prod up-db up-auth up-auth-build up-core up-core-build up-files up-files-build down-dev down-reset down-clean seed-dev

# Start commands
up-dev: # Start all services in compose.dev.yml
	cd deployments && docker compose -f compose.dev.yml up -d

up-dev-build: # Force rebuild dev + start all
	cd deployments && docker compose -f compose.dev.yml up --build -d

up-prod: # Start all services in compose.yml
	cd deployments && docker compose up --build -d

# Individual services (direct access, no nginx)
up-db:
	cd deployments && docker compose -f compose.dev.yml up -d main-db

up-auth: # Start auth + database only
	cd deployments && docker compose -f compose.dev.yml up -d auth-service main-db

up-auth-build: 
	cd deployments && docker compose -f compose.dev.yml up --build -d auth-service main-db

up-core: # Start core + database only
	cd deployments && docker compose -f compose.dev.yml up -d core-service main-db

up-core-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d core-service main-db

up-files: # Start files + database only
	cd deployments && docker compose -f compose.dev.yml up -d files-service main-db

up-files-build:
	cd deployments && docker compose -f compose.dev.yml up --build -d files-service main-db

# Stop commands
down-dev: # Stop all services
	cd deployments && docker compose -f compose.dev.yml down

down-reset: # Stop + wipe database (fresh start)
	cd deployments && docker compose -f compose.dev.yml down -v

down-clean: # Stop + cleanup Docker resources
	cd deployments && docker compose -f compose.dev.yml down -v
	docker container prune -f
	docker volume prune -f
	docker system prune -f

# In-line command override: 
# POSTGRES_USER=abc POSTGRES_DB=xyz make seed
POSTGRES_USER ?= user
POSTGRES_DB ?= tutuplapak
seed-dev:
	cd deployments && docker compose -f compose.dev.yml exec -T main-db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) < ../seeds/001_seeds_data.sql
