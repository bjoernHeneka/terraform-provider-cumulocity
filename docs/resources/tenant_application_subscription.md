---
page_title: "Resource: cumulocity_tenant_application_subscription"
description: |-
  Subscribes a tenant to an application.
---

# cumulocity_tenant_application_subscription

Subscribes a tenant to an application. This resource has no updatable fields; changing either attribute forces a new resource.

Requires `ROLE_APPLICATION_MANAGEMENT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`/`UPDATE`.

Corresponds to `POST/DELETE /tenant/tenants/{tenantId}/applications`.

## Example Usage

### Using direct IDs

```hcl
# Subscribe a tenant to an application using direct IDs.
resource "cumulocity_tenant_application_subscription" "example" {
  tenant_id      = "t0071234"
  application_id = "12345"
}
```

### Using references from other resources

```hcl
resource "cumulocity_tenant" "subtenant" {
  company     = "Example Corp"
  domain      = "example.cumulocity.com"
  admin_email = "admin@example.com"
}

resource "cumulocity_application" "myapp" {
  name         = "my-application"
  type         = "HOSTED"
  key          = "my-app-key"
  availability = "PRIVATE"
}

resource "cumulocity_tenant_application_subscription" "sub" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = cumulocity_application.myapp.id
}
```

## Schema

### Required

- `tenant_id` (String) — The tenant ID to subscribe (e.g. `t0071234`). Changing this forces a new resource.
- `application_id` (String) — The application ID to subscribe to. Changing this forces a new resource.

### Read-Only

- `id` (String) — Composite ID in the format `{tenantId}/{applicationId}`.

## Import

Import an existing subscription using `{tenantId}/{applicationId}`:

```shell
terraform import cumulocity_tenant_application_subscription.example t0071234/12345
```
