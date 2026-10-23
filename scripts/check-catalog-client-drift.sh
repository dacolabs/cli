#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
tmp_dir="$(mktemp -d)"
cp internal/catalogapi/client.gen.go "$tmp_dir/client.gen.go"
restore_client() {
  cp "$tmp_dir/client.gen.go" internal/catalogapi/client.gen.go
  rm -rf "$tmp_dir"
}
trap restore_client EXIT
./scripts/generate-catalog-client.sh >/dev/null
if ! diff -u "$tmp_dir/client.gen.go" internal/catalogapi/client.gen.go; then
  echo "error: internal/catalogapi/client.gen.go is stale; run ./scripts/generate-catalog-client.sh" >&2
  exit 1
fi
# Successful check: leave the regenerated (identical) file; only remove tmp.
trap 'rm -rf "$tmp_dir"' EXIT
echo "catalog client drift check clean"
