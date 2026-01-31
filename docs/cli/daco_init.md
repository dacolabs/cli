## daco init

Initialize a new daco project

### Synopsis

Initialize a new daco project with a daco.yaml configuration file.
Can create a new spec, use an existing one, or extend a parent config.

```
daco init [flags]
```

### Examples

```
  # Interactive mode
  daco init

  # Non-interactive
  daco init --name "my-product" --non-interactive
  daco init --extends ../daco.yaml --non-interactive
```

### Options

```
  -e, --extends string               Path to parent daco.yaml
  -f, --format string                Spec format (yaml or json) (default "yaml")
  -h, --help                         help for init
  -n, --name string                  Project name
      --non-interactive              Run without prompts (requires --name or --extends)
  -p, --path string                  Path to spec folder (default "./spec")
  -s, --schema-organization string   Schema organization (modular, components, or inline) (default "modular")
  -v, --version string               Initial spec version (default "1.0.0")
```

### SEE ALSO

* [daco](daco.md)	 - Data product CLI tool

