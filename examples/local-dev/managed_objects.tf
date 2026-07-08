
resource "cumulocity_managed_object" "test_device" {
  name      = "tf-test-device"
  type      = "c8y_Linux"
  is_device = true
}

resource "cumulocity_managed_object" "test_group" {
  name            = "tf-test-group"
  is_device_group = true
}

output "test_device_id" {
  value       = cumulocity_managed_object.test_device.id
  description = "Use this ID as managed_object_id in cumulocity_user_inventory_role_assignment"
}

output "test_group_id" {
  value = cumulocity_managed_object.test_group.id
}
