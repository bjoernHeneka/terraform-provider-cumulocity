---
page_title: "Resource: cumulocity_user_inventory_role_assignment"
description: |-
  Assigns one or more Cumulocity inventory roles to a user for a specific managed object.
---

# cumulocity_user_inventory_role_assignment

Assigns one or more Cumulocity inventory roles to a user for a specific managed object (device or group). Inventory roles grant fine-grained, per-object permissions, in contrast to global roles which apply tenant-wide.

- Changing `role_names` issues an in-place `PUT` update — no resource recreation needed.
- Changing `user_id`, `managed_object_id`, or `tenant_id` forces a new resource.

Corresponds to `POST/GET/PUT/DELETE /user/{tenantId}/users/{userId}/roles/inventory/{id}`.

## Example Usage

### Assign inventory roles to a user for a device

```hcl
resource "cumulocity_managed_object" "sensor" {
  name      = "temperature-sensor-01"
  type      = "c8y_TemperatureSensor"
  is_device = true
}

resource "cumulocity_user_inventory_role_assignment" "sensor_ops" {
  user_id           = cumulocity_user.alice.username
  managed_object_id = cumulocity_managed_object.sensor.id

  role_names = [
    "Operations: Restart Device",
    "Reader",
  ]
}
```

### Using an inventory role data source

```hcl
data "cumulocity_inventory_role" "restart" {
  name = "Operations: Restart Device"
}

resource "cumulocity_user_inventory_role_assignment" "ops" {
  user_id           = cumulocity_user.alice.username
  managed_object_id = cumulocity_managed_object.sensor.id

  role_names = [
    data.cumulocity_inventory_role.restart.name,
  ]
}
```

## Schema

### Required

- `user_id` (String) — The `username` of the user. Changing this value forces a new resource.
- `managed_object_id` (String) — ID of the managed object (device or group) for which the roles apply. Changing this value forces a new resource.
- `role_names` (List of String) — List of inventory role names to assign, e.g. `["Operations: Restart Device", "Reader"]`. Changing this list issues an in-place update.

### Optional

- `tenant_id` (String) — Cumulocity tenant ID. Defaults to the provider's `tenant_id`. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{userId}/{assignmentId}`.
- `assignment_id` (Number) — Numeric ID of the inventory assignment returned by the API.
- `self` (String) — Self-link URL of the inventory assignment.

## Import

Import an existing inventory role assignment using `{tenantId}/{userId}/{assignmentId}` or `{userId}/{assignmentId}`. The `assignmentId` is the numeric ID returned by the Cumulocity API.

```shell
terraform import cumulocity_user_inventory_role_assignment.ops t0071234/alice/42
# or
terraform import cumulocity_user_inventory_role_assignment.ops alice/42
```

~> **Tip:** After importing, run `terraform plan` to verify that the `role_names` in state match your configuration. The API returns role names in the response, so drift detection works correctly.
