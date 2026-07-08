# A generic asset in the inventory.
resource "cumulocity_managed_object" "building" {
  name = "Main Building"
  type = "c8y_Building"
}

# A device managed object (adds the c8y_IsDevice fragment).
resource "cumulocity_managed_object" "sensor" {
  name      = "Temperature Sensor 1"
  type      = "c8y_TemperatureSensor"
  is_device = true
}

# A device group that can contain devices (adds the c8y_IsDeviceGroup fragment).
resource "cumulocity_managed_object" "fleet" {
  name            = "Field Devices"
  is_device_group = true
}

output "sensor_id" {
  value = cumulocity_managed_object.sensor.id
}
