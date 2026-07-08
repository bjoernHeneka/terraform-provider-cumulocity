# All devices of a given type
data "cumulocity_managed_objects" "mqtt_devices" {
  fragment_type = "c8y_IsDevice"
  type          = "c8y_MQTTDevice"
}

# Full-text search by name
data "cumulocity_managed_objects" "gateways" {
  text = "gateway"
}

output "device_ids" {
  value = [for mo in data.cumulocity_managed_objects.mqtt_devices.managed_objects : mo.id]
}

output "gateway_names" {
  value = [for mo in data.cumulocity_managed_objects.gateways.managed_objects : mo.name]
}
