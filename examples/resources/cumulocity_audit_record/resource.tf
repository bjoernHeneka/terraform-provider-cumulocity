resource "cumulocity_managed_object" "device" {
  name      = "Pump Controller"
  type      = "c8y_Pump"
  is_device = true
}

# Record a custom audit entry against the device.
resource "cumulocity_audit_record" "config_change" {
  source_id = cumulocity_managed_object.device.id
  type      = "Inventory"
  activity  = "Configuration updated"
  text      = "Pump threshold changed from 80 to 90."
  time      = "2024-01-15T10:30:00.000Z"
  user      = "operator@example.com"
}
