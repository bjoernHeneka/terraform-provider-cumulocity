---
page_title: "Resource: cumulocity_device_operation"
description: |-
  Sends an operation to a Cumulocity device.
---

# cumulocity_device_operation

Sends an operation command to a Cumulocity device. The operation is created in `PENDING` status and transitions through `EXECUTING` → `SUCCESSFUL` or `FAILED` as the device processes it. Terraform tracks the current status on every `plan`/`apply`.

All content attributes (`device_id`, `description`, `fragments_json`) are immutable — any change destroys the old operation and sends a new one.

**On destroy:** PENDING operations are cancelled (status set to `FAILED`). Operations in any other status are removed from Terraform state only — they remain in Cumulocity history.

Corresponds to `POST/GET /devicecontrol/operations/{id}`.

## Example Usage

### Restart a device

```hcl
resource "cumulocity_device_operation" "restart" {
  device_id      = cumulocity_managed_object.gateway.id
  description    = "Restart gateway via Terraform"
  fragments_json = jsonencode({
    c8y_Restart = {}
  })
}
```

### Execute a shell command

```hcl
resource "cumulocity_device_operation" "diagnostics" {
  device_id      = cumulocity_managed_object.gateway.id
  description    = "Collect diagnostics"
  fragments_json = jsonencode({
    c8y_Command = { text = "journalctl -u tedge --since '1 hour ago'" }
  })
}
```

### Update firmware

```hcl
resource "cumulocity_device_operation" "firmware_update" {
  device_id   = cumulocity_managed_object.gateway.id
  description = "Update firmware to 2.1.0"
  fragments_json = jsonencode({
    c8y_Firmware = {
      name    = "my-firmware"
      version = "2.1.0"
      url     = "https://firmware.example.com/v2.1.0.bin"
    }
  })
}
```

## Schema

### Required

- `device_id` (String) — ID of the target device. Changing this value forces a new resource.
- `fragments_json` (String) — JSON object with the operation payload. The object is merged with `deviceId` in the API request. Changing this value forces a new resource.

### Optional

- `description` (String) — Human-readable description. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `status` (String) — Current operation status: `PENDING`, `EXECUTING`, `SUCCESSFUL`, or `FAILED`.
- `failure_reason` (String) — Reason for failure, populated when `status = FAILED`.
- `creation_time` (String) — ISO 8601 timestamp when the operation was created.
- `bulk_operation_id` (Number) — ID of the parent bulk operation, if applicable.
- `self` (String) — Self-link URL.

## Import

Import by operation ID.

```shell
terraform import cumulocity_device_operation.restart 123456
```

~> **Note:** After import, `fragments_json` and `description` will be empty in state. Set them explicitly in your configuration to re-enable content tracking.
