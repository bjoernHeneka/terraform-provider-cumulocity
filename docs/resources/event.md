---
page_title: "cumulocity_event Resource"
description: |-
  Manages a Cumulocity event.
---

# cumulocity_event

Manages a Cumulocity event. Events represent time-stamped occurrences on a device or asset, such as location updates or threshold crossings.

Only the `text` description can be updated in-place. Changing `source_id`, `type`, or `time` forces a new resource.

Corresponds to `POST/GET/PUT/DELETE /event/events/{id}`.

## Example Usage

```hcl
resource "cumulocity_event" "location" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_LocationUpdate"
  text      = "Device reported GPS coordinates."
  time      = "2024-01-15T10:30:00.000Z"
}

resource "cumulocity_event" "startup" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_DeviceStartup"
  text      = "Device booted after firmware update."
  time      = "2024-01-15T08:00:00.000Z"
}
```

## Schema

### Required

- `source_id` (String) — The managed object (device/asset) ID to associate the event with. Changing this forces a new resource.
- `type` (String) — Event type identifier, e.g. `c8y_LocationUpdate`. Changing this forces a new resource.
- `text` (String) — Human-readable description of the event. Can be updated in-place.
- `time` (String) — ISO 8601 date-time of when the event occurred, e.g. `2024-01-15T10:30:00.000Z`. Changing this forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `creation_time` (String) — ISO 8601 timestamp when the event was created.
- `last_updated` (String) — ISO 8601 timestamp of the last update.

## Import

Import by event ID:

```shell
terraform import cumulocity_event.example 20200301
```
