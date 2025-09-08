# TutupLapak

Ayuk jualan barangmu sambil tutup lapak offline mu 🤩

## Quick Start

### 1. Start Services
```bash
cd deployments
docker-compose up --build
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

## Next Steps

1. Add database integration (SQLC)
2. Implement authentication (JWT)
3. Add Redis caching
4. Add MinIO file storage
5. Implement business logic