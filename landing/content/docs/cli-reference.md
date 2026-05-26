---
title: CLI Reference
description: Every Daco command, flag, and argument documented in one place.
order: 3
section: CLI Reference
---

## Global behavior

All commands except `daco init` require a `daco.yaml` config file in the current directory (or a parent directory). That file points to your spec:

```yaml
version: 1
path: ./spec
```

Run any command with `--help` for a quick reference:

```bash
daco --help
daco ports --help
daco ports translate --help
```

---

## daco init

Initialize a new data product in the current directory.

```bash
daco init [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--name` | `-n` | Data product name |
| `--version` | `-v` | Initial version (default: `1.0.0`) |
| `--description` | `-d` | Short description |

If `--name` is omitted, the command runs an interactive form. With `--name` set, the remaining flags are optional and default to sensible values.

**Creates:**

```
daco.yaml
spec/
  opendpi.yaml
```

---

## daco describe

Print a full overview of your data product: metadata, all connections, and all ports with their schemas and connection bindings.

```bash
daco describe
```

No flags or arguments. Requires `daco.yaml`.

---

## daco ports

### ports list

List all ports defined in the spec.

```bash
daco ports list
```

Output columns: `NAME`, `DESCRIPTION`, `CONNECTIONS` (as `connection_name:location`).

### ports add

Add a new port to the spec.

```bash
daco ports add [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--name` | `-n` | Port name |
| `--description` | `-d` | Port description |
| `--connection` | `-c` | Connection to bind (must already exist) |
| `--location` | `-l` | Location within the connection |

If `--connection` is provided, `--location` is required. Run without flags for the interactive form.

**Creates** `spec/schemas/<port_name>.schema.yaml` — an empty template you fill in with JSON Schema.

### ports describe

Show detailed information about a port.

```bash
daco ports describe [port_name]
```

If `port_name` is omitted, an interactive selector appears. Output includes the connection table and an inline schema preview.

### ports remove

Remove a port from the spec.

```bash
daco ports remove [port_name] [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--force` | `-f` | Skip the confirmation prompt |

The associated schema file is not automatically deleted — remove it manually if you no longer need it.

### ports translate

Translate one or more port schemas to a target format and write the output to disk.

```bash
daco ports translate [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--name` | `-n` | Port name(s), comma-separated |
| `--all` | `-a` | Translate every port |
| `--format` | | Output format (see below) |
| `--output` | `-o` | Output directory (default: `schemas`) |

`--all` and `--name` are mutually exclusive. Run without flags for the interactive form.

**Available formats:**

| Format | Output | Extension |
|---|---|---|
| `pyspark` | PySpark `StructType` | `.py` |
| `databricks-pyspark` | Databricks-flavored PySpark | `.py` |
| `pydantic` | Pydantic v2 model | `.py` |
| `python` | Python dataclass | `.py` |
| `gotypes` | Go struct with JSON tags | `.go` |
| `scala` | Scala case class | `.scala` |
| `spark-scala` | Spark Scala schema | `.scala` |
| `databricks-scala` | Databricks Scala schema | `.scala` |
| `avro` | Avro schema (JSON) | `.avsc` |
| `protobuf` | Protocol Buffers | `.proto` |
| `spark-sql` | Spark SQL `CREATE TABLE` | `.sql` |
| `databricks-sql` | Databricks SQL `CREATE TABLE` | `.sql` |
| `dqx-yaml` | DQX quality checks | `.yaml` |
| `markdown` | Human-readable table | `.md` |

**Examples:**

```bash
# Translate all ports to PySpark
daco ports translate --all --format pyspark --output models/

# Translate two specific ports to Pydantic
daco ports translate --name users,orders --format pydantic --output models/

# Interactive
daco ports translate
```

---

## daco connections

### connections list

List all connections defined in the spec.

```bash
daco connections list
```

Output columns: `NAME`, `TYPE`, `HOST`, `DESCRIPTION`, `USED BY` (number of ports referencing this connection).

### connections add

Add a new connection to the spec.

```bash
daco connections add [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--name` | `-n` | Connection name |
| `--type` | `-t` | Connection type (e.g., `postgresql`, `kafka`, `s3`) |
| `--host` | | Host or endpoint address |
| `--description` | `-d` | Description |

Non-interactive mode requires `--name`, `--type`, and `--host`. Run without flags for the interactive form.

### connections describe

Show detailed information about a connection.

```bash
daco connections describe [connection_name]
```

If `connection_name` is omitted, an interactive selector appears. Output includes all variables and the list of ports that use this connection.

### connections remove

Remove a connection from the spec.

```bash
daco connections remove [connection_name] [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--force` | `-f` | Skip the confirmation prompt |

Connections referenced by one or more ports cannot be removed. Remove or update those ports first.

---

## daco product

### product version

Print the current version of the data product.

```bash
daco product version
```

### product upgrade

Bump the product version.

```bash
daco product upgrade [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--bump` | `-b` | Bump type: `major`, `minor`, or `patch` |

Run without flags for an interactive prompt. The version follows semantic versioning (`MAJOR.MINOR.PATCH`).

```bash
daco product upgrade --bump patch   # 1.2.3 → 1.2.4
daco product upgrade --bump minor   # 1.2.3 → 1.3.0
daco product upgrade --bump major   # 1.2.3 → 2.0.0
```
