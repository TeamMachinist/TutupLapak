.PHONY: dev build down clean sqlc test

# Development
dev:
	cd deployments && docker-compose -f compose.dev.yml up -d

dev-build:
	cd deployments && docker-compose -f compose.dev.yml up --build -d

# Individual services (dev)
auth-dev:
	cd deployments && docker-compose -f compose.dev.yml up -d nginx auth-service main-db

core-dev:
	cd deployments && docker-compose -f compose.dev.yml up -d nginx core-service main-db

files-dev:
	cd deployments && docker-compose -f compose.dev.yml up -d nginx files-service main-db

# Cleanup
down:
	cd deployments && docker-compose -f compose.dev.yml down

clean:
	docker system prune -f

# Database
sqlc:
	sqlc generate