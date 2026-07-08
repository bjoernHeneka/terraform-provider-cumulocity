---
page_title: "Data Source: cumulocity_inventory_roles"
description: |-
  Lists all Cumulocity inventory roles, with an optional case-insensitive name filter.
---

# cumulocity_inventory_roles

Lists all Cumulocity inventory roles with their full permission details. An optional `name_filter` narrows the results to roles whose name contains the given string (case-insensitive).

Corresponds to `GET /user/inventoryroles` (follows pagination automatically).

## Example Usage

### List all inventory roles

```hcl
data "cumulocity_inventory_roles" "all" {}

output "all_inventory_role_names" {
  value = [for r in data.cumulocity_inventory_roles.all.roles : r.name]
}
```

### Filter to operations roles only

```hcl
data "cumulocity_inventory_roles" "operations" {
  name_filter = "Operations"
}

output "operations_role_names" {
  value = [for r in data.cumulocity_inventory_roles.operations.roles : r.name]
}
```

## Schema

### Optional

- `name_filter` (String) — When set, only roles whose name contains this string (case-insensitive) are returned.

### Read-Only

- `roles` (List of Object) — The list of matching inventory roles. Each object has:
  - `id` (Number) — Numeric ID of the inventory role.
  - `name` (String) — Name of the inventory role.
  - `description` (String) — Description of the inventory role.
  - `self` (String) — Self-link URL of the inventory role.
  - `permissions` (List of Object) — Permissions defined for this role. Each object has:
    - `id` (Number) — Unique identifier of the permission entry.
    - `permission` (String) — Permission level: `ADMIN`, `READ`, or `*` (all).
    - `scope` (String) — Resource scope: `ALARM`, `AUDIT`, `EVENT`, `MANAGED_OBJECT`, `MEASUREMENT`, `OPERATION`, or `*` (all).
    - `type` (String) — Fragment type filter, e.g. `c8y_Restart`. An empty string means all fragment types.
