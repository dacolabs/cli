## daco connections list

List all connections in the OpenDPI spec

### Synopsis

List all connections defined in the OpenDPI spec with their protocols and hosts.

```
daco connections list [flags]
```

### Examples

```
  # List connections in table format
  daco connections list

  # List connections as JSON
  daco connections list -o json

  # List connections as YAML
  daco connections list -o yaml
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format (table, json, yaml) (default "table")
```

### SEE ALSO

* [daco connections](daco_connections.md)	 - Manage data product connections

