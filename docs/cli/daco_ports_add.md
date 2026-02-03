## daco ports add

Add a new port to the OpenDPI spec

### Synopsis

Add a new port to the OpenDPI spec.
An empty schema file is created in the schemas/ directory for you to fill in.

```
daco ports add [flags]
```

### Examples

```
  # Interactive mode
  daco ports add

  # Non-interactive
  daco ports add --name events --connection kafka --location events_topic
```

### Options

```
  -c, --connection string    Connection name
  -d, --description string   Port description
  -h, --help                 help for add
  -l, --location string      Location (table, topic, path, etc.)
  -n, --name string          Port name
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

