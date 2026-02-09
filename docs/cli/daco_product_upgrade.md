## daco product upgrade

Upgrade the data product version

### Synopsis

Upgrade the data product version by bumping major, minor, or patch.
In interactive mode, a form is shown to select the bump type.

```
daco product upgrade [flags]
```

### Examples

```
  # Interactive mode
  daco product upgrade

  # Non-interactive
  daco product upgrade --bump minor
```

### Options

```
  -b, --bump string   Bump type: major, minor, or patch
  -h, --help          help for upgrade
```

### SEE ALSO

* [daco product](daco_product.md)	 - Manage data product metadata

