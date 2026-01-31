## daco ports list

List all ports in the OpenDPI spec

### Synopsis

List all ports defined in the OpenDPI spec.
Displays port names, schema types, descriptions, and connection information.

```
daco ports list [flags]
```

### Examples

```
  # List ports in table format
  daco ports list

  # List ports as JSON
  daco ports list -o json

  # List ports as YAML
  daco ports list -o yaml
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format (table, json, yaml) (default "table")
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

