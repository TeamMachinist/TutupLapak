# TutupLapak Kubernetes Deployment

## Quick Start

```bash
# Deploy with self-hosted infrastructure
./deploy.sh --with-infra

# Deploy apps only (assumes infrastructure exists)
./deploy.sh --apps-only

# Switch to managed infrastructure
./deploy.sh --managed-infra
```

## Files

- `00-infrastructure.yaml` - PostgreSQL, Redis, MinIO
- `01-configmaps.yaml` - Application configurations
- `02-deployments.yaml` - Application deployments
- `03-services.yaml` - Kubernetes services
- `04-hpa.yaml` - Horizontal Pod Autoscaler
- `05-ingress.yaml` - Ingress configuration
- `deploy.sh` - Main deployment script
- `manage-infra.sh` - Infrastructure management

## Infrastructure Management

```bash
# Scale down infrastructure (saves resources)
./manage-infra.sh --scale-down

# Scale up infrastructure
./manage-infra.sh --scale-up

# Check infrastructure status
./manage-infra.sh --status
```

## Load Test Strategy

1. Test with self-hosted infrastructure first
2. Scale down self-hosted infrastructure before load test
3. Update managed infrastructure credentials in ConfigMaps
4. Switch to managed infrastructure for load test

## Health Check

```bash
kubectl port-forward service/auth-service 8001:8001 -n machinist-tutuplapak
curl http://localhost:8001/healthz
```

## Logs

```bash
kubectl logs -f deployment/auth-deployment -n machinist-tutuplapak
kubectl logs -f deployment/core-deployment -n machinist-tutuplapak
kubectl logs -f deployment/files-deployment -n machinist-tutuplapak
```