#!/bin/bash

set -e

NAMESPACE="machinist-tutuplapak"
KUBECTL_CONTEXT=""

usage() {
    echo "Usage: $0 [OPTIONS] ACTION"
    echo "Actions:"
    echo "  --scale-down              Scale infrastructure to 0 replicas"
    echo "  --scale-up                Scale infrastructure to 1 replica"
    echo "  --status                  Show infrastructure status"
    echo ""
    echo "Options:"
    echo "  -c, --context CONTEXT     Kubectl context to use"
    echo "  -h, --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --scale-down          # Scale down for load test (saves resources)"
    echo "  $0 --scale-up            # Scale up after load test"
    echo "  $0 --status              # Check infrastructure status"
}

ACTION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--context)
            KUBECTL_CONTEXT="$2"
            shift 2
            ;;
        --scale-down)
            ACTION="scale-down"
            shift
            ;;
        --scale-up)
            ACTION="scale-up"
            shift
            ;;
        --status)
            ACTION="status"
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

if [[ -z "$ACTION" ]]; then
    echo "Error: No action specified"
    usage
    exit 1
fi

KUBECTL_CMD="kubectl"
if [[ -n "$KUBECTL_CONTEXT" ]]; then
    KUBECTL_CMD="kubectl --context=$KUBECTL_CONTEXT"
fi

case $ACTION in
    "scale-down")
        echo "Scaling down infrastructure to save resources..."
        $KUBECTL_CMD scale deployment postgres --replicas=0 -n $NAMESPACE
        $KUBECTL_CMD scale deployment redis --replicas=0 -n $NAMESPACE
        $KUBECTL_CMD scale deployment minio --replicas=0 -n $NAMESPACE
        echo "Infrastructure scaled down"
        echo "Remember to switch to managed configs: ./deploy.sh --managed-infra"
        ;;
    "scale-up")
        echo "Scaling up infrastructure..."
        $KUBECTL_CMD scale deployment postgres --replicas=1 -n $NAMESPACE
        $KUBECTL_CMD scale deployment redis --replicas=1 -n $NAMESPACE
        $KUBECTL_CMD scale deployment minio --replicas=1 -n $NAMESPACE
        
        echo "Waiting for infrastructure to be ready..."
        $KUBECTL_CMD rollout status deployment/postgres -n $NAMESPACE --timeout=180s
        $KUBECTL_CMD rollout status deployment/redis -n $NAMESPACE --timeout=120s
        $KUBECTL_CMD rollout status deployment/minio -n $NAMESPACE --timeout=120s
        
        echo "Infrastructure scaled up and ready"
        echo "You can now switch back to self-hosted configs if needed"
        ;;
    "status")
        echo "Infrastructure Status:"
        echo ""
        echo "Deployments:"
        $KUBECTL_CMD get deployments postgres redis minio -n $NAMESPACE 2>/dev/null || echo "  Infrastructure not found"
        
        echo ""
        echo "Pods:"
        $KUBECTL_CMD get pods -l 'app in (postgres,redis,minio)' -n $NAMESPACE 2>/dev/null || echo "  No infrastructure pods found"
        
        echo ""
        echo "Services:"
        $KUBECTL_CMD get services main-db redis minio -n $NAMESPACE 2>/dev/null || echo "  Infrastructure services not found"
        
        echo ""
        echo "Storage:"
        $KUBECTL_CMD get pvc postgres-pvc minio-pvc -n $NAMESPACE 2>/dev/null || echo "  Infrastructure PVCs not found"
        ;;
    *)
        echo "Error: Unknown action $ACTION"
        usage
        exit 1
        ;;
esac