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
