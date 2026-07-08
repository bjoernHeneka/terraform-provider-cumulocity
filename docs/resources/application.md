---
page_title: "Resource: cumulocity_application"
description: |-
  Creates and manages a Cumulocity application (HOSTED, EXTERNAL, or MICROSERVICE).
---

# cumulocity_application

Creates and manages a Cumulocity application. Three application types are supported:

| Type | Description |
|------|-------------|
| `HOSTED` | Web application served by Cumulocity. Requires a ZIP upload via `cumulocity_application_binary`. |
| `EXTERNAL` | Link to an application hosted on an external server. |
| `MICROSERVICE` | Backend microservice containerized as a Docker image. Requires a ZIP with a microservice manifest. |

`type` is immutable — changing it forces a new resource.

Corresponds to `POST/GET/PUT/DELETE /application/applications/{id}`.

## Example Usage

### Hosted web application (full workflow)

```hcl
resource "cumulocity_application" "dashboard" {
  key          = "my-dashboard-key"
  name         = "My Dashboard"
  type         = "HOSTED"
  context_path = "my-dashboard"
  availability = "PRIVATE"
  description  = "Custom device dashboard"
}

resource "cumulocity_application_binary" "dashboard_zip" {
  application_id = cumulocity_application.dashboard.id
  file           = "${path.module}/dist/dashboard.zip"
  file_hash      = filemd5("${path.module}/dist/dashboard.zip")
}
```

### Microservice

```hcl
resource "cumulocity_application" "my_service" {
  key  = "my-service-key"
  name = "My Microservice"
  type = "MICROSERVICE"
}

resource "cumulocity_application_binary" "my_service_zip" {
  application_id = cumulocity_application.my_service.id
  file           = "${path.module}/build/my-service.zip"
  file_hash      = filemd5("${path.module}/build/my-service.zip")
}
```

### External application

```hcl
resource "cumulocity_application" "external_tool" {
  key          = "external-tool-key"
  name         = "External Monitoring Tool"
  type         = "EXTERNAL"
  context_path = "external-tool"
}
```

## Schema

### Required

- `key` (String) — Unique application key used as an identifier, e.g. `my-app-key`. Must be unique per tenant.
- `name` (String) — Display name of the application.
- `type` (String) — Application type: `HOSTED`, `EXTERNAL`, or `MICROSERVICE`. Immutable — changing this value forces a new resource.

### Optional

- `context_path` (String) — URL context path, e.g. `myapp`. Required by Cumulocity for `HOSTED` applications.
- `availability` (String) — Access level: `MARKET` (visible in tenant marketplace) or `PRIVATE` (owner only). Defaults to `PRIVATE`.
- `description` (String) — Human-readable description.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `self` (String) — Self-link URL of the application.
- `owner_tenant_id` (String) — ID of the tenant that owns this application.
- `active_version_id` (String) — ID of the currently active binary. Updated automatically when `cumulocity_application_binary` uploads a new ZIP.

## Import

Import an existing application by its Cumulocity ID.

```shell
terraform import cumulocity_application.dashboard 20200301
```
