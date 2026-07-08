---
page_title: "Resource: cumulocity_binary"
description: |-
  Uploads a file to the Cumulocity inventory binary store.
---

# cumulocity_binary

Uploads a file to the Cumulocity inventory binary store. The uploaded file is represented as a binary managed object.

Because all meaningful attributes force replacement, changing the file, its hash, name, or content type deletes the old binary and uploads a new one.

Corresponds to `POST /inventory/binaries` (create), `GET /inventory/managedObjects/{id}` (read) and `DELETE /inventory/binaries/{id}` (delete).

## Example Usage

```hcl
# Upload a firmware image to the Cumulocity inventory binary store.
resource "cumulocity_binary" "firmware" {
  file         = "${path.module}/firmware-1.0.0.bin"
  file_hash    = filemd5("${path.module}/firmware-1.0.0.bin")
  name         = "firmware-1.0.0.bin"
  content_type = "application/octet-stream"
}

output "firmware_id" {
  value = cumulocity_binary.firmware.id
}
```

## Schema

### Required

- `file` (String) — Local filesystem path to the file to upload. Changing this forces a new resource.

### Optional

- `file_hash` (String) — Hash of the file content (e.g. `filemd5("path/to/file")`). When this value changes, the binary is re-uploaded. Use this to trigger a re-upload when the file path stays the same but the content changes. Changing this forces a new resource.
- `name` (String) — Name for the binary managed object. Defaults to the base filename. Changing this forces a new resource.
- `content_type` (String) — MIME content type of the file, e.g. `application/zip`. Defaults to `application/octet-stream`. Changing this forces a new resource.

### Read-Only

- `id` (String) — Binary managed object ID assigned by Cumulocity.
- `length` (Number) — File size in bytes as reported by the platform.
- `owner` (String) — Username of the owner of the binary managed object.
- `self` (String) — Self-link URL of the binary managed object.
- `last_updated` (String) — ISO 8601 timestamp of the last update.

## Import

Import an existing binary by its managed object ID:

```shell
terraform import cumulocity_binary.firmware 12345
```

~> **Note:** The `file` and `file_hash` attributes are not populated on import, as the platform does not return the original filesystem path. Set them in your configuration after importing.
