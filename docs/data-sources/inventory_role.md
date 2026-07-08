---
page_title: "Data Source: cumulocity_inventory_role"
description: |-
  Looks up a single Cumulocity inventory role by name or numeric ID.
---

# cumulocity_inventory_role

Looks up a single Cumulocity inventory role, including its full permission details. Provide either `name` or `id` — at least one is required.

Inventory roles control per-object permissions for managed objects (devices, groups). Their names are used as input to [`cumulocity_user_inventory_role_assignment`](../resources/user_inventory_role_assignment.md).

Corresponds to `GET /user/inventoryroles/{id}`.

## Example Usage

### Look up by name

```hcl
data "cumulocity_inventory_role" "restart" {
  name = "Operations: Restart Device"
}

output "restart_role_id" {
  value = data.cumulocity_inventory_role.restart.id
}

output "restart_role_permissions" {
  value = data.cumulocity_inventory_role.restart.permissions
}
```

### Look up by numeric ID

```hcl
data "cumulocity_inventory_role" "reader" {
  id = 4
}
```

### Use the name as input to an assignment

```hcl
data "cumulocity_inventory_role" "restart" {
  name = "Operations: Restart Device"
}

resource "cumulocity_user_inventory_role_assignment" "ops" {
  user_id           = cumulocity_user.alice.username
  managed_object_id = cumulocity_managed_object.gateway.id

  role_names = [data.cumulocity_inventory_role.restart.name]
}
```

## Schema

### Optional

- `name` (String) — Exact name of the inventory role. Required if `id` is not set.
- `id` (Number) — Numeric ID of the inventory role. Required if `name` is not set.

### Read-Only

- `name` (String) — Name of the inventory role (populated when looked up by `id`).
- `id` (Number) — Numeric ID of the inventory role (populated when looked up by `name`).
- `description` (String) — Description of the inventory role.
- `self` (String) — Self-link URL of the inventory role.
- `permissions` (List of Object) — Permissions defined for this role. Each object has:
  - `id` (Number) — Unique identifier of the permission entry.
  - `permission` (String) — Permission level: `ADMIN`, `READ`, or `*` (all).
  - `scope` (String) — Resource scope: `ALARM`, `AUDIT`, `EVENT`, `MANAGED_OBJECT`, `MEASUREMENT`, `OPERATION`, or `*` (all).
  - `type` (String) — Fragment type filter, e.g. `c8y_Restart`. An empty string means all fragment types.
