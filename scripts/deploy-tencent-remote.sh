#!/usr/bin/env bash
# Run on Tencent CVM after images are pushed to TCR.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

: "${TCR_REGISTRY:?set TCR_REGISTRY}"
: "${IMAGE_TAG:?set IMAGE_TAG}"

export TCR_REGISTRY IMAGE_TAG

echo "Pulling ${TCR_REGISTRY}/*:${IMAGE_TAG} ..."
docker compose -f deploy/docker-compose.prod.yml pull

echo "Starting stack ..."
docker compose -f deploy/docker-compose.prod.yml up -d --remove-orphans

docker compose -f deploy/docker-compose.prod.yml ps
