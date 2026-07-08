# Register a device by its external ID. Cumulocity creates the request in
# WAITING_FOR_CONNECTION; it moves to PENDING_ACCEPTANCE once the device connects.
resource "cumulocity_new_device_request" "gateway" {
  device_id   = "SN-00123456"
  device_type = "c8y_Linux"
}

# Once the device has connected, set status = ACCEPTED to approve it and assign
# it to a device group.
resource "cumulocity_managed_object" "fleet" {
  name            = "Field Devices"
  is_device_group = true
}

resource "cumulocity_new_device_request" "approved_gateway" {
  device_id = "SN-00987654"
  group_id  = cumulocity_managed_object.fleet.id
  status    = "ACCEPTED"
}
