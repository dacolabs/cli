#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

run_codegen() {
  if [[ -n "${OAPI_CODEGEN:-}" ]]; then
    "$OAPI_CODEGEN" "$@"
  elif command -v oapi-codegen >/dev/null 2>&1; then
    oapi-codegen "$@"
  else
    go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 "$@"
  fi
}

run_codegen -config openapi/oapi-codegen.yaml openapi/catalog-api.yaml
gofmt -w internal/catalogapi/client.gen.go
if grep -R "github.com/dacolabs/daco" internal/catalogapi >/dev/null 2>&1; then
  echo "error: generated client imports github.com/dacolabs/daco" >&2
  exit 1
fi
echo "generated internal/catalogapi/client.gen.go"
