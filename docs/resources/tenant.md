---
page_title: "Resource: cumulocity_tenant"
description: |-
  Creates and manages a Cumulocity subtenant.
---

# cumulocity_tenant

Creates and manages a Cumulocity subtenant.

Requires `ROLE_TENANT_MANAGEMENT_ADMIN` or `ROLE_TENANT_MANAGEMENT_CREATE`.

Corresponds to `POST/GET/PUT/DELETE /tenant/tenants/{tenantId}`.

## Example Usage

```hcl
resource "cumulocity_tenant" "example" {
  company     = "ACME AG"
  domain      = "acme.cumulocity.com"
  admin_email = "admin@acme.com"
  admin_name  = "acmeadmin"
  admin_pass  = "S3cur3P@ss!"

  contact_name  = "John Doe"
  contact_phone = "+49 123 456 7890"
}

output "tenant_id" {
  value = cumulocity_tenant.example.id
}

output "tenant_status" {
  value = cumulocity_tenant.example.status
}
```

## Schema

### Required

- `company` (String) — The tenant's company name.
- `domain` (String) — The tenant's domain (e.g. `mytenant.cumulocity.com`). Changing this forces a new resource.
- `admin_email` (String) — Email address of the tenant administrator.

### Optional

- `admin_name` (String) — Username of the tenant administrator. Changing this forces a new resource.
- `admin_pass` (String, Sensitive) — Password of the tenant administrator. Write-only — not returned by the API.
- `contact_name` (String) — Name of the contact person.
- `contact_phone` (String) — Phone number of the contact person in international format.

### Read-Only

- `id` (String) — The tenant ID assigned by Cumulocity (e.g. `t0071234`).
- `parent` (String) — ID of the parent tenant.
- `status` (String) — Current status of the tenant: `ACTIVE` or `SUSPENDED`.
- `creation_time` (String) — Date and time when the tenant was created (RFC 3339).
- `allow_create_tenants` (Boolean) — Whether this tenant is allowed to create subtenants.
- `self` (String) — Self-link URL of the tenant.

## Import

Import an existing tenant by its tenant ID:

```shell
terraform import cumulocity_tenant.example t0071234
```

~> **Note on admin_pass:** Because `admin_pass` is write-only, it is not populated after import. Terraform will not detect drift on this field.
