# TutupLapak

Ayuk jualan barangmu sambil tutup lapak offline mu 🤩

## Quick Start

### 1. Development Commands

```bash
make dev          # Start all services
make dev-build    # Force rebuild + start
make auth-dev     # Start auth + nginx + db only
make core-dev     # Start core + nginx + db only
make files-dev    # Start files + nginx + db only
make down         # Stop all services
make sqlc         # Generate database code
```

### 2. Test Health Endpoints
```bash
curl http://localhost/healthz        # nginx
curl http://localhost/healthz/auth   # auth service
curl http://localhost/healthz/core   # core service  
curl http://localhost/healthz/files  # files service
```

## Services

- **Auth Service**: Port 8001 - User registration/login
- **Core Service**: Port 8002 - Products, users, purchases  
- **Files Service**: Port 8003 - File upload/download
- **Nginx**: Port 80 - API Gateway

## Database Integration

### Setup SQLC (Everyone)
```bash
# Install SQLC - follow platform-specific instructions:
# https://docs.sqlc.dev/en/stable/overview/install.html

# Generate type-safe queries and models (run in project root)
sqlc generate

# IMPORTANT: Commit the generated files
git add internal/database/
git commit -m "Generate SQLC database code"
```

**Rule: Whoever adds/modifies queries runs `sqlc generate` and commits the results.**

### Usage in Services
```go
// In main.go
db, err := internal.NewDatabase(ctx, DatabaseURL)
defer db.Close()

// Initialize layers
repo := NewRepository(db.Queries)
service := NewService(repo)
handler := NewHandler(service)

// Use generated models
var user database.UsersAuth  // SQLC generated model
var file database.Files      // SQLC generated model
```
