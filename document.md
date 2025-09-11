# TutupLapak - Project Instructions & Progress

## Current Project Status ✅

### Architecture Decisions Made:
- **Microservices**: Auth, Core, Files services
- **Single Database**: PostgreSQL shared across services (for bootcamp simplicity)
- **API Gateway**: Nginx for routing and single entry point
- **Authentication**: JWT with local validation (no inter-service auth calls)
- **Live Reload**: Air for development
- **Database**: SQLC for type-safe queries with pgx driver

### Working Components:
- ✅ Basic project structure
- ✅ Docker Compose setup with live reload
- ✅ Nginx routing working
- ✅ Health endpoints (`/healthz/*`)
- ✅ Go workspace setup (go.work)
- ✅ All services starting correctly

## Project Structure (Final)

```
tutuplapak/
├── services/
│   ├── auth/          # Port 8001 - JWT generation, user login
│   ├── core/          # Port 8002 - Products, users, purchases  
│   └── files/         # Port 8003 - File upload/download
├── internal/          # Shared packages
│   ├── database/      # Generated SQLC code (target)
│   ├── queries/       # SQL queries for SQLC
│   ├── jwt.go         # JWT utilities
│   ├── middleware.go  # HTTP middleware
│   └── ...
├── migrations/        # Database schema
├── configs/          # nginx.conf
├── deployments/      # compose.yml
├── sqlc.yaml         # SQLC configuration
├── go.work           # Go workspace
└── README.md
```

How it's going now

```
.
├── configs
│   └── nginx.conf
├── deployments
│   └── compose.yml
├── doc.md
├── go.work
├── internal
│   └── database.go
├── LICENSE
├── migrations
│   └── 001_create_users_auth_table.sql
├── README.md
└── services
    ├── auth
    │   ├── Dockerfile.dev
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   └── tmp
    ├── core
    │   ├── Dockerfile.dev
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   └── tmp
    └── files
        ├── Dockerfile.dev
        ├── go.mod
        ├── go.sum
        ├── main.go
        └── tmp
```

## Key Technical Decisions

### 1. Database Strategy
- **Single PostgreSQL instance** shared by all services
- **SQLC with pgx** for type-safe queries
- **Migrations in root** `/migrations/` directory
- **Generated code** in `/internal/database/`

### 2. Authentication Flow
```
Client → Auth Service → JWT Token
Client → Core/Files Service → Local JWT Validation → Process Request
```
- **No inter-service auth calls**
- **Shared JWT secret** across services
- **JWT validation in each service** using `internal/jwt.go`

### 3. Development Setup
- **Go workspace** for single VS Code window
- **Air for live reload** in each service
- **Individual service development**: `make auth-dev`, `make core-dev`, etc.
- **Health endpoints**: `/healthz` (Kubernetes standard)

## Next Steps: Database Integration with SQLC

### 1. Install SQLC
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 2. Create Sample Query
```sql
-- internal/queries/users.sql
-- name: CreateUserAuth :one
INSERT INTO users_auth (email, phone, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, phone, created_at;
```

### 3. Generate SQLC Code
```bash
# In project root
sqlc generate
```

### 4. Use in Services
```go
// Example in auth service
import "tutuplapak/internal/database"

db, err := database.NewConnection(ctx, os.Getenv("DB_URL"))
user, err := db.CreateUserAuth(ctx, database.CreateUserAuthParams{...})
```

## Important Configurations

### SQLC Config (sqlc.yaml in root):
- **Engine**: postgresql
- **Queries**: `./internal/queries/`
- **Schema**: `./migrations/`
- **Output**: `./internal/database/`
- **SQL Package**: pgx/v5

### Docker Compose Features:
- **Live reload** with Air
- **Single database** (main-db)
- **Nginx gateway** on port 80
- **Service-specific health** endpoints

### Nginx Routing:
- `/v1/login/*` → auth-service
- `/v1/product/*`, `/v1/user/*`, `/v1/purchase/*` → core-service
- `/v1/file/*` → files-service
- `/healthz/*` → service-specific health checks

## Development Commands

```bash
# Full stack
make dev

# Individual services (with nginx)
make auth-dev    # nginx + auth + db
make core-dev    # nginx + core + db  
make files-dev   # nginx + files + db

# Database
make db-reset    # Reset with fresh data
sqlc generate    # Generate type-safe queries

# Testing
curl http://localhost/healthz/auth
curl http://localhost/healthz/core
curl http://localhost/healthz/files
```


## Key Learnings & Gotchas

1. **Go workspace**: Close VS Code completely after creating go.work
2. **Nginx location order**: Specific routes before general ones
3. **Docker context paths**: Use `../` when compose.yml is in subdirectory
4. **Air setup**: Remove go.sum from Dockerfile.dev if no dependencies yet
5. **Health endpoints**: Use `/healthz` for Kubernetes compatibility

---

**Status**: Ready for database integration with SQLC and business logic implementation.