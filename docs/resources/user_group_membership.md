---
page_title: "Resource: cumulocity_user_group_membership"
description: |-
  Assigns a user to a Cumulocity user group.
---

# cumulocity_user_group_membership

Assigns a user to a Cumulocity user group. All attributes are immutable — any change forces the membership to be removed and recreated.

Corresponds to `POST /user/{tenantId}/groups/{groupId}/users` and `DELETE /user/{tenantId}/groups/{groupId}/users/{userId}`.

## Example Usage

```hcl
resource "cumulocity_user_group" "operators" {
  name = "operators"
}

resource "cumulocity_user_group_membership" "alice" {
  group_id = cumulocity_user_group.operators.group_id
  user_id  = cumulocity_user.alice.username
}
```

### Using a for_each to add multiple users

```hcl
locals {
  operator_users = ["alice", "bob", "carol"]
}

resource "cumulocity_user_group_membership" "operators" {
  for_each = toset(local.operator_users)

  group_id = cumulocity_user_group.operators.group_id
  user_id  = each.value
}
```

## Schema

### Required

- `group_id` (Number) — Numeric ID of the target group. Use `cumulocity_user_group.my_group.group_id`. Changing this value forces a new resource.
- `user_id` (String) — Username of the user to add to the group. Changing this value forces a new resource.

### Optional

- `tenant_id` (String) — Cumulocity tenant ID. Defaults to the provider's `tenant_id`. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{groupId}/{userId}`.

## Import

Import an existing membership using `{tenantId}/{groupId}/{userId}` or `{groupId}/{userId}` (using the provider's default tenant).

```shell
terraform import cumulocity_user_group_membership.alice t0071234/42/alice
# or
terraform import cumulocity_user_group_membership.alice 42/alice
```
