## daco ports translate

Translate a port schema to a target format

### Synopsis

Translate a port schema to a target format.

Available formats: avro, databricks-sql, spark-scala, databricks-scala, spark-sql, pyspark, gotypes, databricks-pyspark, pydantic, python, scala, protobuf

```
daco ports translate [flags]
```

### Examples

```
  # Interactive mode
  daco ports translate

  # Translate specific port
  daco ports translate --name my-port --format pyspark --output schema.py

  # Translate to a custom directory (also sets package name for Go/Protobuf/Scala)
  daco ports translate --format gotypes --output-dir models
```

### Options

```
  -f, --format string       Output format (databricks-scala, spark-sql, avro, databricks-sql, spark-scala, pydantic, python, scala, protobuf, pyspark, gotypes, databricks-pyspark)
  -h, --help                help for translate
  -n, --name string         Port name (translates all if not specified)
  -o, --output string       Output file path (only valid when translating a single port)
  -d, --output-dir string   Output directory (also used as package name for Go/Protobuf/Scala) (default "schemas")
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

