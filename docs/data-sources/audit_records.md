---
page_title: "cumulocity_audit_records Data Source"
description: |-
  Lists Cumulocity audit records with optional filters.
---

# cumulocity_audit_records

Retrieves a list of audit records, optionally filtered by source, type, user, and/or application. All pages are followed automatically.

Corresponds to `GET /audit/auditRecords`.

## Example Usage

```hcl
# All audit records for a device
data "cumulocity_audit_records" "device_logs" {
  source_id = cumulocity_managed_object.device.id
}

# All operation-related audit records by a specific user
data "cumulocity_audit_records" "ops" {
  type = "Operation"
  user = "admin"
}

output "audit_count" {
  value = length(data.cumulocity_audit_records.device_logs.audit_records)
}
```

## Schema

### Optional

- `source_id` (String) — Filter by the platform component or managed object ID.
- `type` (String) — Filter by audit record type, e.g. `Operation`, `User`, `Alarm`.
- `user` (String) — Filter by the username who carried out the activity.
- `application` (String) — Filter by the application name from which the audit was carried out.

### Read-Only

- `audit_records` (List of Object) — List of matching audit records. Each object contains:
  - `id` (String) — Audit record ID.
  - `source_id` (String) — Source platform component or managed object ID.
  - `activity` (String) — Summary of the action.
  - `text` (String) — Detailed description.
  - `time` (String) — ISO 8601 time of the audit event.
  - `type` (String) — Platform component type.
  - `user` (String) — User who performed the action.
  - `application` (String) — Application that performed the action.
  - `severity` (String) — Severity level.
  - `creation_time` (String) — ISO 8601 creation timestamp.
