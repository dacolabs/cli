## daco ports translate

Translate a port schema to a target format

### Synopsis

Translate a port schema to a target format.

Available formats: spark-scala, dqx-yaml, pyspark, avro, pydantic, python, databricks-sql, protobuf, gotypes, databricks-pyspark, scala, databricks-scala, spark-sql, markdown

```
daco ports translate [flags]
```

### Examples

```
  # Interactive mode
  daco ports translate

  # Translate specific port
  daco ports translate --name my-port --format pyspark

  # Translate multiple ports
  daco ports translate --name port-a,port-b --format pyspark

  # Translate all ports
  daco ports translate --all --format pyspark

  # Translate to a custom output directory (also sets package name for Go/Protobuf/Scala)
  daco ports translate --all --format gotypes --output models
```

### Options

```
  -a, --all             Translate all ports
      --format string   Output format (spark-scala, dqx-yaml, pyspark, avro, pydantic, python, databricks-sql, protobuf, gotypes, databricks-pyspark, scala, databricks-scala, spark-sql, markdown)
  -h, --help            help for translate
  -n, --name string     Port name(s), comma-separated
  -o, --output string   Output directory (also used as package name for Go/Protobuf/Scala) (default "schemas")
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

