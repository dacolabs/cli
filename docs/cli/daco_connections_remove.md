## daco connections remove

Remove a connection from the OpenDPI spec

### Synopsis

Remove a connection from the OpenDPI spec. Cannot remove connections that are in use by ports. If no connection name is provided, an interactive selection prompt is shown.

```
daco connections remove [CONNECTION_NAME] [flags]
```

### Examples

```
  # Interactive selection
  daco connections remove

  # Remove with confirmation
  daco connections remove unused_conn

  # Remove without confirmation
  daco connections remove unused_conn --force
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for remove
```

### SEE ALSO

* [daco connections](daco_connections.md)	 - Manage data product connections

