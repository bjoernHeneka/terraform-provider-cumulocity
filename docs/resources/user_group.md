---
page_title: "Resource: cumulocity_user_group"
description: |-
  Creates and manages a Cumulocity user group within a tenant.
---

# cumulocity_user_group

Creates and manages a Cumulocity user group within a tenant. User groups bundle roles and aggregate users, providing a convenient way to apply consistent access control to multiple users at once.

Use [`cumulocity_user_group_membership`](user_group_membership.md) to add users to the group and [`cumulocity_user_role_assignment`](user_role_assignment.md) to assign global roles to the group's members.

Corresponds to `POST/GET/PUT/DELETE /user/{tenantId}/groups/{groupId}`.

## Example Usage

```hcl
resource "cumulocity_user_group" "operators" {
  name        = "operators"
  description = "Operator access group — managed by Terraform"
}

resource "cumulocity_user_group_membership" "alice_operators" {
  group_id = cumulocity_user_group.operators.group_id
  user_id  = cumulocity_user.alice.username
}
```

### Multiple users in one group

```hcl
resource "cumulocity_user_group" "readonly" {
  name        = "readonly-viewers"
  description = "Read-only access to all devices"
}

resource "cumulocity_user_group_membership" "alice_readonly" {
  group_id = cumulocity_user_group.readonly.group_id
  user_id  = cumulocity_user.alice.username
}

resource "cumulocity_user_group_membership" "bob_readonly" {
  group_id = cumulocity_user_group.readonly.group_id
  user_id  = cumulocity_user.bob.username
}
```

## Schema

### Required

- `name` (String) — Unique name of the group within the tenant. Duplicate names are rejected by the API with HTTP 409.

### Optional

- `tenant_id` (String) — Cumulocity tenant ID. Defaults to the provider's `tenant_id`. Changing this value forces a new resource.
- `description` (String) — Free-text description of the group.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{groupId}`.
- `group_id` (Number) — Numeric group ID assigned by Cumulocity. Pass this to `cumulocity_user_group_membership.group_id`.
- `self` (String) — Self-link URL of the group.

## Import

Import an existing group using `{tenantId}/{groupId}` or just `{groupId}` (using the provider's default tenant).

```shell
terraform import cumulocity_user_group.operators t0071234/42
# or
terraform import cumulocity_user_group.operators 42
```
