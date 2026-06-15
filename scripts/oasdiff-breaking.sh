#!/usr/bin/env bash
# Compare OpenAPI specs on the current branch against a base ref (default origin/main).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPECS_DIR="$ROOT/platform-contracts/openapi"
BASE_REF="${OASDIFF_BASE_REF:-origin/main}"

if ! git -C "$ROOT" rev-parse --verify "$BASE_REF" >/dev/null 2>&1; then
  echo "oasdiff-breaking: skip (base ref $BASE_REF not found)"
  exit 0
fi

if ! command -v oasdiff >/dev/null 2>&1; then
  echo "oasdiff-breaking: install with: go install github.com/tufin/oasdiff/cmd/oasdiff@latest"
  exit 1
fi

failed=0
for spec in "$SPECS_DIR"/*.yaml; do
  name="$(basename "$spec")"
  rel="platform-contracts/openapi/$name"
  if ! git -C "$ROOT" cat-file -e "$BASE_REF:$rel" 2>/dev/null; then
    echo "oasdiff-breaking: new spec $name (no base on $BASE_REF)"
    continue
  fi
  git -C "$ROOT" show "$BASE_REF:$rel" > /tmp/oas-base-"$name"
  echo "oasdiff-breaking: $name"
  if ! oasdiff breaking "/tmp/oas-base-$name" "$spec"; then
    failed=1
  fi
done

if [ "$failed" -ne 0 ]; then
  echo "oasdiff-breaking: breaking OpenAPI change detected"
  exit 1
fi

echo "oasdiff-breaking: OK"
