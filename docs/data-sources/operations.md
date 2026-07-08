---
page_title: "Data Source: cumulocity_operations"
description: |-
  Retrieves a list of Cumulocity device operations, with optional filters.
---

# cumulocity_operations

Retrieves device operations from the Cumulocity platform, following all pages automatically. Supports optional filtering by device ID and/or status.

Corresponds to `GET /devicecontrol/operations`.

## Example Usage

### All operations for a device

```hcl
data "cumulocity_operations" "gateway_ops" {
  device_id = cumulocity_managed_object.gateway.id
}

output "all_operations" {
  value = data.cumulocity_operations.gateway_ops.operations
}
```

### Pending operations only

```hcl
data "cumulocity_operations" "pending" {
  device_id = cumulocity_managed_object.gateway.id
  status    = "PENDING"
}

output "pending_count" {
  value = length(data.cumulocity_operations.pending.operations)
}
```

### All failed operations across all devices

```hcl
data "cumulocity_operations" "all_failed" {
  status = "FAILED"
}

output "failed_operations" {
  value = [for op in data.cumulocity_operations.all_failed.operations : {
    id        = op.id
    device_id = op.device_id
    reason    = op.failure_reason
  }]
}
```

## Schema

### Optional

- `device_id` (String) — Filter by target device ID.
- `status` (String) — Filter by status: `PENDING`, `EXECUTING`, `SUCCESSFUL`, or `FAILED`.

### Read-Only

- `operations` (List of Object) — Matching operations. Each object has:
  - `id` (String) — Operation ID.
  - `device_id` (String) — Target device ID.
  - `status` (String) — Current status.
  - `failure_reason` (String) — Failure reason, if any.
  - `creation_time` (String) — ISO 8601 creation timestamp.
  - `bulk_operation_id` (Number) — Parent bulk operation ID, if applicable.
  - `self` (String) — Self-link URL.
