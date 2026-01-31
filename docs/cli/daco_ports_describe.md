## daco ports describe

Show detailed information about a port

### Synopsis

Display complete port definition including schema and connections. If no port name is provided, an interactive selection prompt is shown.

```
daco ports describe [PORT_NAME] [flags]
```

### Examples

```
  # Interactive selection
  daco ports describe

  # Show port details in human-readable format
  daco ports describe user_events

  # Show port details as JSON
  daco ports describe user_events -o json

  # Show port details as YAML
  daco ports describe user_events -o yaml
```

### Options

```
  -h, --help            help for describe
  -o, --output string   Output format (text, json, yaml) (default "text")
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

