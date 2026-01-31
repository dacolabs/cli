## daco connections add

Add a new connection to the OpenDPI spec

### Synopsis

Add a new infrastructure connection to the OpenDPI spec.

```
daco connections add [flags]
```

### Examples

```
  # Interactive mode
  daco connections add

  # Non-interactive
  daco connections add -n kafka_prod -p kafka --host broker:9092 --non-interactive
```

### Options

```
  -d, --description string   Description
  -h, --help                 help for add
      --host string          Host/endpoint
  -n, --name string          Connection name
      --non-interactive      Run without prompts
  -p, --protocol string      Protocol (kafka, postgresql, mysql, s3, http, etc.)
```

### SEE ALSO

* [daco connections](daco_connections.md)	 - Manage data product connections

