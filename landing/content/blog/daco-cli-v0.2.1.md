---
title: "Daco CLI Changelog: v0.2.1"
date: "2026-02-09"
tag: "Changelog"
excerpt: "v0.2.1 adds product versioning, upgrade commands, and new translators for DQX YAML and Markdown."
cover: "c1"
glyph: "v0.2.1"
---

Daco CLI v0.2.1 is out. This release adds version management, two new schema translation formats, and fixes for translation reliability.

## Version Management

You can now manage product versions directly from the CLI:

```bash
daco version
daco upgrade
```

The `upgrade` command checks for a newer version of the CLI and applies it in place.

## New Schema Translators

Two new translation targets are available:

**DQX YAML** — converts an OpenDPI port schema to Databricks Data Quality (DQX) format:

```bash
daco ports translate --format dqx-yaml my-port
```

**Markdown** — generates readable documentation from a port schema, useful for wikis and PR descriptions:

```bash
daco ports translate --format markdown my-port
```

## Bug Fixes

- Fixed column order not being preserved during JSON Schema translation
- Restored external schema references in `opendpi.yaml` files when using `daco ports add`

## Upgrading

```bash
# Homebrew
brew upgrade daco

# Scoop
scoop update daco

# Go
go install github.com/dacolabs/cli/cmd/daco@v0.2.1
```

## Contributors

Thanks to Giuseppe Grieco and Gudmundur Orri Palsson for this release.
