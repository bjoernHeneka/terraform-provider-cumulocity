---
page_title: "Resource: cumulocity_user_role_assignment"
description: |-
  Assigns a global role to a Cumulocity user.
---

# cumulocity_user_role_assignment

Assigns a global role (e.g. `ROLE_ALARM_ADMIN`) to a Cumulocity user.

All attributes are immutable — any change forces the old assignment to be deleted and a new one created. To look up the available roles in your tenant, use the [`cumulocity_roles`](../data-sources/roles.md) data source.

Corresponds to `POST/DELETE /user/{tenantId}/users/{userId}/roles`.

## Example Usage

```hcl
resource "cumulocity_user_role_assignment" "alarm_admin" {
  user_id = cumulocity_user.alice.username
  role    = "ROLE_ALARM_ADMIN"
}

resource "cumulocity_user_role_assignment" "device_control" {
  user_id = cumulocity_user.alice.username
  role    = "ROLE_DEVICE_CONTROL_ADMIN"
}
```

### Using a data source to validate the role name

```hcl
data "cumulocity_role" "alarm_admin" {
  name = "ROLE_ALARM_ADMIN"
}

resource "cumulocity_user_role_assignment" "alarm_admin" {
  user_id = cumulocity_user.alice.username
  role    = data.cumulocity_role.alarm_admin.id
}
```

## Schema

### Required

- `user_id` (String) — The `username` of the user to assign the role to. Changing this value forces a new resource.
- `role` (String) — The role identifier, e.g. `ROLE_ALARM_ADMIN`. Changing this value forces a new resource.

### Optional

- `tenant_id` (String) — Cumulocity tenant ID. Defaults to the provider's `tenant_id`. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{userId}/{role}`.
- `self` (String) — Self-link URL of the role assignment.

## Import

Import an existing role assignment using `{tenantId}/{userId}/{roleId}`. If `tenantId` is omitted, the provider's configured `tenant_id` is used.

```shell
terraform import cumulocity_user_role_assignment.alarm_admin t0071234/alice/ROLE_ALARM_ADMIN
# or
terraform import cumulocity_user_role_assignment.alarm_admin alice/ROLE_ALARM_ADMIN
```
