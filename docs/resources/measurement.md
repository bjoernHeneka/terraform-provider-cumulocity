---
page_title: "Resource: cumulocity_measurement"
description: |-
  Creates a Cumulocity measurement.
---

# cumulocity_measurement

Creates a Cumulocity measurement. Measurements are immutable — all attributes trigger a resource replacement when changed.

Corresponds to `POST/GET/DELETE /measurement/measurements/{id}`.

## Example Usage

```hcl
# Create a temperature measurement on a device.
resource "cumulocity_measurement" "temperature" {
  source_id = "12345"
  type      = "c8y_TemperatureMeasurement"
  time      = "2024-01-15T10:30:00.000Z"

  fragments = jsonencode({
    c8y_Temperature = {
      T = {
        value = 22.5
        unit  = "°C"
      }
    }
  })
}

output "measurement_id" {
  value = cumulocity_measurement.temperature.id
}
```

## Schema

### Required

- `source_id` (String) — The managed object (device/asset) ID to which the measurement belongs. Changing this forces a new resource.
- `type` (String) — Measurement type, e.g. `c8y_TemperatureMeasurement`. Changing this forces a new resource.
- `time` (String) — ISO 8601 date-time of when the measurement was taken, e.g. `2024-01-15T10:30:00.000Z`. Changing this forces a new resource.

### Optional

- `fragments` (String) — JSON object containing the measurement fragment data, e.g. `{"c8y_Temperature":{"T":{"value":22.5,"unit":"°C"}}}`. Changing this forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `creation_time` (String) — ISO 8601 timestamp when the measurement was created in Cumulocity.
- `self` (String) — Self-link URL of the measurement.

## Import

Import an existing measurement by its ID:

```shell
terraform import cumulocity_measurement.temperature 20200301
```
