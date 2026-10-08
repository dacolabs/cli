#!/usr/bin/env bash
# Refresh openapi/catalog-api.yaml from the product repo Catalog contract.
# Usage:
#   CATALOG_OPENAPI=/path/to/daco/api/catalog-api.yaml ./scripts/sync-catalog-openapi.sh
# When this CLI is checked out as the daco monorepo submodule at cli/, the
# default source is ../api/catalog-api.yaml.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
src="${CATALOG_OPENAPI:-}"
if [[ -z "$src" ]]; then
  if [[ -f "$root/../api/catalog-api.yaml" ]]; then
    src="$root/../api/catalog-api.yaml"
  else
    echo "set CATALOG_OPENAPI to api/catalog-api.yaml from dacolabs/daco" >&2
    exit 1
  fi
fi
python3 - "$src" "$root/openapi/catalog-api.yaml" <<'PY'
import sys
from pathlib import Path

def strip_go_type_extensions(text: str) -> str:
    lines = text.splitlines(keepends=True)
    out = []
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.lstrip(" ")
        indent = len(line) - len(stripped)
        if stripped.startswith("x-go-type:"):
            i += 1
            continue
        if stripped.startswith("x-go-type-import:"):
            i += 1
            while i < len(lines):
                nxt = lines[i]
                if not nxt.strip():
                    j = i + 1
                    while j < len(lines) and not lines[j].strip():
                        j += 1
                    if j >= len(lines):
                        break
                    nindent = len(lines[j]) - len(lines[j].lstrip(" "))
                    if nindent <= indent:
                        break
                    i += 1
                    continue
                nindent = len(nxt) - len(nxt.lstrip(" "))
                if nindent <= indent:
                    break
                i += 1
            continue
        out.append(line)
        i += 1
    return "".join(out)

src = Path(sys.argv[1]).read_text()
out = strip_go_type_extensions(src)
if "github.com/dacolabs/daco" in out:
    raise SystemExit("sanitized OpenAPI still references github.com/dacolabs/daco")
if "x-go-type" in out:
    raise SystemExit("sanitized OpenAPI still contains x-go-type")
if "DatasetList:" not in out:
    raise SystemExit("sanitized OpenAPI is missing DatasetList schema")
Path(sys.argv[2]).write_text(out)
print(f"synced {sys.argv[1]} -> {sys.argv[2]}")
PY
