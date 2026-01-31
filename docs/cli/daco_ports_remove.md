## daco ports remove

Remove a port from the OpenDPI spec

### Synopsis

Remove a port from the OpenDPI spec.
If no port name is provided, an interactive selection prompt is shown.
Requires confirmation unless --force is specified.

```
daco ports remove [PORT_NAME] [flags]
```

### Examples

```
  # Interactive selection
  daco ports remove

  # Remove with confirmation prompt
  daco ports remove user_events

  # Remove without confirmation
  daco ports remove user_events --force
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for remove
```

### SEE ALSO

* [daco ports](daco_ports.md)	 - Manage data product ports

