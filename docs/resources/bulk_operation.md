---
page_title: "Resource: cumulocity_bulk_operation"
description: |-
  Creates and manages a Cumulocity bulk operation targeting a device group.
---

# cumulocity_bulk_operation

Creates and manages a Cumulocity bulk operation. A bulk operation sends the same command to every device in a group, staggered by `creation_ramp` seconds to avoid overwhelming the platform.

`group_id` and `failed_parent_id` are mutually exclusive. `start_date`, `creation_ramp`, and `operation_prototype_json` can be updated in-place.

Corresponds to `POST/GET/PUT/DELETE /devicecontrol/bulkoperations/{id}`.

## Example Usage

### Restart all devices in a group

```hcl
resource "cumulocity_bulk_operation" "group_restart" {
  group_id                 = cumulocity_managed_object.factory_floor.id
  start_date               = "2025-06-01T06:00:00Z"
  creation_ramp            = 10.0
  operation_prototype_json = jsonencode({ c8y_Restart = {} })
}
```

### Reschedule failed operations from a previous bulk operation

```hcl
resource "cumulocity_bulk_operation" "retry_failed" {
  failed_parent_id         = cumulocity_bulk_operation.group_restart.id
  start_date               = "2025-06-02T06:00:00Z"
  creation_ramp            = 5.0
  operation_prototype_json = jsonencode({ c8y_Restart = {} })
}
```

### Firmware update for a fleet

```hcl
resource "cumulocity_bulk_operation" "firmware_fleet" {
  group_id      = cumulocity_managed_object.fleet_group.id
  start_date    = "2025-07-01T02:00:00Z"
  creation_ramp = 30.0
  operation_prototype_json = jsonencode({
    c8y_Firmware = {
      name    = "device-firmware"
      version = "3.0.0"
      url     = "https://firmware.example.com/3.0.0.bin"
    }
  })
}

output "firmware_update_status" {
  value = cumulocity_bulk_operation.firmware_fleet.general_status
}
```

## Schema

### Required

- `start_date` (String) — ISO 8601 datetime when individual operations should start being created, e.g. `"2025-06-01T08:00:00Z"`.
- `creation_ramp` (Number) — Delay in seconds between creation of consecutive individual operations.
- `operation_prototype_json` (String) — JSON object representing the operation to send to each device.

### Optional

- `group_id` (String) — ID of the device group to target. Mutually exclusive with `failed_parent_id`. Changing this value forces a new resource.
- `failed_parent_id` (String) — ID of a previous bulk operation; reschedules only its failed operations. Mutually exclusive with `group_id`. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `status` (String) — Internal execution status: `ACTIVE`, `IN_PROGRESS`, `COMPLETED`, or `DELETED`.
- `general_status` (String) — End-user visible status: `SCHEDULED`, `EXECUTING`, `EXECUTING_WITH_ERRORS`, `SUCCESSFUL`, `FAILED`, or `CANCELED`.
- `self` (String) — Self-link URL.
- `progress_pending` (Number) — Number of pending individual operations.
- `progress_failed` (Number) — Number of failed individual operations.
- `progress_executing` (Number) — Number of currently executing operations.
- `progress_successful` (Number) — Number of successfully completed operations.
- `progress_all` (Number) — Total number of individual operations.

## Import

Import by bulk operation ID.

```shell
terraform import cumulocity_bulk_operation.group_restart 9876
```
