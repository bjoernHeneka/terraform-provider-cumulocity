---
page_title: "Resource: cumulocity_managed_object"
description: |-
  Creates and manages a Cumulocity managed object (device, group, or generic asset) in the inventory.
---

# cumulocity_managed_object

Creates and manages a Cumulocity managed object in the inventory. Managed objects represent physical or virtual entities such as devices, device groups, or generic assets.

The `id` of a managed object is the value to pass as `managed_object_id` in [`cumulocity_user_inventory_role_assignment`](user_inventory_role_assignment.md).

Corresponds to `POST/GET/PUT/DELETE /inventory/managedObjects/{id}`.

## Example Usage

### Device

```hcl
resource "cumulocity_managed_object" "gateway" {
  name      = "factory-gateway-01"
  type      = "c8y_Linux"
  is_device = true
}
```

### Device group

```hcl
resource "cumulocity_managed_object" "factory_floor" {
  name            = "Factory Floor"
  is_device_group = true
}
```

### Full example with inventory role assignment

```hcl
resource "cumulocity_managed_object" "sensor" {
  name      = "temperature-sensor-01"
  type      = "c8y_TemperatureSensor"
  is_device = true
}

resource "cumulocity_user_inventory_role_assignment" "sensor_access" {
  user_id           = cumulocity_user.alice.username
  managed_object_id = cumulocity_managed_object.sensor.id

  role_names = ["Reader"]
}
```

## Schema

### Required

- `name` (String) — Display name of the managed object.

### Optional

- `type` (String) — Device class type. Devices sharing the same type can receive the same configuration, software, and operations.
- `is_device` (Boolean) — When `true`, adds the `c8y_IsDevice` fragment, marking this object as a device in the Cumulocity device list. Defaults to `false`.
- `is_device_group` (Boolean) — When `true`, adds the `c8y_IsDeviceGroup` fragment, marking this object as a device group. Defaults to `false`.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity. Use this as `managed_object_id` in `cumulocity_user_inventory_role_assignment`.
- `owner` (String) — Username of the managed object's owner, set from the authenticated user at creation.
- `self` (String) — Self-link URL of the managed object.
- `creation_time` (String) — ISO 8601 timestamp when the object was created.
- `last_updated` (String) — ISO 8601 timestamp when the object was last modified.

## Import

Import an existing managed object by its numeric Cumulocity ID.

```shell
terraform import cumulocity_managed_object.gateway 51994
```
