---
page_title: "Data Source: cumulocity_application"
description: |-
  Looks up a single Cumulocity application by name.
---

# cumulocity_application

Looks up a single Cumulocity application by its exact name. Use this data source to reference an existing application (for example, to feed its `id` into a `cumulocity_tenant_application_subscription` resource) without hard-coding the ID.

The lookup fails if no application with the given name exists, or if more than one application shares the name — application names must be unique for a data source lookup to resolve.

Corresponds to `GET /application/applicationsByName/{name}`.

## Example Usage

```hcl
data "cumulocity_application" "myapp" {
  name = "My Dashboard"
}

output "application_id" {
  value = data.cumulocity_application.myapp.id
}
```

## Schema

### Required

- `name` (String) — The application name to look up.

### Read-Only

- `id` (String) — The application ID.
- `key` (String) — The application key.
- `type` (String) — Application type: `HOSTED`, `EXTERNAL`, or `MICROSERVICE`.
- `context_path` (String) — The application context path (for `HOSTED` applications).
- `availability` (String) — Application availability: `MARKET` or `PRIVATE`.
- `description` (String) — Application description.
- `active_version_id` (String) — ID of the active binary version (set after upload).
- `owner_tenant_id` (String) — ID of the tenant that owns this application.
- `self` (String) — The self-link URL of the application.
