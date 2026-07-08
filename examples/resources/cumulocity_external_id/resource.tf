resource "cumulocity_managed_object" "device" {
  name = "My Device"
  type = "c8y_Device"
}

resource "cumulocity_external_id" "serial" {
  managed_object_id = cumulocity_managed_object.device.id
  type              = "c8y_Serial"
  external_id       = "SN-00123456"
}

output "external_id_self" {
  value = cumulocity_external_id.serial.self
}
