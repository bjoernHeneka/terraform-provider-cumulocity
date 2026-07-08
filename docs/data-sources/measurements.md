---
page_title: "Data Source: cumulocity_measurements"
description: |-
  Lists Cumulocity measurements, with optional source, type, date range, and fragment series filters.
---

# cumulocity_measurements

Retrieves a list of measurements from the Cumulocity measurement API, optionally filtered by source device, type, date range, or fragment series. All pages are followed automatically.

Corresponds to `GET /measurement/measurements`.

## Example Usage

```hcl
# Temperature measurements for a device within a date range
data "cumulocity_measurements" "temperature" {
  source_id             = cumulocity_managed_object.device.id
  value_fragment_type   = "c8y_Temperature"
  value_fragment_series = "T"
  date_from             = "2026-01-01T00:00:00Z"
  date_to               = "2026-02-01T00:00:00Z"
}

output "measurement_count" {
  value = length(data.cumulocity_measurements.temperature.measurements)
}
```

## Schema

### Optional

- `source_id` (String) — Filter by the source managed object (device) ID.
- `type` (String) — Filter by measurement type, e.g. `c8y_TemperatureMeasurement`.
- `date_from` (String) — Start of date range (ISO 8601).
- `date_to` (String) — End of date range (ISO 8601).
- `value_fragment_type` (String) — Filter by the fragment type that contains the measurement value, e.g. `c8y_Steam`.
- `value_fragment_series` (String) — Filter by the series name within the fragment, e.g. `Temperature`.

### Read-Only

- `measurements` (List of Object) — List of matching measurements. Each object contains:
  - `id` (String) — Measurement ID.
  - `source_id` (String) — Source managed object ID.
  - `type` (String) — Measurement type.
  - `time` (String) — ISO 8601 time the measurement was taken.
  - `creation_time` (String) — ISO 8601 creation timestamp.
  - `self` (String) — Self-link URL.
