---
page_title: "cumulocity_events Data Source"
description: |-
  Lists Cumulocity events with optional filters.
---

# cumulocity_events

Retrieves a list of events, optionally filtered by source device, type, and date range. All pages are followed automatically.

Corresponds to `GET /event/events`.

## Example Usage

```hcl
# All location events for a device in a time window
data "cumulocity_events" "locations" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_LocationUpdate"
  date_from = "2024-01-01T00:00:00.000Z"
  date_to   = "2024-01-31T23:59:59.999Z"
}

output "location_count" {
  value = length(data.cumulocity_events.locations.events)
}
```

## Schema

### Optional

- `source_id` (String) — Filter events by managed object (device/asset) ID.
- `type` (String) — Filter by event type, e.g. `c8y_LocationUpdate`.
- `date_from` (String) — Start of date range (ISO 8601). Filters by device timestamp.
- `date_to` (String) — End of date range (ISO 8601). Filters by device timestamp.

### Read-Only

- `events` (List of Object) — List of matching events. Each object contains:
  - `id` (String) — Event ID.
  - `source_id` (String) — Source managed object ID.
  - `type` (String) — Event type.
  - `text` (String) — Event description.
  - `time` (String) — ISO 8601 time the event occurred.
  - `creation_time` (String) — ISO 8601 creation timestamp.
  - `last_updated` (String) — ISO 8601 last update timestamp.
