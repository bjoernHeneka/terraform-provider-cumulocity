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
