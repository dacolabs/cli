## daco ports add

Add a new port to the OpenDPI spec

### Synopsis

Add a new port to the OpenDPI spec with schema and connection configuration.
Ports can have optional schemas (from file or created interactively) and
connections that define where data flows.

```
daco ports add [flags]
```

### Examples

```
  # Interactive mode
  daco ports add

  # Non-interactive with existing connection
  daco ports add --name events --schema-file ./schemas/event.json \
    --connection kafka --location events_topic --non-interactive
```

### Options

```
  -c, --connection string    Connection name
  -d, --description string   Port description
  -h, --help                 help for add
  -l, --location string      Location (table, topic, path, etc.)
  -n, --name string          Port name
      --non-interactive      Run without prompts
  -s, --schema-file string   Path to JSON schema file
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

