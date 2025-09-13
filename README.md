# TutupLapak

Ayuk jualan barangmu sambil tutup lapak offline mu 🤩

## Development Options

### Option 1: Docker Compose (Recommended)
Run everything in containers with live reload:
```bash
make up-dev          # Start all services
make up-dev-build    # Force rebuild + start all
```

### Option 2: Individual Services
Run specific services with their dependencies:
```bash
make up-auth         # Start auth + database only
make up-core         # Start core + database only  
make up-files        # Start files + database only
make up-db           # Start database only
```

### Option 3: Local Development
Run database in Docker, services locally:

1. **Start database only:**
   ```bash
   make up-db
   ```

2. **Copy environment file to each service directory:**
   ```bash
   cp deployments/.env services/auth/.env
   cp deployments/.env services/core/.env
   cp deployments/.env services/files/.env
   ```

3. **Run service locally:**
   ```bash
   cd services/auth
   go run .
   ```

### Step 3: Load Test Data (Optional)
```bash
make seed-dev        # Load test users, products, and files
```

See Makefile for all available commands and detailed explanations.

## Health Check Endpoints
```bash
# Health endpoints (requires make up-dev to spawn all services)
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

## Database Setup

### Generate Database Code
```bash
# Install SQLC first: https://docs.sqlc.dev/en/stable/overview/install.html

# Generate type-safe Go code from SQL queries
sqlc generate

# Commit generated files
git add internal/database/
git commit -m "Generate SQLC database code"
```

**Rule: Whoever adds/modifies queries runs `sqlc generate` and commits the results.**

## Architecture

- **Multi-service**: Separate auth, core business logic, and file handling
- **Shared Database**: PostgreSQL with SQLC for type-safe queries
- **API Gateway**: Nginx for request routing
- **Live Reload**: Air for development iterations
- **File Storage**: MinIO for file uploads
- **Authentication**: JWT tokens for API protection