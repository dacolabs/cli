## daco init

Initialize a new daco project

### Synopsis

Initialize a new daco project with a daco.yaml configuration file and an OpenDPI spec.

```
daco init [flags]
```

### Examples

```
  # Interactive mode
  daco init

  # Non-interactive (runs automatically when --name is provided)
  daco init --name "Customer Analytics"
```

### Options

```
  -d, --description string   Data product description
  -h, --help                 help for init
  -n, --name string          Data product name
  -v, --version string       Initial spec version (default "1.0.0")
```

### SEE ALSO

* [daco](daco.md)	 - Data product CLI tool

