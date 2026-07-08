---
page_title: "Data Source: cumulocity_roles"
description: |-
  Lists all Cumulocity global roles, with an optional case-insensitive name filter.
---

# cumulocity_roles

Lists all global roles available in the Cumulocity tenant. An optional `name_filter` narrows the results to roles whose name contains the given string (case-insensitive).

Corresponds to `GET /user/roles` (follows pagination automatically).

## Example Usage

### List all roles

```hcl
data "cumulocity_roles" "all" {}

output "all_role_names" {
  value = [for r in data.cumulocity_roles.all.roles : r.name]
}
```

### Filter by name substring

```hcl
data "cumulocity_roles" "alarm_roles" {
  name_filter = "ALARM"
}

output "alarm_role_names" {
  value = [for r in data.cumulocity_roles.alarm_roles.roles : r.name]
}
```

## Schema

### Optional

- `name_filter` (String) — When set, only roles whose name contains this string (case-insensitive) are returned.

### Read-Only

- `roles` (List of Object) — The list of matching roles. Each object has:
  - `id` (String) — Role identifier.
  - `name` (String) — Role name.
  - `self` (String) — Self-link URL of the role.
