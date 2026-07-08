# Example: Alarm API

resource "cumulocity_managed_object" "sensor" {
  name = "Temperature Sensor"
  type = "c8y_TemperatureSensor"
}

# resource "cumulocity_alarm" "high_temp" {
#   source_id = cumulocity_managed_object.sensor.id
#   type      = "c8y_TemperatureAlarm"
#   text      = "Device temperature exceeded 80°C."
#   severity  = "CRITICAL"
#   time      = "2024-01-15T10:30:00.000Z"
# }
#
# resource "cumulocity_alarm" "unavailable" {
#   source_id = cumulocity_managed_object.sensor.id
#   type      = "c8y_UnavailabilityAlarm"
#   text      = "No data received from the device within the required interval."
#   severity  = "MAJOR"
#   status    = "ACKNOWLEDGED"
#   time      = "2024-01-15T09:00:00.000Z"
# }

# # List all active alarms for the sensor
# data "cumulocity_alarms" "active" {
#   source_id = cumulocity_managed_object.sensor.id
#   status    = "ACTIVE"
# }
#
# output "active_alarm_ids" {
#   value = [for a in data.cumulocity_alarms.active.alarms : a.id]
# }
