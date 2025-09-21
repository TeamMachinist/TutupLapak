#!/bin/bash

set -e

NAMESPACE="machinist-tutuplapak"
KUBECTL_CONTEXT=""

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -c, --context CONTEXT    Kubectl context to use"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 --context my-k8s-cluster"
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--context)
            KUBECTL_CONTEXT="$2"
            shift 2
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

echo "Checking if namespace exists..."
if ! $KUBECTL_CMD get namespace $NAMESPACE >/dev/null 2>&1; then
    echo "Creating namespace: $NAMESPACE"
    $KUBECTL_CMD create namespace $NAMESPACE
else
    echo "Namespace $NAMESPACE already exists"
fi

echo ""
echo "Applying ConfigMaps..."
$KUBECTL_CMD apply -f ./01-configmaps.yaml

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
echo "Waiting for deployments to be ready..."
$KUBECTL_CMD rollout status deployment/auth-deployment -n $NAMESPACE --timeout=300s
$KUBECTL_CMD rollout status deployment/core-deployment -n $NAMESPACE --timeout=300s
$KUBECTL_CMD rollout status deployment/files-deployment -n $NAMESPACE --timeout=300s

echo ""
echo "Checking pod status..."
$KUBECTL_CMD get pods -n $NAMESPACE

echo ""
echo "Checking services..."
$KUBECTL_CMD get services -n $NAMESPACE

echo ""
echo "Checking HPA status..."
$KUBECTL_CMD get hpa -n $NAMESPACE

echo ""
echo "Checking ingress..."
$KUBECTL_CMD get ingress -n $NAMESPACE

echo ""
echo "Getting ingress details..."
$KUBECTL_CMD describe ingress tutuplapak-ingress -n $NAMESPACE

echo ""
echo "Deployment completed successfully!"
echo ""
echo "To check logs:"
echo "  $KUBECTL_CMD logs -f deployment/auth-deployment -n $NAMESPACE"
echo "  $KUBECTL_CMD logs -f deployment/core-deployment -n $NAMESPACE"
echo "  $KUBECTL_CMD logs -f deployment/files-deployment -n $NAMESPACE"
echo ""
echo "To test health endpoints:"
echo "  $KUBECTL_CMD port-forward service/auth-service 8001:8001 -n $NAMESPACE"
echo "  curl http://localhost:8001/healthz"
echo ""
echo "To update ConfigMaps, edit the files and run:"
echo "  $KUBECTL_CMD apply -f k8s/configmaps/"
echo "  $KUBECTL_CMD rollout restart deployment/auth-deployment -n $NAMESPACE"
echo "  $KUBECTL_CMD rollout restart deployment/core-deployment -n $NAMESPACE"
echo "  $KUBECTL_CMD rollout restart deployment/files-deployment -n $NAMESPACE"