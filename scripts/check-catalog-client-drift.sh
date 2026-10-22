#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
cp internal/catalogapi/client.gen.go "$tmp_dir/client.gen.go"
./scripts/generate-catalog-client.sh >/dev/null
if ! diff -u "$tmp_dir/client.gen.go" internal/catalogapi/client.gen.go; then
  echo "error: internal/catalogapi/client.gen.go is stale; run ./scripts/generate-catalog-client.sh" >&2
  cp "$tmp_dir/client.gen.go" internal/catalogapi/client.gen.go
  exit 1
fi
echo "catalog client drift check clean"
