---
page_title: "cumulocity_alarms Data Source"
description: |-
  Lists Cumulocity alarms with optional filters.
---

# cumulocity_alarms

Retrieves a list of alarms, optionally filtered by source device, status, severity, and/or type. All pages are followed automatically.

Corresponds to `GET /alarm/alarms`.

## Example Usage

```hcl
# All active alarms for a device
data "cumulocity_alarms" "active" {
  source_id = cumulocity_managed_object.device.id
  status    = "ACTIVE"
}

# All critical alarms across the tenant
data "cumulocity_alarms" "critical" {
  severity = "CRITICAL"
}

output "alarm_count" {
  value = length(data.cumulocity_alarms.active.alarms)
}
```

## Schema

### Optional

- `source_id` (String) — Filter alarms by managed object (device/asset) ID.
- `status` (String) — Filter by alarm status: `ACTIVE`, `ACKNOWLEDGED`, or `CLEARED`.
- `severity` (String) — Filter by severity: `CRITICAL`, `MAJOR`, `MINOR`, or `WARNING`.
- `type` (String) — Filter by alarm type, e.g. `c8y_UnavailabilityAlarm`.

### Read-Only

- `alarms` (List of Object) — List of matching alarms. Each object contains:
  - `id` (String) — Alarm ID.
  - `source_id` (String) — Source managed object ID.
  - `type` (String) — Alarm type.
  - `text` (String) — Alarm description.
  - `severity` (String) — Severity level.
  - `status` (String) — Current status.
  - `time` (String) — ISO 8601 time the alarm occurred.
  - `occurrence_count` (Number) — Number of times this alarm was triggered.
  - `creation_time` (String) — ISO 8601 creation timestamp.
  - `last_updated` (String) — ISO 8601 last update timestamp.
  - `first_occurrence_time` (String) — ISO 8601 first occurrence timestamp (set when `count` > 1).
