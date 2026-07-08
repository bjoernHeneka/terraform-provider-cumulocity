---
page_title: "Data Source: cumulocity_role"
description: |-
  Looks up a single Cumulocity global role by name.
---

# cumulocity_role

Looks up a single Cumulocity global role by its exact name. Global roles control tenant-wide permissions (e.g. `ROLE_ALARM_ADMIN`, `ROLE_DEVICE_CONTROL_ADMIN`).

Use this data source when you need the role's `id` or `self` URL as an input to other resources.

Corresponds to `GET /user/roles/{name}`.

## Example Usage

```hcl
data "cumulocity_role" "alarm_admin" {
  name = "ROLE_ALARM_ADMIN"
}

output "alarm_admin_self" {
  value = data.cumulocity_role.alarm_admin.self
}
```

## Schema

### Required

- `name` (String) — Exact name of the role, e.g. `ROLE_ALARM_ADMIN`.

### Read-Only

- `id` (String) — Role identifier. In Cumulocity, the role ID equals the role name.
- `self` (String) — Self-link URL of the role.
