# Myport

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes |  |
| `metadata` | [Metadata](#Metadata) | Yes |  |
| `spec` | [Spec](#Spec) | Yes |  |


---

## Metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `owner` | string | Yes |  |
| `environment` | string | No |  (enum: `prod`, `staging`, `dev`) |


---

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes |  |
| `type` | string | Yes |  (enum: `string`, `integer`, `boolean`, `float`) |


---

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `fields` | array([Fields](#Fields)) | No |  |

