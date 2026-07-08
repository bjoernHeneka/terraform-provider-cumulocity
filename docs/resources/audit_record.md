---
page_title: "cumulocity_audit_record Resource"
description: |-
  Creates a Cumulocity audit record.
---

# cumulocity_audit_record

Creates a Cumulocity audit record. Audit records log actions taken on platform components and are **immutable** — they cannot be updated or deleted after creation. Any change to an input attribute forces a new audit record.

Corresponds to `POST /audit/auditRecords`.

## Example Usage

```hcl
resource "cumulocity_audit_record" "operation_log" {
  source_id = cumulocity_managed_object.device.id
  type      = "Operation"
  activity  = "Operation created"
  text      = "Restart operation triggered via Terraform."
  time      = "2024-01-15T10:30:00.000Z"
  user      = "admin"
}
```

## Schema

### Required

- `source_id` (String) — The platform component or managed object ID associated with the audit. Changing this forces a new resource.
- `activity` (String) — Summary of the action, e.g. `Operation created`. Changing this forces a new resource.
- `text` (String) — Detailed description of the action. Changing this forces a new resource.
- `time` (String) — ISO 8601 date-time of the audit event, e.g. `2024-01-15T10:30:00.000Z`. Changing this forces a new resource.
- `type` (String) — Platform component type. Changing this forces a new resource. One of: `Alarm`, `Application`, `BulkOperation`, `CepModule`, `Connector`, `Event`, `Group`, `Inventory`, `InventoryRole`, `Operation`, `Option`, `Report`, `SingleSignOn`, `SmartRule`, `SYSTEM`, `Tenant`, `TenantAuthConfig`, `TrustedCertificates`, `User`, `UserAuthentication`.

### Optional

- `user` (String) — Username of the user who carried out the activity. Changing this forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `application` (String) — Application from which the audit was carried out (set by Cumulocity).
- `severity` (String) — Severity: `CRITICAL`, `MAJOR`, `MINOR`, `WARNING`, or `INFORMATION` (set by Cumulocity).
- `creation_time` (String) — ISO 8601 timestamp when the record was created.

## Note on Destroy

`terraform destroy` does **not** delete the audit record from Cumulocity. Audit records are permanent and retained for compliance purposes.

## Import

Import by audit record ID:

```shell
terraform import cumulocity_audit_record.example 20200301
```
