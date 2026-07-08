# Location update events for a device within a date range
data "cumulocity_events" "locations" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_LocationUpdate"
  date_from = "2026-01-01T00:00:00Z"
  date_to   = "2026-02-01T00:00:00Z"
}

output "event_count" {
  value = length(data.cumulocity_events.locations.events)
}
