---
page_title: "Resource: cumulocity_application_binary"
description: |-
  Uploads a ZIP archive to a Cumulocity application and sets it as the active version.
---

# cumulocity_application_binary

Uploads a ZIP archive to an existing `cumulocity_application` and sets it as the active binary version. Each upload creates a new binary version; the previous binary remains on the server until it is deleted.

All attributes that affect the binary content are immutable — changing `file` or `file_hash` destroys the old binary resource (deletes it from Cumulocity) and uploads a new one.

Corresponds to `POST /application/applications/{id}/binaries` and `DELETE /application/applications/{id}/binaries/{binaryId}`.

## Example Usage

### Upload on first deploy

```hcl
resource "cumulocity_application" "webui" {
  key          = "my-webui-key"
  name         = "My Web UI"
  type         = "HOSTED"
  context_path = "my-webui"
}

resource "cumulocity_application_binary" "webui" {
  application_id = cumulocity_application.webui.id
  file           = "${path.module}/dist/webui.zip"
  file_hash      = filemd5("${path.module}/dist/webui.zip")
}
```

### Re-upload when content changes

The `file_hash` attribute is the key to triggering re-uploads. If you change the ZIP content without changing the file path, Terraform will not detect the change unless `file_hash` is set:

```hcl
resource "cumulocity_application_binary" "webui" {
  application_id = cumulocity_application.webui.id
  file           = "${path.module}/dist/webui.zip"

  # filemd5() / filesha256() re-evaluates every plan — new hash = destroy + upload
  file_hash = filemd5("${path.module}/dist/webui.zip")
}
```

### Microservice ZIP

```hcl
resource "cumulocity_application_binary" "svc" {
  application_id = cumulocity_application.my_service.id
  file           = "${path.module}/build/my-service.zip"
  file_hash      = filesha256("${path.module}/build/my-service.zip")
}
```

## Schema

### Required

- `application_id` (String) — ID of the application to upload to. Changing this value forces a new resource.
- `file` (String) — Local path to the ZIP file. Changing this value forces a new resource.

### Optional

- `file_hash` (String) — Hash of the file content, e.g. `filemd5("path/app.zip")` or `filesha256(...)`. When this value changes (file content changed at the same path), the binary is destroyed and re-uploaded. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{applicationId}/{binaryId}`.
- `binary_id` (String) — ID of the binary attachment in Cumulocity.
- `name` (String) — Filename of the uploaded archive.
- `length` (Number) — Size of the archive in bytes.
- `created` (String) — ISO 8601 timestamp when the binary was uploaded.
- `download_url` (String) — URL to download the uploaded archive.

## Import

Import an existing binary using `{applicationId}/{binaryId}`.

```shell
terraform import cumulocity_application_binary.webui 20200301/30874797
```

~> **Note:** After import, the `file` and `file_hash` attributes will be empty in state. You must set them in your configuration to enable content-change detection on the next `terraform plan`.
