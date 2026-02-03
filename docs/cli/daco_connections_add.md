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
  daco connections add -n kafka_prod -t kafka --host broker:9092
```

### Options

```
  -d, --description string   Description
  -h, --help                 help for add
      --host string          Host/endpoint
  -n, --name string          Connection name
  -t, --type string          Connection type (kafka, postgresql, mysql, s3, http, etc.)
```

### SEE ALSO

* [daco connections](daco_connections.md)	 - Manage data product connections

