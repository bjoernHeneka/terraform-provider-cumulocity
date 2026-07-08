resource "cumulocity_managed_object" "tracker" {
  name      = "Asset Tracker"
  type      = "c8y_Tracker"
  is_device = true
}

# Record a location update event for the device.
resource "cumulocity_event" "location" {
  source_id = cumulocity_managed_object.tracker.id
  type      = "c8y_LocationUpdate"
  text      = "Device moved to the warehouse."
  time      = "2024-01-15T10:30:00.000Z"
}
