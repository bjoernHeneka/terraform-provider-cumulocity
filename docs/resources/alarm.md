---
page_title: "cumulocity_alarm Resource"
description: |-
  Manages a Cumulocity alarm.
---

# cumulocity_alarm

Manages a Cumulocity alarm. Alarms indicate an abnormal condition on a device or asset that requires attention.

Cumulocity deduplicates alarms by `(source, type)` — creating a second alarm with the same source and type increments the `count` of the existing alarm rather than creating a new one.

Corresponds to `POST/GET/PUT/DELETE /alarm/alarms/{id}`.

## Example Usage

```hcl
resource "cumulocity_alarm" "unavailable" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_UnavailabilityAlarm"
  text      = "No data received from the device within the required interval."
  severity  = "MAJOR"
  time      = "2024-01-15T10:30:00.000Z"
}

# Acknowledge an alarm
resource "cumulocity_alarm" "critical" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_TemperatureAlarm"
  text      = "Device temperature exceeded threshold."
  severity  = "CRITICAL"
  status    = "ACKNOWLEDGED"
  time      = "2024-01-15T08:00:00.000Z"
}
```

## Schema

### Required

- `source_id` (String) — The managed object (device/asset) ID to associate the alarm with. Changing this forces a new resource.
- `type` (String) — Alarm type identifier, e.g. `c8y_UnavailabilityAlarm`. Changing this forces a new resource.
- `text` (String) — Human-readable description of the alarm.
- `severity` (String) — Severity of the alarm. One of: `CRITICAL`, `MAJOR`, `MINOR`, `WARNING`.
- `time` (String) — ISO 8601 date-time of when the alarm occurred, e.g. `2024-01-15T10:30:00.000Z`. Changing this forces a new resource.

### Optional

- `status` (String) — Status of the alarm. One of: `ACTIVE`, `ACKNOWLEDGED`, `CLEARED`. Defaults to `ACTIVE` on creation.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `occurrence_count` (Number) — Number of times this alarm has been triggered (incremented on deduplication).
- `creation_time` (String) — ISO 8601 timestamp when the alarm was created.
- `last_updated` (String) — ISO 8601 timestamp of the last update.
- `first_occurrence_time` (String) — ISO 8601 timestamp of the first occurrence (only set when `count` > 1).

## Import

Import by alarm ID:

```shell
terraform import cumulocity_alarm.example 20200301
```
