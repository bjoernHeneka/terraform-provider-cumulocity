---
page_title: "Resource: cumulocity_external_id"
description: |-
  Links a managed object to an identifier in an external system.
---

# cumulocity_external_id

Creates and manages a Cumulocity external ID, linking a managed object to an identifier in an external system (e.g. a device serial number). External IDs are immutable — any change forces replacement.

Corresponds to `POST /identity/globalIds/{id}/externalIds` (create) and `GET/DELETE /identity/externalIds/{type}/{externalId}` (read/delete).

## Example Usage

```hcl
resource "cumulocity_managed_object" "device" {
  name = "My Device"
  type = "c8y_Device"
}

resource "cumulocity_external_id" "serial" {
  managed_object_id = cumulocity_managed_object.device.id
  type              = "c8y_Serial"
  external_id       = "SN-00123456"
}

output "external_id_self" {
  value = cumulocity_external_id.serial.self
}
```

## Schema

### Required

- `external_id` (String) — The identifier value in the external system, e.g. `SN-12345`. Changing this forces a new resource.
- `type` (String) — The type of the external identifier, e.g. `c8y_Serial`. Changing this forces a new resource.
- `managed_object_id` (String) — The ID of the managed object this external ID is linked to. Changing this forces a new resource.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{type}/{external_id}`.
- `managed_object_self` (String) — Self-link URL of the linked managed object.
- `self` (String) — Self-link URL of this external ID.

## Import

Import an existing external ID using `{type}/{external_id}`:

```shell
terraform import cumulocity_external_id.serial c8y_Serial/SN-00123456
```
