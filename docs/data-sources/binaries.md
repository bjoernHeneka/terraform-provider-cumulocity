---
page_title: "Data Source: cumulocity_binaries"
description: |-
  Lists Cumulocity inventory binaries, with optional owner and type filters.
---

# cumulocity_binaries

Retrieves metadata for files stored in the Cumulocity inventory binary store, optionally filtered by owner or managed object type. All pages are followed automatically.

Corresponds to `GET /inventory/binaries`.

## Example Usage

```hcl
# All binaries owned by a given user
data "cumulocity_binaries" "mine" {
  owner = "admin"
}

output "binary_names" {
  value = [for b in data.cumulocity_binaries.mine.binaries : b.name]
}
```

## Schema

### Optional

- `owner` (String) — Filter by the username of the binary owner.
- `type` (String) — Filter by the managed object type of the binary.

### Read-Only

- `binaries` (List of Object) — List of matching binary managed objects. Each object contains:
  - `id` (String) — Binary managed object ID.
  - `name` (String) — File name.
  - `type` (String) — Managed object type.
  - `content_type` (String) — MIME content type of the stored file.
  - `length` (Number) — File size in bytes.
  - `owner` (String) — Username of the owner.
  - `self` (String) — Self-link URL.
  - `last_updated` (String) — ISO 8601 timestamp of the last update.
