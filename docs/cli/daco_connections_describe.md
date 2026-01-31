## daco connections describe

Show detailed information about a connection

### Synopsis

Display complete connection details including which ports use it. If no connection name is provided, an interactive selection prompt is shown.

```
daco connections describe [CONNECTION_NAME] [flags]
```

### Examples

```
  # Interactive selection
  daco connections describe

  # Show connection details
  daco connections describe kafka_prod

  # Show as JSON
  daco connections describe kafka_prod -o json

  # Show as YAML
  daco connections describe kafka_prod -o yaml
```

### Options

```
  -h, --help            help for describe
  -o, --output string   Output format (text, json, yaml) (default "text")
```

### SEE ALSO

* [daco connections](daco_connections.md)	 - Manage data product connections

