# Daco CLI

Official command-line interface for the [Daco](https://dacolabs.com) Catalog API.

This repository is the **CLI product** (`daco`). Language client libraries are generated in [`dacolabs/sdk`](https://github.com/dacolabs/sdk) from Catalog OpenAPI snapshots.

## Status

Early scaffold. The binary will call the remote Catalog HTTP API using WorkOS Connect M2M credentials (`client_credentials`). Interactive login and API keys are out of scope for v1.

## Development

```bash
go build -o daco .
./daco --help
```

## Configuration (planned)

| Variable | Purpose |
|---|---|
| `DACO_BASE_URL` | Catalog API origin |
| `DACO_CLIENT_ID` | WorkOS Connect M2M client ID |
| `DACO_CLIENT_SECRET` | WorkOS Connect M2M client secret |

## Related

- Product / OpenAPI source of truth: [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`)
- Generated language clients: [`dacolabs/sdk`](https://github.com/dacolabs/sdk)

## License

Apache License 2.0
