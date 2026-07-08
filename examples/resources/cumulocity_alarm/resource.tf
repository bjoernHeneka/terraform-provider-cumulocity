resource "cumulocity_managed_object" "device" {
  name      = "Gateway 1"
  type      = "c8y_Gateway"
  is_device = true
}

# Raise an unavailability alarm on the device.
resource "cumulocity_alarm" "unavailable" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_UnavailabilityAlarm"
  text      = "No data received from the device within the required interval."
  severity  = "MAJOR"
  time      = "2024-01-15T10:30:00.000Z"
}

# Raise a temperature alarm and acknowledge it in the same configuration.
resource "cumulocity_alarm" "temperature" {
  source_id = cumulocity_managed_object.device.id
  type      = "c8y_TemperatureAlarm"
  text      = "Device temperature exceeded threshold."
  severity  = "CRITICAL"
  status    = "ACKNOWLEDGED"
  time      = "2024-01-15T08:00:00.000Z"
}
