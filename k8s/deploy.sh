#!/bin/bash

set -e

NAMESPACE="machinist-tutuplapak"
KUBECTL_CONTEXT=""
DEPLOY_INFRA=false
USE_MANAGED_INFRA=false

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -c, --context CONTEXT      Kubectl context to use"
    echo "  --with-infra              Deploy infrastructure (DB, Redis, MinIO)"
    echo "  --managed-infra           Use managed infrastructure configs"
    echo "  --apps-only               Deploy only applications (default)"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --with-infra                    # Deploy everything with self-hosted infra"
    echo "  $0 --apps-only                     # Deploy only apps (use existing infra)"
    echo "  $0 --managed-infra                 # Deploy apps with managed infra configs"
    echo "  $0 --context my-cluster --with-infra  # Use specific kubectl context"
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--context)
            KUBECTL_CONTEXT="$2"
            shift 2
            ;;
        --with-infra)
            DEPLOY_INFRA=true
            shift
            ;;
        --managed-infra)
            USE_MANAGED_INFRA=true
            shift
            ;;
        --apps-only)
            DEPLOY_INFRA=false
            USE_MANAGED_INFRA=false
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            usage
            exit 1
            ;;
    esac
done

KUBECTL_CMD="kubectl"
if [[ -n "$KUBECTL_CONTEXT" ]]; then
    KUBECTL_CMD="kubectl --context=$KUBECTL_CONTEXT"
fi

echo "Deploying TutupLapak to Kubernetes namespace: $NAMESPACE"

if [[ -n "$KUBECTL_CONTEXT" ]]; then
    echo "Using context: $KUBECTL_CONTEXT"
fi

if [[ "$DEPLOY_INFRA" == true ]]; then
    echo "Will deploy infrastructure (DB, Redis, MinIO)"
elif [[ "$USE_MANAGED_INFRA" == true ]]; then
    echo "Will use managed infrastructure configs"
else
    echo "Apps-only deployment (assuming infrastructure exists)"
fi

echo ""
echo "Checking if namespace exists..."
if ! $KUBECTL_CMD get namespace $NAMESPACE >/dev/null 2>&1; then
    echo "Creating namespace: $NAMESPACE"
    $KUBECTL_CMD create namespace $NAMESPACE
else
    echo "Namespace $NAMESPACE already exists"
fi

# Deploy infrastructure first if requested
if [[ "$DEPLOY_INFRA" == true ]]; then
    echo ""
    echo "Deploying Infrastructure..."
    $KUBECTL_CMD apply -f ./00-infrastructure.yaml
    
    echo ""
    echo "Waiting for infrastructure to be ready..."
    echo "  Waiting for PostgreSQL..."
    $KUBECTL_CMD rollout status deployment/postgres -n $NAMESPACE --timeout=180s
    
    echo "  Waiting for Redis..."
    $KUBECTL_CMD rollout status deployment/redis -n $NAMESPACE --timeout=120s
    
    echo "  Waiting for MinIO..."
    $KUBECTL_CMD rollout status deployment/minio -n $NAMESPACE --timeout=120s
    
    echo "Infrastructure is ready!"
    
    # Give databases a moment to fully initialize
    echo "  Allowing 10 seconds for database initialization..."
    sleep 10
fi

echo ""
echo "Applying ConfigMaps..."
$KUBECTL_CMD apply -f ./01-configmaps.yaml

# Switch to managed configs if requested
if [[ "$USE_MANAGED_INFRA" == true ]]; then
    echo ""
    echo "Switching to managed infrastructure configs..."
    
    # Patch deployments to use managed configs
    $KUBECTL_CMD patch deployment auth-deployment -n $NAMESPACE --type='merge' -p='{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","envFrom":[{"configMapRef":{"name":"auth-config-managed"}}]}]}}}}'
    $KUBECTL_CMD patch deployment core-deployment -n $NAMESPACE --type='merge' -p='{"spec":{"template":{"spec":{"containers":[{"name":"core-service","envFrom":[{"configMapRef":{"name":"core-config-managed"}}]}]}}}}'
    $KUBECTL_CMD patch deployment files-deployment -n $NAMESPACE --type='merge' -p='{"spec":{"template":{"spec":{"containers":[{"name":"files-service","envFrom":[{"configMapRef":{"name":"files-config-managed"}}]}]}}}}'
    
    echo "WARNING: Update managed infrastructure credentials in 01-configmaps.yaml before load test!"
fi

echo ""
echo "Applying Deployments..."
$KUBECTL_CMD apply -f ./02-deployments.yaml

echo ""
echo "Applying Services..."
$KUBECTL_CMD apply -f ./03-services.yaml

echo ""
echo "Applying HPA..."
$KUBECTL_CMD apply -f ./04-hpa.yaml

echo ""
echo "Applying Ingress..."
$KUBECTL_CMD apply -f ./05-ingress.yaml

echo ""
echo "Waiting for application deployments to be ready..."
echo "  Waiting for Auth service..."
$KUBECTL_CMD rollout status deployment/auth-deployment -n $NAMESPACE --timeout=300s

echo "  Waiting for Core service..."
$KUBECTL_CMD rollout status deployment/core-deployment -n $NAMESPACE --timeout=300s

echo "  Waiting for Files service..."
$KUBECTL_CMD rollout status deployment/files-deployment -n $NAMESPACE --timeout=300s

echo ""
echo "Checking deployment status..."
echo ""
echo "Pods:"
$KUBECTL_CMD get pods -n $NAMESPACE -o wide

echo ""
echo "Services:"
$KUBECTL_CMD get services -n $NAMESPACE

echo ""
echo "HPA Status:"
$KUBECTL_CMD get hpa -n $NAMESPACE

echo ""
echo "Ingress:"
$KUBECTL_CMD get ingress -n $NAMESPACE

# Scale down infrastructure for load test if it was deployed
if [[ "$DEPLOY_INFRA" == true ]]; then
    echo ""
    echo "Quick infrastructure management commands:"
    echo "  Scale down infra for load test: ./manage-infra.sh --scale-down"
    echo "  Scale up infra after load test:  ./manage-infra.sh --scale-up"
fi

echo ""
echo "Deployment completed successfully!"
echo ""
echo "Useful commands:"
echo "  Check logs:"
echo "    $KUBECTL_CMD logs -f deployment/auth-deployment -n $NAMESPACE"
echo "    $KUBECTL_CMD logs -f deployment/core-deployment -n $NAMESPACE"
echo "    $KUBECTL_CMD logs -f deployment/files-deployment -n $NAMESPACE"
echo ""
echo "  Test health endpoints:"
echo "    $KUBECTL_CMD port-forward service/auth-service 8001:8001 -n $NAMESPACE"
echo "    curl http://localhost:8001/healthz"
echo ""
echo "  Quick restart after config changes:"
echo "    $KUBECTL_CMD rollout restart deployment/auth-deployment -n $NAMESPACE"
echo "    $KUBECTL_CMD rollout restart deployment/core-deployment -n $NAMESPACE"
echo "    $KUBECTL_CMD rollout restart deployment/files-deployment -n $NAMESPACE"
echo ""

if [[ "$USE_MANAGED_INFRA" == true ]]; then
    echo "IMPORTANT: Remember to update managed infrastructure credentials in the ConfigMaps!"
    echo "  Edit the *-config-managed ConfigMaps in 01-configmaps.yaml with actual values"
    echo ""
fi

echo "Ready for testing!"